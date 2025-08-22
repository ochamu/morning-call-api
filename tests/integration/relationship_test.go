package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestFriendRequestFlow(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// テストユーザーの作成
	user1ID := ts.RegisterUser(t, "user1", "user1@example.com", "Password123!")
	user2ID := ts.RegisterUser(t, "user2", "user2@example.com", "Password123!")
	ts.RegisterUser(t, "user3", "user3@example.com", "Password123!")

	session1 := ts.LoginUser(t, "user1", "Password123!")
	session2 := ts.LoginUser(t, "user2", "Password123!")

	// 最初に友達リクエストを送信（後続のテストケースで使用）
	var relationshipID string
	t.Run("友達リクエスト送信", func(t *testing.T) {
		// user1からuser2へリクエスト送信
		reqBody := map[string]string{
			"receiver_id": user2ID,
		}

		resp, err := ts.DoRequest("POST", "/api/v1/relationships/request", reqBody, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("JSONデコードエラー: %v", err)
		}

		relationship, ok := result["relationship"].(map[string]interface{})
		if !ok {
			t.Fatal("relationshipフィールドが存在しません")
		}

		// 後のテストで使用するためIDを保存
		if id, ok := relationship["id"].(string); ok {
			relationshipID = id
			t.Logf("作成されたrelationshipID: %s", relationshipID)
		}

		if relationship["requester_id"] != user1ID {
			t.Errorf("リクエスターIDが不正: expected=%s, actual=%v", user1ID, relationship["requester_id"])
		}
		if relationship["receiver_id"] != user2ID {
			t.Errorf("レシーバーIDが不正: expected=%s, actual=%v", user2ID, relationship["receiver_id"])
		}
		if relationship["status"] != "pending" {
			t.Errorf("ステータスが不正: expected=pending, actual=%v", relationship["status"])
		}
	})

	t.Run("重複リクエストエラー", func(t *testing.T) {
		// 同じリクエストを再度送信
		reqBody := map[string]string{
			"receiver_id": user2ID,
		}

		resp, err := ts.DoRequest("POST", "/api/v1/relationships/request", reqBody, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("自分自身へのリクエストエラー", func(t *testing.T) {
		reqBody := map[string]string{
			"receiver_id": user1ID,
		}

		resp, err := ts.DoRequest("POST", "/api/v1/relationships/request", reqBody, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("リクエスト一覧取得", func(t *testing.T) {
		// user2がリクエスト一覧を取得
		resp, err := ts.DoRequest("GET", "/api/v1/relationships/requests", nil, session2)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("JSONデコードエラー: %v", err)
		}

		requests, ok := result["requests"].([]interface{})
		if !ok {
			t.Fatalf("requestsフィールドが存在しません: %v", result)
		}

		if len(requests) != 1 {
			t.Errorf("リクエスト数が不正: expected=1, actual=%d", len(requests))
		}

		if len(requests) > 0 {
			req := requests[0].(map[string]interface{})
			if req["requester_id"] != user1ID {
				t.Errorf("リクエスターIDが不正: expected=%s, actual=%v", user1ID, req["requester_id"])
			}
		}
	})

	t.Run("友達リクエスト承認", func(t *testing.T) {
		// まず既存のリクエストを取得してIDを確認
		resp, err := ts.DoRequest("GET", "/api/v1/relationships/requests", nil, session2)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("JSONデコードエラー: %v", err)
		}
		
		requests, ok := result["requests"].([]interface{})
		if !ok || len(requests) == 0 {
			t.Fatal("リクエストが取得できません")
		}
		
		req := requests[0].(map[string]interface{})
		reqID, ok := req["id"].(string)
		if !ok {
			t.Fatal("リクエストIDが取得できません")
		}
		relationshipID = reqID

		// リクエストを承認
		acceptResp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/relationships/%s/accept", relationshipID), nil, session2)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer acceptResp.Body.Close()

		AssertStatusCode(t, http.StatusOK, acceptResp.StatusCode)

		var acceptResult map[string]interface{}
		if err := json.NewDecoder(acceptResp.Body).Decode(&acceptResult); err != nil {
			t.Fatalf("JSONデコードエラー: %v", err)
		}

		relationship, ok := acceptResult["relationship"].(map[string]interface{})
		if !ok {
			t.Fatal("relationshipフィールドが存在しません")
		}

		if relationship["status"] != "accepted" {
			t.Errorf("ステータスが不正: expected=accepted, actual=%v", relationship["status"])
		}
	})

	t.Run("友達リスト取得", func(t *testing.T) {
		// user1の友達リスト
		resp1, err := ts.DoRequest("GET", "/api/v1/relationships/friends", nil, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp1.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp1.StatusCode)

		var result1 map[string]interface{}
		json.NewDecoder(resp1.Body).Decode(&result1)
		friends1 := result1["friends"].([]interface{})

		if len(friends1) != 1 {
			t.Errorf("user1の友達数が不正: expected=1, actual=%d", len(friends1))
		}

		// user2の友達リスト
		resp2, err := ts.DoRequest("GET", "/api/v1/relationships/friends", nil, session2)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp2.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp2.StatusCode)

		var result2 map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&result2)
		friends2 := result2["friends"].([]interface{})

		if len(friends2) != 1 {
			t.Errorf("user2の友達数が不正: expected=1, actual=%d", len(friends2))
		}
	})
}

func TestBlockAndDelete(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// テストユーザーの作成
	_ = ts.RegisterUser(t, "blockuser1", "block1@example.com", "Password123!")
	user2ID := ts.RegisterUser(t, "blockuser2", "block2@example.com", "Password123!")

	session1 := ts.LoginUser(t, "blockuser1", "Password123!")
	session2 := ts.LoginUser(t, "blockuser2", "Password123!")

	// 友達関係を作成
	reqBody := map[string]string{
		"receiver_id": user2ID,
	}
	resp, _ := ts.DoRequest("POST", "/api/v1/relationships/request", reqBody, session1)
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	relationshipID := result["relationship"].(map[string]interface{})["id"].(string)

	// リクエストを承認
	ts.DoRequest("PUT", fmt.Sprintf("/api/v1/relationships/%s/accept", relationshipID), nil, session2)

	t.Run("ユーザーをブロック", func(t *testing.T) {
		// user1がuser2をブロック
		blockResp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/relationships/%s/block", relationshipID), nil, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer blockResp.Body.Close()

		AssertStatusCode(t, http.StatusOK, blockResp.StatusCode)

		var blockResult map[string]interface{}
		if err := json.NewDecoder(blockResp.Body).Decode(&blockResult); err != nil {
			t.Fatalf("JSONデコードエラー: %v", err)
		}
		
		relationship, ok := blockResult["relationship"].(map[string]interface{})
		if !ok {
			t.Fatal("relationshipフィールドが存在しません")
		}

		if relationship["status"] != "blocked" {
			t.Errorf("ステータスが不正: expected=blocked, actual=%v", relationship["status"])
		}
	})

	t.Run("ブロック後の友達リスト", func(t *testing.T) {
		// user1の友達リスト（ブロックしたので0件のはず）
		resp, err := ts.DoRequest("GET", "/api/v1/relationships/friends", nil, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("JSONデコードエラー: %v", err)
		}
		friends, ok := result["friends"].([]interface{})
		if !ok {
			t.Fatal("friendsフィールドが存在しません")
		}

		if len(friends) != 0 {
			t.Errorf("ブロック後に友達が残っています: actual=%d", len(friends))
		}
	})

	t.Run("関係を削除", func(t *testing.T) {
		// 新しい友達関係を作成
		user3ID := ts.RegisterUser(t, "deleteuser3", "delete3@example.com", "Password123!")
		session3 := ts.LoginUser(t, "deleteuser3", "Password123!")

		reqBody := map[string]string{
			"receiver_id": user3ID,
		}
		createResp, _ := ts.DoRequest("POST", "/api/v1/relationships/request", reqBody, session1)
		defer createResp.Body.Close()

		var createResult map[string]interface{}
		if err := json.NewDecoder(createResp.Body).Decode(&createResult); err != nil {
			t.Fatalf("JSONデコードエラー: %v", err)
		}
		relationship, ok := createResult["relationship"].(map[string]interface{})
		if !ok {
			t.Fatal("relationshipフィールドが存在しません")
		}
		newRelationshipID, ok := relationship["id"].(string)
		if !ok {
			t.Fatal("relationshipIDが取得できません")
		}

		// 承認
		ts.DoRequest("PUT", fmt.Sprintf("/api/v1/relationships/%s/accept", newRelationshipID), nil, session3)

		// 関係を削除
		deleteResp, err := ts.DoRequest("DELETE", fmt.Sprintf("/api/v1/relationships/%s", newRelationshipID), nil, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer deleteResp.Body.Close()

		AssertStatusCode(t, http.StatusOK, deleteResp.StatusCode)

		// 削除後の友達リスト確認
		checkResp, _ := ts.DoRequest("GET", "/api/v1/relationships/friends", nil, session1)
		defer checkResp.Body.Close()

		var checkResult map[string]interface{}
		json.NewDecoder(checkResp.Body).Decode(&checkResult)
		friends := checkResult["friends"].([]interface{})

		if len(friends) != 0 {
			t.Errorf("削除後に友達が残っています: actual=%d", len(friends))
		}
	})
}

func TestRelationshipAuthorization(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// テストユーザーの作成
	_ = ts.RegisterUser(t, "authuser1", "auth1@example.com", "Password123!")
	user2ID := ts.RegisterUser(t, "authuser2", "auth2@example.com", "Password123!")
	ts.RegisterUser(t, "authuser3", "auth3@example.com", "Password123!")

	session1 := ts.LoginUser(t, "authuser1", "Password123!")
	session2 := ts.LoginUser(t, "authuser2", "Password123!")
	session3 := ts.LoginUser(t, "authuser3", "Password123!")

	// user1からuser2へリクエスト送信
	reqBody := map[string]string{
		"receiver_id": user2ID,
	}
	resp, _ := ts.DoRequest("POST", "/api/v1/relationships/request", reqBody, session1)
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSONデコードエラー: %v", err)
	}
	relationship, ok := result["relationship"].(map[string]interface{})
	if !ok {
		t.Fatal("relationshipフィールドが存在しません")
	}
	relationshipID, ok := relationship["id"].(string)
	if !ok {
		t.Fatal("relationshipIDが取得できません")
	}

	t.Run("関係のない第三者による承認は失敗", func(t *testing.T) {
		// user3が承認しようとする（失敗するはず）
		acceptResp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/relationships/%s/accept", relationshipID), nil, session3)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer acceptResp.Body.Close()

		AssertStatusCode(t, http.StatusForbidden, acceptResp.StatusCode)
	})

	t.Run("リクエスト送信者による承認は失敗", func(t *testing.T) {
		// user1（送信者）が承認しようとする（失敗するはず）
		acceptResp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/relationships/%s/accept", relationshipID), nil, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer acceptResp.Body.Close()

		AssertStatusCode(t, http.StatusForbidden, acceptResp.StatusCode)
	})

	t.Run("正当な受信者による承認は成功", func(t *testing.T) {
		// user2（受信者）が承認
		acceptResp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/relationships/%s/accept", relationshipID), nil, session2)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer acceptResp.Body.Close()

		AssertStatusCode(t, http.StatusOK, acceptResp.StatusCode)
	})
}