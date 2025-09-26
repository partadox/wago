package whatsapp

import (
	domainAccount "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/account"
	infraAccount "github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/account"
	"go.mau.fi/whatsmeow"
)

// GetAccountIDFromClient returns the account ID for a given WhatsApp client
func GetAccountIDFromClient(client *whatsmeow.Client) string {
	if client == nil || client.Store == nil || client.Store.ID == nil {
		return ""
	}

	// Search through all managed clients to find matching account ID
	clients := infraAccount.GlobalAccountManager.ListClients()
	for accountID, managedClient := range clients {
		if managedClient == client {
			return accountID
		}
	}

	return ""
}

// GetAccountRepoFromGlobalVars gets account repository from global variables
// This is a temporary solution until we refactor to dependency injection
func GetAccountRepoFromGlobalVars() domainAccount.IAccountRepository {
	// This will be set from cmd/root.go via a setter function
	return globalAccountRepo
}

// Global variable to hold account repository reference
var globalAccountRepo domainAccount.IAccountRepository

// SetGlobalAccountRepo sets the global account repository
func SetGlobalAccountRepo(repo domainAccount.IAccountRepository) {
	globalAccountRepo = repo
}