package orchestrator

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

func TestOrchestrator(t *testing.T) {
	store := state.NewMemoryStore(nil)
	assert.NotNil(t, store)

	orchestrator := New(store)

	watch, cancel := state.Watch(store.WatchQueue() /*state.EventCreateTask{}, state.EventDeleteTask{}*/)
	defer cancel()

	// Create a job with two instances specified before the orchestrator is
	// started. This should result in two tasks when the orchestrator
	// starts up.
	err := store.Update(func(tx state.Tx) error {
		j1 := &objectspb.Job{
			ID: "id1",
			Spec: &specspb.JobSpec{
				Meta: specspb.Meta{
					Name: "name1",
				},
				Template: &specspb.TaskSpec{},
				Orchestration: &specspb.JobSpec_Service{
					Service: &specspb.JobSpec_ServiceJob{
						Instances: 2,
					},
				},
			},
		}
		assert.NoError(t, tx.Jobs().Create(j1))
		return nil
	})
	assert.NoError(t, err)

	// Start the orchestrator.
	go func() {
		assert.NoError(t, orchestrator.Run())
	}()

	observedTask1 := watchTaskCreate(t, watch)
	assert.Equal(t, observedTask1.Status.State, typespb.TaskStateNew)
	assert.Equal(t, observedTask1.Meta.Name, "name1")

	observedTask2 := watchTaskCreate(t, watch)
	assert.Equal(t, observedTask2.Status.State, typespb.TaskStateNew)
	assert.Equal(t, observedTask2.Meta.Name, "name1")

	// Create a second job.
	err = store.Update(func(tx state.Tx) error {
		j2 := &objectspb.Job{
			ID: "id2",
			Spec: &specspb.JobSpec{
				Meta: specspb.Meta{
					Name: "name2",
				},
				Template: &specspb.TaskSpec{},
				Orchestration: &specspb.JobSpec_Service{
					Service: &specspb.JobSpec_ServiceJob{
						Instances: 1,
					},
				},
			},
		}
		assert.NoError(t, tx.Jobs().Create(j2))
		return nil
	})
	assert.NoError(t, err)

	observedTask3 := watchTaskCreate(t, watch)
	assert.Equal(t, observedTask3.Status.State, typespb.TaskStateNew)
	assert.Equal(t, observedTask3.Meta.Name, "name2")

	// Update a job to scale it out to 3 instances
	err = store.Update(func(tx state.Tx) error {
		j2 := &objectspb.Job{
			ID: "id2",
			Spec: &specspb.JobSpec{
				Meta: specspb.Meta{
					Name: "name2",
				},
				Template: &specspb.TaskSpec{},
				Orchestration: &specspb.JobSpec_Service{
					Service: &specspb.JobSpec_ServiceJob{
						Instances: 3,
					},
				},
			},
		}
		assert.NoError(t, tx.Jobs().Update(j2))
		return nil
	})
	assert.NoError(t, err)

	observedTask4 := watchTaskCreate(t, watch)
	assert.Equal(t, observedTask4.Status.State, typespb.TaskStateNew)
	assert.Equal(t, observedTask4.Meta.Name, "name2")

	observedTask5 := watchTaskCreate(t, watch)
	assert.Equal(t, observedTask5.Status.State, typespb.TaskStateNew)
	assert.Equal(t, observedTask5.Meta.Name, "name2")

	// Now scale it back down to 1 instance
	err = store.Update(func(tx state.Tx) error {
		j2 := &objectspb.Job{
			ID: "id2",
			Spec: &specspb.JobSpec{
				Meta: specspb.Meta{
					Name: "name2",
				},
				Template: &specspb.TaskSpec{},
				Orchestration: &specspb.JobSpec_Service{
					Service: &specspb.JobSpec_ServiceJob{
						Instances: 1,
					},
				},
			},
		}
		assert.NoError(t, tx.Jobs().Update(j2))
		return nil
	})
	assert.NoError(t, err)

	observedDeletion1 := watchTaskDelete(t, watch)
	assert.Equal(t, observedDeletion1.Status.State, typespb.TaskStateNew)
	assert.Equal(t, observedDeletion1.Meta.Name, "name2")

	observedDeletion2 := watchTaskDelete(t, watch)
	assert.Equal(t, observedDeletion2.Status.State, typespb.TaskStateNew)
	assert.Equal(t, observedDeletion2.Meta.Name, "name2")

	// There should be one remaining task attached to job id2/name2.
	var tasks []*objectspb.Task
	err = store.View(func(readTx state.ReadTx) error {
		var err error
		tasks, err = readTx.Tasks().Find(state.ByJobID("id2"))
		return err
	})
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)

	// Delete the remaining task directly. It should be recreated by the
	// orchestrator.
	err = store.Update(func(tx state.Tx) error {
		assert.NoError(t, tx.Tasks().Delete(tasks[0].ID))
		return nil
	})
	assert.NoError(t, err)

	// Consume deletion event from the queue
	watchTaskDelete(t, watch)

	observedTask6 := watchTaskCreate(t, watch)
	assert.Equal(t, observedTask6.Status.State, typespb.TaskStateNew)
	assert.Equal(t, observedTask6.Meta.Name, "name2")

	// Delete the job. Its remaining task should go away.
	err = store.Update(func(tx state.Tx) error {
		assert.NoError(t, tx.Jobs().Delete("id2"))
		return nil
	})
	assert.NoError(t, err)

	observedDeletion3 := watchTaskDelete(t, watch)
	assert.Equal(t, observedDeletion3.Status.State, typespb.TaskStateNew)
	assert.Equal(t, observedDeletion3.Meta.Name, "name2")

	orchestrator.Stop()
}

func watchTaskCreate(t *testing.T, watch chan events.Event) *objectspb.Task {
	for {
		select {
		case event := <-watch:
			if task, ok := event.(state.EventCreateTask); ok {
				return task.Task
			}
			if _, ok := event.(state.EventDeleteTask); ok {
				t.Fatal("got EventDeleteTask when expecting EventCreateTask")
			}
		case <-time.After(time.Second):
			t.Fatal("no task assignment")
		}
	}
}

func watchTaskDelete(t *testing.T, watch chan events.Event) *objectspb.Task {
	for {
		select {
		case event := <-watch:
			if task, ok := event.(state.EventDeleteTask); ok {
				return task.Task
			}
			if _, ok := event.(state.EventCreateTask); ok {
				t.Fatal("got EventCreateTask when expecting EventDeleteTask")
			}
		case <-time.After(time.Second):
			t.Fatal("no task deletion")
		}
	}
}
