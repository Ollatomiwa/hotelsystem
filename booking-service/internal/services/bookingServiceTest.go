package services

import (
	"context"
	"testing"
	"time"

	"github.com/ollatomiwa/hotelsystem/booking-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBookingRepository matches your postgres.BookingRepository
type MockBookingRepository struct {
	mock.Mock
}

func (m *MockBookingRepository) CreateBooking(ctx context.Context, booking *models.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockBookingRepository) GetAvailableRooms(ctx context.Context, req *models.AvailabilityRequest) ([]models.RoomAvailability, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]models.RoomAvailability), args.Error(1)
}

func (m *MockBookingRepository) GetBookingById(ctx context.Context, id string) (*models.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Booking), args.Error(1)
}

func (m *MockBookingRepository) GetUserBookings(ctx context.Context, userId string) ([]models.Booking, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]models.Booking), args.Error(1)
}

func (m *MockBookingRepository) UpdateBookingStatus(ctx context.Context, id string, status models.BookingStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// MockRoomRepository matches your postgres.RoomRepository  
type MockRoomRepository struct {
	mock.Mock
}

func (m *MockRoomRepository) GetRoomById(ctx context.Context, id string) (*models.Room, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Room), args.Error(1)
}

func (m *MockRoomRepository) CreateRoom(ctx context.Context, room *models.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) GetAllRooms(ctx context.Context) ([]models.Room, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Room), args.Error(1)
}

func TestBookingService_CheckAvailability_Success(t *testing.T) {
	// Create mocks
	mockBookingRepo := new(MockBookingRepository)
	mockRoomRepo := new(MockRoomRepository)
	
	// Create service with mocks
	service := &BookingService{
		bookingRepo: mockBookingRepo,
		roomRepo:    mockRoomRepo,
	}

	ctx := context.Background()
	req := &models.AvailabilityRequest{
		RoomType: models.RoomTypeDouble,
		CheckIn:  "2024-01-15",
		CheckOut: "2024-01-20", 
		Guests:   2,
	}

	expectedRooms := []models.RoomAvailability{
		{
			RoomId:        "room-1",
			RoomNumber:    "101",
			RoomType:      models.RoomTypeDouble,
			PricePerNight: 150.0,
			TotalPrice:    750.0,
			MaxGuests:     2,
		},
	}

	// Setup mock expectation
	mockBookingRepo.On("GetAvailableRooms", ctx, req).Return(expectedRooms, nil)

	// Call the method
	response, err := service.CheckAvailability(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.TotalAvailable)
	assert.Equal(t, expectedRooms, response.AvailableRooms)
	
	// Verify mock was called
	mockBookingRepo.AssertExpectations(t)
}

func TestBookingService_CheckAvailability_InvalidDates(t *testing.T) {
	mockBookingRepo := new(MockBookingRepository)
	mockRoomRepo := new(MockRoomRepository)
	
	service := &BookingService{
		bookingRepo: mockBookingRepo,
		roomRepo:    mockRoomRepo,
	}

	ctx := context.Background()
	
	// Test invalid dates (check-out before check-in)
	req := &models.AvailabilityRequest{
		RoomType: models.RoomTypeDouble,
		CheckIn:  "2024-01-20", // Later date
		CheckOut: "2024-01-15", // Earlier date
		Guests:   2,
	}

	response, err := service.CheckAvailability(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "check_out date must be after check_in date")
}

