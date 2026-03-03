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

package common

// InstanceIndexer is a structure used to identify the index information of inference service instances
type InstanceIndexer struct {
	// Service name, identifying owner InferService
	ServiceName string
	// Instance set key, used to associate the same group of instances
	InstanceSetKey string
	// Instance index, identifying the unique sequence number of the instance in InstanceSet
	InstanceIndex string
}

// RequeueError represents an error that requires the controller to requeue the request for reprocessing.
type RequeueError struct {
	// Message describes the reason for requeueing, should be human-readable for debugging.
	Message string
}

// Error implements the error interface for RequeueError.
// It returns the Message field as the error description.
func (req *RequeueError) Error() string {
	return req.Message
}

// NewRequeueError creates a new RequeueError with the given message.
// This is the preferred way to create RequeueError instances to ensure proper initialization.
func NewRequeueError(msg string) *RequeueError {
	return &RequeueError{Message: msg}
}
