package middleware

import (
	"scm/api/app"
	applocale "scm/api/app/locale"
)

func LocaleWrapper(next app.HandlerFunc) app.HandlerFunc {
	return func(ctx *app.Context) error {
		locale := ctx.Query("lang")

		if locale == string(applocale.Bahasa) || locale == string(applocale.English) {
			ctx.UseLocale(applocale.Tag(locale))
		}

		return next(ctx)
	}
}