func TestBookingService_CreateBooking_Success(t *testing.T) {
	mockBookingRepo := new(MockBookingRepository)
	mockRoomRepo := new(MockRoomRepository)
	
	service := &BookingService{
		bookingRepo: mockBookingRepo,
		roomRepo:    mockRoomRepo,
	}

	ctx := context.Background()
	
	// Use future dates
	checkIn := time.Now().AddDate(0, 0, 7)  // 7 days from now
	checkOut := time.Now().AddDate(0, 0, 10) // 10 days from now
	
	req := &models.BookingRequest{
		UserId:  "user-123",
		RoomId:  "room-1", 
		CheckIn: checkIn.Format("2006-01-02"),
		CheckOut: checkOut.Format("2006-01-02"),
		Guests:  2,
	}

	room := &models.Room{
		Id:            "room-1",
		RoomNumber:    "101",
		RoomType:      models.RoomTypeDouble,
		PricePerNight: 150.0,
		MaxGuests:     2,
		Available:     true,
	}

	// Setup mock expectations
	mockRoomRepo.On("GetRoomById", ctx, "room-1").Return(room, nil)
	mockBookingRepo.On("CreateBooking", ctx, mock.AnythingOfType("*models.Booking")).Return(nil)

	// Call the method
	booking, err := service.CreateBooking(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, "user-123", booking.UserId)
	assert.Equal(t, "room-1", booking.RoomId)
	assert.Equal(t, models.RoomTypeDouble, booking.RoomType)
	assert.Equal(t, 2, booking.Guest)
	assert.Equal(t, 450.0, booking.TotalAmount) // 3 nights * 150
	assert.Equal(t, models.StatusConfirmed, booking.Status)
	
	// Verify mocks were called
	mockRoomRepo.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

func TestBookingService_CreateBooking_RoomNotFound(t *testing.T) {
	mockBookingRepo := new(MockBookingRepository)
	mockRoomRepo := new(MockRoomRepository)
	
	service := &BookingService{
		bookingRepo: mockBookingRepo,
		roomRepo:    mockRoomRepo,
	}

	ctx := context.Background()
	
	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	
	req := &models.BookingRequest{
		UserId:  "user-123",
		RoomId:  "non-existent-room",
		CheckIn: checkIn.Format("2006-01-02"),
		CheckOut: checkOut.Format("2006-01-02"),
		Guests:  2,
	}

	// Mock returns error for non-existent room
	mockRoomRepo.On("GetRoomById", ctx, "non-existent-room").Return(nil, assert.AnError)

	booking, err := service.CreateBooking(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, booking)
	assert.Contains(t, err.Error(), "failed to get room")
	mockRoomRepo.AssertExpectations(t)
}

func TestBookingService_CreateBooking_TooManyGuests(t *testing.T) {
	mockBookingRepo := new(MockBookingRepository)
	mockRoomRepo := new(MockRoomRepository)
	
	service := &BookingService{
		bookingRepo: mockBookingRepo,
		roomRepo:    mockRoomRepo,
	}

	ctx := context.Background()
	
	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	
	req := &models.BookingRequest{
		UserId:  "user-123",
		RoomId:  "room-1",
		CheckIn: checkIn.Format("2006-01-02"),
		CheckOut: checkOut.Format("2006-01-02"),
		Guests:  5, // More than room capacity
	}

	room := &models.Room{
		Id:            "room-1",
		RoomNumber:    "101", 
		RoomType:      models.RoomTypeDouble,
		PricePerNight: 150.0,
		MaxGuests:     2, // Room only allows 2 guests
		Available:     true,
	}

	mockRoomRepo.On("GetRoomById", ctx, "room-1").Return(room, nil)

	booking, err := service.CreateBooking(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, booking)
	assert.Contains(t, err.Error(), "room can only accommodate 2 guests")
	mockRoomRepo.AssertExpectations(t)
}

func TestBookingService_GetBooking_Success(t *testing.T) {
	mockBookingRepo := new(MockBookingRepository)
	mockRoomRepo := new(MockRoomRepository)
	
	service := &BookingService{
		bookingRepo: mockBookingRepo,
		roomRepo:    mockRoomRepo,
	}

	ctx := context.Background()
	bookingId := "booking-123"

	expectedBooking := &models.Booking{
		Id:       bookingId,
		UserId:   "user-123",
		RoomId:   "room-1",
		RoomType: models.RoomTypeDouble,
		Status:   models.StatusConfirmed,
	}

	mockBookingRepo.On("GetBookingById", ctx, bookingId).Return(expectedBooking, nil)

	booking, err := service.GetBooking(ctx, bookingId)

	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, bookingId, booking.Id)
	assert.Equal(t, "user-123", booking.UserId)
	mockBookingRepo.AssertExpectations(t)
}

func TestBookingService_CancelBooking_Success(t *testing.T) {
	mockBookingRepo := new(MockBookingRepository)
	mockRoomRepo := new(MockRoomRepository)
	
	service := &BookingService{
		bookingRepo: mockBookingRepo,
		roomRepo:    mockRoomRepo,
	}

	ctx := context.Background()
	bookingId := "booking-123"

	booking := &models.Booking{
		Id:      bookingId,
		UserId:  "user-123",
		CheckIn: time.Now().Add(48 * time.Hour), // 48 hours from now
		Status:  models.StatusConfirmed,
	}

	mockBookingRepo.On("GetBookingById", ctx, bookingId).Return(booking, nil)
	mockBookingRepo.On("UpdateBookingStatus", ctx, bookingId, models.StatusCancelled).Return(nil)

	err := service.CancelBooking(ctx, bookingId)

	assert.NoError(t, err)
	mockBookingRepo.AssertExpectations(t)
}

func TestBookingService_CancelBooking_TooLate(t *testing.T) {
	mockBookingRepo := new(MockBookingRepository)
	mockRoomRepo := new(MockRoomRepository)
	
	service := &BookingService{
		bookingRepo: mockBookingRepo,
		roomRepo:    mockRoomRepo,
	}

	ctx := context.Background()
	bookingId := "booking-123"

	booking := &models.Booking{
		Id:      bookingId,
		UserId:  "user-123", 
		CheckIn: time.Now().Add(12 * time.Hour), // Only 12 hours from now
		Status:  models.StatusConfirmed,
	}

	mockBookingRepo.On("GetBookingById", ctx, bookingId).Return(booking, nil)

	err := service.CancelBooking(ctx, bookingId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "booking can only be canceled at least 24 hours before check in")
	mockBookingRepo.AssertExpectations(t)
}