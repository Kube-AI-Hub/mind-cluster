/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package chip8node8ra64sp for node test
package chip8node8ra64sp

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	itest "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/test"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

// for test cases use
const (
	rackOsNum      = 8
	npuNum8        = 8
	nodeInfoIdx4   = 4
	nodeInfoIdx9   = 9
	nodeInfoIdx185 = 185
	SuperPodSize   = 64 // Define SuperPodSize for test cases
)

var (
	npuList2        = []int{0, 1}
	npuList8        = []int{0, 1, 2, 3, 4, 5, 6, 7}
	testRackNPUTop  = rackNpuTopType{}
	testRackNPUTop1 = rackNpuTopType{
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
	}
	testRackNPUTop2 = rackNpuTopType{
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true},
	}
	testAnnoName         = "test"
	invalidTestAnnoValue = "npu-0,npu-1,npu-2,npu-3,npu-4,npu-5,npu-6,npu-7,npu-8"
)

// checkNodeNPUByTaskTestCase CheckNodeNPUByTask test case
type checkNodeNPUByTaskTestCase struct {
	Task          *api.TaskInfo
	Name          string
	Attr          util.SchedulerJobAttr
	Node          plugin.NPUNode
	WantErr       error
	TpBlockNPUNum int
	MaxNodeNPUNum int
	TaskNodeNPU   string
}

func buildCardAnnotationStr(npuList []int) string {
	var str string
	for index, npu := range npuList {
		if index == 0 {
			str = fmt.Sprintf("%s%d", util.NPUCardNamePre, npu)
		} else {
			str = fmt.Sprintf("%s,%s%d", str, util.NPUCardNamePre, npu)
		}
	}
	return str
}

// TestCheckNodeNPUByTask
func TestCheckNodeNPUByTask(t *testing.T) {
	npu := New(SuperPodx8SchedulerName)
	testCases := buildcheckNodeNPUByTaskTestCases01()
	testCases = append(testCases, buildcheckNodeNPUByTaskTestCases02()...)
	testCases = append(testCases, buildcheckNodeNPUByTaskTestCases03()...)
	testCases = append(testCases, buildcheckNodeNPUByTaskTestCases04()...)
	testCases = append(testCases, buildcheckNodeNPUByTaskTestCases05()...)
	testCases = append(testCases, buildcheckNodeNPUByTaskTestCases06()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			job := test.FakeNormalTestJob("job", 1)
			test.SetFakeJobResRequest(job, util.NPU910CardName, tt.TaskNodeNPU)
			attr := itest.FakeSchedulerJobAttrByJob(job)
			sJob := plugin.SchedulerJob{}
			sJob.SchedulerJobAttr = attr
			env := plugin.ScheduleEnv{
				ClusterCache: plugin.ClusterCache{
					Jobs: map[api.JobID]plugin.SchedulerJob{job.UID: sJob},
				},
			}
			npu.SetSchedulerAttr(attr)
			npu.SetSchedulerEnv(env)
			npu.SetMaxNodeNPUNum(tt.MaxNodeNPUNum)
			npu.tpBlock = tt.TpBlockNPUNum

			if err := npu.CheckNodeNPUByTask(tt.Task, tt.Node); !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("CheckNodeNPUByTask() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

func buildcheckNodeNPUByTaskTestCases01() []checkNodeNPUByTaskTestCase {
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "01-CheckNodeNPUByTask return nil when node npu meet task req",
			Task:          test.FakeTaskWithResReq("pod0", util.NPUCardName, util.NPUIndex8),
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName:    buildCardAnnotationStr(npuList8),
						networkUnhealthyNPU: "",
					},
				},
			},
			WantErr: nil,
		},
		{
			Name:          "02-CheckNodeNPUByTask return err when task is not npu task",
			Task:          test.FakeTaskWithResReq("pod1", util.NPUCardName, util.NPUIndex8),
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName:    buildCardAnnotationStr(npuList8),
						networkUnhealthyNPU: "",
					},
				},
			},
			WantErr: errors.New("task<pod1> is not npu task"),
		},
	}
}

func buildcheckNodeNPUByTaskTestCases02() []checkNodeNPUByTaskTestCase {
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "04-CheckNodeNPUByTask return err when node has no req npu",
			Task:          test.FakeTaskWithResReq("pod0", util.NPUCardName, util.NPUIndex8),
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName:    buildCardAnnotationStr(npuList2),
						networkUnhealthyNPU: "",
					},
				},
			},
			WantErr: errors.New(
				"checkNodeNPUByTask the npus on this node don't satisfy the schedulable topology err: " +
					"node don't have enough npu resource, req<8>, idle<2>"),
		},
		{
			Name:          "05-CheckNodeNPUByTask return err when node has no req npu",
			Task:          test.FakeTaskWithResReq("pod0", util.NPUCardName, util.NPUIndex8),
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName: "npu-0,npu-1,npu-4",
					},
				},
			},
			WantErr: errors.New(
				"checkNodeNPUByTask the npus on this node don't satisfy the schedulable topology err: " +
					"node don't have enough npu resource, req<8>, idle<3>"), // modified expected error
		},
	}
}

