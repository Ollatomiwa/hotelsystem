package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/services"
	"github.com/ollatomiwa/hotelsystem/booking-service/pkg/middleware"
)

func SetupRoutes(router *gin.Engine, bookingService *services.BookingService) {
	bookingHandler := NewBookingHandler(bookingService)
	healthHandler := NewHealthHandler()

	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	router.GET("/health", healthHandler.HealthCheck)

	v1 := router.Group("/api/v1")
	{
		bookings := v1.Group("/bookings")
		{
			bookings.POST("", bookingHandler.CreateBooking)
			bookings.POST("/availability", bookingHandler.CheckAvailability)
			bookings.GET("", bookingHandler.GetUserBookings)
			bookings.GET("/:id", bookingHandler.GetBooking)
			bookings.PUT("/:id/cancel", bookingHandler.CancelBooking)
		}
	}

	router.NoRoute(func(c *gin.Context){
		c.JSON(404, NewErrorResponse("not found", "endpoint not found"))
	})
}
