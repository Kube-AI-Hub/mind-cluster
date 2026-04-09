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

// Package custom is used to filter custom faults defined in job yaml.
// for the mindie server job, custom will automatically filter L2 faults
package custom

import (
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/custom"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/common"
)

// CustomProcessor is used to filter custom faults defined in job yaml
var CustomProcessor *customProcessor

type customProcessor struct{}

func init() {
	CustomProcessor = &customProcessor{}
}

// Process is used to process filter custom fault codes and fault levels
func (processor *customProcessor) Process(info any) any {
	allJobInfoMap, allJobUsedDeviceMap := job.GetCustomFilterFaultJobAndUsedDeviceInfoMap()
	if len(allJobInfoMap) == 0 {
		hwlog.RunLog.Debug("no server job info, skip fault process")
		return info
	}
	if deviceContent, deviceOk := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]); deviceOk {
		deletedFaultCmMap := processor.processDeviceFaults(deviceContent, allJobInfoMap, allJobUsedDeviceMap)
		custom.FaultCache.SetDeletedDevFaultCmForNodeMap(deletedFaultCmMap)
		return deviceContent
	}
	if switchContent, switchOK := info.(constant.OneConfigmapContent[*constant.SwitchInfo]); switchOK {
		deletedFaultCmMap := processor.processSwitchFaults(switchContent, allJobInfoMap)
		custom.FaultCache.SetDeletedSwitchFaultCmForNodeMap(deletedFaultCmMap)
		return switchContent
	}
	return info
}

func (processor *customProcessor) processDeviceFaults(
	deviceContent constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm],
	allJobInfoMap map[string]map[string]constant.JobInfo,
	allJobUsedDeviceMap map[string]map[string]sets.String) map[string]*constant.AdvanceDeviceFaultCm {
	deletedFaultCmMap := make(map[string]*constant.AdvanceDeviceFaultCm)
	for nodeName, advanceDeviceFaultCm := range deviceContent.AllConfigmap {
		jobInfoMap, hasJobInfo := allJobInfoMap[nodeName]
		jobUsedDeviceInfoMap, hasUsedDeviceInfo := allJobUsedDeviceMap[nodeName]
		if !hasJobInfo || !hasUsedDeviceInfo {
			hwlog.RunLog.Debugf("nodeName: %s has no mindie server job info or used device info, "+
				"skip fault process", nodeName)
			continue
		}
		hwlog.RunLog.Debugf("nodeName: %s current advanceDeviceFaultCm.FaultDeviceList: %v",
			nodeName, advanceDeviceFaultCm.FaultDeviceList)

		deletedFaultCm := collectAndRemoveDeviceFaults(advanceDeviceFaultCm,
			jobInfoMap, jobUsedDeviceInfoMap)
		if len(deletedFaultCm.FaultDeviceList) > 0 {
			deletedFaultCmMap[nodeName] = deletedFaultCm
			hwlog.RunLog.Debugf("set nodeName: %s and device deletedFaultCm: %v to deletedFaultCmMap",
				nodeName, deletedFaultCm)
		}
	}
	return deletedFaultCmMap
}

func (processor *customProcessor) processSwitchFaults(switchContent constant.OneConfigmapContent[*constant.SwitchInfo],
	allJobInfoMap map[string]map[string]constant.JobInfo) map[string]*constant.SwitchInfo {
	deletedFaultCmMap := make(map[string]*constant.SwitchInfo)
	for cmName, switchInfo := range switchContent.AllConfigmap {
		hwlog.RunLog.Debugf("cmName: %s current switchInfo: %v", cmName, switchInfo)
		nodeName := strings.TrimPrefix(cmName, constant.SwitchInfoPrefix)
		if _, hasJobInfo := allJobInfoMap[nodeName]; !hasJobInfo {
			hwlog.RunLog.Debugf("node %s (from cm %s) has no mindie server job info, skip fault process",
				nodeName, cmName)
			continue
		}

		deletedSwitchInfo := collectAndRemoveSwitchFaults(switchInfo, allJobInfoMap[nodeName])
		if len(deletedSwitchInfo.FaultInfo) > 0 {
			deletedFaultCmMap[cmName] = deletedSwitchInfo
			hwlog.RunLog.Debugf("set cmName: %s and switch deletedSwitchInfo: %v to deletedFaultCmMap",
				cmName, deletedSwitchInfo)
		}
	}

	return deletedFaultCmMap
}

