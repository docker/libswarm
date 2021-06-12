package main

import (
	"fmt"

	"github.com/coreos/etcd/raft/raftpb"
	"github.com/docker/swarmkit/ca"
	"github.com/pkg/errors"

	"github.com/cloudflare/cfssl/helpers"
	"github.com/docker/swarmkit/api"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

func renewCerts(swarmdir, unlockKey string) error {
	// First, load the existing cert.  We don't actually bother to check if
	// it's expired - this will just obtain a new cert anyway.
	krw, err := getKRW(swarmdir, unlockKey)
	if err != nil {
		return errors.Wrap(err, "could not load swarm certificate")
	}
	cert, _, err := krw.Read()
	if err != nil {
		return errors.Wrap(err, "could not read swarm certificate")
	}
	certificates, err := helpers.ParseCertificatesPEM(cert)
	if err != nil {
		return errors.Wrap(err, "could not parse node certificate")
	}
	// We need to make sure when renewing that we provide the same CN (node ID),
	// OU (role), and org (swarm cluster ID) when getting a new certificate
	cn := string(certificates[0].Subject.CommonName[0])
	ou := string(certificates[0].Subject.OrganizationalUnit[0])
	org := string(certificates[0].Subject.Organization[0])

	// Load up the raft data on disk
	walData, snapshot, err := loadData(swarmdir, unlockKey)
	if err != nil {
		return errors.Wrap(err, "could not load swarm data")
	}
	var cluster *api.Cluster

	// If there's a snapshot, get the cluster from it
	if snapshot != nil {
		s := &api.Snapshot{}
		if err := proto.Unmarshal(snapshot.Data, s); err != nil {
			return err
		}
		if s.Version != api.Snapshot_V0 {
			return fmt.Errorf("unrecognized snapshot version %d", s.Version)
		}
		cluster = s.Store.Clusters[0]
	}

	// It's possible there's no snapshot yet, or the cluster has been updated
	// since the last snapshot, so also read from the WALs
	for _, ent := range walData.Entries {
		if ent.Type != raftpb.EntryNormal {
			continue
		}

		r := &api.InternalRaftRequest{}
		err := proto.Unmarshal(ent.Data, r)
		if err != nil {
			return errors.Wrap(err, "could not read WAL")
		}

		for _, act := range r.Action {
			target := act.GetTarget()
			if actype, ok := target.(*api.StoreAction_Cluster); ok {
				cluster = actype.Cluster
			}
		}
	}

	// There should always be a cluster and CA cert, unless the raft store has been
	// catastrophcially corrupted, but it's possible that there is no CA key because
	// the cluster used an external CA.
	if cluster == nil || cluster.RootCA.CACert == nil || cluster.RootCA.CAKey == nil {
		return errors.New("could not find CA key data in raft logs; cannot renew certs")
	}

	// Issue a new certificate that expires at the configured expiry time.
	expiry := ca.DefaultNodeCertExpiration
	if cluster.Spec.CAConfig.NodeCertExpiry != nil {
		clusterExpiry, err := types.DurationFromProto(cluster.Spec.CAConfig.NodeCertExpiry)
		if err == nil {
			expiry = clusterExpiry
		}
	}
	rootCA, err := ca.RootCAFromAPI(&cluster.RootCA, expiry)
	if err != nil {
		return errors.Wrap(err, "invalid CA info in raft logs; cannot renew certs")
	}

	_, _, err = rootCA.IssueAndSaveNewCertificates(krw, cn, ou, org)
	return err
}
