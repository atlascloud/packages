package api

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/steinfletcher/apitest"
	// "github.com/steinfletcher/apitest-jsonpath"
)

func TestListDistros_Success(t *testing.T) {
	papi := NewPkgRepo()
	e := echo.New()
	RegisterHandlers(e, papi)

	apitest.New().
		Handler(e).
		Get("/distributions").
		Expect(t).
		Body(`["alpine"]`).
		Status(http.StatusOK).
		End()
}

// func TestListOrganizations_Success(t *testing.T) {
// papi := NewPkgRepo()
// e := echo.New()
// RegisterHandlers(e, papi)
//
// apitest.New().
// Handler(e).
// Get("/o").
// Expect(t).
// Body(`["atlascloud"]`).
// Status(http.StatusOK).
// End()
//
// }
