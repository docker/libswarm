package raft

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/pkg/idutil"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
	"github.com/docker/swarm-v2/api"
	pb "github.com/docker/swarm-v2/api"
	"github.com/docker/swarm-v2/ca"
	"github.com/docker/swarm-v2/log"
	"github.com/docker/swarm-v2/manager/state"
	"github.com/docker/swarm-v2/manager/state/store"
	"github.com/gogo/protobuf/proto"
	"github.com/pivotal-golang/clock"
)

var (
	// ErrConfChangeRefused is returned when there is an issue with the configuration change
	ErrConfChangeRefused = errors.New("raft: propose configuration change refused")
	// ErrApplyNotSpecified is returned during the creation of a raft node when no apply method was provided
	ErrApplyNotSpecified = errors.New("raft: apply method was not specified")
	// ErrAppendEntry is thrown when the node fail to append an entry to the logs
	ErrAppendEntry = errors.New("raft: failed to append entry to logs")
	// ErrSetHardState is returned when the node fails to set the hard state
	ErrSetHardState = errors.New("raft: failed to set the hard state for log append entry")
	// ErrApplySnapshot is returned when the node fails to apply a snapshot
	ErrApplySnapshot = errors.New("raft: failed to apply snapshot on raft node")
	// ErrStopped is returned when an operation was submitted but the node was stopped in the meantime
	ErrStopped = errors.New("raft: failed to process the request: node is stopped")
	// ErrLostLeadership is returned when an operation was submitted but the node lost leader status before it became committed
	ErrLostLeadership = errors.New("raft: failed to process the request: node lost leader status")
	// ErrRequestTooLarge is returned when a raft internal message is too large to be sent
	ErrRequestTooLarge = errors.New("raft: raft message is too large and can't be sent")
	// ErrIDExists is thrown when a node wants to join the existing cluster but its ID already exists
	ErrIDExists = errors.New("raft: can't add node to cluster, node id is a duplicate")
	// ErrIDRemoved is thrown when a node tries to perform an operation on an existing cluster but was removed
	ErrIDRemoved = errors.New("raft: node was removed during cluster lifetime")
	// ErrIDNotFound is thrown when we try an operation on a member that does not exist in the cluster list
	ErrIDNotFound = errors.New("raft: member not found in cluster list")
	// ErrCannotRemoveMember is thrown when we try to remove a member from the cluster but this would result in a loss of quorum
	ErrCannotRemoveMember = errors.New("raft: member cannot be removed, because removing it may result in loss of quorum")
	// ErrMemberRemoved is thrown when a node was removed from the cluster
	ErrMemberRemoved = errors.New("raft: member was removed from the cluster")
)

// LeadershipState indicates whether the node is a leader or follower.
type LeadershipState int

const (
	// IsLeader indicates that the node is a raft leader.
	IsLeader LeadershipState = iota
	// IsFollower indicates that the node is a raft follower.
	IsFollower
)

// Node represents the Raft Node useful
// configuration.
type Node struct {
	raft.Node
	cluster *cluster

	Client         *Raft
	Server         *grpc.Server
	Ctx            context.Context
	cancel         func()
	tlsCredentials credentials.TransportAuthenticator

	Address  string
	StateDir string
	Error    error

	raftStore   *raft.MemoryStorage
	memoryStore *store.MemoryStore
	Config      *raft.Config
	reqIDGen    *idutil.Generator
	wait        *wait
	wal         *wal.WAL
	snapshotter *snap.Snapshotter
	wasLeader   bool
	removed     uint32
	joinAddr    string

	// snapshotInterval is the number of log messages after which a new
	// snapshot should be generated.
	snapshotInterval uint64

	// logEntriesForSlowFollowers is the number of log entries to keep
	// around to sync up slow followers after a snapshot is created.
	logEntriesForSlowFollowers uint64

	confState     raftpb.ConfState
	appliedIndex  uint64
	snapshotIndex uint64

	ticker      clock.Ticker
	sendTimeout time.Duration
	stopCh      chan struct{}
	doneCh      chan struct{}

	leadershipCh   chan LeadershipState
	startNodePeers []raft.Peer

	// used to coordinate shutdown
	stopMu sync.RWMutex

	snapshotInProgress chan uint64
	asyncTasks         sync.WaitGroup
}

