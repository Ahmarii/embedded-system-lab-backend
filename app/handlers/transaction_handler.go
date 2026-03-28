package handlers

import (
	"encoding/hex"
	"strings"

	"github.com/KaningNoppasin/embedded-system-lab-backend/app/models"
	"github.com/KaningNoppasin/embedded-system-lab-backend/app/repositories"
	"github.com/gofiber/fiber/v3"
)

type TransactionHandler struct {
	transactionRepo *repositories.TransactionRepository
	userRepo        *repositories.UserRepository
}

type createTransactionRequest struct {
	RFID string `json:"rfid"`
	Type string `json:"type"`
}

func NewTransactionHandler(transactionRepo *repositories.TransactionRepository, userRepo *repositories.UserRepository) *TransactionHandler {
	return &TransactionHandler{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}

func (h *TransactionHandler) CreateTransactionByUserRFID(c fiber.Ctx) error {
	var req createTransactionRequest
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

	req.Type = strings.TrimSpace(strings.ToUpper(req.Type))
	if req.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "type is required",
		})
	}
	if !models.IsValidTransactionType(req.Type) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid transaction type",
			"types":   mapKeys(models.TypeAmounts),
		})
	}

	user, err := h.userRepo.GetByRFID(req.RFID)
	if err == repositories.ErrUserNotFound {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "user not found",
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to get user",
		})
	}

	transaction, err := h.transactionRepo.Create(user, req.Type)
	if err == repositories.ErrInsufficientBalance {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "insufficient balance",
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to create transaction",
		})
	}

	if _, err := h.userRepo.UpdateAmountByID(user.UUID, transaction.RemainingBalance); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to update user balance",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(transactionResponse(transaction))
}

func (h *TransactionHandler) GetAllTransactions(c fiber.Ctx) error {
	transactions, err := h.transactionRepo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to get transactions",
		})
	}

	return c.JSON(fiber.Map{
		"transactions": transactionResponses(transactions),
	})
}

func (h *TransactionHandler) GetTransactionsByUserRFIDHashed(c fiber.Ctx) error {
	userRFIDHashed := strings.TrimSpace(c.Params("userRFIDHashed"))
	if userRFIDHashed == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "userRFIDHashed is required",
		})
	}

	userRFIDHashedBytes, err := hex.DecodeString(userRFIDHashed)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "userRFIDHashed must be hex encoded",
		})
	}

	transactions, err := h.transactionRepo.GetByUserRFIDHashed(userRFIDHashedBytes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to get transactions",
		})
	}

	return c.JSON(fiber.Map{
		"transactions": transactionResponses(transactions),
	})
}

func transactionResponses(transactions []models.Transaction) []fiber.Map {
	response := make([]fiber.Map, 0, len(transactions))
	for _, transaction := range transactions {
		response = append(response, transactionResponse(&transaction))
	}

	return response
}

func transactionResponse(transaction *models.Transaction) fiber.Map {
	return fiber.Map{
		"id":                transaction.UUID.String(),
		"user_id":           transaction.UserUUID.String(),
		"user_rfid_hashed":  hex.EncodeToString(transaction.UserRFIDHashed),
		"type":              transaction.Type,
		"amount":            transaction.Amount,
		"remaining_balance": transaction.RemainingBalance,
		"created_at":        transaction.CreatedAt,
	}
}

func mapKeys(values map[string]float64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	return keys
}
