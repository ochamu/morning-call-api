package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/handler/dto/request"
	"github.com/ochamu/morning-call-api/internal/handler/dto/response"
	"github.com/ochamu/morning-call-api/internal/infrastructure/auth"
	relUseCase "github.com/ochamu/morning-call-api/internal/usecase/relationship"
	"github.com/ochamu/morning-call-api/internal/usecase/user"
)

// RelationshipHandler は友達関係関連のHTTPハンドラー
type RelationshipHandler struct {
	*BaseHandler
	sendFriendRequestUC   *relUseCase.SendFriendRequestUseCase
	acceptFriendRequestUC *relUseCase.AcceptFriendRequestUseCase
	rejectFriendRequestUC *relUseCase.RejectFriendRequestUseCase
	blockUserUC           *relUseCase.BlockUserUseCase
	removeRelationshipUC  *relUseCase.RemoveRelationshipUseCase
	listFriendsUC         *relUseCase.ListFriendsUseCase
	listFriendRequestsUC  *relUseCase.ListFriendRequestsUseCase
	userUC                *user.UserUseCase
	sessionManager        *auth.SessionManager
}

// NewRelationshipHandler は新しいRelationshipHandlerを作成する
func NewRelationshipHandler(
	sendFriendRequestUC *relUseCase.SendFriendRequestUseCase,
	acceptFriendRequestUC *relUseCase.AcceptFriendRequestUseCase,
	rejectFriendRequestUC *relUseCase.RejectFriendRequestUseCase,
	blockUserUC *relUseCase.BlockUserUseCase,
	removeRelationshipUC *relUseCase.RemoveRelationshipUseCase,
	listFriendsUC *relUseCase.ListFriendsUseCase,
	listFriendRequestsUC *relUseCase.ListFriendRequestsUseCase,
	userUC *user.UserUseCase,
	sessionManager *auth.SessionManager,
) *RelationshipHandler {
	return &RelationshipHandler{
		BaseHandler:           &BaseHandler{},
		sendFriendRequestUC:   sendFriendRequestUC,
		acceptFriendRequestUC: acceptFriendRequestUC,
		rejectFriendRequestUC: rejectFriendRequestUC,
		blockUserUC:           blockUserUC,
		removeRelationshipUC:  removeRelationshipUC,
		listFriendsUC:         listFriendsUC,
		listFriendRequestsUC:  listFriendRequestsUC,
		userUC:                userUC,
		sessionManager:        sessionManager,
	}
}

// HandleSendFriendRequest は友達リクエスト送信のハンドラー
func (h *RelationshipHandler) HandleSendFriendRequest(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	currentUser, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// リクエストボディの解析
	var req request.SendFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, http.StatusBadRequest, "PARSE_ERROR", "リクエストの形式が正しくありません", nil)
		return
	}

	// 入力検証
	if req.ReceiverID == "" {
		h.SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", "宛先ユーザーIDが必要です", nil)
		return
	}

	// 友達リクエスト送信
	output, err := h.sendFriendRequestUC.Execute(r.Context(), relUseCase.SendFriendRequestInput{
		RequesterID: currentUser.ID,
		ReceiverID:  req.ReceiverID,
	})
	if err != nil {
		// エラー内容に応じて適切なレスポンスを返す
		if strings.Contains(err.Error(), "既に") || strings.Contains(err.Error(), "ブロック") {
			h.SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		if strings.Contains(err.Error(), "見つかりません") {
			h.SendError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		h.SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "友達リクエストの送信に失敗しました", nil)
		return
	}

	// レスポンス
	h.SendJSON(w, http.StatusCreated, response.NewRelationshipResponse(output.Relationship))
}

// HandleAcceptFriendRequest は友達リクエスト承認のハンドラー
func (h *RelationshipHandler) HandleAcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	currentUser, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// URLパラメータから関係IDを取得
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[len(parts)-1] != "accept" {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "無効なリクエストパスです", nil)
		return
	}
	relationshipID := parts[len(parts)-2]

	// 友達リクエスト承認
	output, err := h.acceptFriendRequestUC.Execute(r.Context(), relUseCase.AcceptFriendRequestInput{
		RelationshipID: relationshipID,
		ReceiverID:     currentUser.ID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			h.SendError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		if strings.Contains(err.Error(), "権限") || strings.Contains(err.Error(), "承認できません") {
			h.SendError(w, http.StatusForbidden, "FORBIDDEN", err.Error(), nil)
			return
		}
		h.SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "友達リクエストの承認に失敗しました", nil)
		return
	}

	// レスポンス
	h.SendJSON(w, http.StatusOK, response.NewRelationshipResponse(output.Relationship))
}