// NewNodeOptions provides arguments for NewNode
type NewNodeOptions struct {
	// Addr is the address of this node's listener
	Addr string
	// JoinAddr is the cluster to join. May be an empty string to create
	// a standalone cluster.
	JoinAddr string
	// Config is the raft config.
	Config *raft.Config
	// StateDir is the directory to store durable state.
	StateDir string
	// TickInterval interval is the time interval between raft ticks.
	TickInterval time.Duration
	// SnapshotInterval is the number of log entries that triggers a new
	// snapshot.
	SnapshotInterval uint64 // optional
	// LogEntriesForSlowFollowers is the number of recent log entries to
	// keep when compacting the log.
	LogEntriesForSlowFollowers *uint64 // optional; pointer because 0 is valid
	// ClockSource is a Clock interface to use as a time base.
	// Leave this nil except for tests that are designed not to run in real
	// time.
	ClockSource clock.Clock
	// SendTimeout is the timeout on the sending messages to other raft
	// nodes. Leave this as 0 to get the default value.
	SendTimeout    time.Duration
	TLSCredentials credentials.TransportAuthenticator
}

func init() {
	// TODO(aaronl): Remove once we're no longer generating random IDs.
	rand.Seed(time.Now().UnixNano())
}

// NewNode generates a new Raft node
func NewNode(ctx context.Context, opts NewNodeOptions, leadershipCh chan LeadershipState) (*Node, error) {
	cfg := opts.Config
	if cfg == nil {
		cfg = DefaultNodeConfig()
	}
	if opts.TickInterval == 0 {
		opts.TickInterval = time.Second
	}

	raftStore := raft.NewMemoryStorage()

	ctx, cancel := context.WithCancel(ctx)

	n := &Node{
		Ctx:            ctx,
		cancel:         cancel,
		cluster:        newCluster(),
		tlsCredentials: opts.TLSCredentials,
		raftStore:      raftStore,
		Address:        opts.Addr,
		Config: &raft.Config{
			ElectionTick:    cfg.ElectionTick,
			HeartbeatTick:   cfg.HeartbeatTick,
			Storage:         raftStore,
			MaxSizePerMsg:   cfg.MaxSizePerMsg,
			MaxInflightMsgs: cfg.MaxInflightMsgs,
			Logger:          cfg.Logger,
		},
		snapshotInterval:           1000,
		logEntriesForSlowFollowers: 500,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
		StateDir:     opts.StateDir,
		joinAddr:     opts.JoinAddr,
		leadershipCh: leadershipCh,
		sendTimeout:  2 * time.Second,
	}
	n.memoryStore = store.NewMemoryStore(n)

	if opts.SnapshotInterval != 0 {
		n.snapshotInterval = opts.SnapshotInterval
	}
	if opts.LogEntriesForSlowFollowers != nil {
		n.logEntriesForSlowFollowers = *opts.LogEntriesForSlowFollowers
	}
	if opts.ClockSource == nil {
		n.ticker = clock.NewClock().NewTicker(opts.TickInterval)
	} else {
		n.ticker = opts.ClockSource.NewTicker(opts.TickInterval)
	}
	if opts.SendTimeout != 0 {
		n.sendTimeout = opts.SendTimeout
	}

	if err := n.loadAndStart(ctx); err != nil {
		n.ticker.Stop()
		return nil, err
	}

	snapshot, err := raftStore.Snapshot()
	// Snapshot never returns an error
	if err != nil {
		panic("could not get snapshot of raft store")
	}

	n.confState = snapshot.Metadata.ConfState
	n.appliedIndex = snapshot.Metadata.Index
	n.snapshotIndex = snapshot.Metadata.Index
	n.reqIDGen = idutil.NewGenerator(uint16(n.Config.ID), time.Now())
	n.wait = newWait()

	if n.startNodePeers != nil {
		if n.joinAddr != "" {
			c, err := n.GetRaftClient(n.joinAddr, 10*time.Second)
			if err != nil {
				return nil, err
			}
			defer func() {
				_ = c.Conn.Close()
			}()

			sid := strconv.FormatUint(n.Config.ID, 16)

			ctx, cancel := context.WithTimeout(n.Ctx, 10*time.Second)
			defer cancel()
			resp, err := c.Join(ctx, &api.JoinRequest{
				Node: &api.Member{
					ID:     sid,
					RaftID: n.Config.ID,
					Addr:   n.Address,
				},
			})
			if err != nil {
				return nil, err
			}

			n.Node = raft.StartNode(n.Config, []raft.Peer{})

			if err := n.registerNodes(resp.Members); err != nil {
				return nil, err
			}
		} else {
			n.Node = raft.StartNode(n.Config, n.startNodePeers)
			if err := n.Campaign(n.Ctx); err != nil {
				return nil, err
			}
		}
		return n, nil
	}

	if n.joinAddr != "" {
		n.Config.Logger.Warning("ignoring request to join cluster, because raft state already exists")
	}
	n.Node = raft.RestartNode(n.Config)
	return n, nil
}

