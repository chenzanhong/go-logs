package logs

import (
	"encoding/json"
	"fmt"
)

type Encoder interface {
	Encode(v ...interface{}) string
}

type PlainEncoder struct{}

func (e *PlainEncoder) Encode(v ...interface{}) string {
	return fmt.Sprint(v...)
}

type JsonEncoder struct{}

func (e *JsonEncoder) Encode(v ...interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("JSON marshal error: %v", err)
	}
	return string(b)
}
