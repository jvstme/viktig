package http_server

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func alwaysOkHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.SetContentType("text/plain")
	ctx.Response.SetBody([]byte("ok"))
}

func makeTestClient(handler fasthttp.RequestHandler) http.Client {
	listener := fasthttputil.NewInmemoryListener()
	go fasthttp.Serve(listener, handler)

	return http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return listener.Dial()
		},
	}}
}

func TestBearerTokenAuth(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		client := makeTestClient(bearerTokenAuth("test-token", alwaysOkHandler))
		resp, _ := client.Get("http://localhost/")
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, fasthttp.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, string(body), "unauthorized")
	})

	t.Run("authorized", func(t *testing.T) {
		client := makeTestClient(bearerTokenAuth("test-token", alwaysOkHandler))
		req, _ := http.NewRequest("GET", "http://localhost/", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		resp, _ := client.Do(req)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, fasthttp.StatusOK, resp.StatusCode)
		assert.Equal(t, string(body), "ok")
	})
}
