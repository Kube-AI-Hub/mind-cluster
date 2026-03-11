/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package utils is common utils
package utils

import (
	"testing"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"ascend-common/api"
	v1 "ascend-operator/pkg/api/v1"
)

func TestGetLogicSuperPodNodes(t *testing.T) {
	convey.Convey("TestGetLogicSuperPodNodes", t, func() {
		convey.Convey("01-when spBlock is smaller than chipsPerNode, should return 1", func() {
			spBlock, chipsPerNode := 1, 2
			res := getLogicSuperPodNodes(spBlock, chipsPerNode)
			convey.So(res, convey.ShouldEqual, 1)
		})
		convey.Convey("02-when spBlock is bigger than chipsPerNode, should return quotient", func() {
			expected := 2
			spBlock, chipsPerNode := 4, 2
			res := getLogicSuperPodNodes(spBlock, chipsPerNode)
			convey.So(res, convey.ShouldEqual, expected)
		})
	})
}

func TestGetLogicSuperPodId(t *testing.T) {
	convey.Convey("TestGetLogicSuperPodId", t, func() {
		convey.Convey("01-when spBlock is 0, should return 0", func() {
			res := GetLogicSuperPodId(1, 0, 0)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("02-when pod rank is 1, spBlock is 1 and chipsPerNode is 1, should return 1", func() {
			res := GetLogicSuperPodId(1, 1, 1)
			convey.So(res, convey.ShouldEqual, 1)
		})
	})
}

func TestGetSpBlock(t *testing.T) {
	convey.Convey("TestGetSpBlock", t, func() {
		convey.Convey("01-job is nil will return 0", func() {
			res := GetSpBlock(nil)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("02-job without sp-block annotation should return 0", func() {
			res := GetSpBlock(newCommonAscendJob())
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("03-job with invalid sp-block annotation should return 0", func() {
			job := newCommonAscendJob()
			job.Annotations[AnnoKeyOfSuperPod] = "xx"
			res := GetSpBlock(newCommonAscendJob())
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("04-job with valid sp-block annotation should return 1", func() {
			job := newCommonAscendJob()
			job.Annotations[AnnoKeyOfSuperPod] = "1"
			res := GetSpBlock(job)
			convey.So(res, convey.ShouldEqual, 1)
		})
	})
}

func TestGetSpBlockNum(t *testing.T) {
	convey.Convey("TestGetSpBlockNum", t, func() {
		convey.Convey("01-job is nil will return 0", func() {
			res := GetSpBlockNum(nil)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("02-job without annotations will return 0", func() {
			job := &v1.AscendJob{}
			res := GetSpBlockNum(job)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("03-job with sp-block 0 will return 0", func() {
			job := newCommonAscendJob()
			job.Annotations[AnnoKeyOfSuperPod] = "0"
			res := GetSpBlockNum(job)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("04-job with valid sp-block should return correct value", func() {
			replicas := int32(2)
			job := newCommonAscendJob()
			job.Annotations[AnnoKeyOfSuperPod] = "2"
			job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				"Worker": {
					Replicas: &replicas,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											api.HuaweiAscend910: resource.MustParse("4"),
										},
									},
								},
							},
						},
					},
				},
			}
			const expectedReplicaNum = 4
			res := GetSpBlockNum(job)
			convey.So(res, convey.ShouldEqual, expectedReplicaNum) // 2 replicas * 4 devices / 2 spBlock = 4
		})
	})
}
