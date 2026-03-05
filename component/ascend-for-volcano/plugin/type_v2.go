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

// Package plugin is using for HuaWei Ascend pin affinity schedule.
package plugin

const (
	// nodeNPUNum indicates that the total npu number of each pod
	nodeNPUNum int = 8
)

// A5Fields for a5 fields
type A5Fields struct {
	TpBlock int
}
