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
	Otp   uint   `json:"otp" validate:"required,min=100000,max=999999" default:"123456"`
}

type UpdateUserRequest struct {
	ID       string `json:"id" validate:"required,uuid4" default:"123e4567-e89b-12d3-a456-426614174000"`
	Username string `json:"username" default:"user123"`
	Bio      string `json:"bio" default:"Updated user bio description"`
	Phone    string `json:"phone" default:"+1234567890"`
	Country  string `json:"country" default:"India"`
	Address  string `json:"address" default:"123 Main Street, City, Country"`
}

type DeleteUserRequest struct {
	ID string `json:"id" validate:"required,uuid4" default:"123e4567-e89b-12d3-a456-426614174000"`
}

type ListUsersRequest struct {
	Page     int `json:"page" validate:"min=1" default:"1"`
	PageSize int `json:"page_size" validate:"min=1,max=100" default:"10"`
}
