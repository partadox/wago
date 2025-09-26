package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainAccount "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/account"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	infraAccount "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/account"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

type accountService struct {
	accountRepo      domainAccount.IAccountRepository
	accountManager   domainAccount.IAccountManager
	chatStorageRepo  domainChatStorage.IChatStorageRepository
}

func NewAccountService(accountRepo domainAccount.IAccountRepository, chatStorageRepo domainChatStorage.IChatStorageRepository) domainAccount.IAccountUsecase {
	return &accountService{
		accountRepo:     accountRepo,
		accountManager:  infraAccount.GlobalAccountManager,
		chatStorageRepo: chatStorageRepo,
	}
}

func (s *accountService) CreateAccount(ctx context.Context, accountID string) (domainAccount.CreateAccountResponse, error) {
	// Check if account already exists
	existing, err := s.accountRepo.GetAccount(accountID)
	if err != nil {
		return domainAccount.CreateAccountResponse{}, pkgError.InternalServerError(fmt.Sprintf("Failed to check account: %v", err))
	}
	if existing != nil {
		return domainAccount.CreateAccountResponse{}, pkgError.BadRequestError("Account already exists")
	}

	// Create account record
	account := &domainAccount.Account{
		ID:        accountID,
		Status:    domainAccount.StatusDisconnected,
		CreatedAt: time.Now(),
	}

	if err := s.accountRepo.CreateAccount(account); err != nil {
		return domainAccount.CreateAccountResponse{}, pkgError.InternalServerError(fmt.Sprintf("Failed to create account: %v", err))
	}

	// Initialize WhatsApp database for this account
	dbURI := fmt.Sprintf("file:storages/whatsapp_%s.db?_foreign_keys=1", accountID)
	logrus.Infof("Initializing database for account %s with URI: %s", accountID, dbURI)

	db := whatsapp.InitWaDB(ctx, dbURI)
	if db == nil {
		return domainAccount.CreateAccountResponse{}, pkgError.InternalServerError("Failed to initialize WhatsApp database")
	}

	var keysDB *sqlstore.Container
	if config.DBKeysURI != "" {
		// Format keysDB URI properly - replace foreign_keys param and add account suffix
		var keysDBURI string
		if strings.Contains(config.DBKeysURI, "?") {
			// Replace foreign_keys=on with foreign_keys=1 and add account suffix before the query params
			baseURI := strings.Split(config.DBKeysURI, "?")[0]
			queryParams := strings.Split(config.DBKeysURI, "?")[1]
			// Replace foreign_keys=on with foreign_keys=1
			queryParams = strings.ReplaceAll(queryParams, "_foreign_keys=on", "_foreign_keys=1")
			keysDBURI = fmt.Sprintf("%s_%s?%s", baseURI, accountID, queryParams)
		} else {
			keysDBURI = fmt.Sprintf("%s_%s?_foreign_keys=1", config.DBKeysURI, accountID)
		}
		logrus.Infof("Initializing keys database for account %s with URI: %s", accountID, keysDBURI)
		keysDB = whatsapp.InitWaDB(ctx, keysDBURI)
	}

	// Initialize WhatsApp client for this account
	logrus.Infof("Initializing WhatsApp client for account %s", accountID)
	client := whatsapp.InitWaCLI(ctx, db, keysDB, s.chatStorageRepo)
	if client == nil {
		return domainAccount.CreateAccountResponse{}, pkgError.InternalServerError("Failed to initialize WhatsApp client")
	}

	// Store client in account manager
	logrus.Infof("Storing client in account manager for account %s", accountID)
	s.accountManager.SetClient(accountID, client, db)

	// Verify client was stored
	if storedClient := s.accountManager.GetClient(accountID); storedClient == nil {
		return domainAccount.CreateAccountResponse{}, pkgError.InternalServerError("Failed to store client in account manager")
	}

	return domainAccount.CreateAccountResponse{
		AccountID: accountID,
		Message:   "Account created successfully",
	}, nil
}

