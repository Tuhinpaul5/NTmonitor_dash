package dto

// Input schema for registration
type RegisterRequest struct {
	Username string `json:"username" validate:"required" default:"user123"`
	Email    string `json:"email" validate:"required,email" default:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" default:"password123"`
	Bio      string `json:"bio" default:"User bio description"`
	Phone    string `json:"phone" validate:"required" default:"+1234567890"`
	Country  string `json:"country" validate:"required" default:"India"`
	Address  string `json:"address" validate:"required" default:"123 Main Street, City, Country"`
	Otp      int    `json:"otp" validate:"required,len=6" default:"123456"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" default:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" default:"password123"`
}

type OTPRequest struct {
	Email string `json:"email" validate:"required,email" default:"user@example.com"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email" default:"user@example.com"`
	OTP   string `json:"otp" validate:"required,len=6" default:"123456"`
}