// DefaultNodeConfig returns the default config for a
// raft node that can be modified and customized
func DefaultNodeConfig() *raft.Config {
	return &raft.Config{
		HeartbeatTick:   1,
		ElectionTick:    3,
		MaxSizePerMsg:   math.MaxUint16,
		MaxInflightMsgs: 256,
		Logger:          log.L,
	}
}

// MemoryStore returns the memory store that is kept in sync with the raft log.
func (n *Node) MemoryStore() state.WatchableStore {
	return n.memoryStore
}

func (n *Node) walDir() string {
	return filepath.Join(n.StateDir, "wal")
}

func (n *Node) snapDir() string {
	return filepath.Join(n.StateDir, "snap")
}

func (n *Node) loadAndStart(ctx context.Context) error {
	walDir := n.walDir()
	snapDir := n.snapDir()

	if err := os.MkdirAll(snapDir, 0700); err != nil {
		return fmt.Errorf("create snapshot directory error: %v", err)
	}

	// Create a snapshotter
	n.snapshotter = snap.New(snapDir)

	if !wal.Exist(walDir) {
		// FIXME(aaronl): Generate unique ID on remote side if joining
		// an existing cluster.
		n.Config.ID = uint64(rand.Int63()) + 1

		sid := strconv.FormatUint(n.Config.ID, 16)

		raftNode := &api.Member{
			ID:     sid,
			RaftID: n.Config.ID,
			Addr:   n.Address,
		}
		metadata, err := raftNode.Marshal()
		if err != nil {
			return fmt.Errorf("error marshalling raft node: %v", err)
		}
		n.wal, err = wal.Create(walDir, metadata)
		if err != nil {
			return fmt.Errorf("create wal error: %v", err)
		}

		n.cluster.addMember(&member{Member: raftNode})
		n.startNodePeers = []raft.Peer{{ID: n.Config.ID, Context: metadata}}

		return nil
	}

	// Load snapshot data
	snapshot, err := n.snapshotter.Load()
	if err != nil && err != snap.ErrNoSnapshot {
		return err
	}

	if snapshot != nil {
		// Load the snapshot data into the store
		if err := n.restoreFromSnapshot(snapshot.Data); err != nil {
			return err
		}
	}

	// Read logs to fully catch up store
	if err := n.readWAL(ctx, snapshot); err != nil {
		return err
	}

	n.Node = raft.RestartNode(n.Config)
	return nil
}

func (n *Node) readWAL(ctx context.Context, snapshot *raftpb.Snapshot) (err error) {
	var (
		walsnap  walpb.Snapshot
		metadata []byte
		st       raftpb.HardState
		ents     []raftpb.Entry
	)

	if snapshot != nil {
		walsnap.Index = snapshot.Metadata.Index
		walsnap.Term = snapshot.Metadata.Term
	}

	repaired := false
	for {
		if n.wal, err = wal.Open(n.walDir(), walsnap); err != nil {
			return fmt.Errorf("open wal error: %v", err)
		}
		if metadata, st, ents, err = n.wal.ReadAll(); err != nil {
			if err := n.wal.Close(); err != nil {
				return err
			}
			// we can only repair ErrUnexpectedEOF and we never repair twice.
			if repaired || err != io.ErrUnexpectedEOF {
				return fmt.Errorf("read wal error (%v) and cannot be repaired", err)
			}
			if !wal.Repair(n.walDir()) {
				return fmt.Errorf("WAL error (%v) cannot be repaired", err)
			}
			log.G(ctx).Infof("repaired WAL error (%v)", err)
			repaired = true
			continue
		}
		break
	}

	defer func() {
		if err != nil {
			if walErr := n.wal.Close(); walErr != nil {
				n.Config.Logger.Errorf("error closing raft WAL: %v", walErr)
			}
		}
	}()

	var raftNode api.Member
	if err := raftNode.Unmarshal(metadata); err != nil {
		return fmt.Errorf("error unmarshalling wal metadata: %v", err)
	}
	n.Config.ID = raftNode.RaftID

	if snapshot != nil {
		if err := n.raftStore.ApplySnapshot(*snapshot); err != nil {
			return err
		}
	}
	if err := n.raftStore.SetHardState(st); err != nil {
		return err
	}
	if err := n.raftStore.Append(ents); err != nil {
		return err
	}

	return nil
}

