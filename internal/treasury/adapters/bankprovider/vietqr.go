package bankprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"sumni-finance-backend/internal/treasury/domain"
)

const defaultVietQrBaseUrl = "https://api.vietqr.io/v2/banks"

func NewVietQrClient() *VietQrClient {
	return &VietQrClient{
		httpClient: http.DefaultClient,
		baseURL:    defaultVietQrBaseUrl,
	}
}

type bankDTO struct {
	Name              string `json:"name"`
	Code              string `json:"code"`
	Bin               string `json:"bin"`
	ShortName         string `json:"shortName"`
	Logo              string `json:"logo"`
	TransferSupported int    `json:"transferSupported"`
	LookupSupported   int    `json:"lookupSupported"`
	SwiftCode         string `json:"swift_code"`
}

type bankListResponse struct {
	Code string    `json:"code"`
	Desc string    `json:"desc"`
	Data []bankDTO `json:"data"`
}

type VietQrClient struct {
	httpClient *http.Client
	baseURL    string
	mu         sync.RWMutex
	cached     []bankDTO
}

func (c *VietQrClient) FindBankInfoByCode(ctx context.Context, bankCode string) (domain.BankInfo, error) {
	banks, err := c.loadCache(ctx)
	if err != nil {
		return domain.BankInfo{}, err
	}

	for _, b := range banks {
		if b.Code == bankCode {
			return domain.NewBankInfo(
				b.Name,
				b.Bin,
				b.Code,
				b.ShortName,
				b.Logo,
			)
		}
	}

	return domain.BankInfo{}, fmt.Errorf("bank code not found")
}

func (c *VietQrClient) loadCache(ctx context.Context) ([]bankDTO, error) {
	c.mu.RLock()
	if c.cached != nil {
		banks := c.cached
		c.mu.RUnlock()
		return banks, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cached != nil {
		return c.cached, nil
	}

	banks, err := c.fetch(ctx)
	if err != nil {
		return nil, err
	}

	c.cached = banks
	return c.cached, nil
}

func (c *VietQrClient) fetch(ctx context.Context) ([]bankDTO, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch banks: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from bank list API", resp.StatusCode)
	}

	var body bankListResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode bank list response: %w", err)
	}

	if body.Code != "00" {
		return nil, fmt.Errorf("bank list API error: %s", body.Desc)
	}

	return body.Data, nil
}