func copyAdvanceDeviceFaultCm(src *constant.AdvanceDeviceFaultCm) (*constant.AdvanceDeviceFaultCm, error) {
	dst := new(constant.AdvanceDeviceFaultCm)
	if err := util.DeepCopy(dst, src); err != nil {
		hwlog.RunLog.Errorf("deep copy AdvanceDeviceFaultCm failed: %v", err)
		return nil, err
	}
	return dst, nil
}

func collectAndRemoveDeviceFaults(src *constant.AdvanceDeviceFaultCm,
	jobInfoMap map[string]constant.JobInfo, usedDeviceMap map[string]sets.String) *constant.AdvanceDeviceFaultCm {
	totalDeleteFaults := make([]constant.DeviceFault, 0)
	for deviceName, faults := range src.FaultDeviceList {
		deleteFaults := getDeletedDeviceFault(faults, deviceName, jobInfoMap, usedDeviceMap)
		totalDeleteFaults = append(totalDeleteFaults, deleteFaults...)
	}
	dst := &constant.AdvanceDeviceFaultCm{
		DeviceType:  src.DeviceType,
		CmName:      src.CmName,
		SuperPodID:  src.SuperPodID,
		RackID:      src.RackID,
		ServerIndex: src.ServerIndex,
		UpdateTime:  src.UpdateTime,
	}
	for _, fault := range totalDeleteFaults {
		src.DelFaultAndFix(fault)
		dst.AddFaultAndFix(fault)
	}
	return dst
}

func copySwitchInfo(src *constant.SwitchInfo) (*constant.SwitchInfo, error) {
	dst := new(constant.SwitchInfo)
	if err := util.DeepCopy(dst, src); err != nil {
		hwlog.RunLog.Errorf("deep copy SwitchInfo failed: %v", err)
		return nil, err
	}
	return dst, nil
}

func shouldReportFault(faultTimeAndLevel constant.FaultTimeAndLevel, jobInfo constant.JobInfo,
	deviceName string, faultCode string, filterCodes, filterLevels map[string]time.Duration) bool {
	nowTime := time.Now().UnixMilli()
	durationMs := nowTime - faultTimeAndLevel.FaultReceivedTime
	hwlog.RunLog.Debugf("deviceName:%s, faultCode:%s, now:%v, faultReceivedTime:%v", deviceName,
		faultCode, time.UnixMilli(nowTime).Format("2006-01-02 15:04:05.000"),
		time.UnixMilli(faultTimeAndLevel.FaultReceivedTime).Format("2006-01-02 15:04:05.000"))
	if filterCodes != nil {
		if thresh, ok := filterCodes[faultCode]; ok {
			isTimeout := durationMs > thresh.Milliseconds()
			if !isTimeout {
				hwlog.RunLog.Infof("job:%s faultCode:%s device:%s, during:%dms thresh:%vms, should not report fault",
					jobInfo.Key, faultCode, deviceName, durationMs, thresh.Milliseconds())
			}
			return isTimeout
		}
	}
	if filterLevels != nil {
		if thresh, ok := filterLevels[faultTimeAndLevel.FaultLevel]; ok {
			isTimeOut := durationMs > thresh.Milliseconds()
			if !isTimeOut {
				hwlog.RunLog.Infof("job:%s faultLevel:%s device:%s, during:%dms thresh:%vms, should not report fault",
					jobInfo.Key, faultTimeAndLevel.FaultLevel, deviceName, durationMs, thresh.Milliseconds())
			}
			return isTimeOut
		}
	}
	hwlog.RunLog.Debugf("not filter job: %s faultCode %s faultLevel %s, using fault npu: %s, should report fault",
		jobInfo.Key, faultCode, faultTimeAndLevel.FaultLevel, deviceName)
	return true
}

