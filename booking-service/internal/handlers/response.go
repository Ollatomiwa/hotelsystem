package handlers

import "time"


type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

 
type SuccessResponse struct {
	Message string `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}


func NewErrorResponse(errorType, message string) ErrorResponse {
	return ErrorResponse{
		Error:     errorType,
		Message:   message,
		Timestamp: time.Now(),
	}
}