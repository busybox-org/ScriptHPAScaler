package plugins

import (
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func MapToStruct(m map[string]string, s interface{}) error {
	jsonStr, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonStr, s)
}
