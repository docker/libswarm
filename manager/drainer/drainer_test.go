package drainer

import (
	"testing"
	"time"

	"github.com/docker/go-events"
	"github.com/docker/swarm-v2/manager/state"
	objectspb "github.com/docker/swarm-v2/pb/docker/cluster/objects"
	specspb "github.com/docker/swarm-v2/pb/docker/cluster/specs"
	typespb "github.com/docker/swarm-v2/pb/docker/cluster/types"
	"github.com/stretchr/testify/assert"
)

func TestDrainer(t *testing.T) {
	initialNodeSet := []*objectspb.Node{
		{
			ID: "id1",
			Spec: &specspb.NodeSpec{
				Meta: specspb.Meta{
					Name: "name1",
				},
				Availability: specspb.NodeAvailabilityActive,
			},
			Status: typespb.NodeStatus{
				State: typespb.NodeStatus_READY,
			},
		},
		{
			ID: "id2",
			Spec: &specspb.NodeSpec{
				Meta: specspb.Meta{
					Name: "name2",
				},
				Availability: specspb.NodeAvailabilityActive,
			},
			Status: typespb.NodeStatus{
				State: typespb.NodeStatus_DOWN,
			},
		},
		{
			ID: "id3",
			Spec: &specspb.NodeSpec{
				Meta: specspb.Meta{
					Name: "name3",
				},
				Availability: specspb.NodeAvailabilityActive,
			},
			Status: typespb.NodeStatus{
				State: typespb.NodeStatus_DISCONNECTED,
			},
		},
		{
			ID: "id4",
			Spec: &specspb.NodeSpec{
				Meta: specspb.Meta{
					Name: "name4",
				},
				Availability: specspb.NodeAvailabilityPause,
			},
			Status: typespb.NodeStatus{
				State: typespb.NodeStatus_READY,
			},
		},
		{
			ID: "id5",
			Spec: &specspb.NodeSpec{
				Meta: specspb.Meta{
					Name: "name5",
				},
				Availability: specspb.NodeAvailabilityDrain,
			},
			Status: typespb.NodeStatus{
				State: typespb.NodeStatus_READY,
			},
		},
	}

	initialTaskSet := []*objectspb.Task{
		// Task not assigned to any node
		{
			ID:   "id0",
			Spec: &specspb.TaskSpec{},
			Meta: specspb.Meta{
				Name: "name0",
			},
		},
		// Tasks assigned to the nodes defined above
		{
			ID:   "id1",
			Spec: &specspb.TaskSpec{},
			Meta: specspb.Meta{
				Name: "name1",
			},
			NodeID: "id1",
		},
		{
			ID:   "id2",
			Spec: &specspb.TaskSpec{},
			Meta: specspb.Meta{
				Name: "name2",
			},
			NodeID: "id2",
		},
		{
			ID:   "id3",
			Spec: &specspb.TaskSpec{},
			Meta: specspb.Meta{
				Name: "name3",
			},
			NodeID: "id3",
		},
		{
			ID:   "id4",
			Spec: &specspb.TaskSpec{},
			Meta: specspb.Meta{
				Name: "name4",
			},
			NodeID: "id4",
		},
		{
			ID:   "id5",
			Spec: &specspb.TaskSpec{},
			Meta: specspb.Meta{
				Name: "name5",
			},
			NodeID: "id5",
		},
	}

	store := state.NewMemoryStore(nil)
	assert.NotNil(t, store)

	err := store.Update(func(tx state.Tx) error {
		// Prepoulate nodes
		for _, n := range initialNodeSet {
			assert.NoError(t, tx.Nodes().Create(n))
		}

		// Prepopulate tasks
		for _, task := range initialTaskSet {
			assert.NoError(t, tx.Tasks().Create(task))
		}
		return nil
	})
	assert.NoError(t, err)

	drainer := New(store)

	watch, cancel := state.Watch(store.WatchQueue(), state.EventDeleteTask{})
	defer cancel()

	go func() {
		assert.NoError(t, drainer.Run())
	}()

	// id2, id3, and id5 should be deleted immediately
	// NOTE: we can assume these will be emitted in lexical order because
	// of the way indexing works in the store. If that ever changes, this
	// part of the test might need to become more flexible.
	deletion1 := watchDeleteTask(t, watch)
	assert.Equal(t, deletion1.ID, "id2")
	assert.Equal(t, deletion1.NodeID, "id2")
	deletion2 := watchDeleteTask(t, watch)
	assert.Equal(t, deletion2.ID, "id3")
	assert.Equal(t, deletion2.NodeID, "id3")
	deletion3 := watchDeleteTask(t, watch)
	assert.Equal(t, deletion3.ID, "id5")
	assert.Equal(t, deletion3.NodeID, "id5")

	// Create a new task, assigned to node id2
	err = store.Update(func(tx state.Tx) error {
		task := initialTaskSet[2].Copy()
		task.ID = "newtask"
		task.NodeID = "id2"
		assert.NoError(t, tx.Tasks().Create(task))
		return nil
	})
	assert.NoError(t, err)

	deletion4 := watchDeleteTask(t, watch)
	assert.Equal(t, deletion4.ID, "newtask")
	assert.Equal(t, deletion4.NodeID, "id2")

	// Set node id4 to the DRAINED state
	err = store.Update(func(tx state.Tx) error {
		n := initialNodeSet[3].Copy()
		n.Spec.Availability = specspb.NodeAvailabilityDrain
		assert.NoError(t, tx.Nodes().Update(n))
		return nil
	})
	assert.NoError(t, err)

	deletion5 := watchDeleteTask(t, watch)
	assert.Equal(t, deletion5.ID, "id4")
	assert.Equal(t, deletion5.NodeID, "id4")

	// Delete node id1
	err = store.Update(func(tx state.Tx) error {
		assert.NoError(t, tx.Nodes().Delete("id1"))
		return nil
	})
	assert.NoError(t, err)

	deletion6 := watchDeleteTask(t, watch)
	assert.Equal(t, deletion6.ID, "id1")
	assert.Equal(t, deletion6.NodeID, "id1")

	drainer.Stop()
}

func watchDeleteTask(t *testing.T, watch chan events.Event) *objectspb.Task {
	for {
		select {
		case event := <-watch:
			if task, ok := event.(state.EventDeleteTask); ok {
				return task.Task
			}
		case <-time.After(time.Second):
			t.Fatalf("no task deletion")
		}
	}
}
