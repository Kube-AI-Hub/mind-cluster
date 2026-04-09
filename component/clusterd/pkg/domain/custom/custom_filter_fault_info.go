// Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

// Package custom a series of job fault code and fault levels filter info function
// for the mindie server job, custom will automatically filter L2 faults
package custom

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const (
	commaChar             = ','
	leftSquareBucketChar  = '['
	rightSquareBucketChar = ']'
	maxAnnoValue          = 24 * time.Hour
)

var (
	mindIeServerFilterLevels = map[string]time.Duration{
		constant.RestartRequestFaultLevelStr: constant.CustomFilterFaultDefaultTimeout,
	}
)

var switchFaultLevelMap = map[string]string{
	// device plugin function "LoadSwitchFaultCode" append resetFaults to separateFaults
	constant.ResetFaultStr:          constant.SeparateFaultLevelStr,
	constant.SeparateFault:          constant.SeparateFaultLevelStr,
	constant.RestartRequestFaultStr: constant.RestartRequestFaultLevelStr,
	constant.PreSeparateFault:       constant.PreSeparateFaultLevelStr,
}

// CustomFault store custom filter fault info, including fault codes and fault levels
var CustomFault *customFault

type customFault struct {
	jobFilterCodes  map[string]map[string]time.Duration
	jobFilterLevels map[string]map[string]time.Duration
	rwMutex         sync.RWMutex
}

func init() {
	CustomFault = &customFault{
		jobFilterCodes:  make(map[string]map[string]time.Duration),
		jobFilterLevels: make(map[string]map[string]time.Duration),
		rwMutex:         sync.RWMutex{},
	}
}

// SetCustomFilterCodes set custom filter codes
func (c *customFault) SetCustomFilterCodes(jobKey, value string) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	filterCodes := parseCustomFilterFaultAnnoValue(value)
	if len(filterCodes) == 0 {
		return
	}
	c.jobFilterCodes[jobKey] = filterCodes
	hwlog.RunLog.Infof("job %v set custom filter codes %v", jobKey, filterCodes)
}

// GetCustomFilterCodes get custom filter codes
func (c *customFault) GetCustomFilterCodes(jobKey string) map[string]time.Duration {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	return c.jobFilterCodes[jobKey]
}

// DeleteCustomFilterCodes delete custom filter codes
func (c *customFault) DeleteCustomFilterCodes(jobKey string) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	delete(c.jobFilterCodes, jobKey)
}

// SetCustomFilterLevels set custom filter levels
func (c *customFault) SetCustomFilterLevels(jobKey, value string, isMindIeServerJob bool) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	filterLevels := parseCustomFilterFaultAnnoValue(value)
	if isMindIeServerJob {
		util.MergeStringMapListOnlyNewKeys(filterLevels, mindIeServerFilterLevels)
	}
	if len(filterLevels) == 0 {
		return
	}
	c.jobFilterLevels[jobKey] = filterLevels
	hwlog.RunLog.Infof("job %v set custom filter levels %v", jobKey, filterLevels)
}

// GetCustomFilterLevels get custom filter levels
func (c *customFault) GetCustomFilterLevels(jobKey string) map[string]time.Duration {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	return c.jobFilterLevels[jobKey]
}

// DeleteCustomFilterLevels delete custom filter levels
func (c *customFault) DeleteCustomFilterLevels(jobKey string) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	delete(c.jobFilterLevels, jobKey)
}

func parseCustomFilterFaultAnnoValue(value string) map[string]time.Duration {
	filterFaultMap := make(map[string]time.Duration)
	if len(value) == 0 {
		return filterFaultMap
	}
	value = strings.Replace(value, constant.InvalidComma, constant.Comma, -1)
	value = strings.Replace(value, constant.InvalidColon, constant.Colon, -1)
	for _, config := range splitValueWithComma(value) {
		configStr := strings.TrimSpace(config)
		if configStr == "" {
			continue
		}
		faultCodeOrLevel := strings.SplitN(configStr, constant.Colon, constant.EachFaultFilterConfigMaxLen)
		key := strings.TrimSpace(faultCodeOrLevel[0])
		if newKey, ok := switchFaultLevelMap[key]; ok {
			key = newKey
		}
		if len(faultCodeOrLevel) == constant.EachFaultFilterConfigMaxLen {
			timeout, err := strconv.ParseInt(strings.TrimSpace(faultCodeOrLevel[1]), constant.FormatBase, constant.FormatBitSize64)
			if err != nil || timeout < 0 || timeout > int64(maxAnnoValue.Seconds()) {
				hwlog.RunLog.Warnf("%v is invalid, set default value, err:%v, maxValue:%v", faultCodeOrLevel[1],
					err, maxAnnoValue.Seconds())
				filterFaultMap[key] = constant.CustomFilterFaultDefaultTimeout
				continue
			}
			filterFaultMap[key] = time.Duration(timeout) * time.Second
			continue
		}
		filterFaultMap[key] = constant.CustomFilterFaultDefaultTimeout
	}
	return filterFaultMap
}

func splitValueWithComma(value string) []string {
	var items []string
	var current strings.Builder
	depth := 0
	for _, ch := range value {
		switch ch {
		case leftSquareBucketChar:
			depth++
			current.WriteRune(ch)
		case rightSquareBucketChar:
			depth--
			current.WriteRune(ch)
		case commaChar:
			if depth == 0 {
				items = append(items, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		items = append(items, current.String())
	}

	return items
}

// JudgeFilterFaultAnnosByJobKey judge whether job should filter fault code or level
func JudgeFilterFaultAnnosByJobKey(jobKey string) bool {
	return len(CustomFault.jobFilterCodes[jobKey]) > 0 || len(CustomFault.jobFilterLevels[jobKey]) > 0
}

// GetDefaultMindIeServerFilterLevels get default mindIeServerFilterLevels
func GetDefaultMindIeServerFilterLevels() map[string]time.Duration {
	return mindIeServerFilterLevels
}
