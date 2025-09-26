# Multi-Account WhatsApp Gateway Guide

## Overview

Proyek ini telah dimodifikasi untuk mendukung multiple WhatsApp accounts dengan webhook per account yang dapat dikonfigurasi secara terpisah.

## Fitur Multi-Account

### 1. **Account Management**
- Buat, hapus, dan kelola multiple WhatsApp accounts
- Setiap account memiliki ID unik
- Status connection per account (connected, disconnected, logged_in)
- Database terpisah untuk setiap account

### 2. **Webhook Per Account**
- Setiap account dapat memiliki webhook URL dan secret sendiri
- Webhook dapat dikonfigurasi independen untuk setiap account
- Backward compatibility dengan global webhook

### 3. **API Endpoints Baru**

#### Account Management
```bash
# Buat account baru
POST /accounts
{
  "account_id": "account1"
}

# List semua accounts
GET /accounts

# Get detail account
GET /accounts/{accountId}

# Delete account
DELETE /accounts/{accountId}

# Login account (QR code)
POST /accounts/{accountId}/login

# Login dengan phone number (pairing code)
POST /accounts/{accountId}/login-with-code
{
  "phone_number": "6281234567890"
}

# Logout account
POST /accounts/{accountId}/logout

# Reconnect account
POST /accounts/{accountId}/reconnect
```

#### Webhook Management
```bash
# Set webhook untuk account
POST /accounts/{accountId}/webhook
{
  "webhook_url": "https://yourdomain.com/webhook",
  "secret": "your-secret-key"
}

# Get webhook account
GET /accounts/{accountId}/webhook
```

### 4. **Modifikasi Send API**

Semua endpoint send sekarang memerlukan `account_id` dalam request body:

```bash
# Send message
POST /send/message
{
  "account_id": "account1",
  "phone": "6281234567890",
  "message": "Hello from account1"
}

# Send image
POST /send/image
{
  "account_id": "account1",
  "phone": "6281234567890",
  "caption": "Image from account1"
}

# Dan seterusnya untuk semua endpoint send...
```

## Cara Penggunaan

### 1. **Setup Multi-Account**

```bash
# 1. Buat account pertama
curl -X POST http://localhost:3000/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": "whatsapp_bisnis"}'

# 2. Buat account kedua
curl -X POST http://localhost:3000/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": "whatsapp_personal"}'

# 3. Setup webhook untuk account bisnis
curl -X POST http://localhost:3000/accounts/whatsapp_bisnis/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://yourbusiness.com/webhook",
    "secret": "business-secret"
  }'

# 4. Setup webhook untuk account personal
curl -X POST http://localhost:3000/accounts/whatsapp_personal/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://yourpersonal.com/webhook",
    "secret": "personal-secret"
  }'
```

### 2. **Login Accounts**

```bash
# Login account bisnis dengan QR
curl -X POST http://localhost:3000/accounts/whatsapp_bisnis/login

# Login account personal dengan phone number
curl -X POST http://localhost:3000/accounts/whatsapp_personal/login-with-code \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "6281234567890"}'
```

### 3. **Send Messages**

```bash
# Send dari account bisnis
curl -X POST http://localhost:3000/send/message \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": "whatsapp_bisnis",
    "phone": "6281234567890",
    "message": "Hello from Business Account"
  }'

# Send dari account personal
curl -X POST http://localhost:3000/send/message \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": "whatsapp_personal",
    "phone": "6281234567890",
    "message": "Hello from Personal Account"
  }'
```

## Database Structure

### Accounts Table
```sql
CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    status TEXT DEFAULT 'disconnected',
    phone_number TEXT DEFAULT '',
    device_id TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_connected DATETIME
);
```

### Account Webhooks Table
```sql
CREATE TABLE account_webhooks (
    account_id TEXT PRIMARY KEY,
    webhook_url TEXT DEFAULT '',
    webhook_secret TEXT DEFAULT '',
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);
```

## File Structure Changes

### Komponen Baru:
- `src/domains/account/` - Domain logic untuk account management
- `src/infrastructure/account/` - Account repository dan manager
- `src/usecase/account.go` - Business logic untuk account operations
- `src/ui/rest/account.go` - REST API endpoints untuk account
- `src/infrastructure/whatsapp/account_helper.go` - Helper functions

### Komponen Yang Dimodifikasi:
- `src/domains/send/base.go` - Tambah AccountID ke BaseRequest
- `src/usecase/send.go` - Gunakan client dari account manager
- `src/infrastructure/whatsapp/webhook.go` - Support webhook per account
- `src/cmd/root.go` - Initialize account services
- `src/cmd/rest.go` - Register account endpoints

## Benefits

1. **Isolasi Account**: Setiap WhatsApp account bekerja independen
2. **Webhook Flexibility**: Setiap account bisa punya webhook berbeda
3. **Scalability**: Mudah menambah account baru tanpa restart
4. **Backward Compatibility**: API lama tetap berfungsi (dengan tambahan account_id)
5. **Resource Efficiency**: Sharing infrastruktur tapi data terpisah

## Resource Requirements & Capacity Planning

### Estimasi Resource Per Account

#### Memory Usage:
- **Per Account**: ~50-100 MB RAM untuk setiap WhatsApp client aktif
- **Base Application**: ~100-200 MB untuk aplikasi inti
- **Example**: 10 accounts = ~700MB - 1.2GB total RAM

#### Storage Requirements:
- **Database per Account**: ~10-50 MB per account (tergantung history chat)
- **Media Storage**: Bervariasi tergantung penggunaan media
- **QR Code Files**: ~5-10 KB per login attempt
- **Example**: 10 accounts = ~100-500 MB database storage

#### CPU Usage:
- **Per Account**: Minimal saat idle, meningkat saat ada aktivitas tinggi
- **Concurrent Connections**: Linear scaling dengan jumlah account aktif
- **Example**: 10 accounts dapat berjalan di 2-4 CPU cores

### Estimasi Kapasitas Berdasarkan Hardware

#### Server Kecil (2GB RAM, 2 CPU cores):
- **Recommended**: 5-10 accounts
- **Maximum**: 15-20 accounts (dengan optimasi)

#### Server Medium (4GB RAM, 4 CPU cores):
- **Recommended**: 20-30 accounts
- **Maximum**: 40-50 accounts

#### Server Besar (8GB RAM, 8 CPU cores):
- **Recommended**: 50-80 accounts
- **Maximum**: 100+ accounts

### Monitoring & Optimization Tips

1. **Memory Monitoring**:
   ```bash
   # Monitor memory usage
   docker stats
   ```

2. **Database Size Monitoring**:
   ```bash
   # Check database sizes
   du -sh storages/whatsapp_*.db
   ```

3. **Connection Health Check**:
   ```bash
   # Check account status
   curl http://localhost:3000/accounts
   ```

4. **Optimization Strategies**:
   - Gunakan database cleanup untuk old messages
   - Implement connection pooling untuk inactive accounts
   - Monitor dan restart accounts yang sering disconnect

## Limitations

- Setiap account tetap memerlukan QR scan atau pairing code terpisah
- Database file terpisah per account (storage requirement meningkat)
- Memory usage meningkat karena multiple clients
- WhatsApp rate limiting berlaku per account
- Maksimal 4 devices per WhatsApp number (termasuk phone)

## Migration dari Single Account

Jika Anda memiliki implementasi single account sebelumnya:

1. Buat account dengan ID "default": `POST /accounts {"account_id": "default"}`
2. Login account default
3. Update semua API calls untuk menambahkan `"account_id": "default"`
4. Pindahkan webhook configuration ke account level jika diperlukan