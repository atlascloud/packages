// package main is this really necessary
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	repoApi "github.com/iggy/packages/internal/openapi"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echomiddleware "github.com/oapi-codegen/echo-middleware"
)

func main() {
	var port = flag.Int("port", 8888, "Port for HTTP server")
	var dir = flag.String("dir", "/srv/packages", "Root directory for packages/config")

	flag.Parse()

	swagger, err := repoApi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// Create an instance of our handler which satisfies the generated interface
	papi := repoApi.NewPkgRepo(*dir)

	// This is how you set up a basic Echo router
	e := echo.New()
	// Enable metrics middleware
	p := prometheus.NewPrometheus("packages", nil)
	p.Use(e)

	// Log all requests
	e.Use(middleware.Logger())

	// secure middleware
	e.Use(middleware.Secure())

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	validatorOptions := &echomiddleware.Options{}

	validatorOptions.Options.AuthenticationFunc = func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		orgName := input.RequestValidationInput.PathParams["org"]
		validTokens := repoApi.GetValidTokens(orgName)
		// they probably forgot to set the auth header or the env var for the token
		if input.RequestValidationInput.Request.Header.Get("Authorization") == "" ||
			input.RequestValidationInput.Request.Header["Authorization"][0] == "Bearer" {
			return errors.New("no auth token")
		}
		token := strings.Split(input.RequestValidationInput.Request.Header["Authorization"][0], " ")[1]
		for _, t := range validTokens {
			if token == t {
				return nil
			}
		}
		return errors.New("invalid auth")
	}
	validatorOptions.Skipper = func(ctx echo.Context) bool {
		// we want the prometheus middleware to handle this, not the normal openapi route
		return ctx.Path() == "/metrics"
	}
	e.Use(echomiddleware.OapiRequestValidatorWithOptions(swagger, validatorOptions))

	// We now register our API above as the handler for the interface
	repoApi.RegisterHandlers(e, papi)

	// And we serve HTTP until the world ends.
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", *port)))
}
