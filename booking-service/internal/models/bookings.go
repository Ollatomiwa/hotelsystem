package models

import (
	"time"
)

//booking status represents the status of a booking
type BookingStatus  string

const (
	StatusPending BookingStatus = "pending"
	StatusConfirmed BookingStatus = "confirmed"
	StatusCancelled BookingStatus = "cancelled"
)

//room type represents different types of rooms
type RoomType string

const (
	RoomTypeSingle RoomType = "single"
	RoomTypeDouble RoomType = "double"
	RoomTypeDeluxe RoomType = "deluxe"
)
//booking struct represents a hotel room booking
type Booking struct {
	Id string `json:"id"`
	UserId string `json:"user_id"`
	RoomId string `json:"room_id"`
	RoomType RoomType `json:"room_type"`
	CheckIn time.Time `json:"check_in"`
	CheckOut time.Time `json:"check_out"`
	Guest int `json:"guests"`
	TotalAmount float64 `json:"total_amount"`
	Status BookingStatus `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

//room represents a hotel room
type Room struct {
	Id string `json:"id"`
	RoomNumber string `json:"room_number"`
	RoomType RoomType `json:"room_type"`
	PricePerNight float64 `json:"price_per_night"`
	MaxGuests int `json:"max_guests"`
	Available bool `json:"available"`
	Description string `json:"description"`
}
//createbooking request represents the payload for creating a booking
type BookingRequest struct {
	UserId string `json:"user_id" binding:"required"`
	RoomId string `json:"room_id" binding:"required"`
	CheckIn string `json:"check_in" binding:"required"`
	CheckOut string `json:"check_out" binding:"required"`
	Guests int `json:"guests" binding:"Required,min=1,max=5"`
}
//availability request represents the payload for check room availability
type AvailabilityRequest struct {
	RoomType RoomType `json:"room_type" binding:"required"`
	CheckIn string `json:"check_in" binding:"required"`
	CheckOut string `json:"check_out" binding:"required"`
	Guests int  `json:"guests" binding:"required,min=1,max=5"`
}

//availability response represents available rooms for a given dates
type AvailabilityResponse struct {
	AvailableRooms []RoomAvailability `json:"available_rooms"`
	TotalAvailable int `json:"total_available"`
}
//room availability represents an available room with pricing
type RoomAvailability struct {
	RoomId string `json:"room_id"`
	RoomNumber string `json:"room_number"`
	RoomType RoomType `json:"room_type"`
	PricePerNight float64 `json:"price_per_night"`
	TotalPrice float64 `json:"total_price"`
	MaxGuests int `json:"max_guests"`
}