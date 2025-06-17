package app

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"reflect"
	errpkg "scm/api/app/errors"
	"scm/api/app/locale"
	"scm/api/app/validator"
	"strconv"
	"strings"
	"time"
)

type HandlerFunc func(*Context) error

type MiddlewareFunc func(HandlerFunc) HandlerFunc

type routeEntry struct {
	handler    HandlerFunc
	middleware []MiddlewareFunc
}

type Router struct {
	prefix     string
	routes     map[string]map[string]routeEntry
	middleware []MiddlewareFunc
}

type App struct {
	router *Router
	mw     []MiddlewareFunc
}

func New() *App {
	r := &Router{
		routes: make(map[string]map[string]routeEntry),
	}
	return &App{router: r}
}

func (app *App) Route() *Router {
	return app.router
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{writer: w, request: r}
	method := r.Method
	path := r.URL.Path

	start := time.Now()

	var entry routeEntry
	var found bool
	for route, e := range app.router.routes[method] {
		if params, ok := matchRoute(route, path); ok {
			ctx.Params = params
			entry = e
			found = true
			break
		}
	}

	if !found {
		http.NotFound(w, r)
		return
	}

	final := entry.handler

	for i := len(entry.middleware) - 1; i >= 0; i-- {
		final = entry.middleware[i](final)
	}
	// Apply global app middleware
	for i := len(app.mw) - 1; i >= 0; i-- {
		final = app.mw[i](final)
	}

	var message = "success"
	if err := final(ctx); err != nil {
		message = err.Error()
	}

	stop := time.Now()
	log.Printf("%s [%d] %s %s (%s) %d milliseconds", ctx.Request().Method,
		ctx.HttpStatus(),
		ctx.Request().URL.Path,
		ctx.Request().RemoteAddr,
		message, stop.Sub(start).Milliseconds())
}

func (app *App) Use(mw ...MiddlewareFunc) {
	app.mw = append(app.mw, mw...)
}

func (r *Router) Group(prefix string, mws ...MiddlewareFunc) *Router {
	return &Router{
		prefix:     r.prefix + prefix,
		routes:     r.routes,
		middleware: append([]MiddlewareFunc{}, append(r.middleware, mws...)...),
	}
}

func (r *Router) handle(method, path string, h HandlerFunc, mws ...MiddlewareFunc) {
	if r.routes[method] == nil {
		r.routes[method] = make(map[string]routeEntry)
	}
	// Simpan route dengan middleware chain (router group + route)
	allMiddleware := append([]MiddlewareFunc{}, r.middleware...)
	allMiddleware = append(allMiddleware, mws...)
	r.routes[method][path] = routeEntry{
		handler:    h,
		middleware: allMiddleware,
	}
}

func (r *Router) Use(mws ...MiddlewareFunc) {
	r.middleware = append(r.middleware, mws...)
}

func (r *Router) GET(path string, h HandlerFunc, mws ...MiddlewareFunc) {
	r.handle("GET", r.prefix+path, h, mws...)
}
func (r *Router) POST(path string, h HandlerFunc, mws ...MiddlewareFunc) {
	r.handle("POST", r.prefix+path, h, mws...)
}

func matchRoute(pattern, path string) (map[string]string, bool) {
	parts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	if len(parts) != len(pathParts) {
		return nil, false
	}
	params := make(map[string]string)
	for i := range parts {
		if strings.HasPrefix(parts[i], ":") {
			params[parts[i][1:]] = pathParts[i]
		} else if parts[i] != pathParts[i] {
			return nil, false
		}
	}
	return params, true
}

type Session interface {
	Get(key string) any
	Set(key string, value any)
	Delete(key string)
}

type Context struct {
	writer     http.ResponseWriter
	httpStatus int
	request    *http.Request
	locale     locale.Tag
	Params     map[string]string
	Session    Session
}

func (c *Context) UseLocale(l locale.Tag) {
	c.locale = l
}

func (c *Context) Locale() locale.Tag {
	return c.locale
}

func (c *Context) JSON(code int, data any) error {
	c.writer.Header().Set("Content-Type", "application/json")
	c.writer.WriteHeader(code)
	return json.NewEncoder(c.writer).Encode(data)
}

func (c *Context) Request() *http.Request {
	return c.request
}

func (c *Context) Writer() http.ResponseWriter {
	return c.writer
}

func (c *Context) HttpStatus() int {
	return c.httpStatus
}

func (c *Context) Success(data any) error {
	c.httpStatus = http.StatusOK
	return c.JSON(c.httpStatus, map[string]any{
		"code": fmt.Sprintf("%d", c.httpStatus),
		"data": data,
	})
}

