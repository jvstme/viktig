package http_server

import (
	"fmt"
	"reflect"

	"github.com/valyala/fasthttp"
)

func BearerTokenAuth(token string, handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	expectedHeader := []byte(fmt.Sprintf("Bearer %s", token))

	return func(ctx *fasthttp.RequestCtx) {
		header := ctx.Request.Header.Peek("Authorization")
		if reflect.DeepEqual(header, expectedHeader) {
			handler(ctx)
		} else {
			ctx.Error("unauthorized", fasthttp.StatusUnauthorized)
		}
	}
}
