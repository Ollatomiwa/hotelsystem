package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/repositories/postgres"
)

type BookingService struct {
	bookingRepo *postgres.BookingRepository
	roomRepo *postgres.RoomRepository
}

func NewBookingService(bookingRepo *postgres.BookingRepository, roomRepo *postgres.RoomRepository) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		roomRepo: roomRepo,
	}
}

//check room availability for a given criteria
func (s *BookingService) CheckAvailability(ctx context.Context, req *models.AvailabilityRequest) (*models.AvailabilityResponse, error) {
	//validating dates
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, fmt.Errorf("invalid check_in date: %w", err)
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, fmt.Errorf("invalid check_out date: %w", err)
	}
	
	if checkIn.Before(time.Now().AddDate(0,0,-1)) {
		return nil, fmt.Errorf("check_in date cannot be in the past")
	}
	if checkOut.Before(checkIn) || checkOut.Equal(checkIn) {
		return nil, fmt.Errorf("check_out date must be after check_in date")
	}

	//get available rooms
	availableRooms, err := s.bookingRepo.GetAvailableRooms(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}

	return &models.AvailabilityResponse{
		AvailableRooms: availableRooms,
		TotalAvailable: len(availableRooms),
	}, nil 
}

//creates a new booking
func (s *BookingService) CreateBooking(ctx context.Context, req *models.BookingRequest) (*models.Booking, error) {
	//validating dates
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, fmt.Errorf("invalid check_in date: %w", err)
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, fmt.Errorf("invalid check_out date: %w", err)
	}
	if checkIn.Before(time.Now().AddDate(0,0,-1)) {
		return nil, fmt.Errorf("check_in date cannot be in the past")
	}
	if checkOut.Before(checkIn.AddDate(0,0,1))  {
		return nil, fmt.Errorf("minimum stay is 1 night")
	}

	//get room details
	room, err := s.roomRepo.GetRoomById(ctx, req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	//validating guests count
	if req.Guests > room.MaxGuests {
		return nil, fmt.Errorf("room can only accomodates %d guests", room.MaxGuests)
	}

	if !room.Available {
		return nil, fmt.Errorf("room is not available")
	}

	//calculate total amount
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	totalAmount := room.PricePerNight *float64(nights)

	//create booking
	booking := &models.Booking {
		Id: uuid.New().String(),
		UserId: req.UserId,
		RoomId: req.RoomId,
		RoomType: room.RoomType,
		CheckIn: checkIn,
		CheckOut: checkOut,
		TotalAmount: totalAmount,
		Status: models.StatusConfirmed,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	//save into db
	err = s.bookingRepo.CreateBooking(ctx, booking)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	return booking, nil
}

//retreive a booking by Id
func (s *BookingService) GetBooking(ctx context.Context, id string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetBookingById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	return booking, nil
}

//retrieve booking for a user
func (s *BookingService) GetUserBooking(ctx context.Context, userId string) ([]models.Booking, error) {
	bookings, err:= s.bookingRepo.GetUserBookings(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %w", err)
	}
	return bookings, nil 
}

//cancek a booking
func (s *BookingService) CancelBooking(ctx context.Context, id string) error {
	//get first to check if it can be canceled
	booking, err := s.bookingRepo.GetBookingById(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}
	
	//check if it can be canceled which will be only before 24hrs before checkin
	if time.Until(booking.CheckIn) <24*time.Hour {
		return fmt.Errorf("booking can only be canceled at least 24 hours before check in")
	}

	//now we updates status
	err = s.bookingRepo.UpdateBookingStatus(ctx, id, models.StatusCancelled)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}
	return nil
}


