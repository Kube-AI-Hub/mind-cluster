// Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

// Package custom a series of job fault code and fault levels filter info function
// for the mindie server job, custom will filter L2 faults
package custom

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

const timeOut = 20

func mockCustomFault() *customFault {
	return &customFault{
		jobFilterCodes:  make(map[string]map[string]time.Duration),
		jobFilterLevels: make(map[string]map[string]time.Duration),
		rwMutex:         sync.RWMutex{},
	}
}

func TestSetAndGetCustomFilterLevels(t *testing.T) {
	convey.Convey("test SetAndGetCustomFilterLevels", t, func() {
		p := mockCustomFault()
		labelValue := "level1,level2:20，"
		convey.Convey("job is training job, get filter levels from label", func() {
			p.SetCustomFilterLevels(job1, labelValue, false)
			res := p.GetCustomFilterLevels(job1)
			convey.So(res, convey.ShouldResemble, map[string]time.Duration{
				"level1": constant.CustomFilterFaultDefaultTimeout,
				"level2": timeOut * time.Second,
			})
		})
		convey.Convey("job is mindie server job, get filter levels from default levels", func() {
			p.SetCustomFilterLevels(job1, "", true)
			res := p.GetCustomFilterLevels(job1)
			convey.So(res, convey.ShouldResemble, GetDefaultMindIeServerFilterLevels())
		})
		convey.Convey("job is mindie server job, levels from label overwrite default levels", func() {
			labelValue = "level1,level2:20，" + constant.RestartRequestFaultLevelStr + ": 20"
			p.SetCustomFilterLevels(job1, labelValue, true)
			res := p.GetCustomFilterLevels(job1)
			convey.So(res, convey.ShouldResemble, map[string]time.Duration{
				"level1":                             constant.CustomFilterFaultDefaultTimeout,
				"level2":                             timeOut * time.Second,
				constant.RestartRequestFaultLevelStr: timeOut * time.Second,
			})
		})
	})
}

func TestSetAndGetCustomFilterCodes(t *testing.T) {
	convey.Convey("test SetAndGetCustomFilterCodes", t, func() {
		p := mockCustomFault()
		labelValue := "code1,code2:20，[0x00f10509,132333,npu,na],"
		convey.Convey("job is training job, get filter codes from label", func() {
			p.SetCustomFilterCodes(job1, labelValue)
			res := p.GetCustomFilterCodes(job1)
			convey.So(res, convey.ShouldResemble, map[string]time.Duration{
				"code1":                      constant.CustomFilterFaultDefaultTimeout,
				"code2":                      timeOut * time.Second,
				"[0x00f10509,132333,npu,na]": constant.CustomFilterFaultDefaultTimeout,
			})
		})
		convey.Convey("job is mindie server job, get filter codes from default codes", func() {
			p.SetCustomFilterCodes(job1, "")
			res := p.GetCustomFilterCodes(job1)
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("job is mindie server job, codes from label overwrite default codes", func() {
			labelValue = "code1," + constant.DevCqeFaultCode + ":20"
			p.SetCustomFilterCodes(job1, labelValue)
			res := p.GetCustomFilterCodes(job1)
			convey.So(res, convey.ShouldResemble, map[string]time.Duration{
				"code1":                  constant.CustomFilterFaultDefaultTimeout,
				constant.DevCqeFaultCode: timeOut * time.Second,
			})
		})
	})
}

func TestParseCustomFilterFaultLabelValue(t *testing.T) {
	type testCase struct {
		input    string
		expected map[string]time.Duration
	}
	testCases := map[string]testCase{
		"fault": {
			input: "code1: 20,code2,code3: 0.2, code4:str,code5:-1, code6：,[0x00f10509,132333,npu,na], [0x00f10509]",
			expected: map[string]time.Duration{
				"code1":                      timeOut * time.Second,
				"code2":                      constant.CustomFilterFaultDefaultTimeout,
				"code3":                      constant.CustomFilterFaultDefaultTimeout,
				"code4":                      constant.CustomFilterFaultDefaultTimeout,
				"code5":                      constant.CustomFilterFaultDefaultTimeout,
				"code6":                      constant.CustomFilterFaultDefaultTimeout,
				"[0x00f10509,132333,npu,na]": constant.CustomFilterFaultDefaultTimeout,
				"[0x00f10509]":               constant.CustomFilterFaultDefaultTimeout,
			},
		},
		"test max value": {
			input: "code1: 86401, code2:86400",
			expected: map[string]time.Duration{
				"code1": constant.CustomFilterFaultDefaultTimeout,
				"code2": maxAnnoValue,
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			convey.Convey("", t, func() {
				result := parseCustomFilterFaultAnnoValue(tc.input)
				convey.So(result, convey.ShouldResemble, tc.expected)
			})

		})
	}
}

func TestJudgeFilterFaultLabelsByJobKey(t *testing.T) {
	convey.Convey("test JudgeFilterFaultAnnosByJobKey", t, func() {
		convey.So(JudgeFilterFaultAnnosByJobKey(job1), convey.ShouldBeFalse)
		convey.Convey("has filter fault codes, should return true", func() {
			CustomFault.jobFilterCodes[job1] = map[string]time.Duration{"code": 2}
			convey.So(JudgeFilterFaultAnnosByJobKey(job1), convey.ShouldBeTrue)
			CustomFault.DeleteCustomFilterCodes(job1)
		})
		convey.Convey("has filter fault levels, should return true", func() {
			CustomFault.jobFilterLevels[job1] = map[string]time.Duration{"level": 2}
			convey.So(JudgeFilterFaultAnnosByJobKey(job1), convey.ShouldBeTrue)
			CustomFault.DeleteCustomFilterCodes(job1)
		})
	})
}