// Run is the main loop for a Raft node, it goes along the state machine,
// acting on the messages received from other Raft nodes in the cluster.
//
// Before running the main loop, it first starts the raft node based on saved
// cluster state. If no saved state exists, it starts a single-node cluster.
func (n *Node) Run(ctx context.Context) error {
	defer close(n.doneCh)

	for {
		select {
		case <-n.ticker.C():
			n.Tick()

		case rd := <-n.Ready():
			// Save entries to storage
			if err := n.saveToStorage(rd.HardState, rd.Entries, rd.Snapshot); err != nil {
				n.Config.Logger.Error(err)
			}

			// Send raft messages to peers
			if err := n.send(rd.Messages); err != nil {
				n.Config.Logger.Error(err)
			}

			// Apply snapshot to memory store. The snapshot
			// was applied to the raft store in
			// saveToStorage.
			if !raft.IsEmptySnap(rd.Snapshot) {
				// Load the snapshot data into the store
				if err := n.restoreFromSnapshot(rd.Snapshot.Data); err != nil {
					n.Config.Logger.Error(err)
				}
				n.appliedIndex = rd.Snapshot.Metadata.Index
				n.snapshotIndex = rd.Snapshot.Metadata.Index
				n.confState = rd.Snapshot.Metadata.ConfState
			}

			// Process committed entries
			for _, entry := range rd.CommittedEntries {
				if err := n.processCommitted(entry); err != nil {
					n.Config.Logger.Error(err)
				}
			}

			// Trigger a snapshot every once in awhile
			if n.snapshotInProgress == nil && n.appliedIndex-n.snapshotIndex >= n.snapshotInterval {
				n.doSnapshot()
			}

			// If we cease to be the leader, we must cancel
			// any proposals that are currently waiting for
			// a quorum to acknowledge them. It is still
			// possible for these to become committed, but
			// if that happens we will apply them as any
			// follower would.
			if rd.SoftState != nil {
				if n.wasLeader && rd.SoftState.RaftState != raft.StateLeader {
					n.wasLeader = false
					n.wait.cancelAll()
					if n.leadershipCh != nil {
						n.leadershipCh <- IsFollower
					}
				} else if !n.wasLeader && rd.SoftState.RaftState == raft.StateLeader {
					n.wasLeader = true
					if n.leadershipCh != nil {
						n.leadershipCh <- IsLeader
					}
				}
			}

			// If the node was removed from other members,
			// send back an error to the caller to start
			// the shutdown process.
			if n.mustStop() {
				close(n.stopCh)
				return ErrMemberRemoved
			}

			// Advance the state machine
			n.Advance()

		case snapshotIndex := <-n.snapshotInProgress:
			if snapshotIndex > n.snapshotIndex {
				n.snapshotIndex = snapshotIndex
			}
			n.snapshotInProgress = nil
		case <-n.stopCh:
			n.stop()
			return nil
		}
	}
}

// Shutdown stops the raft node processing loop.
// Calling Shutdown on an already stopped node
// will result in a panic.
func (n *Node) Shutdown() {
	select {
	case <-n.doneCh:
		n.stop()
	default:
		close(n.stopCh)
		<-n.doneCh
	}
}

func (n *Node) stop() {
	n.stopMu.Lock()
	defer n.stopMu.Unlock()

	n.cancel()
	n.asyncTasks.Wait()

	members := n.cluster.listMembers()
	for _, member := range members {
		if member.Client != nil && member.Client.Conn != nil {
			_ = member.Client.Conn.Close()
		}
	}
	n.Stop()
	if err := n.wal.Close(); err != nil {
		n.Config.Logger.Error(err)
	}
	// TODO(stevvooe): Handle ctx.Done()
}

// IsLeader checks if we are the leader or not
func (n *Node) IsLeader() bool {
	if n.Node.Status().Lead == n.Config.ID {
		return true
	}
	return false
}

// Leader returns the id of the leader
func (n *Node) Leader() uint64 {
	return n.Node.Status().Lead
}

