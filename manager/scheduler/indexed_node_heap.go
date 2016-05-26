package scheduler

import (
	"container/heap"
	"github.com/docker/swarm-v2/api"
)

// A nodeHeap implements heap.Interface for nodes. It also includes an index
// by node id.
type nodeHeap struct {
	heap  []NodeInfo
	index map[string]int // map from node id to heap index
}

func (nh nodeHeap) Len() int {
	return len(nh.heap)
}

func (nh nodeHeap) Less(i, j int) bool {
	return len(nh.heap[i].Tasks) < len(nh.heap[j].Tasks)
}

func (nh nodeHeap) Swap(i, j int) {
	nh.heap[i], nh.heap[j] = nh.heap[j], nh.heap[i]
	nh.index[nh.heap[i].ID] = i
	nh.index[nh.heap[j].ID] = j
}

func (nh *nodeHeap) Push(x interface{}) {
	n := len(nh.heap)
	item := x.(NodeInfo)
	nh.index[item.ID] = n
	nh.heap = append(nh.heap, item)
}

func (nh *nodeHeap) Pop() interface{} {
	old := nh.heap
	n := len(old)
	item := old[n-1]
	delete(nh.index, item.ID)
	nh.heap = old[0 : n-1]
	return item
}

func (nh *nodeHeap) alloc(n int) {
	nh.heap = make([]NodeInfo, 0, n)
	nh.index = make(map[string]int, n)
}

func (nh *nodeHeap) peek() *NodeInfo {
	if len(nh.heap) == 0 {
		return nil
	}
	return &nh.heap[0]
}

// nodeInfo returns the NodeInfo struct for a given node identified by its ID.
func (nh *nodeHeap) nodeInfo(nodeID string) NodeInfo {
	index, ok := nh.index[nodeID]
	if ok {
		return nh.heap[index]
	}
	return NodeInfo{}
}

// addOrUpdateNode sets the number of tasks for a given node. It adds the node
// to the heap if it wasn't already tracked.
func (nh *nodeHeap) addOrUpdateNode(n NodeInfo) {
	if n.Node == nil {
		return
	}
	index, ok := nh.index[n.ID]
	if ok {
		nh.heap[index] = n
		heap.Fix(nh, index)
	} else {
		heap.Push(nh, n)
	}
}

// updateNode sets the number of tasks for a given node. It ignores the update
// if the node isn't already tracked in the heap.
func (nh *nodeHeap) updateNode(n NodeInfo) {
	if n.Node == nil {
		return
	}
	index, ok := nh.index[n.ID]
	if ok {
		nh.heap[index] = n
		heap.Fix(nh, index)
	}
}

func (nh *nodeHeap) remove(nodeID string) {
	index, ok := nh.index[nodeID]
	if ok {
		nh.heap[index].Tasks = nil
		heap.Fix(nh, index)
		heap.Pop(nh)
	}
}

func (nh *nodeHeap) findMin(meetsConstraints func(*NodeInfo) bool, scanAllNodes bool) (*api.Node, int) {
	if scanAllNodes {
		return nh.scanAllToFindMin(meetsConstraints)
	}
	return nh.searchHeapToFindMin(meetsConstraints, 0)
}

// Scan All nodes to find the best node which meets the constraints && has lightest workloads
func (nh *nodeHeap) scanAllToFindMin(meetsConstraints func(*NodeInfo) bool) (*api.Node, int) {
	var bestNode *api.Node
	minTasks := int(^uint(0) >> 1) // max int

	for i := 0; i < len(nh.heap); i++ {
		heapEntry := &nh.heap[i]
		if meetsConstraints(heapEntry) && len(heapEntry.Tasks) < minTasks {
			bestNode = heapEntry.Node
			minTasks = len(heapEntry.Tasks)
		}
	}

	return bestNode, minTasks
}

// Search in heap to find the best node which meets the constraints && has lightest workloads
func (nh *nodeHeap) searchHeapToFindMin(meetsConstraints func(*NodeInfo) bool, index int) (*api.Node, int) {
	if index >= len(nh.heap) {
		return nil, int(^uint(0) >> 1) // max int
	}

	heapEntry := &nh.heap[index]
	if meetsConstraints(heapEntry) {
		// Meet the constraints, return directly
		return heapEntry.Node, len(heapEntry.Tasks)
	}

	// Search in 2 children
	leftBestNode, leftMinTasks := nh.searchHeapToFindMin(meetsConstraints, index*2+1)
	rightBestNode, rightMinTask := nh.searchHeapToFindMin(meetsConstraints, index*2+2)

	// Return the best one
	if leftMinTasks < rightMinTask {
		return leftBestNode, leftMinTasks
	}

	return rightBestNode, rightMinTask
}
