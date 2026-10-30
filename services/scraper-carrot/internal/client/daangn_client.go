package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type DaangnClient struct {
	httpClient *http.Client
	userAgent  string
}

func NewDaangnClient(httpClient *http.Client) *DaangnClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &DaangnClient{httpClient: httpClient, userAgent: defaultUserAgent}
}

func (c *DaangnClient) FetchSearchPage(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return body, nil
}
