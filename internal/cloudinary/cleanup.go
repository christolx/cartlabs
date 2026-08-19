package cloudinary

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const demoUploadTag = "cartlabs_demo_upload"

type CleanupConfig struct {
	CloudName string
	APIKey    string
	APISecret string
	BaseURL   string
}

type CleanupResult struct {
	Deleted int
	Pages   int
}

type deleteResponse struct {
	Deleted    map[string]string `json:"deleted"`
	Partial    bool              `json:"partial"`
	NextCursor string            `json:"next_cursor"`
}

func DeleteDemoUploads(ctx context.Context, client *http.Client, config CleanupConfig) (CleanupResult, error) {
	if config.CloudName == "" || config.APIKey == "" || config.APISecret == "" {
		return CleanupResult{}, fmt.Errorf("Cloudinary cleanup credentials are incomplete")
	}
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.cloudinary.com"
	}
	endpoint := fmt.Sprintf("%s/v1_1/%s/resources/image/tags/%s", baseURL, url.PathEscape(config.CloudName), demoUploadTag)
	result := CleanupResult{}
	cursor := ""
	for {
		requestURL := endpoint
		if cursor != "" {
			requestURL += "?next_cursor=" + url.QueryEscape(cursor)
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURL, nil)
		if err != nil {
			return result, fmt.Errorf("create Cloudinary cleanup request: %w", err)
		}
		request.SetBasicAuth(config.APIKey, config.APISecret)
		response, err := client.Do(request)
		if err != nil {
			return result, fmt.Errorf("delete Cloudinary demo uploads: %w", err)
		}
		var payload deleteResponse
		decodeErr := json.NewDecoder(response.Body).Decode(&payload)
		_ = response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return result, fmt.Errorf("delete Cloudinary demo uploads: status %d", response.StatusCode)
		}
		if decodeErr != nil {
			return result, fmt.Errorf("decode Cloudinary cleanup response: %w", decodeErr)
		}
		result.Pages++
		result.Deleted += len(payload.Deleted)
		if !payload.Partial {
			return result, nil
		}
		if payload.NextCursor == "" {
			return result, fmt.Errorf("Cloudinary cleanup response is partial without next_cursor")
		}
		cursor = payload.NextCursor
	}
}
