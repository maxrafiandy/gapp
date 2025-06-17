package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"

	errpkg "scm/api/app/errors"
)

type Float string

func (f Float) String() string {
	return string(f)
}

func (f *Float) UnmarshalJSON(data []byte) error {
	// Try unmarshaling as float
	var floatVal float64
	if err := json.Unmarshal(data, &floatVal); err == nil {
		*f = Float(fmt.Sprintf("%g", floatVal))
		return nil
	}

	// Try unmarshaling as string
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		*f = Float(strVal)
		return nil
	}

	return errpkg.ErrFieldUnsupportedType
}

func (f *Float) UnmarshalText(text []byte) error {
	s := string(text)
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return errpkg.ErrFieldUnsupportedType
	}
	*f = Float(s)
	return nil
}

func (f Float) Float() float64 {
	val, _ := strconv.ParseFloat(string(f), 64)
	return val
}

func (f Float) Value() (driver.Value, error) {
	return strconv.ParseFloat(string(f), 64)
}

func (f *Float) Scan(value any) error {
	switch v := value.(type) {
	case float64:
		*f = Float(fmt.Sprintf("%g", v))
	case []byte:
		*f = Float(string(v))
	case string:
		*f = Float(v)
	default:
		return errpkg.ErrFieldUnsupportedType
	}
	return nil
}
