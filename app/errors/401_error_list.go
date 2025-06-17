package errors

var (
	ErrUnauthorizedUser        error
	ErrExpiredToken            error
	ErrAccountSuspended        error
	ErrUnauthorizedApplication error
)

func init() {
	loadYamlFile("401_error_list.yaml")

	ErrUnauthorizedUser = registerBuiltinError("ErrUnauthorizedUser")
	ErrExpiredToken = registerBuiltinError("ErrExpiredToken")
	ErrAccountSuspended = registerBuiltinError("ErrAccountSuspended")
	ErrUnauthorizedApplication = registerBuiltinError("ErrUnauthorizedApplication")
}
