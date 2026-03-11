/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package chip1softsharedev is using for HuaWei chip1softsharedev schedule.
package chip1softsharedev

import (
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910b"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
)

const (
	chipMemory32G    = 32
	task1AicoreQuota = 20
	task1HbmQuota    = 10240
	task2AicoreQuota = 30
	task2HbmQuota    = 20480
)

func TestGetChipMemoryFromNodeLabel(t *testing.T) {
	tp := &chip1softsharedev{}
	tests := []struct {
		name      string
		nodeLabel map[string]string
		want      int
		wantErr   bool
	}{
		{
			name:      "missing memory label",
			nodeLabel: map[string]string{},
			want:      0,
			wantErr:   true,
		},
		{
			name: "valid memory label (32G)",
			nodeLabel: map[string]string{
				util.NPUChipMemoryLabelKey: "32G",
			},
			want:    chipMemory32G * util.MBPerGB,
			wantErr: false,
		},
		{
			name: "invalid memory value",
			nodeLabel: map[string]string{
				util.NPUChipMemoryLabelKey: "invalid",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tp.getChipMemoryFromNodeLabel(tt.nodeLabel)
			if (err != nil) != tt.wantErr {
				t.Errorf("getChipMemoryFromNodeLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getChipMemoryFromNodeLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUsedResourceMapFromNodeTasks(t *testing.T) {
	tp := &chip1softsharedev{Base910b: ascend910b.Base910b{
		NPUHandler: base.NPUHandler{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}}}}}
	tp.ReqNPUName = util.NPU910CardName
	tp.SetAnnoPreVal(util.NPU910CardNamePre)
	task1 := &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{
		Annotations: map[string]string{
			util.AscendNPUPodRealUse:                 "Ascend910-0",
			util.SchedulerSoftShareDevAicoreQuotaKey: "20",
			util.SchedulerSoftShareDevHbmQuotaKey:    "10240",
			util.SchedulerSoftShareDevPolicyKey:      util.SoftShareDevPolicyFixedShare,
		},
	}}}
	task2 := &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{
		Annotations: map[string]string{
			util.AscendNPUPodRealUse:                 "Ascend910-0",
			util.SchedulerSoftShareDevAicoreQuotaKey: "30",
			util.SchedulerSoftShareDevHbmQuotaKey:    "20480",
			util.SchedulerSoftShareDevPolicyKey:      util.SoftShareDevPolicyFixedShare,
		},
	}}}
	task3 := &api.TaskInfo{Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}}}
	tasks := map[api.TaskID]*api.TaskInfo{
		"task1": task1,
		"task2": task2,
		"task3": task3,
	}
	got := tp.getUsedResourceMapFromNodeTasks(tasks)
	want := map[int]softShareDevResource{
		0: {
			aicoreQuota:      task1AicoreQuota + task2AicoreQuota,
			hbmQuota:         task1HbmQuota + task2HbmQuota,
			schedulingPolicy: util.SoftShareDevPolicyFixedShare,
		},
	}
	if len(got) != len(want) {
		t.Fatalf("getUsedResourceMapFromNodeTasks() len = %v, want len %v", len(got), len(want))
	}
	for k, v := range got {
		wantV, ok := want[k]
		if !ok {
			t.Errorf("getUsedResourceMapFromNodeTasks() key %v not in want", k)
			continue
		}
		if v.aicoreQuota != wantV.aicoreQuota || v.hbmQuota != wantV.hbmQuota || v.schedulingPolicy != wantV.schedulingPolicy {
			t.Errorf("getUsedResourceMapFromNodeTasks()[%v] = %+v, want %+v", k, v, wantV)
		}
	}
}
