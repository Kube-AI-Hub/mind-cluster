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

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

type Peer struct {
	LocalId int `json:"local_id"`
}

// TopoInfo topo file info
type TopoInfo struct {
	Version      string `json:"version"`
	HardwareType string `json:"hardware_type"`
	PeerCount    int    `json:"peer_count"`
	PeerList     []Peer `json:"peer_list"`
	EdgeList     []Edge `json:"edge_list"`
}

// Edge edge info
type Edge struct {
	NetLayer       int      `json:"net_layer"`
	LinkType       string   `json:"link_type"`
	TopoType       string   `json:"topo_type"`
	TopoInstanceId int      `json:"topo_instance_id"`
	TopoAttr       string   `json:"topo_attr"`
	LocalA         int      `json:"local_a"`
	LocalAPorts    []string `json:"local_a_ports"`
	LocalB         int      `json:"local_b"`
	LocalBPorts    []string `json:"local_b_ports"`
	Protocols      []string `json:"protocols"`
	Position       string   `json:"position"`
}
