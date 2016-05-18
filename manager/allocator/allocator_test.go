package allocator

import (
	"net"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/docker/go-events"
	"github.com/docker/swarm-v2/api"
	"github.com/docker/swarm-v2/manager/state"
	"github.com/docker/swarm-v2/manager/state/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	s = store.NewMemoryStore(nil)
)

func TestAllocator(t *testing.T) {
	assert.NotNil(t, s)

	a, err := New(s)
	assert.NoError(t, err)
	assert.NotNil(t, a)

	// Try adding some objects to store before allocator is started
	assert.NoError(t, s.Update(func(tx store.Tx) error {
		n1 := &api.Network{
			ID: "testID1",
			Spec: api.NetworkSpec{
				Annotations: api.Annotations{
					Name: "test1",
				},
			},
		}
		assert.NoError(t, store.CreateNetwork(tx, n1))

		s1 := &api.Service{
			ID: "testServiceID1",
			Spec: api.ServiceSpec{
				Annotations: api.Annotations{
					Name: "service1",
				},
				Endpoint: &api.Endpoint{},
			},
		}
		assert.NoError(t, store.CreateService(tx, s1))

		t1 := &api.Task{
			ID: "testTaskID1",
			Status: api.TaskStatus{
				State: api.TaskStateNew,
			},
			Runtime: &api.Task_Container{
				Container: &api.Container{
					Spec: api.ContainerSpec{
						Networks: []*api.ContainerSpec_NetworkAttachment{
							{
								Reference: &api.ContainerSpec_NetworkAttachment_NetworkID{
									NetworkID: "testID1",
								},
							},
						},
					},
				},
			},
		}
		assert.NoError(t, store.CreateTask(tx, t1))
		return nil
	}))

	netWatch, cancel := state.Watch(s.WatchQueue(), state.EventUpdateNetwork{}, state.EventDeleteNetwork{})
	defer cancel()
	taskWatch, cancel := state.Watch(s.WatchQueue(), state.EventUpdateTask{}, state.EventDeleteTask{})
	defer cancel()
	serviceWatch, cancel := state.Watch(s.WatchQueue(), state.EventUpdateService{}, state.EventDeleteService{})
	defer cancel()

	// Start allocator
	go func() {
		assert.NoError(t, a.Run(context.Background()))
	}()

	// Now verify if we get network and tasks updated properly
	watchNetwork(t, netWatch, false, isValidNetwork)
	watchTask(t, taskWatch, false, isValidTask)
	watchService(t, serviceWatch, false, nil)

	// Add new networks/tasks/services after allocator is started.
	assert.NoError(t, s.Update(func(tx store.Tx) error {
		n2 := &api.Network{
			ID: "testID2",
			Spec: api.NetworkSpec{
				Annotations: api.Annotations{
					Name: "test2",
				},
			},
		}
		assert.NoError(t, store.CreateNetwork(tx, n2))
		return nil
	}))

	watchNetwork(t, netWatch, false, isValidNetwork)

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		s2 := &api.Service{
			ID: "testServiceID2",
			Spec: api.ServiceSpec{
				Annotations: api.Annotations{
					Name: "service2",
				},
				Endpoint: &api.Endpoint{},
			},
		}
		assert.NoError(t, store.CreateService(tx, s2))
		return nil
	}))

	watchService(t, serviceWatch, false, nil)

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		t2 := &api.Task{
			ID: "testTaskID2",
			Status: api.TaskStatus{
				State: api.TaskStateNew,
			},
			ServiceID:    "testServiceID2",
			DesiredState: api.TaskStateRunning,
			Runtime: &api.Task_Container{
				Container: &api.Container{
					Spec: api.ContainerSpec{
						Networks: []*api.ContainerSpec_NetworkAttachment{
							{
								Reference: &api.ContainerSpec_NetworkAttachment_NetworkID{
									NetworkID: "testID2",
								},
							},
						},
					},
				},
			},
		}
		assert.NoError(t, store.CreateTask(tx, t2))
		return nil
	}))

	watchTask(t, taskWatch, false, isValidTask)

	// Now try adding a task which depends on a network before adding the network.
	assert.NoError(t, s.Update(func(tx store.Tx) error {
		t3 := &api.Task{
			ID: "testTaskID3",
			Status: api.TaskStatus{
				State: api.TaskStateNew,
			},
			DesiredState: api.TaskStateRunning,
			Runtime: &api.Task_Container{
				Container: &api.Container{
					Spec: api.ContainerSpec{
						Networks: []*api.ContainerSpec_NetworkAttachment{
							{
								Reference: &api.ContainerSpec_NetworkAttachment_NetworkID{
									NetworkID: "testID3",
								},
							},
						},
					},
				},
			},
		}
		assert.NoError(t, store.CreateTask(tx, t3))
		return nil
	}))

	// Wait for a little bit of time before adding network just to
	// test network is not available while task allocation is
	// going through
	time.Sleep(10 * time.Millisecond)

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		n3 := &api.Network{
			ID: "testID3",
			Spec: api.NetworkSpec{
				Annotations: api.Annotations{
					Name: "test3",
				},
			},
		}
		assert.NoError(t, store.CreateNetwork(tx, n3))
		return nil
	}))

	watchNetwork(t, netWatch, false, isValidNetwork)
	watchTask(t, taskWatch, false, isValidTask)

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		assert.NoError(t, store.DeleteTask(tx, "testTaskID3"))
		return nil
	}))
	watchTask(t, taskWatch, false, isValidTask)

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		t5 := &api.Task{
			ID: "testTaskID5",
			Status: api.TaskStatus{
				State: api.TaskStateNew,
			},
			DesiredState: api.TaskStateRunning,
			ServiceID:    "testServiceID2",
			Runtime: &api.Task_Container{
				Container: &api.Container{
					Spec: api.ContainerSpec{},
				},
			},
		}
		assert.NoError(t, store.CreateTask(tx, t5))
		return nil
	}))
	watchTask(t, taskWatch, false, isValidTask)

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		assert.NoError(t, store.DeleteNetwork(tx, "testID3"))
		return nil
	}))
	watchNetwork(t, netWatch, false, isValidNetwork)

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		assert.NoError(t, store.DeleteService(tx, "testServiceID2"))
		return nil
	}))
	watchService(t, serviceWatch, false, nil)

	// Try to create a task with no network attachments and test
	// that it moves to ALLOCATED state.
	assert.NoError(t, s.Update(func(tx store.Tx) error {
		t4 := &api.Task{
			ID: "testTaskID4",
			Status: api.TaskStatus{
				State: api.TaskStateNew,
			},
			DesiredState: api.TaskStateRunning,
			Runtime: &api.Task_Container{
				Container: &api.Container{
					Spec: api.ContainerSpec{},
				},
			},
		}
		assert.NoError(t, store.CreateTask(tx, t4))
		return nil
	}))
	watchTask(t, taskWatch, false, isValidTask)

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		n2 := store.GetNetwork(tx, "testID2")
		require.NotEqual(t, nil, n2)
		assert.NoError(t, store.UpdateNetwork(tx, n2))
		return nil
	}))
	watchNetwork(t, netWatch, false, isValidNetwork)
	watchNetwork(t, netWatch, true, nil)

	// Try updating task which is already allocated
	assert.NoError(t, s.Update(func(tx store.Tx) error {
		t2 := store.GetTask(tx, "testTaskID2")
		require.NotEqual(t, nil, t2)
		assert.NoError(t, store.UpdateTask(tx, t2))
		return nil
	}))
	watchTask(t, taskWatch, false, isValidTask)
	watchTask(t, taskWatch, true, nil)

	a.Stop()
}

