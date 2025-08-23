package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestUserProfile(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	t.Run("プロフィール取得", func(t *testing.T) {
		// ユーザー登録とログイン
		userID := ts.RegisterUser(t, "profileuser", "profile@example.com", "Password123!")
		sessionID := ts.LoginUser(t, "profileuser", "Password123!")

		// プロフィール取得
		resp, err := ts.DoRequest("GET", "/api/v1/users/me", nil, sessionID)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("JSONデコードエラー: %v", err)
		}

		user, ok := result["user"].(map[string]interface{})
		if !ok {
			t.Fatal("userフィールドが存在しません")
		}

		if user["id"] != userID {
			t.Errorf("ユーザーIDが一致しません: expected=%s, actual=%v", userID, user["id"])
		}
		if user["username"] != "profileuser" {
			t.Errorf("ユーザー名が一致しません: actual=%v", user["username"])
		}
		if user["email"] != "profile@example.com" {
			t.Errorf("メールアドレスが一致しません: actual=%v", user["email"])
		}
	})

	t.Run("未認証でのプロフィール取得", func(t *testing.T) {
		resp, err := ts.DoRequest("GET", "/api/v1/users/me", nil, "")
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestUserSearch(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// テストユーザーの作成
	ts.RegisterUser(t, "searchuser1", "search1@example.com", "Password123!")
	ts.RegisterUser(t, "searchuser2", "search2@example.com", "Password123!")
	ts.RegisterUser(t, "differentuser", "different@example.com", "Password123!")
	sessionID := ts.LoginUser(t, "searchuser1", "Password123!")

	t.Run("ユーザー名で検索", func(t *testing.T) {
		resp, err := ts.DoRequest("GET", "/api/v1/users/search?query=searchuser", nil, sessionID)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("JSONデコードエラー: %v", err)
		}

		users, ok := result["users"].([]interface{})
		if !ok {
			t.Fatal("usersフィールドが存在しません")
		}

		// searchuser1とsearchuser2の2件が返されるはず（自分を除く）
		if len(users) != 1 {
			t.Errorf("検索結果数が不正: expected=1, actual=%d", len(users))
		}

		// searchuser2が結果に含まれていることを確認
		found := false
		for _, u := range users {
			user := u.(map[string]interface{})
			if user["username"] == "searchuser2" {
				found = true
				break
			}
		}
		if !found {
			t.Error("searchuser2が検索結果に含まれていません")
		}
	})

	t.Run("部分一致検索", func(t *testing.T) {
		resp, err := ts.DoRequest("GET", "/api/v1/users/search?query=user", nil, sessionID)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("JSONデコードエラー: %v", err)
		}

		users, ok := result["users"].([]interface{})
		if !ok {
			t.Fatal("usersフィールドが存在しません")
		}

		// 自分以外の全ユーザーが返されるはず
		if len(users) < 2 {
			t.Errorf("検索結果数が少なすぎます: actual=%d", len(users))
		}
	})

	t.Run("空のクエリ", func(t *testing.T) {
		resp, err := ts.DoRequest("GET", "/api/v1/users/search?query=", nil, sessionID)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("未認証での検索", func(t *testing.T) {
		resp, err := ts.DoRequest("GET", "/api/v1/users/search?query=test", nil, "")
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestUserRegistrationValidation(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	testCases := []struct {
		name        string
		username    string
		email       string
		password    string
		expectedErr bool
	}{
		{"正常な登録", "validuser", "valid@example.com", "Password123!", false},
		{"ユーザー名が短すぎる", "ab", "short@example.com", "Password123!", true},
		{"ユーザー名が長すぎる", "verylongusernamethatshouldnotbeallowed", "long@example.com", "Password123!", true},
		{"無効なメールアドレス", "invalidemail", "notanemail", "Password123!", true},
		{"メールアドレスなし", "noemail", "", "Password123!", true},
		{"ユーザー名なし", "", "nouser@example.com", "Password123!", true},
		{"パスワードなし", "nopassword", "nopass@example.com", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := map[string]string{
				"username": tc.username,
				"email":    tc.email,
				"password": tc.password,
			}

			resp, err := ts.DoRequest("POST", "/api/v1/users/register", reqBody, "")
			if err != nil {
				t.Fatalf("リクエストエラー: %v", err)
			}
			defer resp.Body.Close()

			if tc.expectedErr {
				if resp.StatusCode == http.StatusCreated {
					t.Errorf("無効なデータが受け入れられました")
				}
			} else {
				if resp.StatusCode != http.StatusCreated {
					t.Errorf("有効なデータが拒否されました: status=%d", resp.StatusCode)
				}
			}
		})
	}
}