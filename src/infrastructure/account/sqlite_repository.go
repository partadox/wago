package account

import (
	"database/sql"

	domainAccount "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/account"
	"github.com/sirupsen/logrus"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) domainAccount.IAccountRepository {
	repo := &SQLiteRepository{db: db}
	repo.initTables()
	return repo
}

func (r *SQLiteRepository) initTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			status TEXT DEFAULT 'disconnected',
			phone_number TEXT DEFAULT '',
			device_id TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_connected DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS account_webhooks (
			account_id TEXT PRIMARY KEY,
			webhook_url TEXT DEFAULT '',
			webhook_secret TEXT DEFAULT '',
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(query); err != nil {
			logrus.Errorf("Failed to create table: %v", err)
		}
	}
}

func (r *SQLiteRepository) CreateAccount(account *domainAccount.Account) error {
	query := `INSERT INTO accounts (id, status, phone_number, device_id, created_at, last_connected)
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, account.ID, account.Status, account.PhoneNumber,
					   account.DeviceID, account.CreatedAt, account.LastConnected)

	if err != nil {
		return err
	}

	// Create webhook entry
	webhookQuery := `INSERT INTO account_webhooks (account_id, webhook_url, webhook_secret) VALUES (?, '', '')`
	_, err = r.db.Exec(webhookQuery, account.ID)
	return err
}

func (r *SQLiteRepository) GetAccount(accountID string) (*domainAccount.Account, error) {
	query := `SELECT id, status, phone_number, device_id, created_at, last_connected
			  FROM accounts WHERE id = ?`

	account := &domainAccount.Account{}
	err := r.db.QueryRow(query, accountID).Scan(
		&account.ID, &account.Status, &account.PhoneNumber,
		&account.DeviceID, &account.CreatedAt, &account.LastConnected,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return account, err
}

func (r *SQLiteRepository) UpdateAccount(account *domainAccount.Account) error {
	query := `UPDATE accounts SET status = ?, phone_number = ?, device_id = ?, last_connected = ?
			  WHERE id = ?`

	_, err := r.db.Exec(query, account.Status, account.PhoneNumber,
					   account.DeviceID, account.LastConnected, account.ID)
	return err
}

func (r *SQLiteRepository) DeleteAccount(accountID string) error {
	query := `DELETE FROM accounts WHERE id = ?`
	_, err := r.db.Exec(query, accountID)
	return err
}

func (r *SQLiteRepository) ListAccounts() ([]*domainAccount.Account, error) {
	query := `SELECT id, status, phone_number, device_id, created_at, last_connected
			  FROM accounts ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*domainAccount.Account
	for rows.Next() {
		account := &domainAccount.Account{}
		err := rows.Scan(&account.ID, &account.Status, &account.PhoneNumber,
						 &account.DeviceID, &account.CreatedAt, &account.LastConnected)
		if err != nil {
			continue
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *SQLiteRepository) SetWebhook(accountID string, webhookURL string, secret string) error {
	query := `UPDATE account_webhooks SET webhook_url = ?, webhook_secret = ? WHERE account_id = ?`
	_, err := r.db.Exec(query, webhookURL, secret, accountID)
	return err
}

func (r *SQLiteRepository) GetWebhook(accountID string) (*domainAccount.WebhookInfo, error) {
	query := `SELECT webhook_url, webhook_secret FROM account_webhooks WHERE account_id = ?`

	webhook := &domainAccount.WebhookInfo{}
	err := r.db.QueryRow(query, accountID).Scan(&webhook.URL, &webhook.Secret)

	if err == sql.ErrNoRows {
		return &domainAccount.WebhookInfo{}, nil
	}

	return webhook, err
}