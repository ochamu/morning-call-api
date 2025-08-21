package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/valueobject"
	"github.com/ochamu/morning-call-api/internal/handler/dto/request"
	"github.com/ochamu/morning-call-api/internal/handler/dto/response"
	"github.com/ochamu/morning-call-api/internal/infrastructure/auth"
	mcCreate "github.com/ochamu/morning-call-api/internal/usecase/morning_call"
)

// MorningCallHandler はモーニングコール関連のHTTPハンドラー
type MorningCallHandler struct {
	*BaseHandler
	createUseCase      *mcCreate.CreateUseCase
	updateUseCase      *mcCreate.UpdateUseCase
	deleteUseCase      *mcCreate.DeleteUseCase
	listUseCase        *mcCreate.ListUseCase
	confirmWakeUseCase *mcCreate.ConfirmWakeUseCase
	sessionManager     *auth.SessionManager
}

// NewMorningCallHandler は新しいMorningCallHandlerを作成する
func NewMorningCallHandler(
	createUC *mcCreate.CreateUseCase,
	updateUC *mcCreate.UpdateUseCase,
	deleteUC *mcCreate.DeleteUseCase,
	listUC *mcCreate.ListUseCase,
	confirmWakeUC *mcCreate.ConfirmWakeUseCase,
	sessionManager *auth.SessionManager,
) *MorningCallHandler {
	return &MorningCallHandler{
		BaseHandler:        &BaseHandler{},
		createUseCase:      createUC,
		updateUseCase:      updateUC,
		deleteUseCase:      deleteUC,
		listUseCase:        listUC,
		confirmWakeUseCase: confirmWakeUC,
		sessionManager:     sessionManager,
	}
}

// HandleCreate はモーニングコール作成のハンドラー
func (h *MorningCallHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	user, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// リクエストボディのパース
	var req request.CreateMorningCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, http.StatusBadRequest, "PARSE_ERROR", "リクエストのパースに失敗しました", nil)
		return
	}

	// UseCaseの実行
	input := mcCreate.CreateInput{
		SenderID:      user.ID,
		ReceiverID:    req.ReceiverID,
		ScheduledTime: req.ScheduledTime,
		Message:       req.Message,
	}

	output, err := h.createUseCase.Execute(r.Context(), input)
	if err != nil {
		h.SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	// レスポンスの作成
	resp := h.convertToMorningCallResponse(output.MorningCall)
	h.SendJSON(w, http.StatusCreated, resp)
}

// HandleUpdate はモーニングコール更新のハンドラー
func (h *MorningCallHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	user, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// URLからIDを取得
	morningCallID := h.extractIDFromPath(r.URL.Path, "/api/v1/morning-calls/")
	if morningCallID == "" {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "モーニングコールIDが指定されていません", nil)
		return
	}

	// リクエストボディのパース
	var req request.UpdateMorningCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, http.StatusBadRequest, "PARSE_ERROR", "リクエストのパースに失敗しました", nil)
		return
	}

	// UseCaseの実行
	input := mcCreate.UpdateInput{
		ID:            morningCallID,
		SenderID:      user.ID,
		ScheduledTime: &req.ScheduledTime,
		Message:       &req.Message,
	}

	output, err := h.updateUseCase.Execute(r.Context(), input)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			h.SendError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		} else {
			h.SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		}
		return
	}

	// レスポンスの作成
	resp := h.convertToMorningCallResponse(output.MorningCall)
	h.SendJSON(w, http.StatusOK, resp)
}

// HandleDelete はモーニングコール削除のハンドラー
func (h *MorningCallHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	user, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// URLからIDを取得
	morningCallID := h.extractIDFromPath(r.URL.Path, "/api/v1/morning-calls/")
	if morningCallID == "" {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "モーニングコールIDが指定されていません", nil)
		return
	}

	// UseCaseの実行
	input := mcCreate.DeleteInput{
		ID:       morningCallID,
		SenderID: user.ID,
	}

	_, err = h.deleteUseCase.Execute(r.Context(), input)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			h.SendError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		} else {
			h.SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		}
		return
	}

	// 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// HandleGet はモーニングコール詳細取得のハンドラー
func (h *MorningCallHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	user, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// URLからIDを取得
	morningCallID := h.extractIDFromPath(r.URL.Path, "/api/v1/morning-calls/")
	if morningCallID == "" {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "モーニングコールIDが指定されていません", nil)
		return
	}

	// UseCaseの実行（詳細取得は一覧から絞り込み）
	// 送信と受信の両方を取得するため、2回実行
	inputSent := mcCreate.ListInput{
		UserID:   user.ID,
		ListType: mcCreate.ListTypeSent,
	}
	outputSent, err := h.listUseCase.Execute(r.Context(), inputSent)
	if err != nil {
		h.SendInternalServerError(w, err)
		return
	}

	inputReceived := mcCreate.ListInput{
		UserID:   user.ID,
		ListType: mcCreate.ListTypeReceived,
	}
	outputReceived, err := h.listUseCase.Execute(r.Context(), inputReceived)
	if err != nil {
		h.SendInternalServerError(w, err)
		return
	}

	// 結果をマージ
	allMorningCalls := append(outputSent.MorningCalls, outputReceived.MorningCalls...)

	// IDで絞り込み
	for _, mc := range allMorningCalls {
		if mc.ID == morningCallID {
			// アクセス権限チェック（送信者または受信者）
			if mc.SenderID != user.ID && mc.ReceiverID != user.ID {
				h.SendForbiddenError(w)
				return
			}
			resp := h.convertToMorningCallResponse(mc)
			h.SendJSON(w, http.StatusOK, resp)
			return
		}
	}

	h.SendNotFoundError(w, "モーニングコール")
}

