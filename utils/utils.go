package utils

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"

	"crypto/rand"
	"math/big"
)

var validate = validator.New()

// ParseAndValidate binds the body to type T and runs struct validation
func ParseAndValidate[T any](c fiber.Ctx) (*T, error) {
	payload := new(T)

	// 1. Parse JSON Body using Fiber v3 Bind API
	if err := c.Bind().Body(payload); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid request payload")
	}

	// 2. Validate Struct Tags (e.g., validate:"required")
	if err := validate.Struct(payload); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return payload, nil
}

// APIResponse represents a standardized API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ResponseOptions holds optional parameters for ResponseHandler
type ResponseOptions struct {
	Status  int
	Success *bool
	Message string
	Error   string
	Data    interface{}
}

// ResponseHandler sends a standardized JSON response with optional parameters
// Default values: status=200, success=true, message="", error="", data=[]
func ResponseHandler(c fiber.Ctx, opts ResponseOptions) error {
	// Set defaults
	status := opts.Status
	if status == 0 {
		status = fiber.StatusOK
	}

	success := true
	if opts.Success != nil {
		success = *opts.Success
	}

	data := opts.Data
	if data == nil {
		data = []interface{}{}
	}

	response := APIResponse{
		Success: success,
		Status:  status,
		Message: opts.Message,
		Error:   opts.Error,
		Data:    data,
	}

	c.Response().SetStatusCode(status)
	return c.JSON(response)
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}
