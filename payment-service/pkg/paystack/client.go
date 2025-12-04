package paystack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"github.com/ollatomiwa/hotelsystem/payment-service/internals/models"
	"time"

	"github.com/sirupsen/logrus"
)

type Client struct {
	secretKey  string
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

func NewClient(secretKey, baseURL string, logger *logrus.Logger) *Client {
	return &Client{
		secretKey: secretKey,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// InitializeTransaction initializes a payment transaction
func (c *Client) InitializeTransaction(req *models.InitializePaymentRequest) (*models.PaystackInitializeResponse, error) {
	payload := map[string]interface{}{
		"email":    req.Email,
		"amount":   req.Amount,
		"currency": req.Currency,
	}

	if req.Reference != "" {
		payload["reference"] = req.Reference
	}
	if req.CallbackURL != "" {
		payload["callback_url"] = req.CallbackURL
	}
	if len(req.Metadata) > 0 {
		payload["metadata"] = req.Metadata
	}

	var response models.PaystackInitializeResponse
	err := c.doRequestWithRetry("POST", "/transaction/initialize", payload, &response, 3)
	if err != nil {
		return nil, err
	}

	if !response.Status {
		return nil, fmt.Errorf("paystack error: %s", response.Message)
	}

	return &response, nil
}

// VerifyTransaction verifies a transaction
func (c *Client) VerifyTransaction(reference string) (*models.PaystackVerifyResponse, error) {
	var response models.PaystackVerifyResponse
	err := c.doRequestWithRetry("GET", fmt.Sprintf("/transaction/verify/%s", reference), nil, &response, 3)
	if err != nil {
		return nil, err
	}

	if !response.Status {
		return nil, fmt.Errorf("paystack error: %s", response.Message)
	}

	return &response, nil
}

// doRequestWithRetry performs an HTTP request with exponential backoff retry
func (c *Client) doRequestWithRetry(method, path string, payload interface{}, result interface{}, maxRetries int) error {
	var lastErr error
	backoff := time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Infof("Retrying request (attempt %d/%d) after %v", attempt, maxRetries, backoff)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}

		err := c.doRequest(method, path, payload, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on client errors (4xx)
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
			return err
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doRequest performs an HTTP request
func (c *Client) doRequest(method, path string, payload interface{}, result interface{}) error {
	url := c.baseURL + path

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")

	c.logger.Debugf("Paystack request: %s %s", method, url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.Debugf("Paystack response: %d %s", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// HTTPError represents an HTTP error
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}