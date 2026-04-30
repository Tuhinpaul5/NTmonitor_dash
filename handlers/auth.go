package handlers

import (
	"log"
	"math/rand"
	"time"

	"NTMonitor/config"
	"NTMonitor/dto"
	"NTMonitor/models"
	"NTMonitor/repository"
	"NTMonitor/services"
	"NTMonitor/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	UserRepo    *repository.UserRepository
	OtpRepo     *repository.OtpRepository
	SessionRepo *repository.SessionRepository
	Mailer      *services.Mailer
	Cfg         *config.Config
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	otpRepo *repository.OtpRepository,
	sessionRepo *repository.SessionRepository,
	mailer *services.Mailer,
	cfg *config.Config,
) *AuthHandler {
	return &AuthHandler{
		UserRepo:    userRepo,
		OtpRepo:     otpRepo,
		SessionRepo: sessionRepo,
		Mailer:      mailer,
		Cfg:         cfg,
	}
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
//	@Description	Register a new user with email and profile data. Requires JWT token from OTP verification.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequest	true	"Registration Data"
//	@Success		201		{object}	models.User
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/api/auth/register [post]
func (h *AuthHandler) Register(c fiber.Ctx) error {
	req, err := utils.ParseAndValidate[dto.RegisterRequest](c)
	if err != nil {
		success := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusBadRequest,
			Success: &success,
			Error:   err.Error(),
		})
	}

	profile := models.UserData{
		Bio:     req.Bio,
		Phone:   req.Phone,
		Address: req.Address,
		Ip:      c.IP(),
	}

	var verifiedEmail string

	if h.Cfg.APP_ENV == "local" {
		verifiedEmail = req.Email // In local, trust the body
	} else {
		userToken, ok := c.Locals("user").(*jwt.Token)
		log.Printf("User Token: %v, Ok: %v", userToken, ok)

		if !ok || userToken == nil || !userToken.Valid {
			success := false
			return utils.ResponseHandler(c, utils.ResponseOptions{
				Status:  fiber.StatusUnauthorized,
				Success: &success,
				Error:   "Missing or invalid verification token",
			})
		}

		claims, ok := userToken.Claims.(jwt.MapClaims)
		if !ok {
			success := false
			return utils.ResponseHandler(c, utils.ResponseOptions{
				Status:  fiber.StatusUnauthorized,
				Success: &success,
				Error:   "Invalid token claims",
			})
		}

		verifiedEmail, ok = claims["email"].(string)
		if !ok || verifiedEmail == "" {
			success := false
			return utils.ResponseHandler(c, utils.ResponseOptions{
				Status:  fiber.StatusUnauthorized,
				Success: &success,
				Error:   "Invalid email in token",
			})
		}

		purpose, ok := claims["purpose"].(string)
		if !ok || purpose != "registration_bridge" {
			success := false
			return utils.ResponseHandler(c, utils.ResponseOptions{
				Status:  fiber.StatusForbidden,
				Success: &success,
				Error:   "Token purpose mismatch",
			})
		}

		// Optional but good: expiry check
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				success := false
				return utils.ResponseHandler(c, utils.ResponseOptions{
					Status:  fiber.StatusUnauthorized,
					Success: &success,
					Error:   "Token expired",
				})
			}
		} else {
			success := false
			return utils.ResponseHandler(c, utils.ResponseOptions{
				Status:  fiber.StatusUnauthorized,
				Success: &success,
				Error:   "Invalid token expiry",
			})
		}

		if verifiedEmail != req.Email {
			success := false
			return utils.ResponseHandler(c, utils.ResponseOptions{
				Status:  fiber.StatusUnauthorized,
				Success: &success,
				Error:   "Email verification failed",
			})
		}
	}

	email := req.Email

	var count int64
	h.UserRepo.DB.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		success := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusBadRequest,
			Success: &success,
			Error:   "Email already registered",
		})
	}

	hashedPassword, err := h.HashPassword(req.Password)
	if err != nil {
		success := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusInternalServerError,
			Success: &success,
			Error:   "Internal server error",
		})
	}

	user := models.User{
		Username: req.Username,
		Email:    email,
		Password: hashedPassword,
		Profile:  profile,
	}

	if err := h.UserRepo.Create(&user); err != nil {
		success := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusInternalServerError,
			Success: &success,
			Error:   "Could not create user account",
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
func (h *AuthHandler) Login(c fiber.Ctx) error {
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
	if err := h.UserRepo.DB.Preload("Profile").Where("email = ?", email).First(&user).Error; err != nil {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusUnauthorized,
			Success: &successFlag,
			Error:   "Invalid email or password",
		})
	}

	// log.Print("User found: %v", user)

	if !h.CheckPasswordHash(password, user.Password) {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusUnauthorized,
			Success: &successFlag,
			Error:   "Invalid email or password",
		})
	}

	// Delete old session from this device/browser if it exists
	oldSessionToken := c.Cookies("session_token")
	if oldSessionToken != "" {
		if err := h.SessionRepo.DeleteByToken(oldSessionToken); err != nil {
			log.Printf("Failed to delete old session: %v", err)
		}
	}

	// Generate new session token
	sessionToken, err := utils.GenerateSecureToken(128)
	// log.Println("TOKEN :: %v", sessionToken)
	if err != nil {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusInternalServerError,
			Success: &successFlag,
			Error:   "Failed to generate session token",
		})
	}

	// Create session in database
	userSession := models.UserSession{
		UserID:       user.ID,
		SessionToken: sessionToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour * 7), // 7 days
	}

	if err := h.SessionRepo.Create(&userSession); err != nil {
		log.Printf("Failed to create session in database: %v", err)
	}

	// Set session cookie using Fiber v3
	isSecure := h.Cfg.APP_ENV != "local" && h.Cfg.APP_ENV != "development"
	log.Printf("Setting cookie - Secure: %v, APP_ENV: %s", isSecure, h.Cfg.APP_ENV)

	// Determine domain for cookie - use empty for localhost/127.0.0.1 compatibility
	var cookieDomain string

	cookieDomain = h.Cfg.APP_DOMAIN // Use configured domain in production

	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  userSession.ExpiresAt,
		HTTPOnly: true,
		Secure:   isSecure, // Only secure in production
		SameSite: "Lax",
		Path:     "/", // Explicitly set path
		Domain:   cookieDomain,
	})

	return utils.ResponseHandler(c, utils.ResponseOptions{
		Message: "Login successful",
		Data:    user,
	})
}

