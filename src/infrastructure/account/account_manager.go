package account

import (
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

type AccountManager struct {
	clients map[string]*whatsmeow.Client
	dbs     map[string]*sqlstore.Container
	mutex   sync.RWMutex
}

func NewAccountManager() *AccountManager {
	return &AccountManager{
		clients: make(map[string]*whatsmeow.Client),
		dbs:     make(map[string]*sqlstore.Container),
	}
}

func (am *AccountManager) GetClient(accountID string) *whatsmeow.Client {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.clients[accountID]
}

func (am *AccountManager) SetClient(accountID string, client *whatsmeow.Client, db *sqlstore.Container) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.clients[accountID] = client
	am.dbs[accountID] = db
}

func (am *AccountManager) RemoveClient(accountID string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	if client, exists := am.clients[accountID]; exists {
		client.Disconnect()
		delete(am.clients, accountID)
		delete(am.dbs, accountID)
	}
}

func (am *AccountManager) ListClients() map[string]*whatsmeow.Client {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	result := make(map[string]*whatsmeow.Client)
	for k, v := range am.clients {
		result[k] = v
	}
	return result
}

func (am *AccountManager) GetDB(accountID string) *sqlstore.Container {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.dbs[accountID]
}

// Global account manager instance
var GlobalAccountManager = NewAccountManager()