// Join asks to a member of the raft to propose
// a configuration change and add us as a member thus
// beginning the log replication process. This method
// is called from an aspiring member to an existing member
func (n *Node) Join(ctx context.Context, req *api.JoinRequest) (*api.JoinResponse, error) {
	agentID, err := ca.AuthorizeRole(ctx, []string{ca.ManagerRole})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("(*Join). message from node %s", agentID)

	// can't stop the raft node while an async RPC is in progress
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	if n.Node == nil {
		return nil, ErrStopped
	}

	if n.cluster.isIDRemoved(req.Node.RaftID) {
		return nil, ErrIDRemoved
	}

	// We submit a configuration change only if the node was not registered yet
	if n.cluster.getMember(req.Node.RaftID) == nil {
		err = n.addMember(ctx, req.Node)
		if err != nil {
			return nil, err
		}
	}

	var nodes []*api.Member
	for _, node := range n.cluster.listMembers() {
		nodes = append(nodes, &api.Member{
			ID:     node.ID,
			RaftID: node.RaftID,
			Addr:   node.Addr,
		})
	}

	return &api.JoinResponse{Members: nodes}, nil
}

// addMember submits a configuration change to add a new member on the raft cluster.
func (n *Node) addMember(ctx context.Context, node *api.Member) error {
	meta, err := node.Marshal()
	if err != nil {
		return err
	}

	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  node.RaftID,
		Context: meta,
	}

	// Wait for a raft round to process the configuration change
	err = n.configure(ctx, cc)
	return err
}

// Leave asks to a member of the raft to remove
// us from the raft cluster. This method is called
// from a member who is willing to leave its raft
// membership to an active member of the raft
func (n *Node) Leave(ctx context.Context, req *api.LeaveRequest) (*api.LeaveResponse, error) {
	agentID, err := ca.AuthorizeRole(ctx, []string{ca.ManagerRole})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("(*Leave). message from node %s", agentID)

	// can't stop the raft node while an async RPC is in progress
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	if n.Node == nil {
		return nil, ErrStopped
	}

	err = n.RemoveMember(ctx, req.Node.RaftID)
	if err != nil {
		return nil, err
	}

	return &api.LeaveResponse{}, nil
}

// RemoveMember submits a configuration change to remove a member from the raft cluster.
func (n *Node) RemoveMember(ctx context.Context, id uint64) error {
	// TODO(abronan): this can race if multiple removes are processed, we should
	// send all the requests to the Leader and track pending removals.
	if n.cluster.CanRemoveMember(n.Config.ID, id) {
		cc := raftpb.ConfChange{
			ID:      id,
			Type:    raftpb.ConfChangeRemoveNode,
			NodeID:  id,
			Context: []byte(""),
		}

		err := n.configure(ctx, cc)
		return err
	}

	return ErrCannotRemoveMember
}

// ProcessRaftMessage calls 'Step' which advances the
// raft state machine with the provided message on the
// receiving node
func (n *Node) ProcessRaftMessage(ctx context.Context, msg *api.ProcessRaftMessageRequest) (*api.ProcessRaftMessageResponse, error) {
	agentID, err := ca.AuthorizeRole(ctx, []string{ca.ManagerRole})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("(*ProcessRaftMessage). message from node %s", agentID)

	// Don't process the message if this comes from
	// a node in the remove set
	if n.cluster.isIDRemoved(msg.Message.From) {
		return nil, ErrIDRemoved
	}

	// can't stop the raft node while an async RPC is in progress
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()
	if n.Node == nil {
		return nil, ErrStopped
	}

	err = n.Step(n.Ctx, *msg.Message)
	if err != nil {
		return nil, err
	}

	return &api.ProcessRaftMessageResponse{}, nil
}

// LeaderAddr returns address of current cluster leader.
// With this method Node satisfies raftpicker.AddrSelector interface.
func (n *Node) LeaderAddr() (string, error) {
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()
	if n.Node == nil {
		return "", ErrStopped
	}
	ms := n.cluster.listMembers()
	l := ms[n.Leader()]
	if l == nil {
		return "", fmt.Errorf("incorrect cluster state")
	}
	return l.Addr, nil
}

// registerNode registers a new node on the cluster
func (n *Node) registerNode(node *api.Member) error {
	var client *Raft
	// Avoid opening a connection with ourself
	if node.RaftID != n.Config.ID {
		// We don't want to impose a timeout on the grpc connection. It
		// should keep retrying as long as necessary, in case the peer
		// is temporarily unavailable.
		var err error
		if client, err = n.GetRaftClient(node.Addr, 0); err != nil {
			return err
		}
	}
	n.cluster.addMember(&member{Member: node, Client: client})
	return nil
}

// registerNodes registers a set of nodes in the cluster
func (n *Node) registerNodes(nodes []*api.Member) error {
	for _, node := range nodes {
		if err := n.registerNode(node); err != nil {
			return err
		}
	}

	return nil
}

