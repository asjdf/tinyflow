package utils

import (
	"encoding/json"
	"fmt"
)

// ToJSONStr 对象转换成字符串
// 对象字段必须大写,否则结果为空
func ToJSONStr(data interface{}) (string, error) {
	result, err := json.Marshal(data)
	return fmt.Sprintf("%s", result), err
}
