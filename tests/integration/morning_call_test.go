package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestMorningCallCRUD(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// テストユーザーの作成と友達関係の確立
	user1ID := ts.RegisterUser(t, "mcuser1", "mc1@example.com", "Password123!")
	user2ID := ts.RegisterUser(t, "mcuser2", "mc2@example.com", "Password123!")
	
	session1 := ts.LoginUser(t, "mcuser1", "Password123!")
	session2 := ts.LoginUser(t, "mcuser2", "Password123!")

	// 友達関係を作成
	reqBody := map[string]string{
		"receiver_id": user2ID,
	}
	relResp, _ := ts.DoRequest("POST", "/api/v1/relationships/request", reqBody, session1)
	defer relResp.Body.Close()
	
	var relResult map[string]interface{}
	if err := json.NewDecoder(relResp.Body).Decode(&relResult); err != nil {
		t.Fatalf("友達リクエストレスポンスのデコードエラー: %v", err)
	}
	
	// relationshipのレスポンス構造を確認
	var relationshipID string
	if relationship, ok := relResult["relationship"].(map[string]interface{}); ok {
		if id, ok := relationship["id"].(string); ok {
			relationshipID = id
		} else {
			t.Fatalf("relationshipのIDが取得できません: %v", relationship)
		}
	} else if id, ok := relResult["id"].(string); ok {
		relationshipID = id
	} else {
		t.Fatalf("relationshipIDが取得できません: %v", relResult)
	}
	
	// リクエストを承認
	acceptResp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/relationships/%s/accept", relationshipID), nil, session2)
	if err != nil {
		t.Fatalf("友達リクエスト承認エラー: %v", err)
	}
	defer acceptResp.Body.Close()
	
	// 承認が成功したか確認
	if acceptResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(acceptResp.Body)
		t.Fatalf("友達リクエスト承認失敗: status=%d, body=%s", acceptResp.StatusCode, body)
	}

	// モーニングコールIDを保持する変数
	var morningCallID string

	t.Run("モーニングコール作成", func(t *testing.T) {
		// 明日の朝7時に設定
		tomorrow := time.Now().AddDate(0, 0, 1)
		wakeTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 7, 0, 0, 0, time.Local)

		createReq := map[string]interface{}{
			"receiver_id":    user2ID,
			"scheduled_time": wakeTime.Format(time.RFC3339),
			"message":        "おはよう！今日も一日頑張ろう！",
		}

		resp, err := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		// エラーレスポンスを確認
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("エラーレスポンス: %s", body)
		}

		AssertStatusCode(t, http.StatusCreated, resp.StatusCode)

		var morningCall map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&morningCall); err != nil {
			t.Fatalf("レスポンスのパースに失敗: %v", err)
		}

		morningCallID = morningCall["id"].(string)
		t.Logf("作成されたモーニングコールID: %s", morningCallID)
		if morningCall["sender_id"] != user1ID {
			t.Errorf("送信者IDが不正: expected=%s, actual=%v", user1ID, morningCall["sender_id"])
		}
		if morningCall["receiver_id"] != user2ID {
			t.Errorf("受信者IDが不正: expected=%s, actual=%v", user2ID, morningCall["receiver_id"])
		}
		if morningCall["message"] != "おはよう！今日も一日頑張ろう！" {
			t.Errorf("メッセージが不正: actual=%v", morningCall["message"])
		}
		if morningCall["status"] != "scheduled" {
			t.Errorf("ステータスが不正: expected=scheduled, actual=%v", morningCall["status"])
		}
	})

	t.Run("モーニングコール取得", func(t *testing.T) {
		if morningCallID == "" {
			t.Skip("モーニングコールIDが設定されていません")
		}
		
		resp, err := ts.DoRequest("GET", fmt.Sprintf("/api/v1/morning-calls/%s", morningCallID), nil, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var morningCall map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&morningCall); err != nil {
			t.Fatalf("レスポンスのパースに失敗: %v", err)
		}

		if morningCall["id"] != morningCallID {
			t.Errorf("IDが不正: expected=%s, actual=%v", morningCallID, morningCall["id"])
		}
	})

	t.Run("モーニングコール更新", func(t *testing.T) {
		// 時間とメッセージを更新
		tomorrow := time.Now().AddDate(0, 0, 1)
		newWakeTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 8, 0, 0, 0, time.Local)

		updateReq := map[string]interface{}{
			"scheduled_time": newWakeTime.Format(time.RFC3339),
			"message":   "更新されたメッセージです！",
		}

		resp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/morning-calls/%s", morningCallID), updateReq, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var morningCall map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&morningCall); err != nil {
			t.Fatalf("レスポンスのパースに失敗: %v", err)
		}

		if morningCall["message"] != "更新されたメッセージです！" {
			t.Errorf("メッセージが更新されていません: actual=%v", morningCall["message"])
		}
	})

	t.Run("送信済みモーニングコール一覧", func(t *testing.T) {
		resp, err := ts.DoRequest("GET", "/api/v1/morning-calls/sent", nil, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		morningCalls := result["morning_calls"].([]interface{})

		if len(morningCalls) != 1 {
			t.Errorf("送信済み数が不正: expected=1, actual=%d", len(morningCalls))
		}
	})

	t.Run("受信モーニングコール一覧", func(t *testing.T) {
		resp, err := ts.DoRequest("GET", "/api/v1/morning-calls/received", nil, session2)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		morningCalls := result["morning_calls"].([]interface{})

		if len(morningCalls) != 1 {
			t.Errorf("受信数が不正: expected=1, actual=%d", len(morningCalls))
		}
	})

	t.Run("起床確認", func(t *testing.T) {
		// user2が起床確認
		confirmURL := fmt.Sprintf("/api/v1/morning-calls/%s/confirm", morningCallID)
		t.Logf("起床確認URL: %s", confirmURL)
		
		resp, err := ts.DoRequest("PUT", confirmURL, nil, session2)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		// エラーレスポンスを確認
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("ステータスコード: %d", resp.StatusCode)
			t.Logf("エラーレスポンス: %s", string(body))
			if resp.StatusCode == 301 || resp.StatusCode == 302 {
				t.Logf("リダイレクト先: %s", resp.Header.Get("Location"))
			}
		}

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var morningCall map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&morningCall); err != nil {
			t.Fatalf("レスポンスのパースに失敗: %v", err)
		}

		if morningCall["status"] != "confirmed" {
			t.Errorf("ステータスが不正: expected=confirmed, actual=%v", morningCall["status"])
		}
		if morningCall["confirmed_at"] == nil {
			t.Error("確認時刻が設定されていません")
		}
	})

	t.Run("モーニングコール削除", func(t *testing.T) {
		// 新しいモーニングコールを作成
		tomorrow := time.Now().AddDate(0, 0, 1)
		wakeTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 6, 0, 0, 0, time.Local)

		createReq := map[string]interface{}{
			"receiver_id": user2ID,
			"scheduled_time":   wakeTime.Format(time.RFC3339),
			"message":     "削除テスト用",
		}

		createResp, err := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
		if err != nil {
			t.Fatalf("モーニングコール作成エラー: %v", err)
		}
		defer createResp.Body.Close()

		if createResp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(createResp.Body)
			t.Fatalf("モーニングコール作成失敗: status=%d, body=%s", createResp.StatusCode, body)
		}

		var morningCall map[string]interface{}
		if err := json.NewDecoder(createResp.Body).Decode(&morningCall); err != nil {
			t.Fatalf("レスポンスのパースに失敗: %v", err)
		}
		
		if morningCall["id"] == nil {
			t.Fatalf("モーニングコールIDが取得できません: %v", morningCall)
		}
		deleteID := morningCall["id"].(string)

		// 削除
		deleteResp, err := ts.DoRequest("DELETE", fmt.Sprintf("/api/v1/morning-calls/%s", deleteID), nil, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer deleteResp.Body.Close()

		AssertStatusCode(t, http.StatusOK, deleteResp.StatusCode)

		// 削除確認
		getResp, _ := ts.DoRequest("GET", fmt.Sprintf("/api/v1/morning-calls/%s", deleteID), nil, session1)
		defer getResp.Body.Close()

		AssertStatusCode(t, http.StatusNotFound, getResp.StatusCode)
	})
}

