package mod

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// MockGenerator 负责根据结构体定义生成Mock数据
type MockGenerator struct {
	rand *rand.Rand
}

// NewMockGenerator 创建一个新的Mock数据生成器
func NewMockGenerator() *MockGenerator {
	return &MockGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateMockData 根据类型生成Mock数据
func (m *MockGenerator) GenerateMockData(t reflect.Type) any {
	if t == nil {
		return nil
	}

	// 处理指针类型
	if t.Kind() == reflect.Ptr {
		innerType := t.Elem()
		value := m.GenerateMockData(innerType)
		if value == nil {
			return nil
		}
		result := reflect.New(innerType)
		result.Elem().Set(reflect.ValueOf(value))
		return result.Interface()
	}

	switch t.Kind() {
	case reflect.Bool:
		return m.rand.Intn(2) == 1

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return m.generateIntValue(t.Kind())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return m.generateUintValue(t.Kind())

	case reflect.Float32, reflect.Float64:
		return m.generateFloatValue(t.Kind())

	case reflect.String:
		return m.generateStringValue()

	case reflect.Slice:
		return m.generateSliceValue(t)

	case reflect.Array:
		return m.generateArrayValue(t)

	case reflect.Map:
		return m.generateMapValue(t)

	case reflect.Struct:
		return m.generateStructValue(t)

	case reflect.Interface:
		// 对于any类型，生成一个字符串值
		return "mock_interface_value"

	default:
		return nil
	}
}

// generateIntValue 生成整数类型的Mock值
func (m *MockGenerator) generateIntValue(kind reflect.Kind) any {
	base := m.rand.Int63n(1000) + 1 // 1-1000的随机数

	switch kind {
	case reflect.Int:
		return int(base)
	case reflect.Int8:
		return int8(base % 128)
	case reflect.Int16:
		return int16(base % 32768)
	case reflect.Int32:
		return int32(base)
	case reflect.Int64:
		return base
	default:
		return int(base)
	}
}

// generateUintValue 生成无符号整数类型的Mock值
func (m *MockGenerator) generateUintValue(kind reflect.Kind) any {
	base := uint64(m.rand.Int63n(1000) + 1)

	switch kind {
	case reflect.Uint:
		return uint(base)
	case reflect.Uint8:
		return uint8(base % 256)
	case reflect.Uint16:
		return uint16(base % 65536)
	case reflect.Uint32:
		return uint32(base)
	case reflect.Uint64:
		return base
	default:
		return uint(base)
	}
}

// generateFloatValue 生成浮点数类型的Mock值
func (m *MockGenerator) generateFloatValue(kind reflect.Kind) any {
	base := m.rand.Float64() * 1000

	switch kind {
	case reflect.Float32:
		return float32(base)
	case reflect.Float64:
		return base
	default:
		return base
	}
}

// generateStringValue 生成字符串类型的Mock值
func (m *MockGenerator) generateStringValue() string {
	// 常用的Mock字符串模板
	templates := []string{
		"mock_string_%d",
		"test_value_%d",
		"sample_data_%d",
		"example_%d",
		"demo_text_%d",
	}

	template := templates[m.rand.Intn(len(templates))]
	return fmt.Sprintf(template, m.rand.Intn(10000))
}

// generateSliceValue 生成切片类型的Mock值
func (m *MockGenerator) generateSliceValue(t reflect.Type) any {
	elemType := t.Elem()
	length := m.rand.Intn(5) + 1 // 1-5个元素

	slice := reflect.MakeSlice(t, length, length)
	for i := 0; i < length; i++ {
		elem := m.GenerateMockData(elemType)
		if elem != nil {
			slice.Index(i).Set(reflect.ValueOf(elem))
		}
	}

	return slice.Interface()
}

// generateArrayValue 生成数组类型的Mock值
func (m *MockGenerator) generateArrayValue(t reflect.Type) any {
	elemType := t.Elem()
	length := t.Len()

	array := reflect.New(t).Elem()
	for i := 0; i < length; i++ {
		elem := m.GenerateMockData(elemType)
		if elem != nil {
			array.Index(i).Set(reflect.ValueOf(elem))
		}
	}

	return array.Interface()
}

// generateMapValue 生成Map类型的Mock值
func (m *MockGenerator) generateMapValue(t reflect.Type) any {
	keyType := t.Key()
	valueType := t.Elem()

	mapValue := reflect.MakeMap(t)
	length := m.rand.Intn(3) + 1 // 1-3个键值对

	for i := 0; i < length; i++ {
		key := m.GenerateMockData(keyType)
		value := m.GenerateMockData(valueType)

		if key != nil && value != nil {
			mapValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		}
	}

	return mapValue.Interface()
}

// generateStructValue 生成结构体类型的Mock值
func (m *MockGenerator) generateStructValue(t reflect.Type) any {
	structValue := reflect.New(t).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := structValue.Field(i)

		// 跳过不可设置的字段
		if !fieldValue.CanSet() {
			continue
		}

		// 根据字段标签生成特定类型的数据
		mockValue := m.generateFieldMockValue(field, fieldValue.Type())
		if mockValue != nil {
			fieldValue.Set(reflect.ValueOf(mockValue))
		}
	}

	return structValue.Interface()
}

