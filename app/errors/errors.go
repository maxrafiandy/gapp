package errors

import (
	"fmt"
	"log"
	"net/http"
	"scm/api/app/locale"
	"sort"
	"strings"

	stderrors "errors"
)

type (
	ErrCode int
	Error   struct {
		code          ErrCode
		err           error
		httpStatus    int
		localMessages map[locale.Tag]string
	}

	Errors map[string]error

	ErrAttr struct {
		HttpStatus int
		Code       ErrCode
		Messages   []locale.LangPackage
	}
)

var errorLangPack = make(map[string]ErrAttr)

func registerBuiltinError(key string, args ...any) error {

	newError := Error{
		code:          http.StatusInternalServerError,
		httpStatus:    http.StatusInternalServerError,
		err:           stderrors.New(key),
		localMessages: make(map[locale.Tag]string),
	}

	if langPack, found := errorLangPack[key]; found {
		newError.httpStatus = langPack.HttpStatus
		newError.code = langPack.Code

		for _, msg := range langPack.Messages {
			newError.localMessages[msg.Tag] = fmt.Sprintf(msg.Message, args...)
		}

		newError.err = stderrors.New(newError.localMessages[locale.English])
	} else {
		log.Panic("error package not found: ", key)
	}

	return &newError
}

func New(key string, attr *ErrAttr, args ...any) error {

	newError := Error{
		err:           stderrors.New(key),
		localMessages: make(map[locale.Tag]string),
	}

	if langPack, found := errorLangPack[key]; found {
		newError.httpStatus = langPack.HttpStatus
		newError.code = langPack.Code

		for _, msg := range langPack.Messages {
			newError.localMessages[msg.Tag] = fmt.Sprintf(msg.Message, args...)
		}
	} else {
		newError.httpStatus = http.StatusInternalServerError
		newError.code = http.StatusInternalServerError

		if attr != nil {
			if attr.HttpStatus != 0 {
				newError.httpStatus = attr.HttpStatus
			}

			if attr.Code != 0 {
				newError.code = attr.Code
			}

			for _, msg := range attr.Messages {
				newError.localMessages[msg.Tag] = msg.Message
			}
		}
	}

	return &newError
}

func (err Error) Code() ErrCode {
	return err.code
}

func (err Error) HttpStatus() int {
	return err.httpStatus
}

func (err Error) Error() string {
	if err.err != nil {
		if message, ok := err.localMessages[locale.English]; ok {
			return message
		}

		return err.err.Error()
	}

	return fmt.Sprintf("something went wrong (code %d)", err.code)
}

func (err Error) LocalizedError(tag locale.Tag) string {
	if msg, found := err.localMessages[tag]; found {
		return msg
	}

	return err.Error()
}

func (errs Errors) Error() string {
	if len(errs) == 0 {
		return ""
	}

	keys := make([]string, len(errs))
	i := 0
	for key := range errs {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	var s strings.Builder
	for i, key := range keys {
		if i > 0 {
			s.WriteString("; ")
		}

		if ers, ok := errs[key].(Errors); ok {
			_, _ = fmt.Fprintf(&s, "%v: (%v)", key, ers)
		} else if er, ok := errs[key].(*Error); ok {
			if message, ok := er.localMessages[locale.English]; ok {
				_, _ = fmt.Fprintf(&s, "%v: %v", key, message)
			} else {
				_, _ = fmt.Fprintf(&s, "%v: %v", key, er.Error())
			}
		} else {
			_, _ = fmt.Fprintf(&s, "%v: %v", key, errs[key])
		}
	}
	s.WriteString(".")

	return s.String()
}

func (errs Errors) LocalizedError(tag locale.Tag) map[string]any {
	result := make(map[string]any)

	if len(errs) == 0 {
		return result
	}

	for key, err := range errs {
		if ers, ok := err.(Errors); ok {
			result[key] = ers.LocalizedError(tag)
		} else if er, ok := errs[key].(*Error); ok {
			result[key] = er.LocalizedError(tag)
		} else {
			result[key] = err.Error()
		}
	}

	return result
}
