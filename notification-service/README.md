NOTIFICATION SERVICE - README PAGE
🏨 Hotel Notification Service

A production-ready, scalable notification microservice built in Go for hotel reservation systems. Handles email notifications for booking confirmations, welcome emails, payment receipts, and more.

 🚀 Features

- 📧 Email Delivery - Integrated with Resend for reliable email sending
- 🛡️ Rate Limiting - Configurable limits to prevent abuse (1 email/hour per user)
- ⚡ High Performance - Built with Go and Gin for optimal performance
- 🗄️ PostgreSQL - Persistent storage with proper data modeling
- 🔒 Security - Input validation, sanitization, and security headers
- 🔄 Resilience - Circuit breaker pattern and retry mechanisms
- 📊 Monitoring - Health checks, structured logging, and request tracing
- ☁️ Cloud Ready - Deployed on Railway with auto-scaling

 
 📦 API Endpoints
 Send Email
http
POST /api/v1/notifications/email
Content-Type: application/json

{
  "to": "guest@example.com",
  "subject": "Booking Confirmation",
  "body": "Your room has been booked successfully!",
  "type": "booking_confirmation"
}
Response:
json
{
  "id": "notif_123456789",
  "status": "sent",
  "message": "Email sent successfully",
  "timestamp": "2024-01-15T10:30:00Z"
}


🛠️ Tech Stack

  Language: Go 1.25.1

  Framework: Gin Web Framework

  Database: PostgreSQL

  Email: Resend API

  Deployment: Railway

  Monitoring: Structured logging, health checks

🚀 Quick Start
Prerequisites

  Go 1.21+

  PostgreSQL database

  Resend API account

Environment Variables
bash

# Database
DATABASE_URL=postgresql://user:pass@host:port/db

# Email (Resend)
RESEND_API_KEY=re_your_api_key
FROM_EMAIL=onboarding@resend.dev

# Rate Limiting
RATE_LIMIT_REQUESTS=1
RATE_LIMIT_MINUTES=60

# Server
PORT=8080
ENVIRONMENT=development

Local Development
bash

# Clone repository
git clone https://github.com/ollatomiwa/hotelsystem/notification-service.git
cd notification-service

# Install dependencies
go mod download

# Set environment variables
export RESEND_API_KEY=your_key
export DATABASE_URL=your_db_url

# Run the service
go run ./cmd/server

# Service available at http://localhost:8080

Deployment

The service is configured for easy deployment on Railway:

  Push to GitHub

  Connect repository to Railway

  Set environment variables

  Automatic deployment on push to main

📁 Project Structure
text

notification-service/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── handlers/                   # HTTP handlers
│   ├── services/                   # Business logic
│   ├── repositories/               # Data access layer
│   │   └── postgres/              # PostgreSQL implementation
│   └── models/                     # Data structures
├── pkg/
│   ├── email/                     # Email sending (Resend)
│   ├── config/                    # Configuration management
│   ├── ratelimiter/               # Rate limiting
│   ├── circuitbreaker/            # Circuit breaker pattern
│   ├── middleware/                # HTTP middleware
│   └── logging/                   # Structured logging
└── README.md

🧪 Testing
bash

# Run tests
go test ./...

# Test email sending
curl -X POST http://localhost:8080/api/v1/notifications/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "test@resend.dev",
    "subject": "Test Email",
    "body": "This is a test email from the notification service",
    "type": "test"
  }'

🔧 Configuration
Rate Limiting

  Default: 1 email per hour per recipient

  Configurable via environment variables

  Prevents abuse and manages API costs

Email Types

  booking_confirmation - Room booking confirmations

  welcome_email - New user welcome emails

  payment_receipt - Payment confirmation

  password_reset - Password reset emails

🚢 Deployment
Railway (Recommended)
bash

# The service is configured for zero-config deployment on Railway
# Just connect your GitHub repo and set environment variables

Docker
dockerfile

FROM golang:1.21-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o notification-service ./cmd/server
CMD ["./notification-service"]

📈 Monitoring
Health Endpoints

  /health - Detailed service health with dependency checks

  /ready - Simple readiness probe for load balancers

Logging

Structured JSON logging with:

  Request IDs for tracing

  Performance metrics

  Error tracking

  Audit trails

🤝 Contributing

  Fork the repository

  Create a feature branch (git checkout -b feature/amazing-feature)

  Commit your changes (git commit -m 'Add amazing feature')

  Push to the branch (git push origin feature/amazing-feature)

   Open a Pull Request
🙏 Acknowledgments

   Resend for excellent email API
   Railway for seamless deployment
  Gin Web Framework for high-performance HTTP
