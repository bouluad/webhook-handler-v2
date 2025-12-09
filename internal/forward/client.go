package forward

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Client handles outbound HTTP requests to the target on-premise tool.
type Client struct {
	TargetURL string 
	AuthToken string 
	httpClient *http.Client
}

func NewClient(url, token string) *Client {
	return &Client{
		TargetURL: url,
		AuthToken: token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ForwardPayload sends the raw JSON payload to the target tool.
func (c *Client) ForwardPayload(payload []byte) error {
	req, err := http.NewRequest(http.MethodPost, c.TargetURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
    // Set a generic authorization header
	if c.AuthToken != "" {
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken)) 
    }
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to target tool at %s: %w", c.TargetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body) 
		log.Printf("Target tool responded with non-2xx status: %s. Body snippet: %s", resp.Status, respBody)
		return fmt.Errorf("target tool returned status code %s", resp.Status)
	}

	return nil
}
