package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"

	errpkg "scm/api/app/errors"
)

var (
	digitRe    = regexp.MustCompile(`^\d+$`)
	alphanumRe = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	alphaRe    = regexp.MustCompile(`^[a-zA-Z]+$`)
	emailRe    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func requiredRule(value any, _ string) error {
	if value == nil {
		return errpkg.ErrFieldRequired
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return errpkg.ErrFieldRequired
	}
	if isZero(v) {
		return errpkg.ErrFieldRequired
	}
	return nil
}

func minRule(value any, param string) error {
	minVal, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return errpkg.ErrFieldInvalidParam(param)
	}

	val := reflect.ValueOf(value)

	switch val.Kind() {
	case reflect.String:
		// Check if the string can be parsed to float (e.g. Integer type)
		if f, err := strconv.ParseFloat(val.String(), 64); err == nil {
			if f < minVal {
				return errpkg.ErrFieldBelowMinimum(int64(minVal))
			}
		} else {
			// fallback to length check if not a number
			if float64(len(val.String())) < minVal {
				return errpkg.ErrFieldBelowMinimum(int64(minVal))
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(val.Int()) < minVal {
			return errpkg.ErrFieldBelowMinimum(int64(minVal))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(val.Uint()) < minVal {
			return errpkg.ErrFieldBelowMinimum(int64(minVal))
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() < minVal {
			return errpkg.ErrFieldBelowMinimum(int64(minVal))
		}
	default:
		// Try to parse from string
		s := fmt.Sprintf("%v", value)
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			if f < minVal {
				return errpkg.ErrFieldBelowMinimum(int64(minVal))
			}
		} else {
			return errpkg.ErrFieldUnsupportedType
		}
	}

	return nil
}

func maxRule(value any, param string) error {
	maxVal, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return errpkg.ErrFieldInvalidParam(param)
	}

	val := reflect.ValueOf(value)

	switch val.Kind() {
	case reflect.String:
		// Try parse as float
		if f, err := strconv.ParseFloat(val.String(), 64); err == nil {
			if f > maxVal {
				return errpkg.ErrFieldAboveMaximum(int64(maxVal))
			}
		} else {
			// fallback to string length
			if float64(len(val.String())) > maxVal {
				return errpkg.ErrFieldAboveMaximum(int64(maxVal))
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(val.Int()) > maxVal {
			return errpkg.ErrFieldAboveMaximum(int64(maxVal))
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(val.Uint()) > maxVal {
			return errpkg.ErrFieldAboveMaximum(int64(maxVal))
		}

	case reflect.Float32, reflect.Float64:
		if val.Float() > maxVal {
			return errpkg.ErrFieldAboveMaximum(int64(maxVal))
		}

	default:
		// Try fallback: convert to float via string
		s := fmt.Sprintf("%v", value)
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			if f > maxVal {
				return errpkg.ErrFieldAboveMaximum(int64(maxVal))
			}
		} else {
			return errpkg.ErrFieldUnsupportedType
		}
	}

	return nil
}

func minlenRule(value any, param string) error {
	min, _ := strconv.Atoi(param)
	if s, ok := value.(string); ok && len(s) < min {
		return errpkg.ErrFieldBelowMinimum(int64(min))
	}
	return nil
}

func maxlenRule(value any, param string) error {
	max, _ := strconv.Atoi(param)
	if s, ok := value.(string); ok && len(s) > max {
		return errpkg.ErrFieldAboveMaximum(int64(max))
	}
	return nil
}

func emailRule(value any, _ string) error {
	if s, ok := value.(string); ok && !emailRe.MatchString(s) {
		return errpkg.ErrFieldMustBeEmail
	}
	return nil
}

func digitRule(value any, _ string) error {
	if s, ok := value.(string); !ok || !digitRe.MatchString(s) {
		return errpkg.ErrFieldMustBeDigit
	}
	return nil

}

func alphanunRule(value any, _ string) error {
	if s, ok := value.(string); !ok || !alphanumRe.MatchString(s) {
		return errpkg.ErrFieldMustBeAlphanum
	}
	return nil
}

func alphabetRule(value any, _ string) error {
	if s, ok := value.(string); !ok || !alphaRe.MatchString(s) {
		return errpkg.ErrFieldMustBeAlphabet
	}
	return nil
}

func dateRule(value any, _ string) error {
	var s string

	switch v := value.(type) {
	case string:
		s = v
	case fmt.Stringer:
		s = v.String()
	default:
		return errpkg.ErrFieldMustBeDate
	}

	const format = "2006-01-02"

	if _, err := time.Parse(format, s); err != nil {
		return errpkg.ErrFieldMustBeDate
	}

	return nil
}

func datetimeRule(value any, _ string) error {
	var s string

	switch v := value.(type) {
	case string:
		s = v
	case fmt.Stringer:
		s = v.String()
	default:
		return errpkg.ErrFieldMustBeDatetime
	}

	layout := "2006-01-02 15:04:05"

	_, err := time.Parse(layout, s)
	if err != nil {
		return errpkg.ErrFieldMustBeDatetime
	}
	return nil

}
