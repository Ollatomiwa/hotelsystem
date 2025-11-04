package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/services"
)

type BookingHandler struct {
	bookingService *services.BookingService
}

func NewBookingHandler(bookingService *services.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req models.BookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request:"+ err.Error()})
		return 
	}

	booking, err := h.bookingService.CreateBooking(c.Request.Context(), &req)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error" :" booking creation failed"+ err.Error()})
		return 
	}
	c.JSON(http.StatusCreated, booking)
}

func (h *BookingHandler) GetBooking(c *gin.Context){
	bookingId := c.Param("id")

	if _, err := uuid.Parse(bookingId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking Id"+ err.Error()})
		return 
	}

	booking, err := h.bookingService.GetBooking(c.Request.Context(), bookingId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found" + err.Error()})
		return 
	}

	c.JSON(http.StatusOK, booking)
}

func (h *BookingHandler) GetUserBookings(c *gin.Context) {
	userId := c.Query("user_id")
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return 
	}

	bookings, err := h.bookingService.GetUserBooking(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user bookings" + err.Error()})
		return 
	}
	c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingId := c.Param("id")

	if _,err := uuid.Parse(bookingId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking Id" + err.Error()})
		return 
	}

	err := h.bookingService.CancelBooking(c.Request.Context(), bookingId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking not found"})
		return 
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking Cancelled successfully"}) 

}

// Helper functions
func isValidRoomType(roomType models.RoomType) bool {
	switch roomType {
	case models.RoomTypeSingle, models.RoomTypeDouble, models.RoomTypeDeluxe:
		return true
	default:
		return false
	}
}