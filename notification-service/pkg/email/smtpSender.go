package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/circuitbreaker"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/retry"
)

// SMTPSender handles sending emails via SMTP
type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	from     string
	cb       *circuitbreaker.CircuitBreaker
}

// NewSMTPSender creates a new SMTP email sender
func NewSMTPSender(host string, port int, username, password, from string) *SMTPSender {
	// Circuit breaker: 5 failures open circuit, 2 successes to close, 30 second timeout
	cb := circuitbreaker.NewCircuitBreaker(5, 2, 30*time.Second)
	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		cb:       cb,
	}
}

// SendEmail sends an email using SMTP with timeout and circuit breaker
func (s *SMTPSender) SendEmail(to, subject, body string) error {
	// Use circuit breaker with the internal send method
	err := s.cb.Execute(func() error {
		return s.sendEmailWithRetry(to, subject, body)
	})
	return err
}

// sendEmailWithRetry handles email sending with retry logic
func (s *SMTPSender) sendEmailWithRetry(to, subject, body string) error {
	// Use retry mechanism
	retryConfig := retry.DefaultRetryConfig()
	
	return retry.Retry(retryConfig, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return s.sendEmailWithContext(ctx, to, subject, body)
	})
}

// sendEmailWithContext contains actual email sending logic with context support
func (s *SMTPSender) sendEmailWithContext(ctx context.Context, to, subject, body string) error {
	// Set up authentication information
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Create the email headers and body
	message := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		s.from, to, subject, body,
	)

	// Connect to the SMTP server
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// For port 465, use TLS directly
	if s.port == 465 {
		return s.sendWithTLS(ctx, addr, auth, to, []byte(message))
	}

	// For port 587, use StartTLS
	if s.port == 587 {
		return s.sendWithStartTLS(ctx, addr, auth, to, []byte(message))
	}

	// For other ports, use regular SMTP (with timeout)
	return s.sendWithRegularSMTP(ctx, addr, auth, to, []byte(message))
}

// sendWithTLS handles SMTP with direct TLS for port 465
func (s *SMTPSender) sendWithTLS(ctx context.Context, addr string, auth smtp.Auth, to string, message []byte) error {
	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName: s.host,
	}

	// Dial with timeout using context
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}
	
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set the sender and recipient
	if err = client.Mail(s.from); err != nil {
		return fmt.Errorf("sender failed: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("recipient failed: %w", err)
	}

	// Send the email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data command failed: %w", err)
	}

	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("writing message failed: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("closing data failed: %w", err)
	}

	return client.Quit()
}

// sendWithStartTLS handles SMTP with StartTLS for port 587
func (s *SMTPSender) sendWithStartTLS(ctx context.Context, addr string, auth smtp.Auth, to string, message []byte) error {
	// Connect to the SMTP server with timeout
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}
	
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Send STARTTLS command
	if err = client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set the sender and recipient
	if err = client.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send the email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// sendWithRegularSMTP handles regular SMTP for other ports
func (s *SMTPSender) sendWithRegularSMTP(ctx context.Context, addr string, auth smtp.Auth, to string, message []byte) error {
	// Simple SMTP send with context check
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return smtp.SendMail(addr, auth, s.from, []string{to}, message)
	}
}

// TestConnection tests the SMTP connection and authentication
func (s *SMTPSender) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	if s.port == 465 {
		// Test TLS connection for port 465
		tlsConfig := &tls.Config{
			ServerName: s.host,
		}

		dialer := &net.Dialer{
			Timeout: 5 * time.Second,
		}

		conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS connection failed: %w", err)
		}
		defer conn.Close()
		return nil
	}

	if s.port == 587 {
		// Test StartTLS connection for port 587
		dialer := &net.Dialer{
			Timeout: 5 * time.Second,
		}

		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.host)
		if err != nil {
			return fmt.Errorf("client creation failed: %w", err)
		}
		defer client.Close()

		if err = client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return fmt.Errorf("StartTLS failed: %w", err)
		}

		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		return nil
	}

	// For other ports, simple TCP connection test
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()
	return nil
}