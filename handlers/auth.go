package handlers

import (
	"NTMonitor/models"
	"NTMonitor/repository"
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
//	@Param			user	body		models.User	true	"User object"
//	@Success		201		{object}	models.User
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/api/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		log.Println("💀 here")
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.Repo.Create(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create user"})
	}
	log.Println("User registered successfully:", user.Email)
	return c.Status(201).JSON(user)
}