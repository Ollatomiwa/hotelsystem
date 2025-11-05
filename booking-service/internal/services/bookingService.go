package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/repositories"
	"github.com/ollatomiwa/hotelsystem/booking-service/pkg/notifications"
)
type BookingService struct {
	bookingRepo repositories.BookingRepository
	roomRepo    repositories.RoomRepository
	notifyClient *notifications.Client
	notificationsEnabled bool
}

// Change to accept interfaces
func NewBookingService(bookingRepo repositories.BookingRepository, roomRepo repositories.RoomRepository,notifyClient *notifications.Client,
	notificationsEnabled bool, ) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
		notifyClient: notifyClient,
		notificationsEnabled: notificationsEnabled,
	}
}

// Check room availability for a given criteria
func (s *BookingService) CheckAvailability(ctx context.Context, req *models.AvailabilityRequest) (*models.AvailabilityResponse, error) {
	// Validating dates
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, fmt.Errorf("invalid check_in date: %w", err)
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut) // FIXED: was req.CheckIn
	if err != nil {
		return nil, fmt.Errorf("invalid check_out date: %w", err)
	}

	if checkIn.Before(time.Now().AddDate(0, 0, -1)) {
		return nil, fmt.Errorf("check_in date cannot be in the past")
	}
	if checkOut.Before(checkIn) || checkOut.Equal(checkIn) {
		return nil, fmt.Errorf("check_out date must be after check_in date")
	}

	// Get available rooms
	availableRooms, err := s.bookingRepo.GetAvailableRooms(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}

	return &models.AvailabilityResponse{
		AvailableRooms: availableRooms,
		TotalAvailable: len(availableRooms),
	}, nil
}

// Creates a new booking
func (s *BookingService) CreateBooking(ctx context.Context, req *models.BookingRequest) (*models.Booking, error) {
	// Validating dates
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, fmt.Errorf("invalid check_in date: %w", err)
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut) // FIXED: was req.CheckIn
	if err != nil {
		return nil, fmt.Errorf("invalid check_out date: %w", err)
	}
	if checkIn.Before(time.Now().AddDate(0, 0, -1)) {
		return nil, fmt.Errorf("check_in date cannot be in the past")
	}
	if checkOut.Before(checkIn.AddDate(0, 0, 1)) {
		return nil, fmt.Errorf("minimum stay is 1 night")
	}

	// Get room details
	room, err := s.roomRepo.GetRoomById(ctx, req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// Validating guests count
	if req.Guests > room.MaxGuests {
		return nil, fmt.Errorf("room can only accommodate %d guests", room.MaxGuests)
	}

	if !room.Available {
		return nil, fmt.Errorf("room is not available")
	}

	// Calculate total amount
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	totalAmount := room.PricePerNight * float64(nights)

	// Create booking
	booking := &models.Booking{
		Id:          uuid.New().String(),
		UserId:      req.UserId,
		RoomId:      req.RoomId,
		RoomType:    room.RoomType,
		CheckIn:     checkIn,
		CheckOut:    checkOut,
        Guest:      req.Guests, 
		TotalAmount: totalAmount,
		Status:      models.StatusConfirmed,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save into db
	err = s.bookingRepo.CreateBooking(ctx, booking)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Send notification (async - don't block the response)
	if s.notificationsEnabled {
		go s.sendBookingConfirmation(context.Background(), booking, room, req.UserEmail)
	}	

	return booking, nil
}

// Retrieve a booking by Id
func (s *BookingService) GetBooking(ctx context.Context, id string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetBookingById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	return booking, nil
}

// Retrieve booking for a user
func (s *BookingService) GetUserBooking(ctx context.Context, userId string) ([]models.Booking, error) {
	bookings, err := s.bookingRepo.GetUserBookings(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %w", err)
	}
	return bookings, nil
}

// Cancel a booking
func (s *BookingService) CancelBooking(ctx context.Context, id string) error {
	// Get first to check if it can be canceled
	booking, err := s.bookingRepo.GetBookingById(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}

	// Check if it can be canceled which will be only before 24hrs before checkin
	if time.Until(booking.CheckIn) < 24*time.Hour {
		return fmt.Errorf("booking can only be canceled at least 24 hours before check in")
	}

	// Get room details for notification
	room, err := s.roomRepo.GetRoomById(ctx, booking.RoomId)
	if err != nil {
		return fmt.Errorf("failed to get room details: %w", err)
	}

	// Now we updates status
	err = s.bookingRepo.UpdateBookingStatus(ctx, id, models.StatusCancelled)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	// Send cancellation notification (async)
	if s.notificationsEnabled {
		go s.sendBookingCancellation(context.Background(), booking, room)
	}

	return nil
}

// sendBookingConfirmation sends a confirmation notification
func (s *BookingService) sendBookingConfirmation(ctx context.Context, booking *models.Booking, room *models.Room, userEmail string) {
	bookingData := map[string]interface{}{
		"booking_id":      booking.Id,
		"room_number":     room.RoomNumber,
		"room_type":       string(room.RoomType),
		"check_in":        booking.CheckIn.Format("2006-01-02"),
		"check_out":       booking.CheckOut.Format("2006-01-02"),
		"total_amount":    booking.TotalAmount,
		"guests":          booking.Guest,
		"booking_date":    booking.CreatedAt.Format("2006-01-02"),
	}

	if err := s.notifyClient.SendBookingConfirmation(ctx, userEmail, bookingData); err != nil {
		// Log error but don't fail the booking
		fmt.Printf("Failed to send booking confirmation: %v\n", err)
	}
}

// sendBookingCancellation sends a cancellation notification
func (s *BookingService) sendBookingCancellation(ctx context.Context, booking *models.Booking, room *models.Room) {
	bookingData := map[string]interface{}{
		"booking_id":      booking.Id,
		"room_number":     room.RoomNumber,
		"room_type":       string(room.RoomType),
		"check_in":        booking.CheckIn.Format("2006-01-02"),
		"check_out":       booking.CheckOut.Format("2006-01-02"),
		"total_amount":    booking.TotalAmount,
		"cancellation_date": time.Now().Format("2006-01-02"),
	}

	if err := s.notifyClient.SendBookingCancellation(ctx, booking.UserId, bookingData); err != nil {
		// Log error but don't fail the cancellation
		fmt.Printf("Failed to send booking cancellation: %v\n", err)
	}
}