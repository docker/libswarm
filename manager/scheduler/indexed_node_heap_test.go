package scheduler

import (
	"container/heap"
	"math"
	"math/rand"
	"strconv"
	"testing"

	objectspb "github.com/docker/swarm-v2/pb/docker/cluster/objects"
	specspb "github.com/docker/swarm-v2/pb/docker/cluster/specs"
	"github.com/stretchr/testify/assert"
)

func TestFindMin(t *testing.T) {
	var nh nodeHeap

	for reps := 0; reps < 10; reps++ {
		// Create a bunch of nodes with random numbers of tasks
		numNodes := 10000
		nh.alloc(numNodes)

		for i := 0; i != numNodes; i++ {
			n := &objectspb.Node{
				ID: "id" + strconv.Itoa(i),
				Spec: &specspb.NodeSpec{
					Meta: specspb.Meta{
						Labels: make(map[string]string),
					},
				},
			}

			// Give every hundredth node a special label
			if i%100 == 0 {
				n.Spec.Meta.Labels["special"] = "true"
			}
			nh.heap = append(nh.heap, nodeHeapItem{node: n, numTasks: int(rand.Int())})
			nh.index[n.ID] = i
		}

		heap.Init(&nh)

		isSpecial := func(n *objectspb.Node) bool {
			return n.Spec.Meta.Labels["special"] == "true"
		}

		bestNode, numTasks := nh.findMin(isSpecial, false)
		assert.NotNil(t, bestNode)

		// Verify with manual search
		var manualBestNode *objectspb.Node
		manualBestTasks := uint64(math.MaxUint64)
		for i := 0; i < nh.Len(); i++ {
			if !isSpecial(nh.heap[i].node) {
				continue
			}
			if uint64(nh.heap[i].numTasks) < manualBestTasks {
				manualBestNode = nh.heap[i].node
				manualBestTasks = uint64(nh.heap[i].numTasks)
			} else if uint64(nh.heap[i].numTasks) == manualBestTasks && nh.heap[i].node == bestNode {
				// prefer the node that findMin chose when
				// there are multiple best choices
				manualBestNode = nh.heap[i].node
			}
		}

		assert.Equal(t, bestNode, manualBestNode)
		assert.Equal(t, numTasks, int(manualBestTasks))
	}
}
