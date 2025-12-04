# Payment Backend Service

A production-ready payment backend service built with Go, Gin framework, Paystack API, and SQLite3.

## Features

- ✅ Clean architecture with separation of concerns
- ✅ Paystack payment integration
- ✅ RESTful API with comprehensive endpoints
- ✅ Webhook handling with signature verification
- ✅ SQLite3 database with GORM ORM
- ✅ Request logging and monitoring
- ✅ Rate limiting and security middleware
- ✅ API key authentication
- ✅ Idempotency support
- ✅ Graceful shutdown
- ✅ Docker support
- ✅ Comprehensive error handling

## Project Structure

```
payment-service/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Configuration management
├── internal/
│   ├── handlers/
│   │   ├── payment_handler.go   # HTTP handlers
│   │   └── health_handler.go
│   ├── middleware/
│   │   └── middleware.go        # HTTP middlewares
│   ├── models/
│   │   └── models.go            # Data models
│   ├── repository/
│   │   └── repository.go        # Data access layer
│   ├── router/
│   │   └── router.go            # Route configuration
│   └── service/
│       └── payment_service.go   # Business logic
├── pkg/
│   ├── database/
│   │   └── database.go          # Database connection
│   └── paystack/
│       ├── client.go            # Paystack API client
│       └── webhook.go           # Webhook validation
├── .env.example                 # Environment variables template
├── docker-compose.yml           # Docker Compose configuration
├── Dockerfile                   # Docker configuration
├── go.mod                       # Go module file
├── Makefile                     # Build commands
└── README.md                    # This file
```

## Prerequisites

- Go 1.21 or higher
- SQLite3
- Paystack account with API keys
- Docker (optional)

## Installation

### Local Setup

1. **Clone the repository**
```bash
git clone <repository-url>
cd payment-service
```

2. **Install dependencies**
```bash
go mod download
```

3. **Configure environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Set up your environment variables in `.env`:**
```env
PORT=8080
ENVIRONMENT=development
DATABASE_PATH=./data/payment.db
PAYSTACK_SECRET_KEY=sk_test_your_secret_key
PAYSTACK_PUBLIC_KEY=pk_test_your_public_key
API_KEY=your_secure_api_key
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60
```

5. **Run the application**
```bash
make run
# or
go run cmd/server/main.go
```

### Docker Setup

1. **Build and run with Docker Compose**
```bash
docker-compose up -d
```

2. **View logs**
```bash
docker-compose logs -f
```

3. **Stop containers**
```bash
docker-compose down
```

## API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication

All endpoints (except health check and webhooks) require API key authentication. Include your API key in the request header:

```
X-API-Key: your_api_key
```

Or use Bearer token:

```
Authorization: Bearer your_api_key
```

### Endpoints

#### 1. Health Check
Check service health status.

```http
GET /api/v1/health
```

**Response:**
```json
{
  "status": "success",
  "message": "Health check completed",
  "data": {
    "status": "healthy",
    "database": "ok"
  }
}
```

#### 2. Initialize Payment
Initialize a new payment transaction.

```http
POST /api/v1/payments/initialize
Content-Type: application/json
X-API-Key: your_api_key
Idempotency-Key: unique_key_123 (optional)
```

**Request Body:**
```json
{
  "email": "customer@example.com",
  "amount": 50000,
  "currency": "NGN",
  "callback_url": "https://yoursite.com/callback",
  "metadata": {
    "order_id": "12345",
    "product_name": "Premium Plan"
  }
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Payment initialized successfully",
  "data": {
    "id": 1,
    "reference": "TXN_abc123",
    "amount": 50000,
    "currency": "NGN",
    "status": "pending",
    "customer_email": "customer@example.com",
    "authorization_url": "https://checkout.paystack.com/abc123",
    "access_code": "abc123xyz",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 3. Verify Payment
Verify a payment transaction.

```http
GET /api/v1/payments/verify/:reference
X-API-Key: your_api_key
```

**Response:**
```json
{
  "status": "success",
  "message": "Payment verified successfully",
  "data": {
    "id": 1,
    "reference": "TXN_abc123",
    "amount": 50000,
    "status": "success",
    "customer_email": "customer@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:05:00Z"
  }
}
```

#### 4. Get Payment Details
Get details of a specific payment.

```http
GET /api/v1/payments/:id
X-API-Key: your_api_key
```

#### 5. List Payments
List all payments with pagination and filtering.

```http
GET /api/v1/payments?page=1&page_size=20&status=success&email=customer@example.com
X-API-Key: your_api_key
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20, max: 100)
- `status` (optional): Filter by status (pending, success, failed, abandoned)
- `email` (optional): Filter by customer email