func (c *Context) Unauthorized(err error) error {
	if er, ok := err.(*errpkg.Error); ok {
		c.mapError(er)

		return err
	}

	c.httpStatus = http.StatusUnauthorized

	c.JSON(c.httpStatus, map[string]any{
		"code": fmt.Sprintf("%d", c.httpStatus),
		"data": map[string]any{
			"description": fmt.Sprintf("general unautorized error: %s", err.Error()),
		},
	})

	return err
}

func (c *Context) BadInput(err error) error {
	if er, ok := err.(*errpkg.Error); ok {
		c.mapError(er)

		return err
	}

	c.httpStatus = http.StatusBadRequest

	if ers, ok := err.(errpkg.Errors); ok {
		c.JSON(c.httpStatus, map[string]any{
			"code": fmt.Sprintf("%d", c.httpStatus),
			"data": ers.LocalizedError(c.locale),
		})
	} else {
		c.JSON(c.httpStatus, map[string]any{
			"code": fmt.Sprintf("%d", c.httpStatus),
			"data": map[string]any{
				"description": fmt.Sprintf("general input error: %s", err.Error()),
			},
		})
	}

	return err
}

func (c *Context) NotAllowed(err error) error {
	if er, ok := err.(*errpkg.Error); ok {
		c.mapError(er)

		return err
	}

	c.httpStatus = http.StatusMethodNotAllowed

	c.JSON(c.httpStatus, map[string]any{
		"code": fmt.Sprintf("%d", c.httpStatus),
		"data": map[string]any{
			"description": fmt.Sprintf("general not allowed error: %s", err.Error()),
		},
	})

	return err
}

func (c *Context) BadGateway(err error) error {
	if er, ok := err.(*errpkg.Error); ok {
		c.mapError(er)

		return err
	}

	c.httpStatus = http.StatusBadGateway

	c.JSON(c.httpStatus, map[string]any{
		"code": fmt.Sprintf("%d", c.httpStatus),
		"data": map[string]any{
			"description": fmt.Sprintf("general bad gateway error: %s", err.Error()),
		},
	})

	return err
}

func (c *Context) ServerError(err error) error {
	if er, ok := err.(*errpkg.Error); ok {
		c.mapError(er)

		return err
	}

	c.httpStatus = http.StatusInternalServerError

	c.JSON(c.httpStatus, map[string]any{
		"code": fmt.Sprintf("%d", c.httpStatus),
		"error": map[string]any{
			"description": fmt.Sprintf("general server error: %s", err.Error()),
		},
	})

	return err
}

func (c *Context) mapError(err *errpkg.Error) error {
	c.httpStatus = err.HttpStatus()
	return c.JSON(err.HttpStatus(), map[string]any{
		"code": err.Code(),
		"data": map[string]any{
			"description": err.LocalizedError(c.locale),
		},
	})

}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) Query(key string) string {
	return c.request.URL.Query().Get(key)
}

func (c *Context) Bind(dest any) error {

	defer c.request.Body.Close()
	if err := json.NewDecoder(c.request.Body).Decode(dest); err != nil {
		return err
	}

	log.Println("--------", dest)

	return validator.ValidateStruct(dest)
}

func (c *Context) BindForm(dest any) error {
	if err := c.request.ParseForm(); err != nil {
		return err
	}
	return bindFormValues(c.request.Form, dest)
}

func (c *Context) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.request.FormFile(key)
}

func bindFormValues(values map[string][]string, dest any) error {
	v := reflect.ValueOf(dest).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)
		formKey := structField.Tag.Get("form")
		if formKey == "" {
			continue
		}
		if val, ok := values[formKey]; ok && len(val) > 0 {
			switch field.Kind() {
			case reflect.String:
				field.SetString(val[0])
			case reflect.Int, reflect.Int64:
				i, _ := strconv.ParseInt(val[0], 10, 64)
				field.SetInt(i)
			case reflect.Float64:
				f, _ := strconv.ParseFloat(val[0], 64)
				field.SetFloat(f)
			case reflect.Bool:
				b, _ := strconv.ParseBool(val[0])
				field.SetBool(b)
			case reflect.Ptr:
				ptr := reflect.New(field.Type().Elem())
				switch field.Type().Elem().Kind() {
				case reflect.String:
					ptr.Elem().SetString(val[0])
				case reflect.Int, reflect.Int64:
					i, _ := strconv.ParseInt(val[0], 10, 64)
					ptr.Elem().SetInt(i)
				case reflect.Float64:
					f, _ := strconv.ParseFloat(val[0], 64)
					ptr.Elem().SetFloat(f)
				case reflect.Bool:
					b, _ := strconv.ParseBool(val[0])
					ptr.Elem().SetBool(b)
				}
				field.Set(ptr)
			}
		}
	}
	return validator.ValidateStruct(dest)
}
