package http

import (
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/port"
	"github.com/gofiber/fiber/v2"
)

type RBACHandler struct {
	svc port.RBACService
}

func NewRBACHandler(svc port.RBACService) *RBACHandler {
	return &RBACHandler{svc: svc}
}

func (h *RBACHandler) CreateRole(c *fiber.Ctx) error {
	var req port.CreateRoleReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.svc.CreateRole(&req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Role created"})
}

func (h *RBACHandler) CreatePermission(c *fiber.Ctx) error {
	var req port.CreatePermReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.svc.CreatePermission(&req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Permission created"})
}

func (h *RBACHandler) AssignPermission(c *fiber.Ctx) error {
	var req port.AssignPermReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.svc.AssignPermissionToRole(&req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Permission assigned to Role"})
}

func (h *RBACHandler) AssignRole(c *fiber.Ctx) error {
	var req port.AssignRoleReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.svc.AssignRoleToUser(&req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Role assigned to User"})
}

func (h *RBACHandler) RemovePermission(c *fiber.Ctx) error {
	var req port.UnassignPermReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.svc.RemovePermissionFromRole(&req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Permission removed from Role"})
}

func (h *RBACHandler) RemoveRole(c *fiber.Ctx) error {
	var req port.UnassignRoleReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.svc.RemoveRoleFromUser(&req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Role removed from User"})
}
