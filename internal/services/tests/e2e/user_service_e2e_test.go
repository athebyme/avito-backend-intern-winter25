package e2e

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"errors"`
}

func TestLoginE2E(t *testing.T) {
	baseURL := "http://avito-shop-service-test:8080"
	waitForService(t, baseURL+"/health")

	t.Run("Successful first time login", func(t *testing.T) {
		loginReq := LoginRequest{
			Username: "newuser1",
			Password: "password123",
		}

		resp, body := performLogin(t, baseURL+"/api/auth", loginReq)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var loginResp LoginResponse
		err := json.Unmarshal(body, &loginResp)
		require.NoError(t, err)

		assert.NotEmpty(t, loginResp.Token)

		balanceResp, balanceBody := performAuthorizedRequest(t,
			baseURL+"/api/balance",
			"GET",
			nil,
			loginResp.Token,
		)
		require.Equal(t, http.StatusOK, balanceResp.StatusCode)

		var balanceData struct {
			Coins int `json:"balance"`
		}
		err = json.Unmarshal(balanceBody, &balanceData)
		require.NoError(t, err)
		assert.Equal(t, 1000, balanceData.Coins)
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		loginReq := LoginRequest{
			Username: "newuser2",
			Password: "correctpass",
		}

		_, _ = performLogin(t, baseURL+"/api/auth", loginReq)

		wrongPassReq := LoginRequest{
			Username: "newuser2",
			Password: "wrongpass",
		}

		resp, body := performLogin(t, baseURL+"/api/auth", wrongPassReq)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var errorResp ErrorResponse
		err := json.Unmarshal(body, &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Error, "invalid password")
	})

	t.Run("Subsequent login returns same user", func(t *testing.T) {
		loginReq := LoginRequest{
			Username: "newuser3",
			Password: "password123",
		}

		resp1, body1 := performLogin(t, baseURL+"/api/auth", loginReq)
		require.Equal(t, http.StatusOK, resp1.StatusCode)

		var loginResp1 LoginResponse
		err := json.Unmarshal(body1, &loginResp1)
		require.NoError(t, err)

		resp2, body2 := performLogin(t, baseURL+"/api/auth", loginReq)
		require.Equal(t, http.StatusOK, resp2.StatusCode)

		var loginResp2 LoginResponse
		err = json.Unmarshal(body2, &loginResp2)
		require.NoError(t, err)

	})
}

func waitForService(t *testing.T, healthURL string) {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if err == nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("Service didn't become healthy within deadline")
}

func performLogin(t *testing.T, url string, req LoginRequest) (*http.Response, []byte) {
	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, body
}

func performAuthorizedRequest(t *testing.T, url, method string, body []byte, token string) (*http.Response, []byte) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}
