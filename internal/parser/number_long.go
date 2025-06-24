package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type numberLong int64

//easyjson:json
type mongoNumberLong struct {
	N json.Number `json:"$numberLong,nocopy"`
}

func (nl numberLong) MarshalJSON() ([]byte, error) {
	return strconv.AppendInt(make([]byte, 0, 24), int64(nl), 10), nil
}

func (nl *numberLong) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	switch c := data[0]; {
	case '0' <= c && c <= '9' || c == '-':
		v, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return fmt.Errorf("cant parse int: %w", err)
		}
		*nl = numberLong(v)
		return nil

	case c == '{':
		var tmp mongoNumberLong
		if err := tmp.UnmarshalJSON(data); err != nil {
			return fmt.Errorf("invalid mongo format: %w", err)
		}
		v, err := tmp.N.Int64()
		if err != nil {
			return fmt.Errorf("invalid mongo number: %w", err)
		}
		*nl = numberLong(v)
		return nil

	case c == '"' && len(data) > 1 && data[len(data)-1] == '"':
		v, err := strconv.ParseInt(string(data[1:len(data)-1]), 10, 64)
		if err != nil {
			return fmt.Errorf("string number: %w", err)
		}
		*nl = numberLong(v)
		return nil
	}

	return fmt.Errorf("invalid format: %s", string(data))
}