func isValidNetwork(t assert.TestingT, n *api.Network) bool {
	return assert.NotEqual(t, n.IPAM.Configs, nil) &&
		assert.Equal(t, len(n.IPAM.Configs), 1) &&
		assert.Equal(t, n.IPAM.Configs[0].Range, "") &&
		assert.Equal(t, len(n.IPAM.Configs[0].Reserved), 0) &&
		isValidSubnet(t, n.IPAM.Configs[0].Subnet) &&
		assert.NotEqual(t, net.ParseIP(n.IPAM.Configs[0].Gateway), nil)
}

func isValidTask(t assert.TestingT, task *api.Task) bool {
	return isValidNetworkAttachment(t, task) &&
		isValidEndpoint(t, task) &&
		assert.Equal(t, task.Status.State, api.TaskStateAllocated)
}

func isValidNetworkAttachment(t assert.TestingT, task *api.Task) bool {
	if len(task.GetContainer().Spec.Networks) != 0 {
		return assert.Equal(t, len(task.GetContainer().Networks[0].Addresses), 1) &&
			isValidSubnet(t, task.GetContainer().Networks[0].Addresses[0])
	}

	return true
}

func isValidEndpoint(t assert.TestingT, task *api.Task) bool {
	if task.ServiceID != "" {
		var service *api.Service
		s.View(func(tx store.ReadTx) {
			service = store.GetService(tx, task.ServiceID)
		})

		if service == nil {
			return true
		}

		return assert.Equal(t, service.Endpoint, task.GetContainer().Endpoint)

	}

	return true
}

