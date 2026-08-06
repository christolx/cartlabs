package purchase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPPaymentProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewHTTPPaymentProvider(baseURL, apiKey string) *HTTPPaymentProvider {
	return &HTTPPaymentProvider{baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second}}
}

func (p *HTTPPaymentProvider) CreateIntent(ctx context.Context, input PaymentIntentRequest) (PaymentIntent, error) {
	var output PaymentIntent
	if err := p.request(ctx, http.MethodPost, "/v1/intents", input, &output); err != nil {
		return PaymentIntent{}, err
	}
	return output, nil
}

func (p *HTTPPaymentProvider) CompleteIntent(ctx context.Context, intentID, outcome string) error {
	return p.request(ctx, http.MethodPost, "/v1/intents/"+intentID+"/complete", map[string]string{"outcome": outcome}, nil)
}

func (p *HTTPPaymentProvider) request(ctx context.Context, method, path string, input, output any) error {
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("encode payment request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create payment request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+p.apiKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := p.client.Do(request)
	if err != nil {
		return fmt.Errorf("send payment request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
		return fmt.Errorf("payment provider returned status %d", response.StatusCode)
	}
	if output != nil {
		if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(output); err != nil {
			return fmt.Errorf("decode payment response: %w", err)
		}
	}
	return nil
}