// deregisterNode unregisters a node that has died or
// has gracefully left the raft subsystem
func (n *Node) deregisterNode(id uint64) error {
	// Do not unregister yourself
	if n.Config.ID == id {
		return nil
	}

	peer := n.cluster.getMember(id)
	if peer == nil {
		return ErrIDNotFound
	}

	n.cluster.removeMember(id)

	return nil
}

// ProposeValue calls Propose on the raft and waits
// on the commit log action before returning a result
func (n *Node) ProposeValue(ctx context.Context, storeAction []*api.StoreAction, cb func()) error {
	_, err := n.processInternalRaftRequest(ctx, &api.InternalRaftRequest{Action: storeAction}, cb)
	if err != nil {
		return err
	}
	return nil
}

// GetVersion returns the sequence information for the current raft round.
func (n *Node) GetVersion() *api.Version {
	status := n.Node.Status()
	return &api.Version{Index: status.Commit}
}

// GetMemberlist returns the current list of raft members in the cluster.
func (n *Node) GetMemberlist() map[uint64]*api.Member {
	memberlist := make(map[uint64]*api.Member)
	members := n.cluster.listMembers()
	leaderID := n.Leader()

	for id, member := range members {
		status := api.MemberStatus_REACHABLE
		leader := false

		if member.RaftID != n.Config.ID {
			connState, err := member.Client.Conn.State()
			if err != nil || connState != grpc.Ready {
				status = api.MemberStatus_UNREACHABLE
			}
		}

		if member.RaftID == leaderID {
			leader = true
		}

		sid := strconv.FormatUint(member.RaftID, 16)

		memberlist[id] = &api.Member{
			ID:     sid,
			RaftID: member.RaftID,
			Addr:   member.Addr,
			Status: api.MemberStatus{
				Leader: leader,
				State:  status,
			},
		}
	}

	return memberlist
}

// mustStop checks if the raft node must be stopped
// because it was removed from the cluster from
// other members
func (n *Node) mustStop() bool {
	return atomic.LoadUint32(&n.removed) == 1
}

// Saves a log entry to our Store
func (n *Node) saveToStorage(hardState raftpb.HardState, entries []raftpb.Entry, snapshot raftpb.Snapshot) (err error) {
	if !raft.IsEmptySnap(snapshot) {
		if err := n.saveSnapshot(snapshot); err != nil {
			return ErrApplySnapshot
		}
		if err = n.raftStore.ApplySnapshot(snapshot); err != nil {
			return ErrApplySnapshot
		}
	}

	if err := n.wal.Save(hardState, entries); err != nil {
		// TODO(aaronl): These error types should really wrap more
		// detailed errors.
		return ErrApplySnapshot
	}

	if err = n.raftStore.Append(entries); err != nil {
		return ErrAppendEntry
	}

	return nil
}

func (n *Node) saveSnapshot(snapshot raftpb.Snapshot) error {
	err := n.wal.SaveSnapshot(walpb.Snapshot{
		Index: snapshot.Metadata.Index,
		Term:  snapshot.Metadata.Term,
	})
	if err != nil {
		return err
	}
	err = n.snapshotter.SaveSnap(snapshot)
	if err != nil {
		return err
	}
	err = n.wal.ReleaseLockTo(snapshot.Metadata.Index)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) doSnapshot() {
	snapshot := api.Snapshot{Version: api.Snapshot_V0}
	for _, member := range n.cluster.listMembers() {
		snapshot.Membership.Members = append(snapshot.Membership.Members,
			&api.Member{
				ID:     member.ID,
				RaftID: member.RaftID,
				Addr:   member.Addr,
			})
	}
	snapshot.Membership.Removed = n.cluster.listRemoved()

	viewStarted := make(chan struct{})
	n.asyncTasks.Add(1)
	n.snapshotInProgress = make(chan uint64, 1) // buffered in case Shutdown is called during the snapshot
	go func(appliedIndex, snapshotIndex uint64) {
		defer func() {
			n.asyncTasks.Done()
			n.snapshotInProgress <- snapshotIndex
		}()

		var err error
		n.memoryStore.View(func(tx state.ReadTx) {
			close(viewStarted)

			var storeSnapshot *pb.StoreSnapshot
			storeSnapshot, err = n.memoryStore.Save(tx)
			snapshot.Store = *storeSnapshot
		})
		if err != nil {
			n.Config.Logger.Error(err)
			return
		}

		d, err := snapshot.Marshal()
		if err != nil {
			n.Config.Logger.Error(err)
			return
		}
		snap, err := n.raftStore.CreateSnapshot(appliedIndex, &n.confState, d)
		if err == nil {
			if err := n.saveSnapshot(snap); err != nil {
				n.Config.Logger.Error(err)
				return
			}
			snapshotIndex = appliedIndex

			if appliedIndex > n.logEntriesForSlowFollowers {
				err := n.raftStore.Compact(appliedIndex - n.logEntriesForSlowFollowers)
				if err != nil && err != raft.ErrCompacted {
					n.Config.Logger.Error(err)
				}
			}
		} else if err != raft.ErrSnapOutOfDate {
			n.Config.Logger.Error(err)
		}
	}(n.appliedIndex, n.snapshotIndex)

	// Wait for the goroutine to establish a read transaction, to make
	// sure it sees the state as of this moment.
	<-viewStarted
}