func isValidSubnet(t assert.TestingT, subnet string) bool {
	_, _, err := net.ParseCIDR(subnet)
	return assert.NoError(t, err)
}

type mockTester struct{}

func (m mockTester) Errorf(format string, args ...interface{}) {
}

func watchNetwork(t *testing.T, watch chan events.Event, expectTimeout bool, fn func(t assert.TestingT, n *api.Network) bool) {
	for {
		var network *api.Network
		select {
		case event := <-watch:
			if n, ok := event.(state.EventUpdateNetwork); ok {
				network = n.Network.Copy()
				if fn == nil || (fn != nil && fn(mockTester{}, network)) {
					return
				}
			}

			if n, ok := event.(state.EventDeleteNetwork); ok {
				network = n.Network.Copy()
				if fn == nil || (fn != nil && fn(mockTester{}, network)) {
					return
				}
			}

			//return nil, fmt.Errorf("got event %T when expecting EventUpdateNetwork/EventDeleteNetwork", event)
		case <-time.After(250 * time.Millisecond):
			if !expectTimeout {
				if network != nil && fn != nil {
					fn(t, network)
				}

				t.Fatal("timed out before watchNetwork found expected network state")
			}

			return
		}
	}
}

func watchService(t *testing.T, watch chan events.Event, expectTimeout bool, fn func(t assert.TestingT, n *api.Service) bool) {
	for {
		var service *api.Service
		select {
		case event := <-watch:
			if s, ok := event.(state.EventUpdateService); ok {
				service = s.Service.Copy()
				if fn == nil || (fn != nil && fn(mockTester{}, service)) {
					return
				}
			}

			if s, ok := event.(state.EventDeleteService); ok {
				service = s.Service.Copy()
				if fn == nil || (fn != nil && fn(mockTester{}, service)) {
					return
				}
			}

		case <-time.After(250 * time.Millisecond):
			if !expectTimeout {
				if service != nil && fn != nil {
					fn(t, service)
				}

				t.Fatal("timed out before watchService found expected service state")
			}

			return
		}
	}
}

func watchTask(t *testing.T, watch chan events.Event, expectTimeout bool, fn func(t assert.TestingT, n *api.Task) bool) {
	for {
		var task *api.Task
		select {
		case event := <-watch:
			if t, ok := event.(state.EventUpdateTask); ok {
				task = t.Task.Copy()
				if fn == nil || (fn != nil && fn(mockTester{}, task)) {
					return
				}
			}

			if t, ok := event.(state.EventDeleteTask); ok {
				task = t.Task.Copy()
				if fn == nil || (fn != nil && fn(mockTester{}, task)) {
					return
				}
			}

		case <-time.After(250 * time.Millisecond):
			if !expectTimeout {
				if task != nil && fn != nil {
					fn(t, task)
				}

				t.Fatal("timed out before watchTask found expected task state")
			}

			return
		}
	}
}