func (s *accountService) DeleteAccount(ctx context.Context, accountID string) error {
	// Check if account exists
	account, err := s.accountRepo.GetAccount(accountID)
	if err != nil {
		return pkgError.InternalServerError(fmt.Sprintf("Failed to get account: %v", err))
	}
	if account == nil {
		return pkgError.NotFoundError("Account not found")
	}

	// Disconnect and remove client
	s.accountManager.RemoveClient(accountID)

	// Delete account from database
	if err := s.accountRepo.DeleteAccount(accountID); err != nil {
		return pkgError.InternalServerError(fmt.Sprintf("Failed to delete account: %v", err))
	}

	// Clean up account-specific database files
	go func() {
		dbPath := fmt.Sprintf("storages/whatsapp_%s.db", accountID)
		if err := utils.RemoveFile(0, dbPath); err != nil {
			logrus.Warnf("Failed to remove account database file %s: %v", dbPath, err)
		}
	}()

	return nil
}

func (s *accountService) ListAccounts(ctx context.Context) ([]domainAccount.AccountInfo, error) {
	accounts, err := s.accountRepo.ListAccounts()
	if err != nil {
		return nil, pkgError.InternalServerError(fmt.Sprintf("Failed to list accounts: %v", err))
	}

	var result []domainAccount.AccountInfo
	for _, account := range accounts {
		info := domainAccount.AccountInfo{
			AccountID:     account.ID,
			Status:        account.Status,
			PhoneNumber:   account.PhoneNumber,
			CreatedAt:     account.CreatedAt,
			LastConnected: account.LastConnected,
		}

		// Get connection status from client
		if client := s.accountManager.GetClient(account.ID); client != nil {
			info.IsConnected = client.IsConnected()
			info.IsLoggedIn = client.IsLoggedIn()
			if client.Store != nil && client.Store.ID != nil {
				info.DeviceID = client.Store.ID.String()
			}
		}

		// Get webhook info
		if webhook, err := s.accountRepo.GetWebhook(account.ID); err == nil {
			info.WebhookURL = webhook.URL
		}

		result = append(result, info)
	}

	return result, nil
}

func (s *accountService) GetAccount(ctx context.Context, accountID string) (domainAccount.AccountInfo, error) {
	account, err := s.accountRepo.GetAccount(accountID)
	if err != nil {
		return domainAccount.AccountInfo{}, pkgError.InternalServerError(fmt.Sprintf("Failed to get account: %v", err))
	}
	if account == nil {
		return domainAccount.AccountInfo{}, pkgError.NotFoundError("Account not found")
	}

	info := domainAccount.AccountInfo{
		AccountID:     account.ID,
		Status:        account.Status,
		PhoneNumber:   account.PhoneNumber,
		CreatedAt:     account.CreatedAt,
		LastConnected: account.LastConnected,
	}

	// Get connection status from client
	if client := s.accountManager.GetClient(accountID); client != nil {
		info.IsConnected = client.IsConnected()
		info.IsLoggedIn = client.IsLoggedIn()
		if client.Store != nil && client.Store.ID != nil {
			info.DeviceID = client.Store.ID.String()
		}
	}

	// Get webhook info
	if webhook, err := s.accountRepo.GetWebhook(accountID); err == nil {
		info.WebhookURL = webhook.URL
	}

	return info, nil
}

func (s *accountService) LoginAccount(ctx context.Context, accountID string) (domainAccount.LoginResponse, error) {
	client := s.accountManager.GetClient(accountID)
	if client == nil {
		return domainAccount.LoginResponse{}, pkgError.NotFoundError("Account not found or not initialized")
	}

	if client.IsLoggedIn() {
		return domainAccount.LoginResponse{}, pkgError.BadRequestError("Account is already logged in")
	}

	if client.IsConnected() {
		client.Disconnect()
	}

	// Generate QR code for login
	qrChan, _ := client.GetQRChannel(ctx)
	if err := client.Connect(); err != nil {
		return domainAccount.LoginResponse{}, pkgError.InternalServerError(fmt.Sprintf("Failed to connect: %v", err))
	}

	select {
	case evt := <-qrChan:
		switch evt.Event {
		case "code":
			qrFilename := fmt.Sprintf("scan-%s-%d.png", accountID, time.Now().Unix())
			qrPath := filepath.Join(config.PathQrCode, qrFilename)

			if err := qrcode.WriteFile(evt.Code, qrcode.Medium, 512, qrPath); err != nil {
				return domainAccount.LoginResponse{}, pkgError.InternalServerError(fmt.Sprintf("Failed to generate QR code: %v", err))
			}

			return domainAccount.LoginResponse{
				ImagePath: qrPath,
				Duration:  evt.Timeout,
				Code:      evt.Code,
			}, nil
		case "success":
			// Update account status
			account, _ := s.accountRepo.GetAccount(accountID)
			if account != nil {
				account.Status = domainAccount.StatusLoggedIn
				account.LastConnected = time.Now()
				if client.Store != nil && client.Store.ID != nil {
					account.DeviceID = client.Store.ID.String()
					account.PhoneNumber = client.Store.ID.User
				}
				s.accountRepo.UpdateAccount(account)
			}
			return domainAccount.LoginResponse{Code: "success"}, nil
		}
	case <-time.After(120 * time.Second):
		return domainAccount.LoginResponse{}, pkgError.RequestTimeoutError("QR code generation timeout")
	}

	return domainAccount.LoginResponse{}, pkgError.InternalServerError("Unknown error during login")
}

