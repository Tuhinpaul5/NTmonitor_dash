package handlers

import (
	"NTMonitor/models"
	"NTMonitor/repository"
	"log"
	"github.com/gofiber/fiber/v3"
)

type NodeHandler struct {
	Repo     *repository.NodeRepository
	NodeRepo *repository.NodeRepository
}

func NewNodeHandler(
	nodeRepo *repository.NodeRepository,
) *NodeHandler {
	return &NodeHandler{
		Repo: nodeRepo,
	}
}

// AddNode godoc
//
//	@Summary		Add a new node
//	@Description	Creates a new node with the provided details
//	@Tags			node
//	@Accept			json
//	@Produce		json
//	@Param			node	body		models.Node	true	"Node object to create"
//	@Success		201		{object}	models.Node
//	@Failure		400		{object}	map[string]string	"Invalid request body"
//	@Failure		500		{object}	map[string]string	"Could not create Node"
//	@Router			/api/node/nodes [post]
func (h *NodeHandler) AddNode(c fiber.Ctx) error {
	node := new(models.Node)
	if err := c.Bind().Body(node); err != nil {
		c.Response().SetStatusCode(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.Repo.Create(node); err != nil {
		c.Response().SetStatusCode(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{"error": "Could not create Node"})
	}

	c.Response().SetStatusCode(fiber.StatusCreated)
	return c.JSON(node)
}


// GetAllNodes godoc
//
//	@Summary		Get all nodes
//	@Description	Retrieves a list of all nodes
//	@Tags			node
//	@Produce		json
//	@Success		200	{array}		models.Node
//	@Failure		500	{object}	map[string]string	"Could not retrieve Nodes"
//	@Router			/api/node/nodes [get]
func (h *NodeHandler) GetAllNodes(c fiber.Ctx) error {
	log.Printf("here")
	nodes, err := h.Repo.FindAll()
	log.Printf("NODES: %v, Error: %v", nodes, err)

	if err != nil {
		c.Response().SetStatusCode(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{"error": "Could not retrieve Nodes"})
	}

	return c.JSON(nodes)
}



// GetNodeByID godoc
//
//	@Summary		Get a node by ID
//	@Description	Retrieves a single node by its ID
//	@Tags			node
//	@Produce		json
//	@Param			id	path		string	true	"Node ID"
//	@Success		200	{object}	models.Node
//	@Failure		500	{object}	map[string]string	"Could not retrieve Node"
//	@Router			/api/node/nodes/{id} [get]
func (h *NodeHandler) GetNodeByID(c fiber.Ctx) error {
	id := c.Params("id")
	node, err := h.Repo.FindByID(id)
	log.Printf("NODE: %v, Error: %v", node, err)
	if err != nil {
		c.Response().SetStatusCode(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{"error": "Could not retrieve Node"})
	}
	return c.JSON(node)
}