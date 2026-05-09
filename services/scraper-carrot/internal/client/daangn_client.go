package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"scraper-carrot/internal/models"
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

	return &DaangnClient{
		httpClient: httpClient,
		userAgent:  defaultUserAgent,
	}
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
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return body, nil
}

func (c *DaangnClient) FetchArticlesWithPoW(
	ctx context.Context,
	pow models.RemixContext,
) ([]models.Article, error) {
	route := pow.State.LoaderData.Route
	nonce := computePoWNonce(route.PoW.Challenge, route.PoW.Difficulty)

	params := url.Values{}
	params.Set("region_id", route.CurrentFilters.RegionID)
	params.Set("search", route.CurrentFilters.Search)
	if route.CurrentFilters.CategoryID != "" {
		params.Set("category_id", route.CurrentFilters.CategoryID)
	}
	if route.CurrentFilters.Price != "" {
		params.Set("price", route.CurrentFilters.Price)
	}
	if route.CurrentFilters.OnlyOnSale {
		params.Set("only_on_sale", "true")
	}
	params.Set("uri", route.PoW.URI)
	params.Set("nonce", strconv.Itoa(nonce))
	params.Set("expires_at", strconv.FormatInt(route.PoW.ExpiresAt, 10))

	apiURL := "https://www.daangn.com/kr/api/v1/fleamarket/search?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create pow request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.daangn.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send pow request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pow request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var apiResp models.FleamarketSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode pow response: %w", err)
	}

	return apiResp.FleamarketArticles, nil
}

func computePoWNonce(challenge string, difficulty int) int {
	if difficulty <= 0 {
		return 0
	}

	targetPrefix := strings.Repeat("0", difficulty)
	for nonce := 0; ; nonce++ {
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", challenge, nonce)))
		hashHex := hex.EncodeToString(hash[:])
		if strings.HasPrefix(hashHex, targetPrefix) {
			return nonce
		}
	}
}
