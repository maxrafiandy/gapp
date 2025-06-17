package validator

import (
	"reflect"
)

func isZero(val reflect.Value) bool {
	return reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
}
