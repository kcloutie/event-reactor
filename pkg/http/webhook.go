package http

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

type WebhookConfig struct {
	Log               *zap.Logger
	MaxRetries        int
	Url               string
	Body              string
	HookSecret        string
	SignatureHeader   string
	AdditionalHeaders map[string]string
	Token             string
	TokenType         string
}

// signBody signs the body with the secret using HMAC-SHA256
func SignBody(secret, body []byte) []byte {
	computed := hmac.New(sha256.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}

func (c *WebhookConfig) SendWebhook() error {
	signature := []byte{}
	if c.HookSecret != "" {
		signature = SignBody([]byte(c.HookSecret), []byte(c.Body))
	}

	retryClient := NewHttpRetryClient(c.Log, c.MaxRetries)
	httpReq, err := http.NewRequest("POST", c.Url, bytes.NewBuffer([]byte(c.Body)))
	if err != nil {
		return err
	}
	req := &retryablehttp.Request{
		Request: httpReq,
	}

	if len(signature) > 0 {
		req.Header.Add(c.SignatureHeader, hex.EncodeToString(signature))
	}
	for key, value := range c.AdditionalHeaders {
		req.Header.Add(key, value)
	}
	if c.Token != "" {
		tokenType := c.TokenType
		if tokenType == "" {
			tokenType = "Bearer"
		}
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", tokenType, c.Token))
	}

	response, err := retryClient.Do(req)

	if err != nil {
		return fmt.Errorf("failed to make the request: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read the response body: %v", err)
	}

	if response.StatusCode <= 199 || response.StatusCode >= 400 {
		return fmt.Errorf(CreateHttpErrorMessage(c.Url, string(body), response.StatusCode))
	}

	return nil
}
