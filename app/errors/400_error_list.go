package errors

var (
	ErrFieldRequired error

	ErrFieldBelowMinimum = func(min any) error {
		return registerBuiltinError("ErrFieldBelowMinimum", min)
	}

	ErrFieldAboveMaximum = func(max any) error {
		return registerBuiltinError("ErrFieldAboveMaximum", max)
	}

	ErrFieldMustBeEmail    error
	ErrFieldMustBeDigit    error
	ErrFieldMustBeAlphanum error
	ErrFieldMustBeAlphabet error
	ErrFieldMustBeDate     error
	ErrFieldMustBeDatetime error

	ErrFieldLengthBelowMinimum = func(min int) error {
		return registerBuiltinError("ErrFieldLengthBelowMinimum", min)
	}

	ErrFieldLengthAboveMaximum = func(max int) error {
		return registerBuiltinError("ErrFieldLengthAboveMaximum", max)
	}

	ErrFieldUnsupportedType error

	ErrFieldInvalidParam = func(param any) error {
		return registerBuiltinError("ErrFieldInvalidParam", param)
	}
)

func init() {
	loadYamlFile("400_error_list.yaml")

	ErrFieldRequired = registerBuiltinError("ErrFieldRequired")
	ErrFieldMustBeEmail = registerBuiltinError("ErrFieldMustBeEmail")
	ErrFieldMustBeDigit = registerBuiltinError("ErrFieldMustBeDigit")
	ErrFieldMustBeAlphanum = registerBuiltinError("ErrFieldMustBeAlphanum")
	ErrFieldMustBeAlphabet = registerBuiltinError("ErrFieldMustBeAlphabet")
	ErrFieldMustBeDate = registerBuiltinError("ErrFieldMustBeDate")
	ErrFieldMustBeDatetime = registerBuiltinError("ErrFieldMustBeDatetime")
	ErrFieldUnsupportedType = registerBuiltinError("ErrFieldUnsupportedType")
}
