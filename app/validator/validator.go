package validator

import (
	"reflect"
	"strings"
	"sync"

	errpkg "scm/api/app/errors"
)

type (
	RuleFunc func(value any, param string) error

	Validator interface {
		Validate() error
	}
)

var (
	validators = make(map[string]RuleFunc)
	mu         sync.RWMutex
)

func init() {
	RegisterValidator("required", requiredRule)
	RegisterValidator("minlen", minlenRule)
	RegisterValidator("maxlen", maxlenRule)
	RegisterValidator("email", emailRule)
	RegisterValidator("digit", digitRule)
	RegisterValidator("alphabet", alphabetRule)
	RegisterValidator("alphanum", alphanunRule)
	RegisterValidator("min", minRule)
	RegisterValidator("max", maxRule)
	RegisterValidator("date", dateRule)
	RegisterValidator("datetime", datetimeRule)
}

func RegisterValidator(name string, fn RuleFunc) {
	mu.Lock()
	defer mu.Unlock()
	validators[name] = fn
}

func GetValidator(name string) (RuleFunc, bool) {
	mu.RLock()
	defer mu.RUnlock()
	fn, ok := validators[name]
	return fn, ok
}

func ruleContains(rules []string, target string) bool {
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if strings.HasPrefix(rule, target+"=") || rule == target {
			return true
		}
	}
	return false
}

func ValidateStruct(dest any) error {
	val := reflect.ValueOf(dest)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	errors := make(errpkg.Errors)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)

		if structField.PkgPath != "" {
			continue // unexported
		}

		ruleTag := structField.Tag.Get("validation")
		rules := strings.Split(ruleTag, ",")

		fieldName := structField.Tag.Get("json")
		if fieldName == "" || fieldName == "-" {
			fieldName = structField.Name
		} else {
			fieldName = strings.Split(fieldName, ",")[0]
		}

		fieldValue := field
		isPtr := field.Kind() == reflect.Ptr

		if isPtr && field.IsNil() {
			if ruleContains(rules, "required") {
				if fn, ok := GetValidator("required"); ok {
					if err := fn(nil, ""); err != nil {
						errors[fieldName] = err
					}
				}
			}
			continue
		}

		if isPtr {
			fieldValue = field.Elem()
		}

		for _, rule := range rules {
			if _, exists := errors[fieldName]; exists {
				break // stop if required or any previous rule failed
			}

			rule = strings.TrimSpace(rule)
			parts := strings.SplitN(rule, "=", 2)
			name := parts[0]
			param := ""
			if len(parts) == 2 {
				param = parts[1]
			}

			fn, ok := GetValidator(name)
			if !ok {
				continue // unregistered validator
			}

			if err := fn(fieldValue.Interface(), param); err != nil {
				errors[fieldName] = err
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}
