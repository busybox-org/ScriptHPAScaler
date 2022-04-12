package plugins

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"strconv"
)

var (
	ErrInvalidConfig = errors.New("invalid config")
	json             = jsoniter.ConfigCompatibleWithStandardLibrary
)

func MapToStruct(m map[string]string, s interface{}) error {
	jsonStr, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonStr, s)
}

func ParseInt64(s string) (int64, error) {
	return ParseInt(s, 64)
}

func ParseInt(s string, bitSize int) (int64, error) {
	if s == "" {
		return 0, nil
	}
	base := 10
	if len(s) > 2 && s[0:2] == "0x" {
		base = 16
		s = s[2:]
	}
	return strconv.ParseInt(s, base, bitSize)
}