// generateFieldMockValue 根据字段信息生成特定的Mock值
func (m *MockGenerator) generateFieldMockValue(field reflect.StructField, fieldType reflect.Type) any {
	fieldName := strings.ToLower(field.Name)
	jsonTag := field.Tag.Get("json")
	descTag := field.Tag.Get("desc")

	// 根据字段名或标签生成特定的值
	if jsonTag != "" && jsonTag != "-" {
		// 使用json标签名
		parts := strings.Split(jsonTag, ",")
		if len(parts) > 0 {
			fieldName = strings.ToLower(parts[0])
		}
	}

	// 根据字段名生成特定类型的数据
	if mockValue := m.generateSpecificMockValue(fieldName, descTag, fieldType); mockValue != nil {
		return mockValue
	}

	// 使用通用的类型生成
	return m.GenerateMockData(fieldType)
}

// generateSpecificMockValue 根据字段名生成特定的Mock值
func (m *MockGenerator) generateSpecificMockValue(fieldName, desc string, fieldType reflect.Type) any {
	if fieldType.Kind() != reflect.String {
		return nil
	}

	// 根据字段名模式生成特定值
	switch {
	case strings.Contains(fieldName, "id") || strings.Contains(fieldName, "uid"):
		return fmt.Sprintf("mock_id_%d", m.rand.Intn(100000))

	case strings.Contains(fieldName, "name"):
		names := []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank"}
		return names[m.rand.Intn(len(names))]

	case strings.Contains(fieldName, "email"):
		domains := []string{"example.com", "test.org", "mock.net"}
		names := []string{"user", "test", "demo", "sample"}
		return fmt.Sprintf("%s%d@%s",
			names[m.rand.Intn(len(names))],
			m.rand.Intn(1000),
			domains[m.rand.Intn(len(domains))])

	case strings.Contains(fieldName, "phone"):
		return fmt.Sprintf("138%08d", m.rand.Intn(100000000))

	case strings.Contains(fieldName, "url") || strings.Contains(fieldName, "link"):
		return fmt.Sprintf("https://example.com/mock/%d", m.rand.Intn(10000))

	case strings.Contains(fieldName, "token"):
		return fmt.Sprintf("mock_token_%s", m.generateRandomString(16))

	case strings.Contains(fieldName, "address"):
		addresses := []string{
			"北京市朝阳区", "上海市浦东新区", "广州市天河区",
			"深圳市南山区", "杭州市西湖区", "成都市高新区"}
		return addresses[m.rand.Intn(len(addresses))]

	case strings.Contains(fieldName, "message") || strings.Contains(fieldName, "msg"):
		messages := []string{
			"这是一条Mock消息", "测试数据", "示例内容",
			"模拟数据内容", "演示文本信息"}
		return messages[m.rand.Intn(len(messages))]

	case strings.Contains(fieldName, "status"):
		statuses := []string{"active", "inactive", "pending", "completed", "processing"}
		return statuses[m.rand.Intn(len(statuses))]

	default:
		return nil
	}
}

// generateRandomString 生成指定长度的随机字符串
func (m *MockGenerator) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[m.rand.Intn(len(charset))]
	}
	return string(result)
}

// isMockEnabled 检查给定的服务是否启用了Mock
func (app *App) isMockEnabled(service *Service) bool {
	config := app.GetModConfig()
	if config == nil {
		return false
	}

	mockConfig := &config.Mock

	// 1. 检查服务级别的Mock设置（最高优先级）
	if serviceConfig, exists := mockConfig.Services[service.Name]; exists {
		return serviceConfig.Enabled
	}

	// 2. 检查分组级别的Mock设置
	if service.Group != "" {
		if groupConfig, exists := mockConfig.Groups[service.Group]; exists {
			return groupConfig.Enabled
		}
	}

	// 3. 检查全局Mock设置（最低优先级）
	return mockConfig.Global.Enabled
}

// generateMockResponse 为服务生成Mock响应
func (app *App) generateMockResponse(service *Service) any {
	if service.Handler.OutputType == nil {
		return nil
	}

	generator := NewMockGenerator()
	return generator.GenerateMockData(service.Handler.OutputType)
}
