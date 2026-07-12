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

// GetMe godoc
//
//	@Summary		Get Me
//	@Description	Retrieve the profile of the currently authenticated user.
//	@Tags			users
//	@Produce		json
//	@Success		200	{object}	models.User
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Security		SessionAuth
//	@Router			/api/users/get-me [get]
func (h *UserHandler) GetMe(c fiber.Ctx) error {
	// Get authenticated user ID from context
	authUserID, ok := c.Locals("user_id").(string)
	if !ok || authUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Fetch the authenticated user's profile
	var user models.User
	if err := h.Repo.DB.Preload("Profile").First(&user, "id = ?", authUserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User profile retrieved successfully",
		"data":    user,
	})
}


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
//	@Summary		Get all users (Admin only)
//	@Description	Retrieve a list of all users. Requires admin role.
//	@Tags			users
//	@Produce		json
//	@Success		200	{array}		models.User
//	@Failure		403	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		SessionAuth
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
//	@Description	Retrieve a user by their unique ID. Users can only view their own profile unless they are admin.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"User ID"
//	@Success		200		{object}	models.User
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		SessionAuth
//	@Router			/api/users/{id} [get]
func (h *UserHandler) GetUser(c fiber.Ctx) error {
	// Get authenticated user ID from context
	authUserID, ok := c.Locals("user_id").(string)
	if !ok || authUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Get target user ID from route parameter
	targetUserID := c.Params("id")

	// Fetch authenticated user to check their role
	var authUser models.User
	if err := h.Repo.DB.Select("id, type").First(&authUser, "id = ?", authUserID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Check if user is admin or accessing their own profile
	if authUser.Type != models.UserTypeAdmin && authUserID != targetUserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only access your own profile",
		})
	}

	// Fetch the target user
	var user models.User
	if err := h.Repo.DB.Preload("Profile").First(&user, "id = ?", targetUserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User retrieved successfully",
		"data":    user,
	})
}
