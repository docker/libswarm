package ca_test

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	cfconfig "github.com/cloudflare/cfssl/config"
	"github.com/docker/swarmkit/ca"
	"github.com/docker/swarmkit/ca/testutils"
	"github.com/docker/swarmkit/ioutils"
	"github.com/docker/swarmkit/manager/state/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadRootCASuccess(t *testing.T) {
	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	// Remove the CA cert
	os.RemoveAll(tc.Paths.RootCA.Cert)

	rootCA, err := ca.DownloadRootCA(tc.Context, tc.Paths.RootCA, tc.WorkerToken, tc.Remotes)
	require.NoError(t, err)
	require.NotNil(t, rootCA.Pool)
	require.NotNil(t, rootCA.Cert)
	require.Nil(t, rootCA.Signer)
	require.False(t, rootCA.CanSign())
	require.Equal(t, tc.RootCA.Cert, rootCA.Cert)

	// Remove the CA cert
	os.RemoveAll(tc.Paths.RootCA.Cert)

	// downloading without a join token also succeeds
	rootCA, err = ca.DownloadRootCA(tc.Context, tc.Paths.RootCA, "", tc.Remotes)
	require.NoError(t, err)
	require.NotNil(t, rootCA.Pool)
	require.NotNil(t, rootCA.Cert)
	require.Nil(t, rootCA.Signer)
	require.False(t, rootCA.CanSign())
	require.Equal(t, tc.RootCA.Cert, rootCA.Cert)
}

func TestDownloadRootCAWrongCAHash(t *testing.T) {
	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	// Remove the CA cert
	os.RemoveAll(tc.Paths.RootCA.Cert)

	// invalid token
	_, err := ca.DownloadRootCA(tc.Context, tc.Paths.RootCA, "invalidtoken", tc.Remotes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid join token")

	// invalid hash token
	splitToken := strings.Split(tc.ManagerToken, "-")
	splitToken[2] = "1kxftv4ofnc6mt30lmgipg6ngf9luhwqopfk1tz6bdmnkubg0e"
	replacementToken := strings.Join(splitToken, "-")

	os.RemoveAll(tc.Paths.RootCA.Cert)

	_, err = ca.DownloadRootCA(tc.Context, tc.Paths.RootCA, replacementToken, tc.Remotes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "remote CA does not match fingerprint.")
}

func TestCreateSecurityConfigEmptyDir(t *testing.T) {
	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	// Remove all the contents from the temp dir and try again with a new node
	os.RemoveAll(tc.TempDir)
	krw := ca.NewKeyReadWriter(tc.Paths.Node, nil, nil)
	nodeConfig, err := tc.RootCA.CreateSecurityConfig(tc.Context, krw,
		ca.CertificateRequestConfig{
			Token:   tc.WorkerToken,
			Remotes: tc.Remotes,
		})
	assert.NoError(t, err)
	assert.NotNil(t, nodeConfig)
	assert.NotNil(t, nodeConfig.ClientTLSCreds)
	assert.NotNil(t, nodeConfig.ServerTLSCreds)
	assert.Equal(t, tc.RootCA, *nodeConfig.RootCA())
}

func TestCreateSecurityConfigNoCerts(t *testing.T) {
	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	// Remove only the node certificates form the directory, and attest that we get
	// new certificates that are locally signed
	os.RemoveAll(tc.Paths.Node.Cert)
	krw := ca.NewKeyReadWriter(tc.Paths.Node, nil, nil)
	nodeConfig, err := tc.RootCA.CreateSecurityConfig(tc.Context, krw,
		ca.CertificateRequestConfig{
			Token:   tc.WorkerToken,
			Remotes: tc.Remotes,
		})
	assert.NoError(t, err)
	assert.NotNil(t, nodeConfig)
	assert.NotNil(t, nodeConfig.ClientTLSCreds)
	assert.NotNil(t, nodeConfig.ServerTLSCreds)
	assert.Equal(t, tc.RootCA, *nodeConfig.RootCA())

	// Remove only the node certificates form the directory, get a new rootCA, and attest that we get
	// new certificates that are issued by the remote CA
	os.RemoveAll(tc.Paths.Node.Cert)
	rootCA, err := ca.GetLocalRootCA(tc.Paths.RootCA)
	assert.NoError(t, err)
	nodeConfig, err = rootCA.CreateSecurityConfig(tc.Context, krw,
		ca.CertificateRequestConfig{
			Token:   tc.WorkerToken,
			Remotes: tc.Remotes,
		})
	assert.NoError(t, err)
	assert.NotNil(t, nodeConfig)
	assert.NotNil(t, nodeConfig.ClientTLSCreds)
	assert.NotNil(t, nodeConfig.ServerTLSCreds)
	assert.Equal(t, rootCA, *nodeConfig.RootCA())
}

func TestCreateSecurityConfigNewRootCA(t *testing.T) {
	t.Parallel()

	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	tempDir, err := ioutil.TempDir("", "test-create-security-config")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	paths := ca.NewConfigPaths(tempDir)

	// Create new root CA material on the server
	cert, _, err := testutils.CreateRootCertAndKey("rootCNnew")
	require.NoError(t, err)
	CABundle := append(cert, tc.RootCA.Cert...)

	rootCA, err := ca.NewRootCA(CABundle, nil, paths.Node, ca.DefaultNodeCertExpiration)
	require.NoError(t, err)

	krw := ca.NewKeyReadWriter(paths.Node, nil, nil)

	securityConfig, err := rootCA.CreateSecurityConfig(
		tc.Context, krw,
		ca.CertificateRequestConfig{
			Token:   tc.WorkerToken,
			Remotes: tc.Remotes,
		})
	require.NoError(t, err)

	_, err = tc.WriteNewNodeConfig(ca.ManagerRole)
	require.NoError(t, err)
	tcCert, err := ioutil.ReadFile(tc.Paths.Node.Cert)
	require.NoError(t, err)
	certBlock, _ := pem.Decode(tcCert)
	require.NotNil(t, certBlock)
	x509Cert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)

	for _, pool := range []*x509.CertPool{
		securityConfig.ClientTLSCreds.Config().RootCAs,
		securityConfig.ServerTLSCreds.Config().RootCAs,
	} {
		_, err := x509Cert.Verify(x509.VerifyOptions{Roots: pool})
		require.NoError(t, err)
	}
}

