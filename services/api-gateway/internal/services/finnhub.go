package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type FinnhubClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type FinnhubQuote struct {
	CurrentPrice       float64 `json:"c"`
	Change             float64 `json:"d"`
	PercentChange      float64 `json:"dp"`
	HighPriceOfDay     float64 `json:"h"`
	LowPriceOfDay      float64 `json:"l"`
	OpenPriceOfDay     float64 `json:"o"`
	PreviousClosePrice float64 `json:"pc"`
	Timestamp          int64   `json:"t"`
}

type FinnhubProfile struct {
	Country         string  `json:"country"`
	Currency        string  `json:"currency"`
	Exchange        string  `json:"exchange"`
	IPO             string  `json:"ipo"`
	MarketCap       float64 `json:"marketCapitalization"`
	Name            string  `json:"name"`
	Phone           string  `json:"phone"`
	SharesOut       float64 `json:"shareOutstanding"`
	Ticker          string  `json:"ticker"`
	WebURL          string  `json:"weburl"`
	Logo            string  `json:"logo"`
	FinnhubIndustry string  `json:"finnhubIndustry"`
}

func NewFinnhubClient(apiKey string) *FinnhubClient {
	return &FinnhubClient{
		apiKey:  apiKey,
		baseURL: "https://finnhub.io/api/v1",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f *FinnhubClient) GetQuote(symbol string) (*FinnhubQuote, error) {
	url := fmt.Sprintf("%s/quote?symbol=%s&token=%s", f.baseURL, symbol, f.apiKey)

	resp, err := f.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote for %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var quote FinnhubQuote
	if err := json.Unmarshal(body, &quote); err != nil {
		return nil, fmt.Errorf("failed to parse quote response: %w", err)
	}

	return &quote, nil
}

func (f *FinnhubClient) GetCompanyProfile(symbol string) (*FinnhubProfile, error) {
	url := fmt.Sprintf("%s/stock/profile2?symbol=%s&token=%s", f.baseURL, symbol, f.apiKey)

	resp, err := f.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile for %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var profile FinnhubProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	return &profile, nil
}
