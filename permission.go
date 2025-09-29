package mod

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// CheckServicePermission 检查服务权限
func (app *App) CheckServicePermission(token string, permission *PermissionConfig) bool {
	if permission == nil || len(permission.Rules) == 0 {
		return true // 没有配置权限规则，默认允许访问
	}

	// 获取Token缓存数据
	tokenData, err := app.GetTokenData(token)
	if err != nil {
		app.logger.WithField("error", err.Error()).Debug("Failed to get token data for permission check")
		return false
	}

	// 解析Token数据为map
	var data map[string]interface{}
	if err := json.Unmarshal(tokenData, &data); err != nil {
		app.logger.WithField("error", err.Error()).Debug("Failed to unmarshal token data for permission check")
		return false
	}

	// 默认逻辑为AND
	logic := permission.Logic
	if logic == "" {
		logic = "AND"
	}

	// 评估权限规则
	if logic == "OR" {
		// OR逻辑：任一规则满足即可
		for _, rule := range permission.Rules {
			if app.evaluatePermissionRule(data, rule) {
				return true
			}
		}
		return false
	} else {
		// AND逻辑：所有规则都必须满足
		for _, rule := range permission.Rules {
			if !app.evaluatePermissionRule(data, rule) {
				return false
			}
		}
		return true
	}
}

// evaluatePermissionRule 评估单个权限规则
func (app *App) evaluatePermissionRule(data map[string]interface{}, rule PermissionRule) bool {
	// 获取字段值
	fieldValue := getNestedValue(data, rule.Field)

	switch rule.Operator {
	case "eq":
		return compareValues(fieldValue, rule.Value, "eq")
	case "ne":
		return compareValues(fieldValue, rule.Value, "ne")
	case "gt":
		return compareValues(fieldValue, rule.Value, "gt")
	case "gte":
		return compareValues(fieldValue, rule.Value, "gte")
	case "lt":
		return compareValues(fieldValue, rule.Value, "lt")
	case "lte":
		return compareValues(fieldValue, rule.Value, "lte")
	case "in":
		return valueInSlice(fieldValue, rule.Value)
	case "not_in":
		return !valueInSlice(fieldValue, rule.Value)
	case "contains":
		return stringContains(fieldValue, rule.Value)
	case "exists":
		return fieldValue != nil
	default:
		app.logger.WithField("operator", rule.Operator).Warn("Unknown permission operator")
		return false
	}
}

// getNestedValue 获取嵌套字段的值，支持点分隔路径如 "user.role"
func getNestedValue(data map[string]interface{}, fieldPath string) interface{} {
	if fieldPath == "" {
		return nil
	}

	parts := strings.Split(fieldPath, ".")
	current := data

	for i, part := range parts {
		if current == nil {
			return nil
		}

		if i == len(parts)-1 {
			// 最后一个字段
			if val, ok := current[part]; ok {
				return val
			}
			return nil
		}

		// 中间字段，继续向下查找
		if val, ok := current[part]; ok {
			if nextMap, ok := val.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return nil
			}
		} else {
			return nil
		}
	}

	return nil
}

// compareValues 比较两个值
func compareValues(fieldValue, expectedValue interface{}, operator string) bool {
	if fieldValue == nil && expectedValue == nil {
		return operator == "eq"
	}
	if fieldValue == nil || expectedValue == nil {
		return operator == "ne"
	}

	// 尝试转换为相同类型进行比较
	switch operator {
	case "eq":
		return valuesEqual(fieldValue, expectedValue)
	case "ne":
		return !valuesEqual(fieldValue, expectedValue)
	case "gt", "gte", "lt", "lte":
		return compareNumbers(fieldValue, expectedValue, operator)
	default:
		return false
	}
}

// valuesEqual 判断两个值是否相等
func valuesEqual(a, b interface{}) bool {
	// 直接比较
	if a == b {
		return true
	}

	// 尝试转换为字符串比较
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	if aStr == bStr {
		return true
	}

	// 尝试转换为数字比较
	if aNum, aOk := toFloat64(a); aOk {
		if bNum, bOk := toFloat64(b); bOk {
			return aNum == bNum
		}
	}

	return false
}

// compareNumbers 比较数字
func compareNumbers(fieldValue, expectedValue interface{}, operator string) bool {
	fieldNum, fieldOk := toFloat64(fieldValue)
	expectedNum, expectedOk := toFloat64(expectedValue)

	if !fieldOk || !expectedOk {
		return false
	}

	switch operator {
	case "gt":
		return fieldNum > expectedNum
	case "gte":
		return fieldNum >= expectedNum
	case "lt":
		return fieldNum < expectedNum
	case "lte":
		return fieldNum <= expectedNum
	default:
		return false
	}
}

// toFloat64 尝试将值转换为float64
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// valueInSlice 检查值是否在切片中
func valueInSlice(fieldValue, expectedValue interface{}) bool {
	expectedSlice := reflect.ValueOf(expectedValue)
	if expectedSlice.Kind() != reflect.Slice && expectedSlice.Kind() != reflect.Array {
		return false
	}

	for i := 0; i < expectedSlice.Len(); i++ {
		if valuesEqual(fieldValue, expectedSlice.Index(i).Interface()) {
			return true
		}
	}

	return false
}

// stringContains 检查字符串包含关系
func stringContains(fieldValue, expectedValue interface{}) bool {
	fieldStr := fmt.Sprintf("%v", fieldValue)
	expectedStr := fmt.Sprintf("%v", expectedValue)
	return strings.Contains(fieldStr, expectedStr)
}
