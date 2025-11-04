package repositories

import (
	"context"
	

	"github.com/ollatomiwa/hotelsystem/booking-service/internal/models"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *models.Booking) error
	GetAvailableRooms(ctx context.Context, req *models.AvailabilityRequest) ([]models.RoomAvailability, error)
	GetBookingById(ctx context.Context, id string) (*models.Booking, error)
	GetUserBookings(ctx context.Context, userId string) ([]models.Booking, error)
	UpdateBookingStatus(ctx context.Context, id string, status models.BookingStatus) error
}

type RoomRepository interface {
	GetRoomById(ctx context.Context, id string) (*models.Room, error)
	CreateRoom(ctx context.Context, room *models.Room) error
	GetAllRooms(ctx context.Context) ([]models.Room, error)
}
