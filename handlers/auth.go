package handlers

import (
	"NTMonitor/dto"
	"NTMonitor/models"
	"NTMonitor/repository"
	"NTMonitor/utils"

	"log"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	Repo *repository.UserRepository
}

func NewAuthHandler(repo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{Repo: repo}
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
