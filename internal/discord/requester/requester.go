package requester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/marouane-souiri/vocalize/internal/discord/models"
	"github.com/marouane-souiri/vocalize/internal/ratelimiter"
)

const (
	APIBaseURL = "https://discord.com/api/v10"
)

type APIRequester interface {
	SendMessage(channelID string, message *models.SendMessage) error
}

type APIRequesterImpl struct {
	client      *http.Client
	token       string
	rateLimiter ratelimiter.RateLimiter
}

func NewAPIRequester(token string, rateLimiter ratelimiter.RateLimiter) APIRequester {
	return &APIRequesterImpl{
		client:      &http.Client{Timeout: 30 * time.Second},
		token:       token,
		rateLimiter: rateLimiter,
	}
}

func (api *APIRequesterImpl) BaseReq(method, endpoint string, body any, result any) error {
	route := method + endpoint

	if api.rateLimiter.IsRateLimited(route) {
		waitTime := api.rateLimiter.RetryAfter(route)
		time.Sleep(waitTime)
	}

	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("[Requester] failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, APIBaseURL+endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("[Requester] failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+api.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "Vocalize")

	resp, err := api.client.Do(req)
	if err != nil {
		return fmt.Errorf("[Requester] failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("[Requester] failed to read response body: %w", err)
	}

	api.processRateLimitHeaders(route, resp.Header)

	// Handle rate limiting responses
	if resp.StatusCode == http.StatusTooManyRequests {
		var rateLimitResponse struct {
			Message    string  `json:"message"`
			RetryAfter float64 `json:"retry_after"`
			Global     bool    `json:"global"`
		}
		if err := json.Unmarshal(respBody, &rateLimitResponse); err != nil {
			return fmt.Errorf("[Requester] hit rate limit but failed to parse response: %w", err)
		}

		limit := &ratelimiter.RateLimit{
			Remaining:  0,
			ResetAfter: time.Duration(rateLimitResponse.RetryAfter * float64(time.Second)),
			Global:     rateLimitResponse.Global,
		}

		api.rateLimiter.UpdateLimit(route, limit)

		return fmt.Errorf("[Requester] rate limited: %s, retry after %.2f seconds",
			rateLimitResponse.Message, rateLimitResponse.RetryAfter)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("[Requester] API error: %d - %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("[Requester] failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (api *APIRequesterImpl) processRateLimitHeaders(route string, headers http.Header) {
	remaining, _ := strconv.Atoi(headers.Get("X-RateLimit-Remaining"))
	resetAfterStr := headers.Get("X-RateLimit-Reset-After")

	if resetAfterStr != "" {
		resetAfter, err := strconv.ParseFloat(resetAfterStr, 64)
		if err != nil {
			return
		}

		isGlobal := headers.Get("X-RateLimit-Global") == "true"

		limit := &ratelimiter.RateLimit{
			Remaining:  remaining,
			ResetAfter: time.Duration(resetAfter * float64(time.Second)),
			Global:     isGlobal,
		}

		bucket := headers.Get("X-RateLimit-Bucket")
		if bucket != "" {
			api.rateLimiter.UpdateLimit(bucket, limit)
		} else {
			api.rateLimiter.UpdateLimit(route, limit)
		}
	}
}