func TestMorningCallValidation(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// テストユーザーの作成
	user1ID := ts.RegisterUser(t, "valuser1", "val1@example.com", "Password123!")
	user2ID := ts.RegisterUser(t, "valuser2", "val2@example.com", "Password123!")
	user3ID := ts.RegisterUser(t, "valuser3", "val3@example.com", "Password123!")
	
	session1 := ts.LoginUser(t, "valuser1", "Password123!")

	// user1とuser2を友達にする
	reqBody := map[string]string{
		"receiver_id": user2ID,
	}
	relResp, _ := ts.DoRequest("POST", "/api/v1/relationships/request", reqBody, session1)
	defer relResp.Body.Close()
	
	var relResult map[string]interface{}
	if err := json.NewDecoder(relResp.Body).Decode(&relResult); err != nil {
		t.Fatalf("友達リクエストレスポンスのデコードエラー: %v", err)
	}
	
	// relationshipのレスポンス構造を確認
	var relationshipID string
	if relationship, ok := relResult["relationship"].(map[string]interface{}); ok {
		if id, ok := relationship["id"].(string); ok {
			relationshipID = id
		} else {
			t.Fatalf("relationshipのIDが取得できません: %v", relationship)
		}
	} else if id, ok := relResult["id"].(string); ok {
		relationshipID = id
	} else {
		t.Fatalf("relationshipIDが取得できません: %v", relResult)
	}
	
	session2 := ts.LoginUser(t, "valuser2", "Password123!")
	ts.DoRequest("PUT", fmt.Sprintf("/api/v1/relationships/%s/accept", relationshipID), nil, session2)

	t.Run("過去の時刻でのモーニングコール作成エラー", func(t *testing.T) {
		yesterday := time.Now().AddDate(0, 0, -1)

		createReq := map[string]interface{}{
			"receiver_id": user2ID,
			"scheduled_time":   yesterday.Format(time.RFC3339),
			"message":     "過去の時刻",
		}

		resp, err := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("30日以上先のモーニングコール作成エラー", func(t *testing.T) {
		farFuture := time.Now().AddDate(0, 0, 31)

		createReq := map[string]interface{}{
			"receiver_id": user2ID,
			"scheduled_time":   farFuture.Format(time.RFC3339),
			"message":     "遠すぎる未来",
		}

		resp, err := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("友達でないユーザーへのモーニングコール作成エラー", func(t *testing.T) {
		tomorrow := time.Now().AddDate(0, 0, 1)

		createReq := map[string]interface{}{
			"receiver_id": user3ID,
			"scheduled_time":   tomorrow.Format(time.RFC3339),
			"message":     "友達でない",
		}

		resp, err := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		// 友達でないのでエラーになるはず
		if resp.StatusCode == http.StatusCreated {
			t.Error("友達でないユーザーへのモーニングコールが作成されました")
		}
	})

	t.Run("自分自身へのモーニングコール作成エラー", func(t *testing.T) {
		tomorrow := time.Now().AddDate(0, 0, 1)

		createReq := map[string]interface{}{
			"receiver_id": user1ID,
			"scheduled_time":   tomorrow.Format(time.RFC3339),
			"message":     "自分宛て",
		}

		resp, err := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("空のメッセージでのモーニングコール作成", func(t *testing.T) {
		tomorrow := time.Now().AddDate(0, 0, 1)

		createReq := map[string]interface{}{
			"receiver_id": user2ID,
			"scheduled_time":   tomorrow.Format(time.RFC3339),
			"message":     "",
		}

		resp, err := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		// 空のメッセージは許可される（デフォルトメッセージが使われる）
		AssertStatusCode(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestMorningCallAuthorization(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// テストユーザーの作成
	_ = ts.RegisterUser(t, "authmc1", "authmc1@example.com", "Password123!")
	user2ID := ts.RegisterUser(t, "authmc2", "authmc2@example.com", "Password123!")
	_ = ts.RegisterUser(t, "authmc3", "authmc3@example.com", "Password123!")
	
	session1 := ts.LoginUser(t, "authmc1", "Password123!")
	session2 := ts.LoginUser(t, "authmc2", "Password123!")
	session3 := ts.LoginUser(t, "authmc3", "Password123!")

	// user1とuser2を友達にする
	reqBody := map[string]string{
		"receiver_id": user2ID,
	}
	relResp, _ := ts.DoRequest("POST", "/api/v1/relationships/request", reqBody, session1)
	defer relResp.Body.Close()
	
	var relResult map[string]interface{}
	if err := json.NewDecoder(relResp.Body).Decode(&relResult); err != nil {
		t.Fatalf("友達リクエストレスポンスのデコードエラー: %v", err)
	}
	
	// relationshipのレスポンス構造を確認
	var relationshipID string
	if relationship, ok := relResult["relationship"].(map[string]interface{}); ok {
		if id, ok := relationship["id"].(string); ok {
			relationshipID = id
		} else {
			t.Fatalf("relationshipのIDが取得できません: %v", relationship)
		}
	} else if id, ok := relResult["id"].(string); ok {
		relationshipID = id
	} else {
		t.Fatalf("relationshipIDが取得できません: %v", relResult)
	}
	
	ts.DoRequest("PUT", fmt.Sprintf("/api/v1/relationships/%s/accept", relationshipID), nil, session2)

	// user1からuser2へモーニングコールを作成（時刻をずらして作成）
	tomorrow := time.Now().AddDate(0, 0, 1)
	wakeTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 30, 0, 0, time.Local)
	createReq := map[string]interface{}{
		"receiver_id": user2ID,
		"scheduled_time":   wakeTime.Format(time.RFC3339),
		"message":     "権限テスト",
	}

	createResp, _ := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
	defer createResp.Body.Close()

	var createResult map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createResult)
	morningCallID := createResult["id"].(string)

	t.Run("第三者による更新は失敗", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"message": "不正な更新",
		}

		resp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/morning-calls/%s", morningCallID), updateReq, session3)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("受信者による更新は失敗", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"message": "受信者からの更新",
		}

		resp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/morning-calls/%s", morningCallID), updateReq, session2)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("送信者による削除は成功", func(t *testing.T) {
		// 新しいモーニングコールを作成（時刻をずらして作成）
		wakeTime2 := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.Local)
		createReq := map[string]interface{}{
			"receiver_id": user2ID,
			"scheduled_time":   wakeTime2.Format(time.RFC3339),
			"message":     "削除テスト",
		}

		createResp, err := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
		if err != nil {
			t.Fatalf("モーニングコール作成エラー: %v", err)
		}
		defer createResp.Body.Close()

		if createResp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(createResp.Body)
			t.Fatalf("モーニングコール作成失敗: status=%d, body=%s", createResp.StatusCode, body)
		}

		var morningCall map[string]interface{}
		if err := json.NewDecoder(createResp.Body).Decode(&morningCall); err != nil {
			t.Fatalf("レスポンスのパースに失敗: %v", err)
		}
		
		if morningCall["id"] == nil {
			t.Fatalf("モーニングコールIDが取得できません: %v", morningCall)
		}
		deleteID := morningCall["id"].(string)

		// 送信者が削除
		deleteResp, err := ts.DoRequest("DELETE", fmt.Sprintf("/api/v1/morning-calls/%s", deleteID), nil, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer deleteResp.Body.Close()

		AssertStatusCode(t, http.StatusOK, deleteResp.StatusCode)
	})

	t.Run("送信者による起床確認は失敗", func(t *testing.T) {
		// 新しいモーニングコールを作成（時刻をずらして作成）
		wakeTime3 := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 30, 0, 0, time.Local)
		createReq := map[string]interface{}{
			"receiver_id": user2ID,
			"scheduled_time":   wakeTime3.Format(time.RFC3339),
			"message":     "起床確認テスト用",
		}

		createResp, err := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
		if err != nil {
			t.Fatalf("モーニングコール作成エラー: %v", err)
		}
		defer createResp.Body.Close()

		if createResp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(createResp.Body)
			t.Fatalf("モーニングコール作成失敗: status=%d, body=%s", createResp.StatusCode, body)
		}

		var morningCall map[string]interface{}
		if err := json.NewDecoder(createResp.Body).Decode(&morningCall); err != nil {
			t.Fatalf("レスポンスのパースに失敗: %v", err)
		}
		
		if morningCall["id"] == nil {
			t.Fatalf("モーニングコールIDが取得できません: %v", morningCall)
		}
		confirmID := morningCall["id"].(string)
		
		// 送信者による起床確認（失敗するはず）
		resp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/morning-calls/%s/confirm", confirmID), nil, session1)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("受信者による起床確認は成功", func(t *testing.T) {
		// 新しいモーニングコールを作成（時刻をずらして作成）
		wakeTime4 := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 11, 0, 0, 0, time.Local)
		createReq := map[string]interface{}{
			"receiver_id": user2ID,
			"scheduled_time":   wakeTime4.Format(time.RFC3339),
			"message":     "起床確認成功テスト用",
		}

		createResp, err := ts.DoRequest("POST", "/api/v1/morning-calls", createReq, session1)
		if err != nil {
			t.Fatalf("モーニングコール作成エラー: %v", err)
		}
		defer createResp.Body.Close()

		if createResp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(createResp.Body)
			t.Fatalf("モーニングコール作成失敗: status=%d, body=%s", createResp.StatusCode, body)
		}

		var morningCall map[string]interface{}
		if err := json.NewDecoder(createResp.Body).Decode(&morningCall); err != nil {
			t.Fatalf("レスポンスのパースに失敗: %v", err)
		}
		
		if morningCall["id"] == nil {
			t.Fatalf("モーニングコールIDが取得できません: %v", morningCall)
		}
		confirmID := morningCall["id"].(string)
		
		// 受信者による起床確認（成功するはず）
		resp, err := ts.DoRequest("PUT", fmt.Sprintf("/api/v1/morning-calls/%s/confirm", confirmID), nil, session2)
		if err != nil {
			t.Fatalf("リクエストエラー: %v", err)
		}
		defer resp.Body.Close()

		AssertStatusCode(t, http.StatusOK, resp.StatusCode)
	})
}