// HandleListSent は送信済みモーニングコール一覧取得のハンドラー
func (h *MorningCallHandler) HandleListSent(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	user, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// UseCaseの実行
	input := mcCreate.ListInput{
		UserID:   user.ID,
		ListType: mcCreate.ListTypeSent,
	}

	output, err := h.listUseCase.Execute(r.Context(), input)
	if err != nil {
		h.SendInternalServerError(w, err)
		return
	}

	// レスポンスの作成
	morningCalls := make([]response.MorningCallResponse, len(output.MorningCalls))
	for i, mc := range output.MorningCalls {
		morningCalls[i] = h.convertToMorningCallResponse(mc)
	}

	resp := response.MorningCallListResponse{
		MorningCalls: morningCalls,
		Total:        len(morningCalls),
		Limit:        0, // 現在はページネーション未実装
		Offset:       0,
	}

	h.SendJSON(w, http.StatusOK, resp)
}

// HandleListReceived は受信モーニングコール一覧取得のハンドラー
func (h *MorningCallHandler) HandleListReceived(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	user, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// UseCaseの実行
	input := mcCreate.ListInput{
		UserID:   user.ID,
		ListType: mcCreate.ListTypeReceived,
	}

	output, err := h.listUseCase.Execute(r.Context(), input)
	if err != nil {
		h.SendInternalServerError(w, err)
		return
	}

	// レスポンスの作成
	morningCalls := make([]response.MorningCallResponse, len(output.MorningCalls))
	for i, mc := range output.MorningCalls {
		morningCalls[i] = h.convertToMorningCallResponse(mc)
	}

	resp := response.MorningCallListResponse{
		MorningCalls: morningCalls,
		Total:        len(morningCalls),
		Limit:        0, // 現在はページネーション未実装
		Offset:       0,
	}

	h.SendJSON(w, http.StatusOK, resp)
}

// HandleConfirmWake は起床確認のハンドラー
func (h *MorningCallHandler) HandleConfirmWake(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	user, err := h.GetUserFromContext(r.Context())
	if err != nil {
		h.SendAuthenticationError(w)
		return
	}

	// URLからIDを取得
	morningCallID := h.extractIDFromPath(r.URL.Path, "/api/v1/morning-calls/")
	morningCallID = strings.TrimSuffix(morningCallID, "/confirm")
	if morningCallID == "" {
		h.SendError(w, http.StatusBadRequest, "INVALID_REQUEST", "モーニングコールIDが指定されていません", nil)
		return
	}

	// UseCaseの実行
	input := mcCreate.ConfirmWakeInput{
		MorningCallID: morningCallID,
		ReceiverID:    user.ID,
	}

	output, err := h.confirmWakeUseCase.Execute(r.Context(), input)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			h.SendError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		} else {
			h.SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		}
		return
	}

	// レスポンスの作成
	resp := h.convertToMorningCallResponse(output.MorningCall)
	h.SendJSON(w, http.StatusOK, resp)
}

// convertToMorningCallResponse はエンティティをレスポンスDTOに変換する
func (h *MorningCallHandler) convertToMorningCallResponse(mc *entity.MorningCall) response.MorningCallResponse {
	resp := response.MorningCallResponse{
		ID:            mc.ID,
		SenderID:      mc.SenderID,
		ReceiverID:    mc.ReceiverID,
		ScheduledTime: mc.ScheduledTime,
		Message:       mc.Message,
		Status:        string(mc.Status),
		CreatedAt:     mc.CreatedAt,
		UpdatedAt:     mc.UpdatedAt,
	}

	// ConfirmedAtフィールドは現在のエンティティには存在しないため、
	// ステータスがConfirmedの場合はUpdatedAtを使用
	if mc.Status == valueobject.MorningCallStatusConfirmed {
		confirmedAt := mc.UpdatedAt
		resp.ConfirmedAt = &confirmedAt
	}

	return resp
}

// extractIDFromPath はURLパスからIDを抽出する
func (h *MorningCallHandler) extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	idPath := strings.TrimPrefix(path, prefix)
	// /confirm などのサフィックスを除去
	if idx := strings.Index(idPath, "/"); idx != -1 {
		idPath = idPath[:idx]
	}

	return idPath
}
