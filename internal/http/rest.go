package http

import (
	"github.com/labstack/echo/v4"
	"rockets/internal/http/gen"
	"rockets/internal/rocket"
)

type ServerOpts struct {
	Echo   *echo.Echo
	Rocket rocket.Service
}

// NewServer creates a new HTTP server with the provided options and attaches the API routes.
func NewServer(opts *ServerOpts) (*StrictServer, *echo.Echo) {
	api := NewStrictServer(opts)

	AttachHttpAPIRoutes(
		opts.Echo,
		gen.NewStrictHandler(api, nil),
	)

	return api, opts.Echo
}

func NewStrictServer(opts *ServerOpts) *StrictServer {
	return &StrictServer{
		rocket: opts.Rocket,
	}
}

type ServerInterfaceWrapper struct {
	Handler gen.ServerInterface
}

// AttachHttpAPIRoutes attaches the HTTP API routes to the provided Echo router.
func AttachHttpAPIRoutes(router gen.EchoRouter, si gen.ServerInterface) {
	hnd := gen.ServerInterfaceWrapper{Handler: si}
	router.GET(
		"/v1/rockets",
		hnd.ListRockets,
	)
	router.GET(
		"/v1/rockets/:id",
		hnd.GetRocketState,
	)

	router.POST(
		"/messages",
		hnd.IngestMessage,
	)
}
