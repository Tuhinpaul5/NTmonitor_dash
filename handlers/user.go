package handlers

import (
	"NTMonitor/models"
	"NTMonitor/repository"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	Repo *repository.UserRepository
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
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.Repo.Create(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create user"})
	}

	return c.Status(201).JSON(user)
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
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	var users []models.User
	// Use GORM to find all
	if err := h.Repo.DB.Find(&users).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(users)
}

// GetUser godoc
//
//	@Summary		Get user by ID
//	@Description	Retrieve a single user by their ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	models.User
//	@Failure		404	{object}	map[string]string
//	@Router			/api/users/id/{id} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User
	if err := h.Repo.DB.First(&user, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}
	return c.JSON(user)
}