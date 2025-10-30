package security

import (
	"html"
	"net/mail"
	"strings"
	"unicode/utf8"
)

//sanitize email sanitize and validates an email address
func SanitizeEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	//basic html escaping
	email = html.EscapeString(email)

	//validate email format
	_, err := mail.ParseAddress(email) 
	if err != nil {
		return "", err 
	}

	return email, nil
}

//sanitize text sanitizes text inout
func SanitizeText(text string, maxLength int) string {
	text = strings.TrimSpace(text)

	//limit length
	if utf8.RuneCountInString(text) > maxLength {
		runes := []rune(text)
		text = string(runes[:maxLength])
	}

	//escape html to prevent xss
	text = html.EscapeString(text)
	return text
}

//sanitize subject sanitizes email subject
func SanitizeSubject(subject string) string {
	return SanitizeText(subject, 255) //rf limit
}

//sanitize body sanitizes email body
func SanitizeBody(body string) string {
	return SanitizeText(body, 10000) //reasonable limit
}