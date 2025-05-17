package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Encoder interface {
	Encode(v ...interface{}) string
}

type PlainEncoder struct{}

func (e *PlainEncoder) Encode(v ...interface{}) string {
	if len(v) == 1 {
		if s, ok := v[0].(string); ok {
			return s
		}
	}
	var b strings.Builder
	for i, val := range v {
		if i > 0 {
			b.WriteString(" ")
		}
		fmt.Fprint(&b, val)
	}

	return b.String()
}

type StructuredEncoder interface {
	EncodeWithFields(fields ...Field) string
	EncodeWithFieldsOrder(fields ...Field) string
}

// Field 表示一个结构化字段
type Field struct {
	Key   string
	Value interface{}
}

// 辅助函数构造 Field
func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

type JsonEncoder struct{}

func (e *JsonEncoder) Encode(v ...interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("JSON marshal error: %v", err)
	}
	return string(b)
}

// EncodeWithFields 将字段和消息编码为 JSON 字符串，ASCII 排序字段键。
func (e *JsonEncoder) EncodeWithFields(fields ...Field) string {
	// 初始化 map
	m := make(map[string]interface{}, len(fields))

	for _, f := range fields {
		m[f.Key] = f.Value
	}
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("JSON marshal error: %v", err)
	}
	return string(b)
}

// 将字段和消息编码为 JSON 字符串，按输入顺序排序字段键。(更快)
func (e *JsonEncoder) EncodeWithFieldsOrder(fields ...Field) string {
	var orderedFields []Field
	for _, f := range fields {
		orderedFields = append(orderedFields, f)
	}

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, f := range orderedFields {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(`"` + f.Key + `":`)
		jsonValue, _ := json.Marshal(f.Value)
		// if err!= nil {
		// 	return fmt.Sprintf("JSON marshal error: %v", err)
		// }
		buf.Write(jsonValue)
	}

	buf.WriteString("}\n")
	return buf.String()
}