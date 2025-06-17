package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	errpkg "scm/api/app/errors"
)

type Date string

const dateLayout = "2006-01-02"

func (d Date) String() string {
	return string(d)
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return errpkg.ErrFieldUnsupportedType
	}
	*d = Date(str)
	return nil
}

func (d *Date) UnmarshalText(text []byte) error {
	s := string(text)
	*d = Date(s)
	return nil
}

func (d Date) Time() (time.Time, error) {
	return time.Parse(dateLayout, string(d))
}

func (d Date) Value() (driver.Value, error) {
	return string(d), nil
}

func (d *Date) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		*d = Date(v.Format(dateLayout))
	case []byte:
		*d = Date(string(v))
	case string:
		*d = Date(v)
	default:
		return errpkg.ErrFieldUnsupportedType
	}
	return nil
}
