package router_test

import (
	"testing"

	"flux/apps/backend/internal/router"

	"github.com/labstack/echo/v4"
)

func TestInitRouter(t *testing.T) {
	e := echo.New()
	
	// Ensure it initializes without panicking with nil dependencies (except Echo)
	router.InitRouter(e, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
}
