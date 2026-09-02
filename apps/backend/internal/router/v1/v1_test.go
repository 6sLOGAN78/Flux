package v1_test

import (
	"testing"

	"flux/apps/backend/internal/router/v1"

	"github.com/labstack/echo/v4"
)

func TestRegisterV1Routes(t *testing.T) {
	e := echo.New()
	
	// Ensure it initializes without panicking
	v1.RegisterV1Routes(e, nil, nil, nil, nil, nil, nil, nil)
}
