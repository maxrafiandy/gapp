package middleware

import (
	"fmt"
	"log"
	"runtime/debug"
	"scm/api/app"
)

func Recover(next app.HandlerFunc) app.HandlerFunc {
	return func(ctx *app.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic — you can use your own logger here
				log.Printf("[PANIC RECOVER] %v\n%s", r, debug.Stack())

				// Optionally: wrap panic as an error if your context expects it
				err = fmt.Errorf("internal panic recover")

				// Optionally: write response immediately
				ctx.ServerError(err)
			}
		}()

		// Continue to next middleware/handler
		return next(ctx)
	}
}
