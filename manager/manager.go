package manager

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/swarm-v2/api"
	"github.com/docker/swarm-v2/ca"
	"github.com/docker/swarm-v2/identity"
	"github.com/docker/swarm-v2/log"
	"github.com/docker/swarm-v2/manager/allocator"
	"github.com/docker/swarm-v2/manager/controlapi"
	"github.com/docker/swarm-v2/manager/dispatcher"
	"github.com/docker/swarm-v2/manager/orchestrator"
	"github.com/docker/swarm-v2/manager/raftpicker"
	"github.com/docker/swarm-v2/manager/scheduler"
	"github.com/docker/swarm-v2/manager/state/raft"
	"github.com/docker/swarm-v2/manager/state/store"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	// defaultTaskHistoryRetentionLimit is the number of tasks to keep.
	defaultTaskHistoryRetentionLimit = 10
)

// Config is used to tune the Manager.
type Config struct {
	SecurityConfig *ca.SecurityConfig

	ListenProto string
	ListenAddr  string
	// Listener will be used for grpc serving if it's not nil, ListenProto and
	// ListenAddr fields will be used otherwise.
	Listener net.Listener

	// JoinRaft is an optional address of a node in an existing raft
	// cluster to join.
	JoinRaft string

	// Top-level state directory
	StateDir string

	// ForceNewCluster defines if we have to force a new cluster
	// because we are recovering from a backup data directory.
	ForceNewCluster bool
}

// Manager is the cluster manager for Swarm.
// This is the high-level object holding and initializing all the manager
// subsystems.
type Manager struct {
	config   *Config
	listener net.Listener

	caserver         *ca.Server
	dispatcher       *dispatcher.Dispatcher
	orchestrator     *orchestrator.Orchestrator
	fillOrchestrator *orchestrator.FillOrchestrator
	taskReaper       *orchestrator.TaskReaper
	scheduler        *scheduler.Scheduler
	allocator        *allocator.Allocator
	server           *grpc.Server
	raftNode         *raft.Node

	mu sync.Mutex

	started chan struct{}
	stopped bool
}