// Logout godoc
//
//	@Summary		User logout
//	@Description	Log out the authenticated user by destroying the session
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/api/auth/logout [post]
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	// Get session token from cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken != "" {
		// Delete session from database
		if err := h.SessionRepo.DeleteByToken(sessionToken); err != nil {
			log.Printf("Failed to delete session from database: %v", err)
		}

		// Clear the cookie using Fiber v3
		isSecure := h.Cfg.APP_ENV != "local" && h.Cfg.APP_ENV != "development"

		// Use same domain logic as login
		var cookieDomain string
		if h.Cfg.APP_ENV == "local" || h.Cfg.APP_ENV == "development" {
			cookieDomain = ""
		} else {
			cookieDomain = h.Cfg.APP_DOMAIN
		}

		c.Cookie(&fiber.Cookie{
			Name:     "session_token",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: true,
			Secure:   isSecure,
			SameSite: "Lax",
			Path:     "/",
			Domain:   cookieDomain,
		})
	}

	return utils.ResponseHandler(c, utils.ResponseOptions{
		Message: "Logout successful",
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
func (h *AuthHandler) SendOTP(c fiber.Ctx) error {

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

	otps := models.OTP{
		Email:     email,
		Otp:       uint(otp),
		Type:      models.OtpTypeRegister,
		Used:      false,
		ExpiresAt: time.Now().Add(time.Minute * 15), // OTP expires in 15 minutes
	}

	if err := h.OtpRepo.Create(&otps); err != nil {
		log.Print("Error creating otps:", err)
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusInternalServerError,
			Success: &successFlag,
			Error:   "Could not create otps",
		})
	}

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
func (h *AuthHandler) VerifyOTP(c fiber.Ctx) error {
	req, err := utils.ParseAndValidate[dto.VerifyOTPRequest](c)
	if err != nil {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusBadRequest,
			Success: &successFlag,
			Error:   err.Error(),
		})
	}

	email := req.Email
	otp := req.Otp

	var otpRecord models.OTP
	if err := h.OtpRepo.DB.Where("email = ? AND type = ? AND otp = ? AND used = ? AND expires_at > ?", email, models.OtpTypeRegister, otp, false, time.Now()).First(&otpRecord).Error; err != nil {
		successFlag := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusUnauthorized,
			Success: &successFlag,
			Error:   "Invalid or expired OTP",
		})
	}

	// Mark OTP as used
	otpRecord.Used = true
	h.OtpRepo.DB.Save(&otpRecord)

	// Define claims specific to the registration bridge
	claims := jwt.MapClaims{
		"email":   email, // Get from DB/Request
		"purpose": "registration_bridge",
		"exp":     time.Now().Add(time.Minute * 1).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(h.Cfg.JWT_SECRET))
	if err != nil {
		success := false
		return utils.ResponseHandler(c, utils.ResponseOptions{
			Status:  fiber.StatusInternalServerError,
			Success: &success,
			Error:   "Internal server error",
		})
	}

	return utils.ResponseHandler(c, utils.ResponseOptions{
		Status:  fiber.StatusOK,
		Message: "OTP verified successfully",
		Data:    fiber.Map{"verification_token": t},
	})
}
