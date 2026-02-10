package sheets

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

type serviceAccountKey struct {
	Type         string `json:"type"`
	ClientEmail  string `json:"client_email"`
	PrivateKey   string `json:"private_key"`
	TokenURI     string `json:"token_uri"`
}

type tokenSource struct {
	key      serviceAccountKey
	privKey  *rsa.PrivateKey
	mu       sync.Mutex
	token    string
	expiry   time.Time
}

func newTokenSource(credentialsFile string) (*tokenSource, error) {
	data, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("認証ファイルの読み込みに失敗: %w", err)
	}

	var key serviceAccountKey
	if err := json.Unmarshal(data, &key); err != nil {
		return nil, fmt.Errorf("認証ファイルの解析に失敗: %w", err)
	}

	if key.TokenURI == "" {
		key.TokenURI = "https://oauth2.googleapis.com/token"
	}

	block, _ := pem.Decode([]byte(key.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("秘密鍵のデコードに失敗")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("秘密鍵の解析に失敗: %w", err)
	}

	rsaKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("RSA秘密鍵ではありません")
	}

	return &tokenSource{key: key, privKey: rsaKey}, nil
}

func (ts *tokenSource) getToken() (string, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.token != "" && time.Now().Before(ts.expiry) {
		return ts.token, nil
	}

	now := time.Now()
	claims := map[string]interface{}{
		"iss":   ts.key.ClientEmail,
		"scope": "https://www.googleapis.com/auth/spreadsheets",
		"aud":   ts.key.TokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	}

	jwt, err := signJWT(claims, ts.privKey)
	if err != nil {
		return "", fmt.Errorf("JWT署名に失敗: %w", err)
	}

	resp, err := http.PostForm(ts.key.TokenURI, url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {jwt},
	})
	if err != nil {
		return "", fmt.Errorf("トークン取得に失敗: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("トークン応答の読み込みに失敗: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("トークン取得失敗 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("トークン応答の解析に失敗: %w", err)
	}

	ts.token = tokenResp.AccessToken
	ts.expiry = now.Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return ts.token, nil
}

func signJWT(claims map[string]interface{}, key *rsa.PrivateKey) (string, error) {
	header := base64URLEncode([]byte(`{"alg":"RS256","typ":"JWT"}`))

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64URLEncode(claimsJSON)

	signingInput := header + "." + payload
	h := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}

	return signingInput + "." + base64URLEncode(sig), nil
}

func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}
