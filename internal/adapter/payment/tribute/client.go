package tribute

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Option func(*APIClient)

func WithHTTPClient(c *http.Client) Option { return func(cl *APIClient) { cl.httpClient = c } }

func WithBaseURL(url string) Option { return func(cl *APIClient) { cl.baseURL = url } }

// OrderDTO represents order info returned by Tribute.
type OrderDTO struct {
	ID         int    `json:"id"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	TelegramID int64  `json:"telegram_id"`
}

type APIClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

const defaultBaseURL = "https://tribute.tg/api/v1"

func New(apiKey string, opts ...Option) *APIClient {
	c := &APIClient{apiKey: apiKey, httpClient: http.DefaultClient, baseURL: defaultBaseURL}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *APIClient) ListOrders(ctx context.Context, limit, lastID int) ([]OrderDTO, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/orders", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if lastID > 0 {
		q.Set("last_id", fmt.Sprintf("%d", lastID))
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var out []OrderDTO
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
