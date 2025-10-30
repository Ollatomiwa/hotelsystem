//gin routes and HTTP logic
package handlers

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/services"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/security"
)

//Notificationhandler to handle http requests for notifs
type NotificationHandler struct {
	notificationService *services.NotificationService
}

//constructor: 
func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}
// sendemail handles post /ap/..
func (h *NotificationHandler) SendEmail(c *gin.Context) {
	var req models.SendEmailRequest

	//binding and validate the request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invaid request payload:" + err.Error()})
		return 
	}

	//sanitize inputs
	sanitizedTo, err := security.SanitizeEmail(req.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address" + err.Error()})
		return
	}
	sanitizedSubject := security.SanitizeSubject(req.Subject)
	sanitizedBody := security.SanitizeBody(req.Body)

	//update requests with sanitized data
	req.To = sanitizedTo
	req.Subject = sanitizedSubject
	req.Body = sanitizedBody


	//calling the server
	ctx := c.Request.Context()
	response, err := h.notificationService.SendEmail(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email:" + err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, response)
}

//getnotificationstatus handler to get by id
func (h *NotificationHandler) GetNotificationStatus(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Notification ID is required"})
		return 
	}

	//call the service layer
	ctx:= c.Request.Context()
	notification, err := h.notificationService.GetNotificationStatus(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound,gin.H{"error": "notification not found:" + err.Error()})
		return 
	}
	c.JSON(http.StatusOK, notification)
}
//healthcheck handler
func (h *NotificationHandler) HealthCheck(c *gin.Context){
	healthResponse := models.HealthResponse{
		Status: "healthy",
		TimeStamp: time.Now().Format(time.RFC3339),
		Service: "notification-service",
	}
	c.JSON(http.StatusOK, healthResponse)
}