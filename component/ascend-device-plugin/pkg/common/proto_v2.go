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

// Package common a series of common function
package common

// DpuCMData  dpu data structure in configmap
type DpuCMData struct {
	// Name iface name
	Name string
	// Operstate iface operstate
	Operstate string
	// DeviceID iface device id
	DeviceID string
	// VendorID iface vendor id
	VendorID string
}

// DpuInfo DPU Info
type DpuInfo struct {
	BusType      string
	DPUList      []DpuCMData
	NpuToDpusMap map[string][]string
	UpdateTime   int64
}