func buildcheckNodeNPUByTaskTestCases03() []checkNodeNPUByTaskTestCase {
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "07-CheckNodeNPUByTask return err when task is nil",
			Task:          nil,
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0,Ascend310P-1"},
				},
			},
			WantErr: errors.New(util.ArgumentError),
		},
		{
			Name:          "08-CheckNodeNPUByTask return err when node annotation is nil",
			Task:          test.FakeTaskWithResReq("pod1", util.NPU310PCardName, util.NPUIndex2),
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "node1",
					Annotation: nil,
				},
			},
			WantErr: errors.New(util.ArgumentError),
		},
	}
}

func buildcheckNodeNPUByTaskTestCases04() []checkNodeNPUByTaskTestCase {
	task := test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex8)
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "9-CheckNodeNPUByTask return nil when tp-block is valid",
			Task:          task,
			TpBlockNPUNum: 8,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName:    buildCardAnnotationStr(npuList8),
						networkUnhealthyNPU: "",
					},
				},
			},
			WantErr: nil,
		},
	}
}

func buildcheckNodeNPUByTaskTestCases05() []checkNodeNPUByTaskTestCase {
	task := test.FakeTaskWithResReq("pod0", util.NPU910CardName, util.NPUIndex8)
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "10-CheckNodeNPUByTask return err when node has no req npu",
			Task:          task,
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName:    "npu-0,npu-1,npu-2,npu-3,npu-4,npu-5,npu-6,npu-7",
						networkUnhealthyNPU: "npu-5",
					},
				},
			},
			WantErr: nil, // modified expected error, no longer filter network unhealthy cards under non-multi-SuperPod strategy
		},
	}
}

func buildcheckNodeNPUByTaskTestCases06() []checkNodeNPUByTaskTestCase {
	task := test.FakeTaskWithResReq("pod0", util.NPU910CardName, npuNum8)
	return []checkNodeNPUByTaskTestCase{
		{
			Name:          "11-CheckNodeNPUByTask return err when node RackID invalid",
			Task:          task,
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName: "npu-0,npu-1,npu-2,npu-3,npu-4,npu-5,npu-6,npu-7",
					},
					RackID: -1,
				},
			},
			WantErr: errors.New("node rack-id is invalid for node=node1, id=-1"),
		},
		{
			Name:          "16-CheckNodeNPUByTask return err when node SuperPodID invalid",
			Task:          task,
			TpBlockNPUNum: 1,
			MaxNodeNPUNum: npuNum8,
			TaskNodeNPU:   "8",
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName: "npu-0,npu-1,npu-2,npu-3,npu-4,npu-5,npu-6,npu-7",
					},
					SuperPodID: -1,
				},
			},
			WantErr: errors.New("the super-pod-id of node is invalid for node=node1, id=-1"),
		},
	}
}

func buildNodeBaseInfoArr(count int) []nodeBaseInfo {
	var nodes []nodeBaseInfo
	for i := 0; i < count; i++ {
		nodes = append(nodes, nodeBaseInfo{
			name:       fmt.Sprintf("node%d", i),
			superPodID: 1,
			rackID:     1,
		})
	}
	return nodes
}

type getUsableTopFromNode struct {
	Name    string
	Node    plugin.NPUNode
	WantErr error
}

func buildGetUsableTopFromNodeTest1() []getUsableTopFromNode {
	return []getUsableTopFromNode{
		{
			Name:    "Case1-annotation is empty",
			WantErr: errors.New(util.ArgumentError),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Annotation: map[string]string{},
				},
			},
		},
		{
			Name:    "Case2-annotation without the key of huawei.com/npu",
			WantErr: fmt.Errorf("getUsableTopFromNode node1 don't have %s", util.NPUCardName),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						testAnnoName: "npu-0,npu-1",
					},
				},
			},
		},
		{
			Name:    "Case3-node npu num more than 8",
			WantErr: fmt.Errorf("node npu num is invalid, and the npus index: [0 1 2 3 4 5 6 7 8]"),
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName: invalidTestAnnoValue,
					},
				},
			},
		},
	}
}

func buildGetUsableTopFromNodeTest2() []getUsableTopFromNode {
	return []getUsableTopFromNode{
		{
			Name:    "Case4-node networkUnHealthy key is empty",
			WantErr: nil, // No longer expect error since we don't check networkUnHealthy key in getUsableTopFromNode
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName: "npu-0,npu-1",
					},
				},
			},
		},
		{
			Name:    "Case5-node networkUnHealthy npu num is more than 8",
			WantErr: nil, // No longer expect error since we don't check networkUnHealthy key in getUsableTopFromNode
			Node: plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name: "node1",
					Annotation: map[string]string{
						util.NPUCardName:    "npu-0,npu-1",
						networkUnhealthyNPU: invalidTestAnnoValue,
					},
				},
			},
		},
	}
}

