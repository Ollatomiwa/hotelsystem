package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
)

//SMTP sender handles sebding emails via smtp
type SMTPSender struct {
	host string
	port int
	username string
	password string
	from string 
}

//new smtp sender to create a new smtp email sender
func NewSMTPSender (host string, port int, username, password, from string) *SMTPSender {
	return &SMTPSender{
		host: host,
		port: port,
		username: username,
		password: password,
		from: from, 
	}
}

//sendemail method to send email using SMTP
func (s *SMTPSender) SendEmail(to, subject, body string) error {
	//setting up authnentication info
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	//create the email header and body
	message := fmt.Sprintf("From: %s\r\n" + "To: %s\r\n" + "Subject: %s\r\n" + "Content-Type: text/plain; charser=UTF-8\r\n" + "\r\n" + "%s", s.from, to, subject, body,)

	//connect to the smtp server
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	//for port 587, we need to use starttls
	if s.port == 587 {
		return s.sendWithStartTLS(addr, auth, to, []byte(message))
	}

	//for other ports use regular smtp
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(message))
}

//sendwithstarttls methosd handls smpt for port 587
func (s *SMTPSender) sendWithStartTLS(addr string, auth smtp.Auth, to string, message[]byte) error {
	//coonect to the smtp server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	//send STARTTLS command
	if err = client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	//authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	//set the sender and the recipent
	if err = client.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %W",err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set the recipent: %w", err)
	}

	//send the email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("failed to get write email body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}
	return client.Quit()
}

//testconn tests smtp conn and auth
// TestConnection tests the SMTP connection and authentication
func (s *SMTPSender) TestConnection() error {
    // Mask password in logs for security
    maskedPassword := "***"
    if len(s.password) > 0 {
        maskedPassword = "****" + s.password[len(s.password)-4:]
    }
    
    log.Printf("Testing SMTP connection to %s:%d with user: %s", s.host, s.port, s.username)
    log.Printf("Password: %s", maskedPassword)
    
    auth := smtp.PlainAuth("", s.username, s.password, s.host)
    addr := fmt.Sprintf("%s:%d", s.host, s.port)
    
    if s.port == 587 {
        client, err := smtp.Dial(addr)
        if err != nil {
            return fmt.Errorf("failed to dial SMTP server: %w", err)
        }
        defer client.Close()
        
        if err = client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
            return fmt.Errorf("failed to start TLS: %w", err)
        }
        
        if err = client.Auth(auth); err != nil {
            return fmt.Errorf("authentication failed: %w", err)
        }
        
        return nil
    }
    
    // For port 465
    return smtp.SendMail(addr, auth, s.from, []string{s.username}, []byte("Test connection"))
}
