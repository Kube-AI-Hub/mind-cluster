/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package v1 is using for reconcile AscendJob.
package v1

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"

	"ascend-operator/pkg/api/v1"
)

func newVcjobWithRankTable(uid string) *v1alpha1.Job {
	return &v1alpha1.Job{
		ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid)},
		Spec: v1alpha1.JobSpec{
			Tasks: []v1alpha1.TaskSpec{
				{Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{Volumes: []corev1.Volume{{Name: rankTableName}}},
				}},
			},
		},
	}
}

func newDeployWithRankTable(uid string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid)},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Volumes: []corev1.Volume{{Name: rankTableName}}},
			},
		},
	}
}

func newStsWithRankTable(uid string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid)},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Volumes: []corev1.Volume{{Name: rankTableName}}},
			},
		},
	}
}

// TestOnOwnerCreateFuncForA5 test case for TestOnOwnerCreateFunc in A5
func TestOnOwnerCreateFuncForA5(t *testing.T) {
	convey.Convey("TestOnOwnerCreateFunc for A5", t, func() {
		r := newCommonReconciler()
		fn := r.onOwnerCreateFunc()
		convey.Convey("02-ascend job with scaleout-type=roce labels should return false", func() {
			job := newCommonAscendJob()
			job.Labels = map[string]string{v1.ScaleOutTypeLabel: "roce"}
			res := fn(event.CreateEvent{Object: job})
			convey.So(res, convey.ShouldEqual, true)
		})
		convey.Convey("03-ascend job with scaleout-type=uboe  labels should return false", func() {
			job := newCommonAscendJob()
			job.Labels = map[string]string{v1.ScaleOutTypeLabel: "uboe"}
			res := fn(event.CreateEvent{Object: job})
			convey.So(res, convey.ShouldEqual, true)
		})
		convey.Convey("04-vcjob with ranktable mount should return true", func() {
			vcjob := newVcjobWithRankTable("vcjob-uid-rt")
			res := fn(event.CreateEvent{Object: vcjob})
			convey.So(res, convey.ShouldEqual, true)
			convey.So(r.rtGenerators[vcjob.UID], convey.ShouldNotBeNil)
		})
		convey.Convey("05-deployment with ranktable mount should return true", func() {
			deploy := newDeployWithRankTable("deploy-uid-rt")
			res := fn(event.CreateEvent{Object: deploy})
			convey.So(res, convey.ShouldEqual, true)
			convey.So(r.rtGenerators[deploy.UID], convey.ShouldNotBeNil)
		})
		convey.Convey("06-statefulset with ranktable mount should return true", func() {
			sts := newStsWithRankTable("sts-uid-rt")
			res := fn(event.CreateEvent{Object: sts})
			convey.So(res, convey.ShouldEqual, true)
			convey.So(r.rtGenerators[sts.UID], convey.ShouldNotBeNil)
		})
	})
}
