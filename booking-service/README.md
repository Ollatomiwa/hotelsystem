
### Booking Service
- ✅ Real-time room availability checking
- ✅ Booking creation with validation
- ✅ Date conflict prevention
- ✅ Automatic price calculation
- ✅ Booking cancellation
- ✅ User booking history

### Integration
- ✅ Microservices architecture
- ✅ PostgreSQL database
- ✅ Automatic email notifications
- ✅ Railway deployment
- ✅ Production-ready error handling

## Start Docker Integration
docker run -d --name booking-postgres \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=booking_service \
  postgres:15-alpine

## Run the booking service
    go run cmd/server/main.go

# Check health
curl http://localhost:8080/health

# Check availability
curl -X POST http://localhost:8080/api/v1/bookings/availability \
  -H "Content-Type: application/json" \
  -d '{
    "room_type": "double",
    "check_in": "2024-12-15",
    "check_out": "2024-12-20",
    "guests": 2
  }'
Database Schema
Rooms Table
sql

CREATE TABLE rooms (
    id TEXT PRIMARY KEY,
    room_number TEXT UNIQUE NOT NULL,
    room_type TEXT NOT NULL CHECK (room_type IN ('single', 'double', 'suite', 'deluxe')),
    price_per_night DECIMAL(10,2) NOT NULL,
    max_guests INTEGER NOT NULL CHECK (max_guests > 0),
    available BOOLEAN DEFAULT TRUE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

Bookings Table
sql

CREATE TABLE bookings (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    room_type TEXT NOT NULL CHECK (room_type IN ('single', 'double', 'suite', 'deluxe')),
    check_in TIMESTAMPTZ NOT NULL,
    check_out TIMESTAMPTZ NOT NULL,
    guests INTEGER NOT NULL CHECK (guests > 0),
    total_amount DECIMAL(10,2) NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT valid_dates CHECK (check_out > check_in)
);

🐳 Docker Deployment
Build and run with Docker
bash

docker build -t booking-service .
docker run -p 8080:8080 booking-service

Docker Compose (Full stack)
bash
docker-compose up -d

🧪 Testing

Run the test suite:
bash
# Unit tests
go test ./internal/services/...
# Integration tests  
go test ./tests/integration/...

🌐 Deployment
Railway Deployment
    Connect your GitHub repository to Railway
    Set environment variables in Railway dashboard
    Deploy automatically on git push

Environment Variables
env

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=booking_service
DB_SSL_MODE=disable

# Notification Service
NOTIFICATION_SERVICE_URL=https://notification-services.up.railway.app
NOTIFICATIONS_ENABLED=true

# Server
PORT=8080
ENV=development

🤝 Contributing
    Fork the repository
    Create a feature branch (git checkout -b feature/amazing-feature)
    Commit your changes (git commit -m 'Add some amazing feature')
    Push to the branch (git push origin feature/amazing-feature)
    Open a Pull Request

📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
👥 Authors
    OLlatomiwa

🙏 Acknowledgments
    Gin Web Framework
    PostgreSQL
    Railway for deployment
    Render for email services
