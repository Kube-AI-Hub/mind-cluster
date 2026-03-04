/*
Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package configManager

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/common"
	util "infer-operator/pkg/common/client-go"
)

const (
	int2               = 2
	testInvalidKeyType = 123 // invalid key type for testing
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestNewConfigManager(t *testing.T) {
	convey.Convey("Test NewConfigManager", t, func() {
		cfgMgr := NewConfigManager()
		convey.So(cfgMgr, convey.ShouldNotBeNil)
	})
}

func TestConfigManagerStart(t *testing.T) {
	convey.Convey("Test ConfigManager Start", t, func() {
		cfgMgr := NewConfigManager()

		patch := gomonkey.ApplyFunc(util.AddInferOperatorCfgEventFuncs,
			func(_ ...func(interface{}, interface{}, string)) {
			})
		defer patch.Reset()

		cfgMgr.Start()
	})
}

func TestConfigManagerFaultStrategyHandler(t *testing.T) {
	convey.Convey("Test ConfigManager faultStrategyHandler", t, func() {
		cfgMgr := NewConfigManager()
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "default",
			},
			Data: map[string]string{
				"key1": "value1",
			},
		}

		convey.Convey("add operator", func() {
			cfgMgr.faultStrategyHandler(nil, configMap, common.AddOperator)
			value, exists := cfgMgr.GetConfigByCMAndKey("test-cm", "key1")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "value1")
		})

		convey.Convey("update operator", func() {
			newConfigMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"key1": "new-value1",
					"key2": "value2",
				},
			}
			cfgMgr.faultStrategyHandler(configMap, newConfigMap, common.UpdateOperator)
			value, exists := cfgMgr.GetConfigByCMAndKey("test-cm", "key1")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "new-value1")
			value, exists = cfgMgr.GetConfigByCMAndKey("test-cm", "key2")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "value2")
		})

		convey.Convey("delete operator", func() {
			cfgMgr.faultStrategyHandler(configMap, nil, common.DeleteOperator)
			_, exists := cfgMgr.GetConfigByCMAndKey("test-cm", "key1")
			convey.So(exists, convey.ShouldBeFalse)
		})

		convey.Convey("unknown operator", func() {
			// Should not panic, just log error
			cfgMgr.faultStrategyHandler(nil, nil, "unknown")
		})
	})
}

func TestConfigManagerHandleAdd(t *testing.T) {
	convey.Convey("Test ConfigManager handleAdd", t, func() {
		cfgMgr := NewConfigManager()

		convey.Convey("normal add", func() {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			}
			cfgMgr.handleAdd(configMap)

			value, exists := cfgMgr.GetConfigByCMAndKey("test-cm", "key1")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "value1")

			value, exists = cfgMgr.GetConfigByCMAndKey("test-cm", "key2")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "value2")
		})

		convey.Convey("add with invalid type", func() {
			// Should not panic
			cfgMgr.handleAdd("not a configmap")
		})

		convey.Convey("add with empty data", func() {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-cm",
					Namespace: "default",
				},
				Data: map[string]string{},
			}
			cfgMgr.handleAdd(configMap)

			result := cfgMgr.GetConfigByCMName("empty-cm")
			convey.So(result, convey.ShouldEqual, "{}")
		})
	})
}

func TestConfigManagerHandleUpdateNormalCases(t *testing.T) {
	convey.Convey("Test ConfigManager handleUpdate normal cases", t, func() {
		cfgMgr := NewConfigManager()

		// Prepare initial data
		oldCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "default",
			},
			Data: map[string]string{
				"key1": "old-value1",
				"key2": "old-value2",
			},
		}
		cfgMgr.handleAdd(oldCM)

		convey.Convey("normal update", func() {
			newCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"key1": "new-value1",
					"key3": "value3",
				},
			}
			cfgMgr.handleUpdate(oldCM, newCM)

			value, exists := cfgMgr.GetConfigByCMAndKey("test-cm", "key1")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "new-value1")

			_, exists = cfgMgr.GetConfigByCMAndKey("test-cm", "key2")
			convey.So(exists, convey.ShouldBeFalse)

			value, exists = cfgMgr.GetConfigByCMAndKey("test-cm", "key3")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "value3")
		})

		convey.Convey("update non-existent configmap", func() {
			newCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "new-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "value",
				},
			}
			cfgMgr.handleUpdate(nil, newCM)

			value, exists := cfgMgr.GetConfigByCMAndKey("new-cm", "key")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "value")
		})
	})
}

func TestConfigManagerHandleUpdateErrorCases(t *testing.T) {
	convey.Convey("Test ConfigManager handleUpdate error cases", t, func() {
		cfgMgr := NewConfigManager()

		// Prepare initial data
		oldCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "default",
			},
			Data: map[string]string{
				"key1": "old-value1",
				"key2": "old-value2",
			},
		}
		cfgMgr.handleAdd(oldCM)

		convey.Convey("update with invalid new object", func() {
			// Should not panic
			cfgMgr.handleUpdate(oldCM, "invalid object")
		})

		convey.Convey("update with invalid cache type", func() {
			// Manually set invalid cache type
			cfgMgr.configCache.Store("test-cm", "not a sync.Map")

			newCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{"key": "value"},
			}
			// Should not panic
			cfgMgr.handleUpdate(oldCM, newCM)
		})
	})
}

func TestConfigManagerHandleDelete(t *testing.T) {
	convey.Convey("Test ConfigManager handleDelete", t, func() {
		cfgMgr := NewConfigManager()

		// Prepare initial data
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "default",
			},
			Data: map[string]string{
				"key1": "value1",
			},
		}
		cfgMgr.handleAdd(configMap)

		convey.Convey("normal delete", func() {
			cfgMgr.handleDelete(configMap)

			_, exists := cfgMgr.GetConfigByCMAndKey("test-cm", "key1")
			convey.So(exists, convey.ShouldBeFalse)
		})

		convey.Convey("delete with invalid type", func() {
			// Should not panic
			cfgMgr.handleDelete("not a configmap")
		})

		convey.Convey("delete non-existent configmap", func() {
			nonExistentCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "non-existent",
				},
			}
			// Should not panic
			cfgMgr.handleDelete(nonExistentCM)
		})
	})
}

func TestConfigManagerGetConfig(t *testing.T) {
	convey.Convey("Test ConfigManager GetConfig", t, func() {
		cfgMgr := NewConfigManager()

		convey.Convey("empty cache", func() {
			result := cfgMgr.GetConfig()
			convey.So(result, convey.ShouldEqual, "{}")
		})

		convey.Convey("with data", func() {
			// Add test data
			cm1 := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "cm1"},
				Data: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			}
			cm2 := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "cm2"},
				Data: map[string]string{
					"key3": "value3",
				},
			}
			cfgMgr.handleAdd(cm1)
			cfgMgr.handleAdd(cm2)

			result := cfgMgr.GetConfig()

			var parsed map[string]map[string]interface{}
			err := json.Unmarshal([]byte(result), &parsed)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(parsed), convey.ShouldEqual, int2)

			// Check cm1 exists and has correct length
			cm1Data, exists := parsed["cm1"]
			if !exists {
				t.Fatalf("Expected cm1 to exist in parsed result")
			}
			convey.So(len(cm1Data), convey.ShouldEqual, int2)

			// Check cm2 exists and has correct length
			cm2Data, exists := parsed["cm2"]
			if !exists {
				t.Fatalf("Expected cm2 to exist in parsed result")
			}
			convey.So(len(cm2Data), convey.ShouldEqual, 1)
		})

		convey.Convey("with invalid cache types", func() {
			// Store invalid types in cache
			cfgMgr.configCache.Store(testInvalidKeyType, "invalid key type")
			cfgMgr.configCache.Store("invalid-value", "not a sync.Map")

			// Should not panic and should skip invalid entries
			result := cfgMgr.GetConfig()
			var parsed map[string]map[string]interface{}
			err := json.Unmarshal([]byte(result), &parsed)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestConfigManagerGetConfigByCMName(t *testing.T) {
	convey.Convey("Test ConfigManager GetConfigByCMName", t, func() {
		cfgMgr := NewConfigManager()

		// Prepare test data
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "test-cm"},
			Data: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		}
		cfgMgr.handleAdd(configMap)

		convey.Convey("get existing configmap", func() {
			result := cfgMgr.GetConfigByCMName("test-cm")

			var parsed map[string]interface{}
			err := json.Unmarshal([]byte(result), &parsed)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(parsed), convey.ShouldEqual, int2)

			// Check key1 exists and has correct value
			value, exists := parsed["key1"]
			if !exists {
				t.Fatalf("Expected key1 to exist in parsed result")
			}
			convey.So(value, convey.ShouldEqual, "value1")

			// Check key2 exists and has correct value
			value, exists = parsed["key2"]
			if !exists {
				t.Fatalf("Expected key2 to exist in parsed result")
			}
			convey.So(value, convey.ShouldEqual, "value2")
		})

		convey.Convey("get non-existent configmap", func() {
			result := cfgMgr.GetConfigByCMName("non-existent")
			convey.So(result, convey.ShouldEqual, "{}")
		})

		convey.Convey("get with invalid cache type", func() {
			cfgMgr.configCache.Store("invalid-cm", "not a sync.Map")
			result := cfgMgr.GetConfigByCMName("invalid-cm")
			convey.So(result, convey.ShouldEqual, "{}")
		})
	})
}

func TestConfigManagerGetConfigByCMAndKey(t *testing.T) {
	convey.Convey("Test ConfigManager GetConfigByCMAndKey", t, func() {
		cfgMgr := NewConfigManager()

		// Prepare test data
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "test-cm"},
			Data: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		}
		cfgMgr.handleAdd(configMap)

		convey.Convey("get existing key", func() {
			value, exists := cfgMgr.GetConfigByCMAndKey("test-cm", "key1")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "value1")
		})

		convey.Convey("get non-existent key", func() {
			value, exists := cfgMgr.GetConfigByCMAndKey("test-cm", "non-existent")
			convey.So(exists, convey.ShouldBeFalse)
			convey.So(value, convey.ShouldBeNil)
		})

		convey.Convey("get from non-existent configmap", func() {
			value, exists := cfgMgr.GetConfigByCMAndKey("non-existent", "key1")
			convey.So(exists, convey.ShouldBeFalse)
			convey.So(value, convey.ShouldBeNil)
		})

		convey.Convey("get with invalid cache type", func() {
			cfgMgr.configCache.Store("invalid-cm", "not a sync.Map")
			value, exists := cfgMgr.GetConfigByCMAndKey("invalid-cm", "key")
			convey.So(exists, convey.ShouldBeFalse)
			convey.So(value, convey.ShouldBeNil)
		})
	})
}

func TestConfigManagerConcurrency(t *testing.T) {
	convey.Convey("Test ConfigManager concurrency safety", t, func() {
		cfgMgr := NewConfigManager()
		done := make(chan bool)

		// Concurrent operations
		for i := 0; i < 10; i++ {
			go func(idx int) {
				cmName := "test-cm"
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: cmName},
					Data:       map[string]string{"key": "value"},
				}

				cfgMgr.handleAdd(configMap)
				cfgMgr.GetConfig()
				cfgMgr.GetConfigByCMName(cmName)
				cfgMgr.GetConfigByCMAndKey(cmName, "key")
				cfgMgr.handleUpdate(configMap, configMap)
				cfgMgr.handleDelete(configMap)

				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
		// Test passes if no deadlock occurs
	})
}

func TestConfigManagerEdgeCases(t *testing.T) {
	convey.Convey("Test ConfigManager edge cases", t, func() {
		cfgMgr := NewConfigManager()

		convey.Convey("configmap with large data", func() {
			largeData := make(map[string]string)
			for i := 0; i < 1000; i++ {
				largeData["key"] = "value"
			}

			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "large-cm"},
				Data:       largeData,
			}
			cfgMgr.handleAdd(configMap)

			result := cfgMgr.GetConfigByCMName("large-cm")
			convey.So(result, convey.ShouldNotEqual, "{}")
		})

		convey.Convey("configmap with special characters", func() {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "special-cm"},
				Data: map[string]string{
					"key.with.dots":    "value1",
					"key-with-dashes":  "value2",
					"key/with/slashes": "value3",
					"chinese键":         "chinese值",
				},
			}
			cfgMgr.handleAdd(configMap)

			value, exists := cfgMgr.GetConfigByCMAndKey("special-cm", "key.with.dots")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "value1")

			value, exists = cfgMgr.GetConfigByCMAndKey("special-cm", "chinese键")
			convey.So(exists, convey.ShouldBeTrue)
			convey.So(value, convey.ShouldEqual, "chinese值")
		})

		convey.Convey("empty configmap name", func() {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: ""},
				Data:       map[string]string{"key": "value"},
			}
			// Should handle empty name
			cfgMgr.handleAdd(configMap)

			_, exists := cfgMgr.GetConfigByCMAndKey("", "key")
			convey.So(exists, convey.ShouldBeTrue)
		})
	})
}