func (n *Node) restoreFromSnapshot(data []byte) error {
	var snapshot api.Snapshot
	if err := snapshot.Unmarshal(data); err != nil {
		return err
	}
	if snapshot.Version != api.Snapshot_V0 {
		return fmt.Errorf("unrecognized snapshot version %d", snapshot.Version)
	}

	if err := n.memoryStore.Restore(&snapshot.Store); err != nil {
		return err
	}

	n.cluster.clear()
	for _, member := range snapshot.Membership.Members {
		if err := n.registerNode(&api.Member{ID: member.ID, RaftID: member.RaftID, Addr: member.Addr}); err != nil {
			return err
		}
	}
	for _, removedMember := range snapshot.Membership.Removed {
		n.cluster.removeMember(removedMember)
	}

	return nil
}

// Sends a series of messages to members in the raft
func (n *Node) send(messages []raftpb.Message) error {
	members := n.cluster.listMembers()

	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	for _, m := range messages {
		// Process locally
		if m.To == n.Config.ID {
			if err := n.Step(n.Ctx, m); err != nil {
				return err
			}
			continue
		}

		// If node is an active raft member send the message
		if member, ok := members[m.To]; ok {
			n.asyncTasks.Add(1)
			go n.sendToMember(member, m)
		}
	}

	return nil
}

func (n *Node) sendToMember(member *member, m raftpb.Message) {
	ctx, cancel := context.WithTimeout(n.Ctx, n.sendTimeout)

	_, err := member.Client.ProcessRaftMessage(ctx, &api.ProcessRaftMessageRequest{Message: &m})
	if err != nil {
		if grpc.ErrorDesc(err) == ErrIDRemoved.Error() {
			atomic.StoreUint32(&n.removed, 1)
		}
		if m.Type == raftpb.MsgSnap {
			n.ReportSnapshot(m.To, raft.SnapshotFailure)
		}
		if member == nil {
			panic("member is nil")
		}
		if n.Node == nil {
			panic("node is nil")
		}
		n.ReportUnreachable(member.RaftID)
	} else if m.Type == raftpb.MsgSnap {
		n.ReportSnapshot(m.To, raft.SnapshotFinish)
	}
	cancel()
	n.asyncTasks.Done()
}

type applyResult struct {
	resp proto.Message
	err  error
}

// processInternalRaftRequest sends a message through consensus
// and then waits for it to be applies to the server. It will
// block until the change is performed or there is an error
func (n *Node) processInternalRaftRequest(ctx context.Context, r *api.InternalRaftRequest, cb func()) (proto.Message, error) {
	r.ID = n.reqIDGen.Next()

	ch := n.wait.register(r.ID, cb)

	// Do this check after calling register to avoid a race.
	if !n.IsLeader() {
		n.wait.cancel(r.ID)
		return nil, ErrLostLeadership
	}

	data, err := r.Marshal()
	if err != nil {
		n.wait.cancel(r.ID)
		return nil, err
	}

	if len(data) > state.MaxTransactionBytes {
		n.wait.cancel(r.ID)
		return nil, ErrRequestTooLarge
	}

	err = n.Propose(ctx, data)
	if err != nil {
		n.wait.cancel(r.ID)
		return nil, err
	}

	select {
	case x, ok := <-ch:
		if ok {
			res := x.(*applyResult)
			return res.resp, res.err
		}
		return nil, ErrLostLeadership
	case <-n.stopCh:
		n.wait.cancel(r.ID)
		return nil, ErrStopped
	case <-ctx.Done():
		n.wait.cancel(r.ID)
		return nil, ctx.Err()
	}
}