func TestLoadSecurityConfigInvalidCert(t *testing.T) {
	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	// Write some garbage to the cert
	ioutil.WriteFile(tc.Paths.Node.Cert, []byte(`-----BEGIN CERTIFICATE-----\n
some random garbage\n
-----END CERTIFICATE-----`), 0644)

	krw := ca.NewKeyReadWriter(tc.Paths.Node, nil, nil)

	_, err := ca.LoadSecurityConfig(tc.Context, tc.RootCA, krw)
	assert.Error(t, err)

	nodeConfig, err := tc.RootCA.CreateSecurityConfig(tc.Context, krw,
		ca.CertificateRequestConfig{
			Remotes: tc.Remotes,
		})

	assert.NoError(t, err)
	assert.NotNil(t, nodeConfig)
	assert.NotNil(t, nodeConfig.ClientTLSCreds)
	assert.NotNil(t, nodeConfig.ServerTLSCreds)
	assert.Equal(t, tc.RootCA, *nodeConfig.RootCA())
}

func TestLoadSecurityConfigInvalidKey(t *testing.T) {
	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	// Write some garbage to the Key
	ioutil.WriteFile(tc.Paths.Node.Key, []byte(`-----BEGIN EC PRIVATE KEY-----\n
some random garbage\n
-----END EC PRIVATE KEY-----`), 0644)

	krw := ca.NewKeyReadWriter(tc.Paths.Node, nil, nil)

	_, err := ca.LoadSecurityConfig(tc.Context, tc.RootCA, krw)
	assert.Error(t, err)

	nodeConfig, err := tc.RootCA.CreateSecurityConfig(tc.Context, krw,
		ca.CertificateRequestConfig{
			Remotes: tc.Remotes,
		})
	assert.NoError(t, err)
	assert.NotNil(t, nodeConfig)
	assert.NotNil(t, nodeConfig.ClientTLSCreds)
	assert.NotNil(t, nodeConfig.ServerTLSCreds)
	assert.Equal(t, tc.RootCA, *nodeConfig.RootCA())
}

func TestLoadSecurityConfigIncorrectPassphrase(t *testing.T) {
	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	paths := ca.NewConfigPaths(tc.TempDir)
	_, err := tc.RootCA.IssueAndSaveNewCertificates(ca.NewKeyReadWriter(paths.Node, []byte("kek"), nil),
		"nodeID", ca.WorkerRole, tc.Organization)
	require.NoError(t, err)

	_, err = ca.LoadSecurityConfig(tc.Context, tc.RootCA, ca.NewKeyReadWriter(paths.Node, nil, nil))
	require.IsType(t, ca.ErrInvalidKEK{}, err)
}