// HandleRejectFriendRequest は友達リクエスト拒否のハンドラー
func (h *RelationshipHandler) HandleRejectFriendRequest(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	currentUser, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// URLパラメータから関係IDを取得
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[len(parts)-1] != "reject" {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "無効なリクエストパスです", nil)
		return
	}
	relationshipID := parts[len(parts)-2]

	// 友達リクエスト拒否
	output, err := h.rejectFriendRequestUC.Execute(r.Context(), relUseCase.RejectFriendRequestInput{
		RelationshipID: relationshipID,
		ReceiverID:     currentUser.ID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			h.SendError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		if strings.Contains(err.Error(), "権限") || strings.Contains(err.Error(), "拒否できません") {
			h.SendError(w, http.StatusForbidden, "FORBIDDEN", err.Error(), nil)
			return
		}
		h.SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "友達リクエストの拒否に失敗しました", nil)
		return
	}

	// レスポンス
	if output != nil {
		h.SendJSON(w, http.StatusOK, response.NewRelationshipResponse(output.Relationship))
	} else {
		h.SendJSON(w, http.StatusNoContent, nil)
	}
}

// HandleBlockUser はユーザーブロックのハンドラー
func (h *RelationshipHandler) HandleBlockUser(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	currentUser, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// URLパラメータからユーザーIDを取得
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[len(parts)-1] != "block" {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "無効なリクエストパスです", nil)
		return
	}
	targetUserID := parts[len(parts)-2]

	// ユーザーブロック
	output, err := h.blockUserUC.Execute(r.Context(), relUseCase.BlockUserInput{
		BlockerID: currentUser.ID,
		BlockedID: targetUserID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			h.SendError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		if strings.Contains(err.Error(), "自分自身") {
			h.SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		h.SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "ユーザーのブロックに失敗しました", nil)
		return
	}

	// レスポンス
	h.SendJSON(w, http.StatusOK, response.NewRelationshipResponse(output.Relationship))
}

// HandleRemoveRelationship は関係削除のハンドラー
func (h *RelationshipHandler) HandleRemoveRelationship(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	currentUser, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// URLパラメータから関係IDを取得
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "無効なリクエストパスです", nil)
		return
	}
	relationshipID := parts[len(parts)-1]

	// 関係削除
	_, err = h.removeRelationshipUC.Execute(r.Context(), relUseCase.RemoveRelationshipInput{
		RelationshipID: relationshipID,
		UserID:         currentUser.ID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			h.SendError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		if strings.Contains(err.Error(), "権限") {
			h.SendError(w, http.StatusForbidden, "FORBIDDEN", err.Error(), nil)
			return
		}
		h.SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "関係の削除に失敗しました", nil)
		return
	}

	// レスポンス
	h.SendJSON(w, http.StatusNoContent, nil)
}

// HandleListFriends は友達一覧取得のハンドラー
func (h *RelationshipHandler) HandleListFriends(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	currentUser, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// 友達一覧取得
	output, err := h.listFriendsUC.Execute(r.Context(), relUseCase.ListFriendsInput{
		UserID: currentUser.ID,
	})
	if err != nil {
		h.SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "友達一覧の取得に失敗しました", nil)
		return
	}

	// 友達情報を取得して詳細なレスポンスを作成
	friendResponses := make([]*response.FriendResponse, 0, len(output.Friends))
	for _, friendInfo := range output.Friends {
		friendResponses = append(friendResponses, &response.FriendResponse{
			ID:          friendInfo.User.ID,
			Username:    friendInfo.User.Username,
			Email:       friendInfo.User.Email,
			FriendSince: friendInfo.Relationship.UpdatedAt, // 友達になった日時
		})
	}

	// レスポンス
	h.SendJSON(w, http.StatusOK, &response.FriendListResponse{
		Friends: friendResponses,
		Total:   len(friendResponses),
	})
}

// HandleListFriendRequests は友達リクエスト一覧取得のハンドラー
func (h *RelationshipHandler) HandleListFriendRequests(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	currentUser, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// クエリパラメータで方向を指定（sent/received）
	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "received" // デフォルトは受信したリクエスト
	}

	// 友達リクエスト一覧取得
	var relationships []*entity.Relationship
	switch direction {
	case "sent":
		// 送信した友達リクエスト
		output, err := h.listFriendRequestsUC.Execute(r.Context(), relUseCase.ListFriendRequestsInput{
			UserID: currentUser.ID,
			Type:   "sent",
		})
		if err != nil {
			h.SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "友達リクエスト一覧の取得に失敗しました", nil)
			return
		}
		for _, reqInfo := range output.Requests {
			relationships = append(relationships, reqInfo.Relationship)
		}
	case "received":
		// 受信した友達リクエスト
		output, err := h.listFriendRequestsUC.Execute(r.Context(), relUseCase.ListFriendRequestsInput{
			UserID: currentUser.ID,
			Type:   "received",
		})
		if err != nil {
			h.SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "友達リクエスト一覧の取得に失敗しました", nil)
			return
		}
		for _, reqInfo := range output.Requests {
			relationships = append(relationships, reqInfo.Relationship)
		}
	default:
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "無効な方向指定です（sent/receivedのいずれかを指定してください）", nil)
		return
	}

	// レスポンス
	h.SendJSON(w, http.StatusOK, response.NewRelationshipListResponse(relationships))
}
