#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

import argparse
import json
from typing import NoReturn, List, Tuple

from cluster_simulator.lib.utils import render

NODE_TEMPLATE = '''apiVersion: v1
kind: Node
metadata:
  annotations:
    baseDeviceInfos: '{{base_device_infos}}'
    node.alpha.kubernetes.io/ttl: "0"
    kwok.x-k8s.io/node: fake
    huawei.com/Ascend910: "0,1,2,3,4,5,6,7"
    huawei.com/Ascend910-Fault: "[]"
    huawei.com/Ascend910-NetworkUnhealthy: ""
  labels:
    beta.kubernetes.io/arch: arm64
    beta.kubernetes.io/os: linux
    kubernetes.io/arch: arm64
    kubernetes.io/hostname: {{node_name}}
    kubernetes.io/os: linux
    kubernetes.io/role: agent
    node-role.kubernetes.io/agent: ""
    node-role.kubernetes.io/worker: worker
    workerselector: dls-worker-node
    accelerator: huawei-Ascend910
    accelerator-type: module
    host-arch: huawei-arm
    type: kwok
  name: {{node_name}}
status:
  allocatable:
    cpu: 96
    memory: 720Gi
    huawei.com/Ascend910: 8
    pods: 110
  capacity:
    cpu: 96
    memory: 720Gi
    huawei.com/Ascend910: 8
    pods: 110
  nodeInfo:
    architecture: arm64
    bootID: ""
    containerRuntimeVersion: ""
    kernelVersion: ""
    kubeProxyVersion: fake
    kubeletVersion: fake
    machineID: ""
    operatingSystem: linux
    osImage: ""
    systemUUID: ""
  phase: Running'''

DEVICE_INFO_CM_TEMPLATE = '''apiVersion: v1
kind: ConfigMap
metadata:
  name: mindx-dl-deviceinfo-{{node_name}}
  namespace: kube-system
  labels:
    mx-consumer-cim: "true"
data:
  DeviceInfoCfg: |
    {
      "DeviceInfo": {{deviceinfo_template}},
      "CheckCode": "{{deviceinfo_checkcode}}"
    }
    '''


DEFAULT_DEVICEINFO_TEMPLATE = '{' +\
  '"DeviceList": {"huawei.com/Ascend910": "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-NetworkUnhealthy": ""},' +\
  '"UpdateTime": 1720679215' +\
  '}'

DEFAULT_DEVICEINFO_CHECKCODE = "0c1aaed0a363a68edfa245b38dad622e9ac6e3a73751217ca7b6198fa6b14a9e"


CARD_NUMBER = 8


def get_base_device_infos(server_index: int) -> str:
    base_device_infos_dict = {}
    for m in range(CARD_NUMBER):
        base_device_infos_dict[f"Ascend910-{m}"] = {
            "IP": f'10.0.{server_index}.{m}'
        }
    return json.dumps(base_device_infos_dict)


def get_k8s_resources(args: argparse.Namespace) -> List[Tuple[str, str]]:
    node_num = args.node_num
    deviceinfo_template = args.deviceinfo_template
    deviceinfo_checkcode = args.deviceinfo_checkcode

    manifests = []
    for i in range(node_num):
        node_name = f'kwok-node-910ax8-{i}'
        node_info = render(NODE_TEMPLATE, node_name=node_name, base_device_infos=get_base_device_infos(i))
        manifests.append((f'node/{node_name}', node_info, ))

        device_info = render(DEVICE_INFO_CM_TEMPLATE, node_name=node_name,
                            deviceinfo_template=deviceinfo_template, deviceinfo_checkcode=deviceinfo_checkcode)
        manifests.append((f'configmap/mindx-dl-deviceinfo-{node_name}', device_info, ))

    return manifests


def setup_arguments(parser: argparse.ArgumentParser) -> NoReturn:
    parser.add_argument('--node_num', default=4, type=int, help='node number')
    parser.add_argument('--deviceinfo_template', default=DEFAULT_DEVICEINFO_TEMPLATE, type=str, help='device info template')
    parser.add_argument('--deviceinfo_checkcode', default=DEFAULT_DEVICEINFO_CHECKCODE, type=str, help='checkcode for device info')