func TestRenewTLSConfigWorker(t *testing.T) {
	t.Parallel()

	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get a new nodeConfig with a TLS cert that has the default Cert duration
	nodeConfig, err := tc.WriteNewNodeConfig(ca.WorkerRole)
	assert.NoError(t, err)

	// Create a new RootCA, and change the policy to issue 6 minute certificates
	// Because of the default backdate of 5 minutes, this issues certificates
	// valid for 1 minute.
	newRootCA, err := ca.NewRootCA(tc.RootCA.Cert, tc.RootCA.Key, tc.Paths.RootCA, ca.DefaultNodeCertExpiration)
	assert.NoError(t, err)
	newRootCA.Signer.SetPolicy(&cfconfig.Signing{
		Default: &cfconfig.SigningProfile{
			Usage:  []string{"signing", "key encipherment", "server auth", "client auth"},
			Expiry: 6 * time.Minute,
		},
	})

	// Create a new CSR and overwrite the key on disk
	csr, key, err := ca.GenerateNewCSR()
	assert.NoError(t, err)

	// Issue a new certificate with the same details as the current config, but with 1 min expiration time
	c := nodeConfig.ClientTLSCreds
	signedCert, err := newRootCA.ParseValidateAndSignCSR(csr, c.NodeID(), c.Role(), c.Organization())
	assert.NoError(t, err)
	assert.NotNil(t, signedCert)

	// Overwrite the certificate on disk with one that expires in 1 minute
	err = ioutils.AtomicWriteFile(tc.Paths.Node.Cert, signedCert, 0644)
	assert.NoError(t, err)

	err = ioutils.AtomicWriteFile(tc.Paths.Node.Key, key, 0600)
	assert.NoError(t, err)

	renew := make(chan struct{})
	updates := ca.RenewTLSConfig(ctx, nodeConfig, tc.Remotes, renew)
	select {
	case <-time.After(10 * time.Second):
		assert.Fail(t, "TestRenewTLSConfig timed-out")
	case certUpdate := <-updates:
		assert.NoError(t, certUpdate.Err)
		assert.NotNil(t, certUpdate)
		assert.Equal(t, ca.WorkerRole, certUpdate.Role)
	}
}

func TestRenewTLSConfigManager(t *testing.T) {
	t.Parallel()

	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get a new nodeConfig with a TLS cert that has the default Cert duration
	nodeConfig, err := tc.WriteNewNodeConfig(ca.ManagerRole)
	assert.NoError(t, err)

	// Create a new RootCA, and change the policy to issue 6 minute certificates
	// Because of the default backdate of 5 minutes, this issues certificates
	// valid for 1 minute.
	newRootCA, err := ca.NewRootCA(tc.RootCA.Cert, tc.RootCA.Key, tc.Paths.RootCA, ca.DefaultNodeCertExpiration)
	assert.NoError(t, err)
	newRootCA.Signer.SetPolicy(&cfconfig.Signing{
		Default: &cfconfig.SigningProfile{
			Usage:  []string{"signing", "key encipherment", "server auth", "client auth"},
			Expiry: 6 * time.Minute,
		},
	})

	// Create a new CSR and overwrite the key on disk
	csr, key, err := ca.GenerateNewCSR()
	assert.NoError(t, err)

	// Issue a new certificate with the same details as the current config, but with 1 min expiration time
	c := nodeConfig.ClientTLSCreds
	signedCert, err := newRootCA.ParseValidateAndSignCSR(csr, c.NodeID(), c.Role(), c.Organization())
	assert.NoError(t, err)
	assert.NotNil(t, signedCert)

	// Overwrite the certificate on disk with one that expires in 1 minute
	err = ioutils.AtomicWriteFile(tc.Paths.Node.Cert, signedCert, 0644)
	assert.NoError(t, err)

	err = ioutils.AtomicWriteFile(tc.Paths.Node.Key, key, 0600)
	assert.NoError(t, err)

	// Get a new nodeConfig with a TLS cert that has 1 minute to live
	renew := make(chan struct{})

	updates := ca.RenewTLSConfig(ctx, nodeConfig, tc.Remotes, renew)
	select {
	case <-time.After(10 * time.Second):
		assert.Fail(t, "TestRenewTLSConfig timed-out")
	case certUpdate := <-updates:
		assert.NoError(t, certUpdate.Err)
		assert.NotNil(t, certUpdate)
		assert.Equal(t, ca.ManagerRole, certUpdate.Role)
	}
}

