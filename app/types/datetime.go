package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	errpkg "scm/api/app/errors"
)

type Datetime string

const datetimeLayout = "2006-01-02 15:04:05"

func (dt Datetime) String() string {
	return string(dt)
}

func (dt *Datetime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return errpkg.ErrFieldUnsupportedType
	}

	*dt = Datetime(str)
	return nil
}

func (dt *Datetime) UnmarshalText(text []byte) error {
	s := string(text)
	*dt = Datetime(s)
	return nil
}

func (dt Datetime) Time() (time.Time, error) {
	return time.Parse(datetimeLayout, string(dt))
}

func (dt Datetime) Value() (driver.Value, error) {
	return string(dt), nil
}

func (dt *Datetime) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		*dt = Datetime(v.Format(datetimeLayout))
	case []byte:
		*dt = Datetime(string(v))
	case string:
		*dt = Datetime(v)
	default:
		return errpkg.ErrFieldUnsupportedType
	}
	return nil
}
