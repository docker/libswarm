package manager

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/docker/swarm-v2/agent"
	"github.com/docker/swarm-v2/agent/exec"
	"github.com/docker/swarm-v2/api"
	"github.com/docker/swarm-v2/ca/testutils"
	"github.com/docker/swarm-v2/manager/dispatcher"
	"github.com/docker/swarm-v2/manager/state/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NoopExecutor is a dummy executor that implements enough to get the agent started.
type NoopExecutor struct {
}

func (e *NoopExecutor) Describe(ctx context.Context) (*api.NodeDescription, error) {
	return &api.NodeDescription{}, nil
}

func (e *NoopExecutor) Runner(t *api.Task) (exec.Runner, error) {
	return nil, exec.ErrRuntimeUnsupported
}

func TestManager(t *testing.T) {
	ctx := context.TODO()
	store := store.NewMemoryStore(nil)
	assert.NotNil(t, store)

	temp, err := ioutil.TempFile("", "test-socket")
	assert.NoError(t, err)
	assert.NoError(t, temp.Close())
	assert.NoError(t, os.Remove(temp.Name()))

	stateDir, err := ioutil.TempDir("", "test-raft")
	assert.NoError(t, err)
	defer os.RemoveAll(stateDir)

	agentSecurityConfigs, managerSecurityConfig, tmpDir, err := testutils.GenerateAgentAndManagerSecurityConfig(1)
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	m, err := New(&Config{
		ListenProto:    "unix",
		ListenAddr:     temp.Name(),
		StateDir:       stateDir,
		SecurityConfig: managerSecurityConfig,
	})
	assert.NoError(t, err)
	assert.NotNil(t, m)

	done := make(chan error)
	defer close(done)
	go func() {
		done <- m.Run(ctx)
	}()

	opts := []grpc.DialOption{grpc.WithTimeout(10 * time.Second)}
	opts = append(opts, grpc.WithTransportCredentials(agentSecurityConfigs[0].ClientTLSCreds))
	opts = append(opts, grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
		return net.DialTimeout("unix", addr, timeout)
	}))

	conn, err := grpc.Dial(temp.Name(), opts...)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, conn.Close())
	}()

	// We have to send a dummy request to verify if the connection is actually up.
	client := api.NewDispatcherClient(conn)
	_, err = client.Heartbeat(context.Background(), &api.HeartbeatRequest{})
	assert.Equal(t, grpc.ErrorDesc(err), dispatcher.ErrNodeNotRegistered.Error())

	m.Stop()

	// After stopping we should receive an error from ListenAndServe.
	assert.Error(t, <-done)
}

func TestManagerNodeCount(t *testing.T) {
	ctx := context.TODO()
	store := store.NewMemoryStore(nil)
	assert.NotNil(t, store)

	l, err := net.Listen("tcp", "127.0.0.1:0")

	stateDir, err := ioutil.TempDir("", "test-raft")
	assert.NoError(t, err)
	defer os.RemoveAll(stateDir)

	agentSecurityConfigs, managerSecurityConfig, tmpDir, err := testutils.GenerateAgentAndManagerSecurityConfig(2)
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	m, err := New(&Config{
		Listener:       l,
		StateDir:       stateDir,
		SecurityConfig: managerSecurityConfig,
	})
	assert.NoError(t, err)
	assert.NotNil(t, m)
	go m.Run(ctx)
	defer m.Stop()

	opts := []grpc.DialOption{grpc.WithTimeout(10 * time.Second)}
	opts = append(opts, grpc.WithTransportCredentials(managerSecurityConfig.ClientTLSCreds))

	conn, err := grpc.Dial(l.Addr().String(), opts...)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, conn.Close())
	}()

	// We have to send a dummy request to verify if the connection is actually up.
	mClient := api.NewManagerClient(conn)

	managers := agent.NewManagers(l.Addr().String())
	a1, err := agent.New(&agent.Config{
		Hostname:       "hostname1",
		Managers:       managers,
		Executor:       &NoopExecutor{},
		SecurityConfig: agentSecurityConfigs[0],
	})
	require.NoError(t, err)
	a2, err := agent.New(&agent.Config{
		Hostname:       "hostname2",
		Managers:       managers,
		Executor:       &NoopExecutor{},
		SecurityConfig: agentSecurityConfigs[1],
	})
	require.NoError(t, err)

	require.NoError(t, a1.Start(context.Background()))
	require.NoError(t, a2.Start(context.Background()))

	defer func() {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		a1.Stop(ctx)
		a2.Stop(ctx)
	}()

	time.Sleep(1500 * time.Millisecond)

	resp, err := mClient.NodeCount(context.Background(), &api.NodeCountRequest{})
	assert.NoError(t, err)
	assert.Equal(t, 2, resp.Count)
}