func (s *accountService) LoginAccountWithCode(ctx context.Context, accountID string, phoneNumber string) (string, error) {
	client := s.accountManager.GetClient(accountID)
	if client == nil {
		return "", pkgError.NotFoundError("Account not found or not initialized")
	}

	if client.IsLoggedIn() {
		return "", pkgError.BadRequestError("Account is already logged in")
	}

	// Connect client first before pairing
	if client.IsConnected() {
		client.Disconnect()
	}

	if err := client.Connect(); err != nil {
		return "", pkgError.InternalServerError(fmt.Sprintf("Failed to connect client: %v", err))
	}

	// Use PairPhone method with proper client type and display name
	code, err := client.PairPhone(ctx, phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		return "", pkgError.InternalServerError(fmt.Sprintf("Failed to request pairing code: %v", err))
	}

	return code, nil
}

func (s *accountService) LogoutAccount(ctx context.Context, accountID string) error {
	client := s.accountManager.GetClient(accountID)
	if client == nil {
		return pkgError.NotFoundError("Account not found or not initialized")
	}

	if !client.IsLoggedIn() {
		return pkgError.BadRequestError("Account is not logged in")
	}

	if err := client.Logout(ctx); err != nil {
		return pkgError.InternalServerError(fmt.Sprintf("Failed to logout: %v", err))
	}

	// Update account status
	account, _ := s.accountRepo.GetAccount(accountID)
	if account != nil {
		account.Status = domainAccount.StatusDisconnected
		s.accountRepo.UpdateAccount(account)
	}

	return nil
}

func (s *accountService) ReconnectAccount(ctx context.Context, accountID string) error {
	client := s.accountManager.GetClient(accountID)
	if client == nil {
		return pkgError.NotFoundError("Account not found or not initialized")
	}

	if !client.IsLoggedIn() {
		return pkgError.BadRequestError("Account is not logged in")
	}

	if client.IsConnected() {
		client.Disconnect()
	}

	if err := client.Connect(); err != nil {
		return pkgError.InternalServerError(fmt.Sprintf("Failed to reconnect: %v", err))
	}

	// Update account status
	account, _ := s.accountRepo.GetAccount(accountID)
	if account != nil {
		account.Status = domainAccount.StatusConnected
		account.LastConnected = time.Now()
		s.accountRepo.UpdateAccount(account)
	}

	return nil
}

func (s *accountService) SetAccountWebhook(ctx context.Context, accountID string, webhookURL string, secret string) error {
	// Check if account exists
	account, err := s.accountRepo.GetAccount(accountID)
	if err != nil {
		return pkgError.InternalServerError(fmt.Sprintf("Failed to get account: %v", err))
	}
	if account == nil {
		return pkgError.NotFoundError("Account not found")
	}

	if err := s.accountRepo.SetWebhook(accountID, webhookURL, secret); err != nil {
		return pkgError.InternalServerError(fmt.Sprintf("Failed to set webhook: %v", err))
	}

	return nil
}

func (s *accountService) GetAccountWebhook(ctx context.Context, accountID string) (domainAccount.WebhookInfo, error) {
	// Check if account exists
	account, err := s.accountRepo.GetAccount(accountID)
	if err != nil {
		return domainAccount.WebhookInfo{}, pkgError.InternalServerError(fmt.Sprintf("Failed to get account: %v", err))
	}
	if account == nil {
		return domainAccount.WebhookInfo{}, pkgError.NotFoundError("Account not found")
	}

	webhook, err := s.accountRepo.GetWebhook(accountID)
	if err != nil {
		return domainAccount.WebhookInfo{}, pkgError.InternalServerError(fmt.Sprintf("Failed to get webhook: %v", err))
	}

	return *webhook, nil
}