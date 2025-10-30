package email

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// ResendSender handles sending emails via Resend API
type ResendSender struct {
    apiKey string
    from   string
}

// NewResendSender creates a new Resend email sender
func NewResendSender(apiKey, from string) *ResendSender {
    return &ResendSender{
        apiKey: apiKey,
        from:   from,
    }
}

// SendEmail sends an email using Resend API
func (r *ResendSender) SendEmail(to, subject, body string) error {
    payload := map[string]interface{}{
        "from":    r.from,
        "to":      []string{to},
        "subject": subject,
        "text":    body,
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    client := &http.Client{Timeout: 30 * time.Second}
    req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+r.apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("resend API error: %d %s", resp.StatusCode, resp.Status)
    }

    return nil
}

// TestConnection tests the Resend API connection
func (r *ResendSender) TestConnection() error {
    client := &http.Client{Timeout: 10 * time.Second}
    req, err := http.NewRequest("GET", "https://api.resend.com/domains", nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+r.apiKey)
    
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return fmt.Errorf("resend API test failed: %d %s", resp.StatusCode, resp.Status)
    }
    
    return nil
}