// configure sends a configuration change through consensus and
// then waits for it to be applied to the server. It will block
// until the change is performed or there is an error.
func (n *Node) configure(ctx context.Context, cc raftpb.ConfChange) error {
	cc.ID = n.reqIDGen.Next()
	ch := n.wait.register(cc.ID, nil)

	if err := n.ProposeConfChange(ctx, cc); err != nil {
		n.wait.trigger(cc.ID, nil)
		return err
	}

	select {
	case x := <-ch:
		if err, ok := x.(error); ok {
			return err
		}
		if x != nil {
			log.G(ctx).Panic("raft: configuration change error, return type should always be error")
		}
		return nil
	case <-ctx.Done():
		n.wait.trigger(cc.ID, nil)
		return ctx.Err()
	case <-n.stopCh:
		return ErrStopped
	}
}

func (n *Node) processCommitted(entry raftpb.Entry) error {
	// Process a normal entry
	if entry.Type == raftpb.EntryNormal && entry.Data != nil {
		if err := n.processEntry(entry); err != nil {
			return err
		}
	}

	// Process a configuration change (add/remove node)
	if entry.Type == raftpb.EntryConfChange {
		n.processConfChange(entry)
	}

	n.appliedIndex = entry.Index
	return nil
}

func (n *Node) processEntry(entry raftpb.Entry) error {
	r := &api.InternalRaftRequest{}
	err := proto.Unmarshal(entry.Data, r)
	if err != nil {
		return err
	}

	if r.Action == nil {
		return nil
	}

	if !n.wait.trigger(r.ID, &applyResult{resp: r, err: nil}) {
		// There was no wait on this ID, meaning we don't have a
		// transaction in progress that would be committed to the
		// memory store by the "trigger" call. Either a different node
		// wrote this to raft, or we wrote it before losing the leader
		// position and cancelling the transaction. Create a new
		// transaction to commit the data.

		err := n.memoryStore.ApplyStoreActions(r.Action)
		if err != nil {
			log.G(context.Background()).Errorf("error applying actions from raft: %v", err)
		}
	}
	return nil
}

func (n *Node) processConfChange(entry raftpb.Entry) {
	var (
		err error
		cc  raftpb.ConfChange
	)

	if err = cc.Unmarshal(entry.Data); err != nil {
		n.wait.trigger(cc.ID, err)
	}

	if n.cluster.isIDRemoved(cc.NodeID) {
		n.wait.trigger(cc.ID, ErrIDRemoved)
	}

	switch cc.Type {
	case raftpb.ConfChangeAddNode:
		err = n.applyAddNode(cc)
	case raftpb.ConfChangeRemoveNode:
		err = n.applyRemoveNode(cc)
	}

	if err != nil {
		n.wait.trigger(cc.ID, err)
	}

	n.confState = *n.ApplyConfChange(cc)
	n.wait.trigger(cc.ID, nil)
}

// applyAddNode is called when we receive a ConfChange
// from a member in the raft cluster, this adds a new
// node to the existing raft cluster
func (n *Node) applyAddNode(cc raftpb.ConfChange) error {
	if n.cluster.getMember(cc.NodeID) != nil {
		return ErrIDExists
	}

	member := &api.Member{}
	err := proto.Unmarshal(cc.Context, member)
	if err != nil {
		return err
	}

	// ID must be non zero
	if member.RaftID == 0 {
		return nil
	}

	if err = n.registerNode(member); err != nil {
		return err
	}
	return nil
}

// applyRemoveNode is called when we receive a ConfChange
// from a member in the raft cluster, this removes a node
// from the existing raft cluster
func (n *Node) applyRemoveNode(cc raftpb.ConfChange) (err error) {
	if n.cluster.getMember(cc.NodeID) == nil {
		return ErrIDNotFound
	}

	// If the node from where the remove is issued is
	// a follower and the leader steps down, Campaign
	// to be the leader.
	if cc.NodeID == n.Leader() {
		if err = n.Campaign(n.Ctx); err != nil {
			return err
		}
	}

	// Unregister the node
	if err = n.deregisterNode(cc.NodeID); err != nil {
		return err
	}

	return nil
}

// GetRaftClient returns a raft client object to communicate
// with other raft members
func (n *Node) GetRaftClient(addr string, timeout time.Duration) (*Raft, error) {
	conn, err := dial(addr, "tcp", n.tlsCredentials, timeout)
	if err != nil {
		return nil, err
	}

	return &Raft{
		RaftClient: api.NewRaftClient(conn),
		Conn:       conn,
	}, nil
}
