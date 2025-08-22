package integration

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestAuthFlow(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	t.Run("ユーザー登録", func(t *testing.T) {
		// 正常な登録
		reqBody := map[string]string{
			"username": "newuser",
			"email":    "newuser@example.com",
			"password": "Password123!",
		}

		resp, err := ts.DoRequest("POST", "/api/v1/users/register", reqBody, "")
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusCreated, resp.StatusCode)

		// 重複ユーザー名でのエラー
		resp2, err := ts.DoRequest("POST", "/api/v1/users/register", reqBody, "")
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp2.Body.Close()

		AssertStatusCode(t, http.StatusConflict, resp2.StatusCode)
	})

	t.Run("ログイン", func(t *testing.T) {
		// ユーザー登録
		ts.RegisterUser(t, "loginuser", "login@example.com", "Password123!")

		// 正常なログイン
		loginReq := map[string]string{
			"username": "loginuser",
			"password": "Password123!",
		}

		resp, err := ts.DoRequest("POST", "/api/v1/auth/login", loginReq, "")
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		// セッションIDの確認
		var sessionID string
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "session_id" {
				sessionID = cookie.Value
				break
			}
		}
		if sessionID == "" {
			t.Error("セッションIDが設定されていません")
		}

		// 間違ったパスワードでのログイン
		wrongLoginReq := map[string]string{
			"username": "loginuser",
			"password": "WrongPassword!",
		}

		resp2, err := ts.DoRequest("POST", "/api/v1/auth/login", wrongLoginReq, "")
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp2.Body.Close()

		AssertStatusCode(t, http.StatusUnauthorized, resp2.StatusCode)
	})

	t.Run("ログアウト", func(t *testing.T) {
		// ユーザー登録とログイン
		ts.RegisterUser(t, "logoutuser", "logout@example.com", "Password123!")
		sessionID := ts.LoginUser(t, "logoutuser", "Password123!")

		// ログアウト
		resp, err := ts.DoRequest("POST", "/api/v1/auth/logout", nil, sessionID)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		// ログアウト後のアクセス確認
		resp2, err := ts.DoRequest("GET", "/api/v1/users/me", nil, sessionID)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp2.Body.Close()

		AssertStatusCode(t, http.StatusUnauthorized, resp2.StatusCode)
	})

	t.Run("セッション検証", func(t *testing.T) {
		// ユーザー登録とログイン
		ts.RegisterUser(t, "sessionuser", "session@example.com", "Password123!")
		sessionID := ts.LoginUser(t, "sessionuser", "Password123!")

		// デバッグ: セッションIDの確認
		t.Logf("取得したセッションID: %s", sessionID)

		// 有効なセッションでのアクセス
		resp, err := ts.DoRequest("GET", "/api/v1/users/me", nil, sessionID)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		// デバッグ: レスポンスボディの確認
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("レスポンスボディ: %s", body)
		}

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		// 無効なセッションでのアクセス
		resp2, err := ts.DoRequest("GET", "/api/v1/users/me", nil, "invalid-session")
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp2.Body.Close()

		AssertStatusCode(t, http.StatusUnauthorized, resp2.StatusCode)
	})
}

func TestPasswordValidation(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	testCases := []struct {
		name       string
		password   string
		shouldFail bool
	}{
		{"弱いパスワード（小文字のみ）", "password", true},
		{"弱いパスワード（数字なし）", "Password!", true},
		{"弱いパスワード（特殊文字なし）", "Password123", true},
		{"弱いパスワード（大文字なし）", "password123!", true},
		{"短すぎるパスワード", "P@ss1", true},
		{"長すぎるパスワード", "P@ssword12345678901234567890123456789012345678901234567890123456789012345", true},
		{"強いパスワード", "Password123!", false},
		{"強いパスワード（別パターン）", "MyP@ssw0rd", false},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := map[string]string{
				"username": fmt.Sprintf("user%d", i),
				"email":    fmt.Sprintf("test%d@example.com", i),
				"password": tc.password,
			}

			resp, err := ts.DoRequest("POST", "/api/v1/users/register", reqBody, "")
			if err != nil {
				t.Fatalf("リクエストエラー: %v", err)
			}
			defer resp.Body.Close()

			if tc.shouldFail {
				if resp.StatusCode == http.StatusCreated {
					t.Errorf("弱いパスワードが受け入れられました: %s", tc.password)
				}
			} else {
				if resp.StatusCode != http.StatusCreated {
					body, _ := io.ReadAll(resp.Body)
					t.Errorf("強いパスワードが拒否されました: %s, status=%d, body=%s", tc.password, resp.StatusCode, body)
				}
			}
		})
	}
}