**Response:**
```json
{
  "status": "success",
  "message": "Payments retrieved successfully",
  "data": [...],
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

#### 6. Get Customer
Get customer details by email.

```http
GET /api/v1/customers/:email
X-API-Key: your_api_key
```

#### 7. Paystack Webhook
Handle Paystack webhook events.

```http
POST /api/v1/webhooks/paystack
Content-Type: application/json
X-Paystack-Signature: signature_from_paystack
```

**Note:** This endpoint validates the webhook signature automatically. Configure your webhook URL in your Paystack dashboard.

## Webhook Setup

1. **Log in to your Paystack dashboard**
2. **Navigate to Settings > Webhooks**
3. **Add your webhook URL:**
   ```
   https://your-domain.com/api/v1/webhooks/paystack
   ```
4. **Select events to listen to:**
   - charge.success
   - charge.failed

## Idempotency

To prevent duplicate payments, include an `Idempotency-Key` header in your payment initialization requests:

```http
POST /api/v1/payments/initialize
Idempotency-Key: unique-key-for-this-request
```

If you retry with the same key, you'll receive the original transaction instead of creating a duplicate.

## Error Handling

All errors follow the standard response format:

```json
{
  "status": "error",
  "message": "Error description"
}
```

**Common HTTP Status Codes:**
- `200` - Success
- `400` - Bad Request (validation error)
- `401` - Unauthorized (invalid API key)
- `404` - Not Found
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

## Rate Limiting

Default rate limits:
- 100 requests per 60 seconds per IP address

Exceeding the limit returns `429 Too Many Requests`.

## Testing

### Run tests
```bash
make test
```

### Run tests with coverage
```bash
make test-coverage
```

### Example test with cURL

**Initialize Payment:**
```bash
curl -X POST http://localhost:8080/api/v1/payments/initialize \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_api_key" \
  -d '{
    "email": "test@example.com",
    "amount": 50000,
    "currency": "NGN"
  }'
```

**Verify Payment:**
```bash
curl -X GET http://localhost:8080/api/v1/payments/verify/TXN_abc123 \
  -H "X-API-Key: your_api_key"
```

## Postman Collection

Import this sample collection to test the API:

```json
{
  "info": {
    "name": "Payment Service API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Initialize Payment",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "X-API-Key",
            "value": "{{api_key}}"
          },
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"email\": \"test@example.com\",\n  \"amount\": 50000,\n  \"currency\": \"NGN\"\n}"
        },
        "url": {
          "raw": "{{base_url}}/payments/initialize",
          "host": ["{{base_url}}"],
          "path": ["payments", "initialize"]
        }
      }
    }
  ],
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080/api/v1"
    },
    {
      "key": "api_key",
      "value": "your_api_key"
    }
  ]
}
```

## Production Deployment

### Environment Variables for Production

```env
ENVIRONMENT=production
DATABASE_PATH=/var/lib/payment-service/payment.db
PAYSTACK_SECRET_KEY=sk_live_your_live_secret_key
PAYSTACK_PUBLIC_KEY=pk_live_your_live_public_key
API_KEY=strong_random_api_key
```

### Security Recommendations

1. **Use HTTPS in production** - Deploy behind a reverse proxy (nginx, Caddy)
2. **Strong API keys** - Generate cryptographically secure API keys
3. **Database backups** - Regular automated backups of SQLite database
4. **Rate limiting** - Adjust based on your traffic patterns
5. **Monitoring** - Set up logging and monitoring (Prometheus, Grafana)
6. **Environment variables** - Never commit secrets to version control

### HTTPS Configuration (Nginx Example)

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Troubleshooting

### Database locked error
If you get "database locked" errors:
```bash
# Check for other processes using the database
lsof | grep payment.db

# Increase busy timeout in database.go
```

### Paystack connection errors
- Verify your API keys are correct
- Check your internet connection
- Ensure Paystack API is accessible (check status.paystack.com)

### Rate limit issues
Adjust rate limits in `.env`:
```env
RATE_LIMIT_REQUESTS=200
RATE_LIMIT_WINDOW=60
```

## Development

### Available Make Commands

```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run the application
make test          # Run tests
make clean         # Clean build artifacts
make docker-build  # Build Docker image
make docker-up     # Start with Docker Compose
make lint          # Run linter
make fmt           # Format code
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions:
- Create an issue in the repository
- Check existing documentation
- Review Paystack documentation: https://paystack.com/docs

## Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Paystack](https://paystack.com/)
- [Logrus](https://github.com/sirupsen/logrus)