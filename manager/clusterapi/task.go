package clusterapi

import (
	"github.com/docker/swarm-v2/api"
	"github.com/docker/swarm-v2/manager/state"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// GetTask returns a Task given a TaskID.
// - Returns `InvalidArgument` if TaskID is not provided.
// - Returns `NotFound` if the Task is not found.
func (s *Server) GetTask(ctx context.Context, request *api.GetTaskRequest) (*api.GetTaskResponse, error) {
	if request.TaskID == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, errInvalidArgument.Error())
	}

	var task *api.Task
	s.store.View(func(tx state.ReadTx) {
		task = tx.Tasks().Get(request.TaskID)
	})
	if task == nil {
		return nil, grpc.Errorf(codes.NotFound, "task %s not found", request.TaskID)
	}
	return &api.GetTaskResponse{
		Task: task,
	}, nil
}

// RemoveTask removes a Task referenced by TaskID.
// - Returns `InvalidArgument` if TaskID is not provided.
// - Returns `NotFound` if the Task is not found.
// - Returns an error if the deletion fails.
func (s *Server) RemoveTask(ctx context.Context, request *api.RemoveTaskRequest) (*api.RemoveTaskResponse, error) {
	if request.TaskID == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, errInvalidArgument.Error())
	}

	err := s.store.Update(func(tx state.Tx) error {
		return tx.Tasks().Delete(request.TaskID)
	})
	if err != nil {
		if err == state.ErrNotExist {
			return nil, grpc.Errorf(codes.NotFound, "task %s not found", request.TaskID)
		}
		return nil, err
	}
	return &api.RemoveTaskResponse{}, nil
}

// ListTasks returns a list of all tasks.
func (s *Server) ListTasks(ctx context.Context, request *api.ListTasksRequest) (*api.ListTasksResponse, error) {
	var (
		tasks []*api.Task
		err   error
	)
	s.store.View(func(tx state.ReadTx) {
		tasks, err = tx.Tasks().Find(state.All)
	})
	if err != nil {
		return nil, err
	}
	return &api.ListTasksResponse{
		Tasks: tasks,
	}, nil
}
