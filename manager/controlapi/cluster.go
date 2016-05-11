package controlapi

import (
	"github.com/docker/swarm-v2/api"
	"github.com/docker/swarm-v2/manager/state/store"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func validateClusterSpec(spec *api.ClusterSpec) error {
	if spec == nil {
		return grpc.Errorf(codes.InvalidArgument, errInvalidArgument.Error())
	}
	return nil
}

// GetCluster returns a Cluster given a ClusterID.
// - Returns `InvalidArgument` if ClusterID is not provided.
// - Returns `NotFound` if the Cluster is not found.
func (s *Server) GetCluster(ctx context.Context, request *api.GetClusterRequest) (*api.GetClusterResponse, error) {
	if request.ClusterID == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, errInvalidArgument.Error())
	}

	var cluster *api.Cluster
	s.store.View(func(tx store.ReadTx) {
		cluster = store.GetCluster(tx, request.ClusterID)
	})
	if cluster == nil {
		return nil, grpc.Errorf(codes.NotFound, "cluster %s not found", request.ClusterID)
	}

	return &api.GetClusterResponse{
		Cluster: cluster,
	}, nil
}

// UpdateCluster updates a Cluster referenced by ClusterID with the given ClusterSpec.
// - Returns `NotFound` if the Cluster is not found.
// - Returns `InvalidArgument` if the ClusterSpec is malformed.
// - Returns `Unimplemented` if the ClusterSpec references unimplemented features.
// - Returns an error if the update fails.
func (s *Server) UpdateCluster(ctx context.Context, request *api.UpdateClusterRequest) (*api.UpdateClusterResponse, error) {
	if request.ClusterID == "" || request.ClusterVersion == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, errInvalidArgument.Error())
	}
	if err := validateClusterSpec(request.Spec); err != nil {
		return nil, err
	}

	var cluster *api.Cluster
	err := s.store.Update(func(tx store.Tx) error {
		cluster = store.GetCluster(tx, request.ClusterID)
		if cluster == nil {
			return nil
		}
		cluster.Meta.Version = *request.ClusterVersion
		cluster.Spec = *request.Spec.Copy()
		return store.UpdateCluster(tx, cluster)
	})
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, grpc.Errorf(codes.NotFound, "cluster %s not found", request.ClusterID)
	}
	return &api.UpdateClusterResponse{
		Cluster: cluster,
	}, nil
}

// ListClusters returns a list of all clusters.
func (s *Server) ListClusters(ctx context.Context, request *api.ListClustersRequest) (*api.ListClustersResponse, error) {
	var (
		clusters []*api.Cluster
		err      error
	)
	s.store.View(func(tx store.ReadTx) {
		if request.Options == nil || request.Options.Query == "" {
			clusters, err = store.FindClusters(tx, store.All)
		} else {
			clusters, err = store.FindClusters(tx, store.ByQuery(request.Options.Query))
		}
	})
	if err != nil {
		return nil, err
	}
	return &api.ListClustersResponse{
		Clusters: clusters,
	}, nil
}
