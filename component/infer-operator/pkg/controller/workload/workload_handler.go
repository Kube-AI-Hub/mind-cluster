/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package workload

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

type WorkLoadHandler interface {
	// CheckOrCreateWorkLoad checks if the workload exists and creates it if not
	CheckOrCreateWorkLoad(ctx context.Context, instanceSet *v1.InstanceSet, indexer common.InstanceIndexer) error
	// DeleteExtraWorkLoad deletes workloads that exceed the specified index limit
	DeleteExtraWorkLoad(ctx context.Context, indexer common.InstanceIndexer, indexLimit int) error
	// GetWorkLoadReadyReplicas returns the number of ready replicas of the workload
	GetWorkLoadReadyReplicas(ctx context.Context, indexer common.InstanceIndexer) (int, error)
	// Validate checks if the workload specification is valid
	Validate(spec runtime.RawExtension) error
	// GetReplicas retrieves the number of replicas from the workload specification
	GetReplicas(spec runtime.RawExtension) (int32, error)
}
