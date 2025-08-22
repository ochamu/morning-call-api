package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ochamu/morning-call-api/internal/handler"
	"github.com/ochamu/morning-call-api/internal/infrastructure/auth"
	"github.com/ochamu/morning-call-api/internal/infrastructure/memory"
	authUC "github.com/ochamu/morning-call-api/internal/usecase/auth"
	morningCallUC "github.com/ochamu/morning-call-api/internal/usecase/morning_call"
	relationshipUC "github.com/ochamu/morning-call-api/internal/usecase/relationship"
	userUC "github.com/ochamu/morning-call-api/internal/usecase/user"
)

// TestServer はテスト用のHTTPサーバーをセットアップします
type TestServer struct {
	Server         *httptest.Server
	UserRepo       *memory.UserRepository
	MorningRepo    *memory.MorningCallRepository
	RelationRepo   *memory.RelationshipRepository
	PasswordService *auth.PasswordService
	SessionManager *auth.SessionManager
}

// NewTestServer はテスト用サーバーを初期化します
func NewTestServer(t *testing.T) *TestServer {
	// リポジトリの初期化
	userRepo := memory.NewUserRepository()
	morningCallRepo := memory.NewMorningCallRepository()
	relationshipRepo := memory.NewRelationshipRepository()
	
	// サービスの初期化
	passwordService := auth.NewPasswordService()
	sessionManager := auth.NewSessionManager(24 * time.Hour)

	// ユースケースの初期化
	authUseCase := authUC.NewAuthUseCase(userRepo, passwordService)
	userUseCase := userUC.NewUserUseCase(userRepo, passwordService)
	
	// モーニングコールユースケースの初期化
	createMorningCallUC := morningCallUC.NewCreateUseCase(morningCallRepo, userRepo, relationshipRepo)
	updateMorningCallUC := morningCallUC.NewUpdateUseCase(morningCallRepo, userRepo)
	deleteMorningCallUC := morningCallUC.NewDeleteUseCase(morningCallRepo)
	listMorningCallUC := morningCallUC.NewListUseCase(morningCallRepo, userRepo)
	confirmWakeUC := morningCallUC.NewConfirmWakeUseCase(morningCallRepo, userRepo)
	
	// 関係性ユースケースの初期化
	sendFriendRequestUC := relationshipUC.NewSendFriendRequestUseCase(relationshipRepo, userRepo)
	acceptFriendRequestUC := relationshipUC.NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	rejectFriendRequestUC := relationshipUC.NewRejectFriendRequestUseCase(relationshipRepo, userRepo)
	blockUserUC := relationshipUC.NewBlockUserUseCase(relationshipRepo, userRepo)
	blockRelationshipUC := relationshipUC.NewBlockRelationshipUseCase(relationshipRepo, userRepo)
	removeRelationshipUC := relationshipUC.NewRemoveRelationshipUseCase(relationshipRepo, userRepo)
	listFriendsUC := relationshipUC.NewListFriendsUseCase(relationshipRepo, userRepo)
	listFriendRequestsUC := relationshipUC.NewListFriendRequestsUseCase(relationshipRepo, userRepo)

	// Handlerの初期化
	authHandler := handler.NewAuthHandler(authUseCase, sessionManager)
	userHandler := handler.NewUserHandler(userUseCase, sessionManager)
	morningCallHandler := handler.NewMorningCallHandler(
		createMorningCallUC,
		updateMorningCallUC,
		deleteMorningCallUC,
		listMorningCallUC,
		confirmWakeUC,
		sessionManager,
	)
	relationshipHandler := handler.NewRelationshipHandler(
		sendFriendRequestUC,
		acceptFriendRequestUC,
		rejectFriendRequestUC,
		blockUserUC,
		blockRelationshipUC,
		removeRelationshipUC,
		listFriendsUC,
		listFriendRequestsUC,
		userUseCase,
		sessionManager,
	)

	// ルーターのセットアップ
	router := SetupTestRouter(
		authHandler,
		userHandler,
		morningCallHandler,
		relationshipHandler,
		sessionManager,
		userRepo,
	)

	// テストサーバーの作成
	ts := httptest.NewServer(router)

	return &TestServer{
		Server:         ts,
		UserRepo:       userRepo,
		MorningRepo:    morningCallRepo,
		RelationRepo:   relationshipRepo,
		PasswordService: passwordService,
		SessionManager: sessionManager,
	}
}

// Close はテストサーバーをクリーンアップします
func (ts *TestServer) Close() {
	ts.Server.Close()
}

// DoRequest はHTTPリクエストを実行します
func (ts *TestServer) DoRequest(method, path string, body interface{}, sessionID string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, ts.Server.URL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionID,
		})
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return client.Do(req)
}

// RegisterUser はテスト用ユーザーを登録します
func (ts *TestServer) RegisterUser(t *testing.T, username, email, password string) string {
	reqBody := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}

	resp, err := ts.DoRequest("POST", "/api/v1/users/register", reqBody, "")
	if err != nil {
		t.Fatalf("ユーザー登録リクエストエラー: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("ユーザー登録失敗: status=%d, body=%s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("レスポンスデコードエラー: %v", err)
	}

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatal("レスポンスにuserフィールドがありません")
	}

	userID, ok := user["id"].(string)
	if !ok {
		t.Fatal("ユーザーIDが取得できません")
	}

	return userID
}

// LoginUser はテスト用ユーザーでログインします
func (ts *TestServer) LoginUser(t *testing.T, username, password string) string {
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := ts.DoRequest("POST", "/api/v1/auth/login", reqBody, "")
	if err != nil {
		t.Fatalf("ログインリクエストエラー: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("ログイン失敗: status=%d, body=%s", resp.StatusCode, body)
	}

	// セッションIDをCookieから取得
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session_id" {
			return cookie.Value
		}
	}

	t.Fatal("セッションIDが取得できません")
	return ""
}

// AssertStatusCode はステータスコードを検証します
func AssertStatusCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("ステータスコードが一致しません: expected=%d, actual=%d", expected, actual)
	}
}

// AssertJSONResponse はJSONレスポンスを検証します
func AssertJSONResponse(t *testing.T, resp *http.Response, expectedKey string, expectedValue interface{}) {
	t.Helper()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSONデコードエラー: %v", err)
	}

	actualValue, exists := result[expectedKey]
	if !exists {
		t.Errorf("期待したキーが存在しません: %s", expectedKey)
		return
	}

	if expectedValue != nil {
		expected := fmt.Sprintf("%v", expectedValue)
		actual := fmt.Sprintf("%v", actualValue)
		if expected != actual {
			t.Errorf("値が一致しません: key=%s, expected=%v, actual=%v", expectedKey, expected, actual)
		}
	}
}

// CreateTestUsers は複数のテストユーザーを作成します
func (ts *TestServer) CreateTestUsers(t *testing.T, count int) []string {
	userIDs := make([]string, count)
	for i := 0; i < count; i++ {
		username := fmt.Sprintf("testuser%d", i+1)
		email := fmt.Sprintf("test%d@example.com", i+1)
		password := "Password123!"
		userIDs[i] = ts.RegisterUser(t, username, email, password)
	}
	return userIDs
}