func TestRenewTLSConfigNewRootCA(t *testing.T) {
	t.Parallel()

	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodeConfig, err := tc.WriteNewNodeConfig(ca.ManagerRole)
	require.NoError(t, err)

	// Create new root CA material on the server
	cert, key, err := testutils.CreateRootCertAndKey("rootCNnew")
	require.NoError(t, err)
	// Append to a bundle
	CABundle := append(cert, tc.RootCA.Cert...)

	assert.NoError(t, tc.MemoryStore.Update(func(tx store.Tx) error {
		clusters, err := store.FindClusters(tx, store.ByIDPrefix(tc.Organization))
		assert.NoError(t, err)
		clusters[0].RootCA.CACert = CABundle
		clusters[0].RootCA.CAKey = key
		return store.UpdateCluster(tx, clusters[0])
	}))

	// Ensure that we can update, and that the root CA in the security config
	// is updated and that the new TLS creds have a rootCA that includes the bundle
	err = ca.RenewTLSConfigNow(ctx, nodeConfig, tc.Remotes)
	require.NoError(t, err)

	require.Equal(t, CABundle, nodeConfig.RootCA().Cert)

	// use this to generate a new cert and key, so it can be verified using the client
	// and server TLS bundle
	newRootCA, err := ca.NewRootCA(cert, key, tc.RootCA.Path, ca.DefaultNodeCertExpiration)
	require.NoError(t, err)

	csr, _, err := ca.GenerateNewCSR()
	require.NoError(t, err)
	certChain, err := newRootCA.ParseValidateAndSignCSR(csr, "cn", "ou", "org")
	require.NoError(t, err)

	certBlock, _ := pem.Decode(certChain)
	require.NotNil(t, certBlock)
	x509Cert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)

	for _, pool := range []*x509.CertPool{
		nodeConfig.ClientTLSCreds.Config().RootCAs,
		nodeConfig.ServerTLSCreds.Config().RootCAs,
	} {
		_, err := x509Cert.Verify(x509.VerifyOptions{Roots: pool})
		require.NoError(t, err)
	}
}

func TestRenewTLSConfigWithNoNode(t *testing.T) {
	t.Parallel()

	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get a new nodeConfig with a TLS cert that has the default Cert duration
	nodeConfig, err := tc.WriteNewNodeConfig(ca.ManagerRole)
	assert.NoError(t, err)

	// Create a new RootCA, and change the policy to issue 6 minute certificates.
	// Because of the default backdate of 5 minutes, this issues certificates
	// valid for 1 minute.
	newRootCA, err := ca.NewRootCA(tc.RootCA.Cert, tc.RootCA.Key, tc.Paths.RootCA, ca.DefaultNodeCertExpiration)
	assert.NoError(t, err)
	newRootCA.Signer.SetPolicy(&cfconfig.Signing{
		Default: &cfconfig.SigningProfile{
			Usage:  []string{"signing", "key encipherment", "server auth", "client auth"},
			Expiry: 6 * time.Minute,
		},
	})

	// Create a new CSR and overwrite the key on disk
	csr, key, err := ca.GenerateNewCSR()
	assert.NoError(t, err)

	// Issue a new certificate with the same details as the current config, but with 1 min expiration time
	c := nodeConfig.ClientTLSCreds
	signedCert, err := newRootCA.ParseValidateAndSignCSR(csr, c.NodeID(), c.Role(), c.Organization())
	assert.NoError(t, err)
	assert.NotNil(t, signedCert)

	// Overwrite the certificate on disk with one that expires in 1 minute
	err = ioutils.AtomicWriteFile(tc.Paths.Node.Cert, signedCert, 0644)
	assert.NoError(t, err)

	err = ioutils.AtomicWriteFile(tc.Paths.Node.Key, key, 0600)
	assert.NoError(t, err)

	// Delete the node from the backend store
	err = tc.MemoryStore.Update(func(tx store.Tx) error {
		node := store.GetNode(tx, nodeConfig.ClientTLSCreds.NodeID())
		assert.NotNil(t, node)
		return store.DeleteNode(tx, nodeConfig.ClientTLSCreds.NodeID())
	})
	assert.NoError(t, err)

	renew := make(chan struct{})
	updates := ca.RenewTLSConfig(ctx, nodeConfig, tc.Remotes, renew)
	select {
	case <-time.After(10 * time.Second):
		assert.Fail(t, "TestRenewTLSConfig timed-out")
	case certUpdate := <-updates:
		assert.Error(t, certUpdate.Err)
		assert.Contains(t, certUpdate.Err.Error(), "not found when attempting to renew certificate")
	}
}

func TestForceRenewTLSConfig(t *testing.T) {
	t.Parallel()

	tc := testutils.NewTestCA(t)
	defer tc.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get a new managerConfig with a TLS cert that has 15 minutes to live
	nodeConfig, err := tc.WriteNewNodeConfig(ca.ManagerRole)
	assert.NoError(t, err)

	renew := make(chan struct{}, 1)
	updates := ca.RenewTLSConfig(ctx, nodeConfig, tc.Remotes, renew)
	renew <- struct{}{}
	select {
	case <-time.After(10 * time.Second):
		assert.Fail(t, "TestForceRenewTLSConfig timed-out")
	case certUpdate := <-updates:
		assert.NoError(t, certUpdate.Err)
		assert.NotNil(t, certUpdate)
		assert.Equal(t, certUpdate.Role, ca.ManagerRole)
	}
}
