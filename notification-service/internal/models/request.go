package models

//type sendemailrequest represents the payload for sending an email
type SendEmailRequest struct {
	To string `json:"to" binding:"required,email"`
	Subject string `json:"subject" binding:"required,min=1,max=255"`
	Body string `json:"body" binding:"required,min=1"`
	Type string `json:"type" binding:"required,oneof=booking_confirmation welcome_email payment_receipt password_reset"`
}

//type sendemailresponse represents the Api response after sending an email
type SendEmailResponse struct {
	Id string `json:"id"`
	Status string `json:"status"`
	Message string `json:"message"`
	TimeStamp string `json:"timestamp"`
}

//type errorresponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
	Code int `json:"code"`
	Message string `json:"message"`
}

//health response check
type HealthResponse struct {
	Status string `json:"status"`
	Service string `json:"service"`
	TimeStamp string `json:"timestamp"`
}

