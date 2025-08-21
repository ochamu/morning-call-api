package server

import (
	"github.com/ochamu/morning-call-api/internal/config"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/handler"
	"github.com/ochamu/morning-call-api/internal/handler/middleware"
	"github.com/ochamu/morning-call-api/internal/infrastructure/auth"
	authUC "github.com/ochamu/morning-call-api/internal/usecase/auth"
	// morningCallUC "github.com/ochamu/morning-call-api/internal/usecase/morning_call" // TODO: ユースケース実装後に有効化
	relationshipUC "github.com/ochamu/morning-call-api/internal/usecase/relationship"
	userUC "github.com/ochamu/morning-call-api/internal/usecase/user"
)

// Dependencies はアプリケーションの依存性を管理する構造体
type Dependencies struct {
	Config            *config.Config
	RepositoryFactory repository.RepositoryFactory
	PasswordService   *auth.PasswordService
	SessionManager    *auth.SessionManager
	Handlers          Handlers
	AuthMiddleware    *middleware.AuthMiddleware
	UseCases          UseCases
}

// Handlers はHTTPハンドラーをまとめた構造体
type Handlers struct {
	Auth *handler.AuthHandler
	User *handler.UserHandler
	// TODO: 他のハンドラーを追加
	// Relationship *handler.RelationshipHandler // TODO: 実装予定
	// MorningCall  *handler.MorningCallHandler  // TODO: 実装予定
}

// UseCases はユースケースをまとめた構造体
type UseCases struct {
	Auth *authUC.AuthUseCase
	User *userUC.UserUseCase
	// TODO: モーニングコールユースケース（未実装）
	// CreateMorningCall   *morningCallUC.CreateMorningCallUseCase
	// UpdateMorningCall   *morningCallUC.UpdateMorningCallUseCase
	// DeleteMorningCall   *morningCallUC.DeleteMorningCallUseCase
	// ListMorningCalls    *morningCallUC.ListMorningCallsUseCase
	// ConfirmWake         *morningCallUC.ConfirmWakeUseCase
	SendFriendRequest   *relationshipUC.SendFriendRequestUseCase
	AcceptFriendRequest *relationshipUC.AcceptFriendRequestUseCase
	RejectFriendRequest *relationshipUC.RejectFriendRequestUseCase
	BlockUser           *relationshipUC.BlockUserUseCase
	RemoveRelationship  *relationshipUC.RemoveRelationshipUseCase
	ListFriends         *relationshipUC.ListFriendsUseCase
	ListFriendRequests  *relationshipUC.ListFriendRequestsUseCase
}
