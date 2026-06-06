package lookup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"sumni-finance-backend/internal/treasury/domain"
)

var ErrBankCodeNotFound = errors.New("bank not found")

type Bank struct {
	Name              string
	Code              string
	Bin               int
	ShortName         string
	Logo              string
	LookupSupported   int
	TransferSupported int
}

func (b Bank) toDomain() (domain.BankInfo, error) {
	return domain.NewBankInfo(
		b.Name,
		b.Bin,
		b.Code,
		b.ShortName,
		b.Logo,
		b.Logo,
		b.LookupSupported,
		b.TransferSupported,
	)
}

type bankDTO struct {
	Name              string `json:"name"`
	Code              string `json:"code"`
	Bin               string `json:"bin"`
	ShortName         string `json:"shortName"`
	Logo              string `json:"logo"`
	TransferSupported int    `json:"transferSupported"`
	LookupSupported   int    `json:"lookupSupported"`
}

type bankListResponse struct {
	Code string    `json:"code"`
	Desc string    `json:"desc"`
	Data []bankDTO `json:"data"`
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	mu         sync.RWMutex
	cached     []Bank
}

func NewClient(httpClient *http.Client, baseURL string) *Client {
	if httpClient == nil {
		panic("http client can't be empty")
	}

	if baseURL == "" {
		panic("baseUrl can't be empty")
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (c *Client) ListBanks(ctx context.Context) ([]domain.BankInfo, error) {
	banks, err := c.loadCache(ctx)
	if err != nil {
		return nil, err
	}
	return mapToDomain(banks)
}

func (c *Client) FindByBankCode(ctx context.Context, bankCode string) (domain.BankInfo, error) {
	banks, err := c.loadCache(ctx)
	if err != nil {
		return domain.BankInfo{}, err
	}

	for _, b := range banks {
		if b.Code == bankCode {
			return b.toDomain()
		}
	}

	return domain.BankInfo{}, ErrBankCodeNotFound
}

func (c *Client) loadCache(ctx context.Context) ([]Bank, error) {
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

func (c *Client) fetch(ctx context.Context) ([]Bank, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v2/banks", nil)
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

	banks := make([]Bank, 0, len(body.Data))
	for _, dto := range body.Data {
		bin, err := strconv.Atoi(dto.Bin)
		if err != nil {
			return nil, fmt.Errorf("invalid bin %q for bank %q: %w", dto.Bin, dto.Name, err)
		}

		banks = append(banks, Bank{
			Name:              dto.Name,
			Code:              dto.Code,
			Bin:               bin,
			ShortName:         dto.ShortName,
			Logo:              dto.Logo,
			LookupSupported:   dto.LookupSupported,
			TransferSupported: dto.TransferSupported,
		})
	}

	return banks, nil
}

func mapToDomain(banks []Bank) ([]domain.BankInfo, error) {
	result := make([]domain.BankInfo, 0, len(banks))
	for _, b := range banks {
		bankInfo, err := b.toDomain()
		if err != nil {
			return nil, fmt.Errorf("invalid bank info for %q: %w", b.Name, err)
		}
		result = append(result, bankInfo)
	}
	return result, nil
}
