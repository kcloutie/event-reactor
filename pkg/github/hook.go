package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type Hook struct {
	Signature string
	Payload   []byte
}

const signaturePrefix = "sha256="
const prefixLength = len(signaturePrefix)
const signatureLength = prefixLength + (sha256.Size * 2)

func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha256.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}

func (h *Hook) SignedBy(secret []byte) bool {
	if len(h.Signature) != signatureLength || !strings.HasPrefix(h.Signature, signaturePrefix) {
		return false
	}

	actual := make([]byte, sha256.Size)
	hex.Decode(actual, []byte(h.Signature[prefixLength:]))

	expected := signBody(secret, h.Payload)

	return hmac.Equal(expected, actual)
}

func (h *Hook) Extract(dst interface{}) error {
	return json.Unmarshal(h.Payload, dst)
}

func NewHookRequest(req *http.Request) (hook *Hook, err error) {
	hook = new(Hook)
	if !strings.EqualFold(req.Method, "POST") {
		return nil, errors.New("unknown method")
	}

	if hook.Signature = req.Header.Get("x-hub-signature-256"); len(hook.Signature) == 0 {
		return nil, errors.New("no signature")
	}

	hook.Payload, err = io.ReadAll(req.Body)
	return
}

func Parse(secret []byte, req *http.Request) (hook *Hook, err error) {
	hook, err = NewHookRequest(req)
	if err == nil && !hook.SignedBy(secret) {
		err = errors.New("invalid signature")
	}
	return
}