// TestAllocatorVolumes - test the allocator for volumes
func TestAllocatorVolumes(t *testing.T) {
	s = store.NewMemoryStore(nil)
	assert.NotNil(t, s)

	a, err := New(s)
	assert.NoError(t, err)
	assert.NotNil(t, a)

	// Add some objects to store before allocator is started
	// Try adding some objects to store before allocator is started
	assert.NoError(t, s.Update(func(tx store.Tx) error {
		v1 := &api.Volume{
			ID: "testVolID1",
			Spec: api.VolumeSpec{
				Annotations: api.Annotations{
					Name: "testVol1",
				},
			},
		}
		assert.NoError(t, store.CreateVolume(tx, v1))
		return nil
	}))

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		t1 := &api.Task{
			ID: "vTestTaskID1",
			Status: api.TaskStatus{
				State: api.TaskStateNew,
			},
			Runtime: &api.Task_Container{
				Container: &api.Container{
					Spec: api.ContainerSpec{
						Mounts: []*api.Mount{
							{
								Target:     "/foo",
								Type:       api.MountTypeVolume,
								VolumeName: "testVol1",
							},
						},
					},
				},
			},
		}
		assert.NoError(t, store.CreateTask(tx, t1))
		return nil
	}))

	taskWatch, cancel := state.Watch(s.WatchQueue(), state.EventUpdateTask{}, state.EventDeleteTask{})
	defer cancel()

	// Start allocator
	go func() {
		assert.NoError(t, a.Run(context.Background()))
	}()

	// Now verify that tasks are updated properly
	watchTask(t, taskWatch, false, isValidTaskWithVolume)

	// Add a new task after allocator is started.
	assert.NoError(t, s.Update(func(tx store.Tx) error {
		v1 := &api.Volume{
			ID: "testVolID2",
			Spec: api.VolumeSpec{
				Annotations: api.Annotations{
					Name: "testVol2",
				},
			},
		}
		assert.NoError(t, store.CreateVolume(tx, v1))
		return nil
	}))

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		t1 := &api.Task{
			ID: "vTestTaskID2",
			Status: api.TaskStatus{
				State: api.TaskStateNew,
			},
			DesiredState: api.TaskStateRunning,
			Runtime: &api.Task_Container{
				Container: &api.Container{
					Spec: api.ContainerSpec{
						Mounts: []*api.Mount{
							{
								Target:     "/foo",
								Type:       api.MountTypeVolume,
								VolumeName: "testVol2",
							},
						},
					},
				},
			},
		}
		assert.NoError(t, store.CreateTask(tx, t1))
		return nil
	}))

	watchTask(t, taskWatch, false, isValidTaskWithVolume)

	// Add a task which depends on a volume before adding the volume.
	assert.NoError(t, s.Update(func(tx store.Tx) error {
		t3 := &api.Task{
			ID: "vTestTaskID3",
			Status: api.TaskStatus{
				State: api.TaskStateNew,
			},
			DesiredState: api.TaskStateRunning,
			Runtime: &api.Task_Container{
				Container: &api.Container{
					Spec: api.ContainerSpec{
						Mounts: []*api.Mount{
							{
								Target:     "/foo",
								Type:       api.MountTypeVolume,
								VolumeName: "testVol3",
							},
						},
					},
				},
			},
		}
		assert.NoError(t, store.CreateTask(tx, t3))
		return nil
	}))

	// Wait for a little bit of time before adding volume just to
	// test the case where the volume is not available during task allocation
	time.Sleep(10 * time.Millisecond)

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		v3 := &api.Volume{
			ID: "testVolID3",
			Spec: api.VolumeSpec{
				Annotations: api.Annotations{
					Name: "testVol3",
				},
			},
		}
		assert.NoError(t, store.CreateVolume(tx, v3))
		return nil
	}))
	watchTask(t, taskWatch, false, isValidTaskWithVolume)

	assert.NoError(t, s.Update(func(tx store.Tx) error {
		assert.NoError(t, store.DeleteTask(tx, "vTestTaskID3"))
		return nil
	}))
	watchTask(t, taskWatch, false, isValidTaskWithVolume)

	// Create a task with no volumes. Ensure that it moves to
	// ALLOCATED state.
	assert.NoError(t, s.Update(func(tx store.Tx) error {
		t1 := &api.Task{
			ID: "vTestTaskID4",
			Status: api.TaskStatus{
				State: api.TaskStateNew,
			},
			DesiredState: api.TaskStateRunning,
			Runtime: &api.Task_Container{
				Container: &api.Container{
					Spec: api.ContainerSpec{
						Mounts: []*api.Mount{
							{
								Target: "/foo",
								Source: "/bar",
								Type:   api.MountTypeBind,
							},
						},
					},
				},
			},
		}
		assert.NoError(t, store.CreateTask(tx, t1))
		return nil
	}))
	watchTask(t, taskWatch, false, isValidTaskWithVolume)

	// Try updating task which is already allocated
	assert.NoError(t, s.Update(func(tx store.Tx) error {
		t2 := store.GetTask(tx, "vTestTaskID2")
		require.NotEqual(t, nil, t2)
		assert.NoError(t, store.UpdateTask(tx, t2))
		return nil
	}))
	watchTask(t, taskWatch, false, isValidTaskWithVolume)
	watchTask(t, taskWatch, true, nil)

	a.Stop()
}

func isValidTaskWithVolume(t assert.TestingT, task *api.Task) bool {
	if !assert.Equal(t, task.Status.State, api.TaskStateAllocated) {
		return false
	}

	// If there are no BindType == Volume in the task, we are ok
	hasVolumeMounts := false
	for _, m := range task.GetContainer().Spec.Mounts {
		if m.Type == api.MountTypeVolume {
			hasVolumeMounts = true
			break
		}
	}
	if !hasVolumeMounts {
		return true
	}

	if len(task.GetContainer().Volumes) != 0 &&
		len(task.GetContainer().Volumes[0].Spec.Annotations.Name) != 0 {
		return true
	}
	return false
}