// createTestTask creates a test task for TestCheckNodeNPUNums
func createTestTask() *api.TaskInfo {
	return &api.TaskInfo{Job: "vcjob/job", Name: "test-task", UID: "test-task-0"}
}

// createNPUJob creates a test NPUJob for TestCheckNodeNPUNums
func createNPUJob(task *api.TaskInfo) *util.NPUJob {
	npuJob := &util.NPUJob{
		Tasks:      make(map[api.TaskID]util.NPUTask),
		ReqNPUName: util.NPUCardName,
	}
	npuJob.Tasks[task.UID] = util.NPUTask{
		ReqNPUNum:  8,
		ReqNPUName: util.NPUCardName,
	}
	return npuJob
}

// createStrategy creates a test strategy for TestCheckNodeNPUNums
func createStrategy(task *api.TaskInfo, npuJob *util.NPUJob) *chip8node8ra64sp {
	jobs := make(map[api.JobID]plugin.SchedulerJob)
	jobs[task.Job] = plugin.SchedulerJob{
		SchedulerJobAttr: util.SchedulerJobAttr{
			NPUJob: npuJob,
		},
	}

	strategy := &chip8node8ra64sp{
		NPUHandler: base.NPUHandler{
			ScheduleEnv: plugin.ScheduleEnv{
				ClusterCache: plugin.ClusterCache{
					Jobs: jobs,
				},
			},
			SchedulerJobAttr: util.SchedulerJobAttr{
				NPUJob: npuJob,
			},
		},
		jobParams: jobParams{
			netUnhealthyKey: networkUnhealthyNPU,
		},
	}

	strategy.SetMaxNodeNPUNum(npuNumber8)
	strategy.SetMaxCardNPUNum(npuNumber8)
	strategy.tpBlock = 1
	strategy.spBlock = 1

	return strategy
}

// TestCheckNodeNPUNums tests the checkNodeNPUNums function
func TestCheckNodeNPUNums(t *testing.T) {
	task := createTestTask()
	npuJob := createNPUJob(task)
	strategy := createStrategy(task, npuJob)

	cases := []struct {
		name           string
		nodeAnnotation map[string]string
		wantErr        bool
		errorContains  string
	}{
		{
			"01 multi super pod and network healthy",
			map[string]string{networkUnhealthyNPU: "", util.NPUCardName: "npu-0,npu-1,npu-2,npu-3,npu-4,npu-5,npu-6,npu-7"},
			false, "",
		},
		{
			"02 multi super pod and no network annotation",
			map[string]string{util.NPUCardName: "npu-0,npu-1,npu-2,npu-3,npu-4,npu-5,npu-6,npu-7"},
			true, "don't have resource",
		},
		{
			"03 multi super pod and network unhealthy npu",
			map[string]string{
				networkUnhealthyNPU: "npu-0,npu-1",
				util.NPUCardName:    "npu-0,npu-1,npu-2,npu-3,npu-4,npu-5,npu-6,npu-7",
			},
			true, "don't have enough npu resource",
		},
		{
			"04 not multi super pod and network unhealthy npu",
			map[string]string{
				networkUnhealthyNPU: "npu-0,npu-1",
				util.NPUCardName:    "npu-0,npu-1,npu-2,npu-3,npu-4,npu-5,npu-6,npu-7",
			},
			false, "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "04 not multi super pod and network unhealthy npu" {
				strategy.NPUTaskNum = 1
			} else {
				strategy.NPUTaskNum = SuperPodSize + 1
			}

			node := plugin.NPUNode{
				CommonNode: plugin.CommonNode{
					Name:       "test-node",
					Annotation: tt.nodeAnnotation,
				},
			}

			err := strategy.checkNodeNPUNums(task, node)
			checkTestResult(err, tt.wantErr, tt.errorContains, t)
		})
	}
}

// checkTestResult checks if the test result matches the expected result
func checkTestResult(err error, wantErr bool, errorContains string, t *testing.T) {
	if wantErr {
		if err == nil {
			t.Fatalf("Expected error but got nil")
		}
		if !strings.Contains(err.Error(), errorContains) {
			t.Fatalf("Expected error to contain '%s', but got '%s'", errorContains, err.Error())
		}
	} else {
		if err != nil {
			t.Fatalf("Expected no error but got '%s'", err.Error())
		}
	}
}

// TestGetUsableTopFromNode tests the getUsableTopFromNode function
func TestGetUsableTopFromNode(t *testing.T) {
	npu := New(SuperPodx8SchedulerName)
	// Set up NPUJob with ReqNPUName to avoid nil pointer
	npu.SetSchedulerAttr(util.SchedulerJobAttr{
		NPUJob: &util.NPUJob{
			ReqNPUName: util.NPUCardName,
		},
	})
	testCases := buildGetUsableTopFromNodeTest1()
	testCases = append(testCases, buildGetUsableTopFromNodeTest2()...)
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := npu.getUsableTopFromNode(tt.Node)
			if tt.WantErr != nil {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if err.Error() != tt.WantErr.Error() {
					t.Errorf("getUsableTopFromNode() error = %v, wantErr %v", err, tt.WantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}
			}
		})
	}
}
