package handlers

import (
	"strings"

	"github.com/KaningNoppasin/embedded-system-lab-backend/app/repositories"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	repo *repositories.UserRepository
}

type createUserRequest struct {
	RFID string `json:"rfid"`
}

type updateAmountRequest struct {
	Amount float64 `json:"amount"`
}

func NewUserHandler(repo *repositories.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) GetAllUsers(c fiber.Ctx) error {
	users, err := h.repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to get users",
		})
	}

	response := make([]fiber.Map, 0, len(users))
	for _, user := range users {
		response = append(response, fiber.Map{
			"id":     user.UUID.String(),
			"amount": user.Amount,
		})
	}

	return c.JSON(fiber.Map{
		"users": response,
	})
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	var req createUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	req.RFID = strings.TrimSpace(req.RFID)
	if req.RFID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "rfid is required",
		})
	}

	user, err := h.repo.CreateByRFID(req.RFID)
	if err == repositories.ErrUserAlreadyExists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "user already exists",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":     user.UUID.String(),
		"amount": user.Amount,
	})
}

func (h *UserHandler) GetAmountByRFID(c fiber.Ctx) error {
	rfid := strings.TrimSpace(c.Params("rfid"))
	if rfid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "rfid is required",
		})
	}

	user, err := h.repo.GetByRFID(rfid)
	if err == repositories.ErrUserNotFound {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "user not found",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to get user amount",
		})
	}

	return c.JSON(fiber.Map{
		"rfid":   rfid,
		"amount": user.Amount,
	})
}

func (h *UserHandler) UpdateAmountByRFID(c fiber.Ctx) error {
	rfid := strings.TrimSpace(c.Params("rfid"))
	if rfid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "rfid is required",
		})
	}

	var req updateAmountRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	user, err := h.repo.UpdateAmountByRFID(rfid, req.Amount)
	if err == repositories.ErrUserNotFound {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "user not found",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to update user amount",
		})
	}

	return c.JSON(fiber.Map{
		"rfid":   rfid,
		"amount": user.Amount,
	})
}
