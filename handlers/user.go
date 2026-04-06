package handlers

import (
	"NTMonitor/models"
	"NTMonitor/repository"
	"log"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	Repo     *repository.UserRepository
	UserRepo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{Repo: repo}
}

// CreateUser godoc
//
//	@Summary		Create a new user
//	@Description	Create a new user with email and profile data
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			user	body		models.User	true	"User object"
//	@Success		201		{object}	models.User
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/api/users [post]
func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	user := new(models.User)
	if err := c.Bind().Body(user); err != nil {
		c.Response().SetStatusCode(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.Repo.Create(user); err != nil {
		c.Response().SetStatusCode(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{"error": "Could not create user"})
	}

	c.Response().SetStatusCode(fiber.StatusCreated)
	return c.JSON(user)
}

// GetUsers godoc
//
//	@Summary		Get all users
//	@Description	Retrieve a list of all users
//	@Tags			users
//	@Produce		json
//	@Success		200	{array}		models.User
//	@Failure		500	{object}	map[string]string
//	@Router			/api/users [get]
func (h *UserHandler) GetUsers(c fiber.Ctx) error {
	// Get authenticated user ID from middleware
	userID := c.Locals("user_id").(string)
	log.Printf("Authenticated user %s is requesting all users", userID)

	var users []models.User
	// Use GORM to find all
	if err := h.Repo.DB.Find(&users).Error; err != nil {
		c.Response().SetStatusCode(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(users)
}

// GetUser godoc
//
//	@Summary		Get user by ID
//	@Description	Retrieve a user by their unique ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"User ID"
//	@Success		200		{object}	models.User
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/api/users/{id} [get]
func (h *UserHandler) GetUser(c fiber.Ctx) error {
	id := c.Params("id")

	var user models.User
	if err := h.Repo.DB.Preload("Profile").First(&user, "id = ?", id).Error; err != nil {
		c.Response().SetStatusCode(fiber.StatusNotFound)
		return c.JSON(fiber.Map{"error": "User not found"})
	}

	c.Response().SetStatusCode(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "User retrieved successfully",
		"data":    user,
	})
}
