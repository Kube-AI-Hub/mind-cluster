/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package utils is common utils
package utils

import (
	"strconv"

	"ascend-common/api"
	"ascend-operator/pkg/api/v1"
)

const (
	// AnnoKeyOfSuperPod annotation key of utils
	AnnoKeyOfSuperPod = "sp-block"
)

const (
	// SuperPodEnvPath super pod env path
	SuperPodEnvPath = `metadata.annotations['super-pod-rank']`
	// SuperPodAffinity super pod affinity key
	SuperPodAffinity = "super-pod-affinity"
	// SoftStrategy soft strategy
	SoftStrategy = "soft"
	// HardStrategy hard strategy
	HardStrategy = "hard"
	// SuperPodRankAnno super pod rank annotation key
	SuperPodRankAnno = "super-pod-rank"
	// Chip2Node16Sp a3x16 super pod schedule policy
	Chip2Node16Sp = "chip2-node16-sp"
	// Chip2Node8Sp a3x8 super pod schedule policy
	Chip2Node8Sp = "chip2-node8-sp"
)

// GetLogicSuperPodNodes Return the number of computational nodes contained in the logical utils
func getLogicSuperPodNodes(spBlock, chipsPerNode int) int {
	if spBlock < chipsPerNode {
		return 1
	}
	return spBlock / chipsPerNode
}

// GetLogicSuperPodId Return the logical utils ID
func GetLogicSuperPodId(podRank, spBlock, chipsPerNode int) int {
	if spBlock <= 0 || chipsPerNode <= 0 {
		return 0
	}
	return podRank / getLogicSuperPodNodes(spBlock, chipsPerNode)
}

// GetSpBlock get logic superpod sp-block value
func GetSpBlock(job *v1.AscendJob) int {
	if job == nil || job.Annotations == nil {
		return 0
	}

	spBlockStr := job.Annotations[AnnoKeyOfSuperPod]
	spBlock, err := strconv.Atoi(spBlockStr)
	if err != nil {
		spBlock = 0
	}
	return spBlock
}

func getAllDevicesForJob(job *v1.AscendJob) int {
	if job == nil || job.Spec.ReplicaSpecs == nil {
		return 0
	}

	totalDevices := 0
	for _, spec := range job.Spec.ReplicaSpecs {
		if spec == nil || spec.Replicas == nil {
			continue
		}
		replicas := int(*spec.Replicas)

		devicesPerPod := 0
		for _, container := range spec.Template.Spec.Containers {
			if quantity, ok := container.Resources.Requests[api.HuaweiAscend910]; ok {
				devices, _ := quantity.AsInt64()
				devicesPerPod = int(devices)
				break
			}
		}

		totalDevices += replicas * devicesPerPod
	}
	return totalDevices
}

// GetSpBlockNum get spblock num for job
func GetSpBlockNum(job *v1.AscendJob) int {
	if job == nil || job.Annotations == nil {
		return 0
	}

	allDevices := getAllDevicesForJob(job)
	spBlock := GetSpBlock(job)
	if spBlock == 0 {
		return 0
	}
	return allDevices / spBlock
}