func getDeletedDeviceFault(faults []constant.DeviceFault, deviceName string, jobInfoMap map[string]constant.JobInfo,
	jobUsedDeviceInfoMap map[string]sets.String) []constant.DeviceFault {
	deleteFaults := make([]constant.DeviceFault, 0, len(faults))
	for jobId, jobInfo := range jobInfoMap {
		jobUsedDeviceInfo, hasUsedDevices := jobUsedDeviceInfoMap[jobId]
		if !hasUsedDevices {
			hwlog.RunLog.Debugf("job %s has no used device info, report all faults", jobId)
			continue
		}
		if job.IsMindIeServerJob(&jobInfo) &&
			!common.Publisher.IsSubscribed(jobInfo.MultiInstanceJobId, constant.ControllerAppType) {
			hwlog.RunLog.Debugf("mindie job:%s not subscribed to grpc interface, should report fault", jobInfo.Key)
			continue
		}
		filterCodes := custom.CustomFault.GetCustomFilterCodes(jobInfo.Key)
		filterLevels := custom.CustomFault.GetCustomFilterLevels(jobInfo.Key)
		for _, faultInfo := range faults {
			faultTimeAndLevel, hasTimeLevel := faultInfo.FaultTimeAndLevelMap[faultInfo.FaultCode]
			if !hasTimeLevel {
				hwlog.RunLog.Warnf("fault %s has no time and level info, report it", faultInfo.FaultCode)
				continue
			}
			if deviceName != "" && !jobUsedDeviceInfo.Has(deviceName) {
				hwlog.RunLog.Debugf("job:%s does not use fault npu:%s, report fault:%v",
					jobInfo.Key, deviceName, faultInfo.FaultCode)
				continue
			}
			if !shouldReportFault(faultTimeAndLevel, jobInfo, deviceName, faultInfo.FaultCode, filterCodes, filterLevels) {
				deleteFaults = append(deleteFaults, faultInfo)
			}
		}
	}
	return deleteFaults
}

func collectAndRemoveSwitchFaults(switchInfo *constant.SwitchInfo, jobInfoMap map[string]constant.JobInfo) *constant.SwitchInfo {
	deletedFaults := make([]constant.SimpleSwitchFaultInfo, 0, len(switchInfo.FaultInfo))
	delFaultTimeAndLevelMap := make(map[string]constant.FaultTimeAndLevel, len(switchInfo.FaultTimeAndLevelMap))
	for _, jobInfo := range jobInfoMap {
		if job.IsMindIeServerJob(&jobInfo) &&
			!common.Publisher.IsSubscribed(jobInfo.MultiInstanceJobId, constant.ControllerAppType) {
			hwlog.RunLog.Debugf("mindie job:%s not subscribed to grpc interface, should report fault", jobInfo.Key)
			continue
		}
		filterCodes := custom.CustomFault.GetCustomFilterCodes(jobInfo.Key)
		filterLevels := custom.CustomFault.GetCustomFilterLevels(jobInfo.Key)
		for _, faultInfo := range switchInfo.FaultInfo {
			faultTimeAndLevelKey := faultInfo.GetFaultTimeAndLevelKey()
			faultTimeAndLevel, ok := switchInfo.FaultTimeAndLevelMap[faultTimeAndLevelKey]
			if !ok {
				hwlog.RunLog.Warnf("switchInfo has no faultTimeAndLevel for faultTimeAndLevelKey:%s, "+
					"report fault:%v", faultTimeAndLevelKey, faultInfo)
				continue
			}
			if !shouldReportFault(faultTimeAndLevel, jobInfo, "", faultInfo.AssembledFaultCode,
				filterCodes, filterLevels) {
				deletedFaults = append(deletedFaults, faultInfo)
				delFaultTimeAndLevelMap[faultTimeAndLevelKey] = faultTimeAndLevel
			}
		}
	}
	deletedSwitchInfo := &constant.SwitchInfo{
		SwitchFaultInfo: constant.SwitchFaultInfo{
			FaultLevel: constant.NotHandleFault,
			UpdateTime: switchInfo.UpdateTime,
			NodeStatus: constant.HealthyState,
		},
		CmName: switchInfo.CmName,
	}
	for _, faultInfo := range deletedFaults {
		switchInfo.DelFaultAndFix(faultInfo)
		deletedSwitchInfo.AddFaultAndFix1(faultInfo, delFaultTimeAndLevelMap[faultInfo.GetFaultTimeAndLevelKey()])
	}
	return deletedSwitchInfo
}