// New creates a Manager which has not started to accept requests yet.
func New(config *Config) (*Manager, error) {
	dispatcherConfig := dispatcher.DefaultConfig()

	if config.Listener != nil {
		config.ListenAddr = config.Listener.Addr().String()
	}

	listenAddr := config.ListenAddr

	listenHost, listenPort, err := net.SplitHostPort(config.ListenAddr)
	if err == nil {
		ip := net.ParseIP(listenHost)
		if ip != nil && ip.IsUnspecified() {
			// Find our local IP address associated with the default route.
			// This may not be the appropriate address to use for internal
			// cluster communications, but it seems like the best default.
			// The admin can override this address if necessary.
			conn, err := net.Dial("udp", "8.8.8.8:53")
			if err != nil {
				return nil, fmt.Errorf("could not determine local IP address: %v", err)
			}
			localAddr := conn.LocalAddr().String()
			conn.Close()

			listenHost, _, err = net.SplitHostPort(localAddr)
			if err != nil {
				return nil, fmt.Errorf("could not split local IP address: %v", err)
			}

			listenAddr = net.JoinHostPort(listenHost, listenPort)
		}
	}

	// TODO(stevvooe): Reported address of manager is plumbed to listen addr
	// for now, may want to make this separate. This can be tricky to get right
	// so we need to make it easy to override. This needs to be the address
	// through which agent nodes access the manager.
	dispatcherConfig.Addr = listenAddr

	err = os.MkdirAll(config.StateDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create state directory: %v", err)
	}

	raftStateDir := filepath.Join(config.StateDir, "raft")
	err = os.MkdirAll(raftStateDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft state directory: %v", err)
	}

	lis := config.Listener
	if lis == nil {
		l, err := net.Listen(config.ListenProto, config.ListenAddr)
		if err != nil {
			return nil, err
		}
		lis = l
	}

	raftCfg := raft.DefaultNodeConfig()

	newNodeOpts := raft.NewNodeOptions{
		Addr:            listenAddr,
		JoinAddr:        config.JoinRaft,
		Config:          raftCfg,
		StateDir:        raftStateDir,
		ForceNewCluster: config.ForceNewCluster,
		TLSCredentials:  config.SecurityConfig.ClientTLSCreds,
	}
	raftNode, err := raft.NewNode(context.TODO(), newNodeOpts)
	if err != nil {
		lis.Close()
		return nil, fmt.Errorf("can't create raft node: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.Creds(config.SecurityConfig.ServerTLSCreds)}

	m := &Manager{
		config:     config,
		listener:   lis,
		caserver:   ca.NewServer(raftNode.MemoryStore(), config.SecurityConfig),
		dispatcher: dispatcher.New(raftNode, dispatcherConfig),
		server:     grpc.NewServer(opts...),
		raftNode:   raftNode,
		started:    make(chan struct{}),
	}

	return m, nil
}

// Run starts all manager sub-systems and the gRPC server at the configured
// address.
// The call never returns unless an error occurs or `Stop()` is called.
func (m *Manager) Run(ctx context.Context) error {
	go func() {
		leadershipCh, cancel := m.raftNode.SubscribeLeadership()
		defer cancel()
		for leadershipEvent := range leadershipCh {
			m.mu.Lock()
			if m.stopped {
				m.mu.Unlock()
				continue
			}
			newState := leadershipEvent.(raft.LeadershipState)

			if newState == raft.IsLeader {
				s := m.raftNode.MemoryStore()

				// Add a default cluster object to the store. Don't check the error
				// because we expect this to fail unless this is a brand new cluster.
				s.Update(func(tx store.Tx) error {
					store.CreateCluster(tx, &api.Cluster{
						ID: identity.NewID(),
						Spec: api.ClusterSpec{
							Annotations: api.Annotations{
								Name: store.DefaultClusterName,
							},
							AcceptancePolicy: ca.DefaultAcceptancePolicy(),
							Orchestration: api.OrchestrationConfig{
								TaskHistoryRetentionLimit: defaultTaskHistoryRetentionLimit,
							},
							Raft: raft.DefaultRaftConfig(),
						},
					})
					return nil
				})

				m.orchestrator = orchestrator.New(s)
				m.fillOrchestrator = orchestrator.NewFillOrchestrator(s)
				m.taskReaper = orchestrator.NewTaskReaper(s)
				m.scheduler = scheduler.New(s)

				// TODO(stevvooe): Allocate a context that can be used to
				// shutdown underlying manager processes when leadership is
				// lost.

				var err error
				m.allocator, err = allocator.New(s)
				if err != nil {
					log.G(ctx).WithError(err).Error("failed to create allocator")
					// TODO(stevvooe): It doesn't seem correct here to fail
					// creating the allocator but then use it anyways.
				}

				go func(d *dispatcher.Dispatcher) {
					if err := d.Run(ctx); err != nil {
						log.G(ctx).WithError(err).Error("dispatcher exited with an error")
					}
				}(m.dispatcher)

				go func(server *ca.Server) {
					if err := server.Run(ctx); err != nil {
						log.G(ctx).WithError(err).Error("CA signer exited with an error")
					}
				}(m.caserver)

				// Start all sub-components in separate goroutines.
				// TODO(aluzzardi): This should have some kind of error handling so that
				// any component that goes down would bring the entire manager down.

				if m.allocator != nil {
					go func(allocator *allocator.Allocator) {
						if err := allocator.Run(ctx); err != nil {
							log.G(ctx).WithError(err).Error("allocator exited with an error")
						}
					}(m.allocator)
				}

				go func(scheduler *scheduler.Scheduler) {
					if err := scheduler.Run(ctx); err != nil {
						log.G(ctx).WithError(err).Error("scheduler exited with an error")
					}
				}(m.scheduler)
				go func(taskReaper *orchestrator.TaskReaper) {
					taskReaper.Run()
				}(m.taskReaper)
				go func(orchestrator *orchestrator.Orchestrator) {
					if err := orchestrator.Run(ctx); err != nil {
						log.G(ctx).WithError(err).Error("orchestrator exited with an error")
					}
				}(m.orchestrator)
				go func(fillOrchestrator *orchestrator.FillOrchestrator) {
					if err := fillOrchestrator.Run(ctx); err != nil {
						log.G(ctx).WithError(err).Error("fillOrchestrator exited with an error")
					}
				}(m.fillOrchestrator)
			} else if newState == raft.IsFollower {
				m.dispatcher.Stop()
				m.caserver.Stop()

				if m.allocator != nil {
					m.allocator.Stop()
					m.allocator = nil
				}

				m.orchestrator.Stop()
				m.orchestrator = nil

				m.fillOrchestrator.Stop()
				m.fillOrchestrator = nil

				m.taskReaper.Stop()
				m.taskReaper = nil

				m.scheduler.Stop()
				m.scheduler = nil
			}
			m.mu.Unlock()
		}
	}()

	ctx = log.WithLogger(ctx, log.G(ctx).WithFields(
		logrus.Fields{
			"proto": m.listener.Addr().Network(),
			"addr":  m.listener.Addr().String()}))
	log.G(ctx).Info("listening")

	go func() {
		err := m.raftNode.Run(ctx)
		if err != nil {
			log.G(ctx).Error(err)
			m.Stop()
		}
	}()

	proxyOpts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(2 * time.Second),
		grpc.WithTransportCredentials(m.config.SecurityConfig.ClientTLSCreds),
	}

	cs := raftpicker.NewConnSelector(m.raftNode, proxyOpts...)

	localAPI := controlapi.NewServer(m.raftNode.MemoryStore(), m.raftNode)
	proxyAPI := api.NewRaftProxyControlServer(localAPI, cs, m.raftNode)
	proxyDispatcher := api.NewRaftProxyDispatcherServer(m.dispatcher, cs, m.raftNode, ca.WithMetadataForwardCN)
	proxyCA := api.NewRaftProxyCAServer(m.caserver, cs, m.raftNode, ca.WithMetadataForwardCN)

	api.RegisterRaftServer(m.server, m.raftNode)
	api.RegisterCAServer(m.server, proxyCA)
	api.RegisterControlServer(m.server, proxyAPI)
	api.RegisterDispatcherServer(m.server, proxyDispatcher)

	errServe := make(chan error, 1)
	go func() {
		errServe <- m.server.Serve(m.listener)
	}()

	if err := raft.WaitForLeader(ctx, m.raftNode); err != nil {
		m.server.Stop()
		return err
	}
	close(m.started)
	return <-errServe
}

// Stop stops the manager. It immediately closes all open connections and
// active RPCs as well as stopping the scheduler.
func (m *Manager) Stop() {
	// Don't shut things down while the manager is still starting up.
	<-m.started

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return
	}

	m.dispatcher.Stop()
	m.caserver.Stop()

	if m.allocator != nil {
		m.allocator.Stop()
	}
	if m.orchestrator != nil {
		m.orchestrator.Stop()
	}
	if m.fillOrchestrator != nil {
		m.fillOrchestrator.Stop()
	}
	if m.taskReaper != nil {
		m.taskReaper.Stop()
	}
	if m.scheduler != nil {
		m.scheduler.Stop()
	}
	m.raftNode.Shutdown()
	m.server.Stop()
	m.stopped = true
}
