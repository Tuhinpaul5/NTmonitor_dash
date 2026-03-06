package handlers

import (
	"NTMonitor/dto"
	"NTMonitor/models"
	"NTMonitor/repository"
	mailer "NTMonitor/services"
	"NTMonitor/utils"
	"math/rand"

	"log"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Repo *repository.UserRepository
}

func NewAuthHandler(repo *repository.UserRepository, mailer *mailer.Mailer) *AuthHandler {
	return &AuthHandler{Repo: repo}
}

func (h *AuthHandler) HashPassword(password string) (string, error) {
	// Implement password hashing logic here (e.g., using bcrypt)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil // Return the hashed password as a string
}

func (h *AuthHandler) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Register a new user with email and profile data
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequest	true	"Registration Data"
//	@Success		201		{object}	models.User
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/api/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	// Single line to parse and validate
	var count int64

	req, err := utils.ParseAndValidate[dto.RegisterRequest](c)
	if err != nil {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusBadRequest,
			Success: &successFlag,
			Error:   err.Error(),
		})
	}

	ip := c.IP()

	h.Repo.DB.Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusBadRequest,
			Success: &successFlag,
			Error:   "Email already registered",
		})
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Profile: models.UserData{
			Bio:     req.Bio,
			Phone:   req.Phone,
			Address: req.Address,
			Ip:      ip,
		},
	}

	if err := h.Repo.Create(&user); err != nil {
		log.Print("Error creating user:", err)
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusInternalServerError,
			Success: &successFlag,
			Error:   "Could not create user",
		})
	}

	return utils.ResponseHandler(c, utils.ResponseOptions{
		Status:  fiber.StatusCreated,
		Message: "User registered successfully",
		Data:    user,
	})
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticate user with email and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest	true	"Login Credentials"
//	@Success		200		{object}	models.User
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/api/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req, err := utils.ParseAndValidate[dto.LoginRequest](c)
	if err != nil {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusBadRequest,
			Success: &successFlag,
			Error:   err.Error(),
		})
	}

	email := req.Email
	password := req.Password

	var user models.User
	if err := h.Repo.DB.Where("email = ?", email).First(&user).Error; err != nil {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusUnauthorized,
			Success: &successFlag,
			Error:   "Invalid email or password",
		})
	}

	if !h.CheckPasswordHash(password, user.Password) {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusUnauthorized,
			Success: &successFlag,
			Error:   "Invalid email or password",
		})
	}

	return utils.ResponseHandler(c, utils.ResponseOptions{
		Message: "Login successful",
		Data:    user,
	})
}

// SendOTP godoc
//
//	@Summary		Send OTP to user email
//	@Description	Generate and send a one-time password to the user's email address
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.OTPRequest	true	"Email Address"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Router			/api/auth/send-otp [post]
func (h *AuthHandler) SendOTP(c *fiber.Ctx) error {

	req, err := utils.ParseAndValidate[dto.OTPRequest](c)
	if err != nil {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusBadRequest,
			Success: &successFlag,
			Error:   err.Error(),
		})
	}

	otp := rand.Intn(900000) + 100000
	email := req.Email

	// log.Println(otp)
	

	// Placeholder for sending OTP logic (e.g., using an email service)
	log.Printf("Sending OTP %d to email %s\n", otp, email)

	return utils.ResponseHandler(c, utils.ResponseOptions{
		Message: "OTP sent successfully",
	})
}

// VerifyOTP godoc
//
//	@Summary		Verify OTP
//	@Description	Verify the one-time password sent to user's email
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.VerifyOTPRequest	true	"OTP Verification Data"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/api/auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	// Placeholder for verifying OTP logic
	return utils.ResponseHandler(c, utils.ResponseOptions{
		Message: "OTP verified successfully (placeholder)",
	})
}
