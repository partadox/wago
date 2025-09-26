package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/account"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

func InitRestAccount(app fiber.Router, accountService account.IAccountUsecase) {
	app.Post("/accounts", createAccount(accountService))
	app.Get("/accounts", listAccounts(accountService))
	app.Get("/accounts/:accountId", getAccount(accountService))
	app.Delete("/accounts/:accountId", deleteAccount(accountService))
	app.Post("/accounts/:accountId/login", loginAccount(accountService))
	app.Post("/accounts/:accountId/login-with-code", loginAccountWithCode(accountService))
	app.Post("/accounts/:accountId/logout", logoutAccount(accountService))
	app.Post("/accounts/:accountId/reconnect", reconnectAccount(accountService))
	app.Post("/accounts/:accountId/webhook", setAccountWebhook(accountService))
	app.Get("/accounts/:accountId/webhook", getAccountWebhook(accountService))
}

func createAccount(service account.IAccountUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			AccountID string `json:"account_id" validate:"required,min=1,max=50,alphanum"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Invalid request body",
			})
		}

		if req.AccountID == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "account_id is required",
			})
		}
		if len(req.AccountID) < 1 || len(req.AccountID) > 50 {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "account_id must be between 1 and 50 characters",
			})
		}

		response, err := service.CreateAccount(c.Context(), req.AccountID)
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: err.Error(),
			})
		}

		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Account created successfully",
			Results: response,
		})
	}
}

func listAccounts(service account.IAccountUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accounts, err := service.ListAccounts(c.Context())
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: err.Error(),
			})
		}

		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Success get accounts",
			Results: accounts,
		})
	}
}

func getAccount(service account.IAccountUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accountID := c.Params("accountId")
		if accountID == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Account ID is required",
			})
		}

		account, err := service.GetAccount(c.Context(), accountID)
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: err.Error(),
			})
		}

		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Success get account",
			Results: account,
		})
	}
}

func deleteAccount(service account.IAccountUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accountID := c.Params("accountId")
		if accountID == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Account ID is required",
			})
		}

		err := service.DeleteAccount(c.Context(), accountID)
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: err.Error(),
			})
		}

		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Account deleted successfully",
		})
	}
}

func loginAccount(service account.IAccountUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accountID := c.Params("accountId")
		if accountID == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Account ID is required",
			})
		}

		response, err := service.LoginAccount(c.Context(), accountID)
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: err.Error(),
			})
		}

		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Please scan the QR code",
			Results: response,
		})
	}
}

func loginAccountWithCode(service account.IAccountUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accountID := c.Params("accountId")
		if accountID == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Account ID is required",
			})
		}

		var req struct {
			PhoneNumber string `json:"phone_number" validate:"required,min=10,max=15"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Invalid request body",
			})
		}

		if req.PhoneNumber == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "phone_number is required",
			})
		}
		if len(req.PhoneNumber) < 10 || len(req.PhoneNumber) > 15 {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "phone_number must be between 10 and 15 characters",
			})
		}

		code, err := service.LoginAccountWithCode(c.Context(), accountID, req.PhoneNumber)
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: err.Error(),
			})
		}

		response := map[string]string{
			"code":    code,
			"message": "Please enter this code in your WhatsApp app",
		}

		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Login code generated successfully",
			Results: response,
		})
	}
}

func logoutAccount(service account.IAccountUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accountID := c.Params("accountId")
		if accountID == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Account ID is required",
			})
		}

		err := service.LogoutAccount(c.Context(), accountID)
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: err.Error(),
			})
		}

		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Account logged out successfully",
		})
	}
}

func reconnectAccount(service account.IAccountUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accountID := c.Params("accountId")
		if accountID == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Account ID is required",
			})
		}

		err := service.ReconnectAccount(c.Context(), accountID)
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: err.Error(),
			})
		}

		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Account reconnected successfully",
		})
	}
}

func setAccountWebhook(service account.IAccountUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accountID := c.Params("accountId")
		if accountID == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Account ID is required",
			})
		}

		var req struct {
			WebhookURL string `json:"webhook_url" validate:"required,url"`
			Secret     string `json:"secret"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Invalid request body",
			})
		}

		if req.WebhookURL == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "webhook_url is required",
			})
		}

		err := service.SetAccountWebhook(c.Context(), accountID, req.WebhookURL, req.Secret)
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: err.Error(),
			})
		}

		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Webhook set successfully",
		})
	}
}

func getAccountWebhook(service account.IAccountUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accountID := c.Params("accountId")
		if accountID == "" {
			return c.Status(400).JSON(utils.ResponseData{
				Status:  400,
				Code:    "BAD_REQUEST",
				Message: "Account ID is required",
			})
		}

		webhook, err := service.GetAccountWebhook(c.Context(), accountID)
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "ERROR",
				Message: err.Error(),
			})
		}

		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Success get webhook",
			Results: webhook,
		})
	}
}