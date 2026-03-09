#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

"""
诊断报告Sheet生成器
"""

from dataclasses import dataclass
from typing import List, Dict, Tuple

from ascend_fd_tk.core.report.sheet.base import BaseSheetGenerator
from ascend_fd_tk.core.report.threshold_report import ThresholdConfig, create_threshold_report, generate_threshold_excel


@dataclass
class DiagReportData:
    """诊断报告数据类，用于存储诊断报告信息"""
    # 诊断结果信息
    fault_domain: str  # 故障域
    fault_code: str  # 故障码
    fault_info: str  # 故障信息
    solution: str  # 处理建议


class DiagReportSheetGenerator(BaseSheetGenerator):
    """诊断报告Sheet生成器"""

    def __init__(self, cluster_info, excel_gen=None, diag_results=None):
        """
        初始化诊断报告Sheet生成器

        :param cluster_info: 集群信息缓存对象
        :param excel_gen: Excel生成器对象，如果不提供则创建新实例
        :param diag_results: 诊断结果列表
        """
        super().__init__(cluster_info, excel_gen)
        self.diag_results = diag_results or []

    @staticmethod
    def _create_threshold_configs() -> List[ThresholdConfig]:
        """
        创建阈值配置（诊断报告可能不需要阈值检查，返回空列表）

        :return: 阈值配置列表
        """
        return []

    @staticmethod
    def _create_header_config() -> Tuple[Dict[str, str], List[str]]:
        """
        创建header映射和顺序

        :return: (header_mapping, header_order)
            header_mapping: {field_name: header_name}
            header_order: [header_name]
        """
        header_mapping = {
            "fault_domain": "故障域",
            "fault_code": "故障码",
            "fault_info": "故障信息",
            "solution": "处理建议"
        }

        header_order = ["故障域", "故障码", "故障信息", "处理建议"]

        return header_mapping, header_order

    def generate_sheet(self) -> None:
        """
        生成诊断报告Excel Sheet
        """
        # 收集诊断报告数据
        diag_report_data_list = self._collect_diag_report_data()

        # 如果没有数据，跳过生成Sheet
        if not diag_report_data_list:
            return

        # 创建阈值配置（诊断报告可能不需要阈值检查）
        threshold_configs = self._create_threshold_configs()

        # 创建header映射和顺序
        header_mapping, header_order = self._create_header_config()

        # 创建报告Sheet
        sheet = create_threshold_report(
            sheet_name="诊断报告",
            data_list=diag_report_data_list,
            header_mapping=header_mapping,
            header_order=header_order,
            threshold_configs=threshold_configs,
            na_rep="-"
        )

        # 生成Excel
        generate_threshold_excel(
            excel_gen=self.excel_gen,
            sheets=[sheet]
        )

    def _collect_diag_report_data(self) -> List[DiagReportData]:
        """
        收集诊断报告数据

        :return: 诊断报告数据列表
        """
        data_list = []

        # 遍历所有诊断结果
        for diag_result in self.diag_results:
            # 转换为字典格式
            diag_dict = diag_result.to_dict()

            # 创建诊断报告数据对象
            data = DiagReportData(
                fault_domain=diag_dict.get("故障域", ""),
                fault_code=diag_dict.get("故障码", ""),
                fault_info=diag_dict.get("故障信息", ""),
                solution=diag_dict.get("处理建议", "")
            )

            data_list.append(data)

        # 按故障域排序
        data_list.sort(key=lambda x: x.fault_domain)

        return data_list
