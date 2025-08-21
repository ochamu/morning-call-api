package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ochamu/morning-call-api/internal/config"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/handler"
	"github.com/ochamu/morning-call-api/internal/handler/middleware"
	"github.com/ochamu/morning-call-api/internal/infrastructure/auth"
	"github.com/ochamu/morning-call-api/internal/infrastructure/memory"
	"github.com/ochamu/morning-call-api/internal/infrastructure/server"
	authUC "github.com/ochamu/morning-call-api/internal/usecase/auth"
	morningCallUC "github.com/ochamu/morning-call-api/internal/usecase/morning_call"
	relationshipUC "github.com/ochamu/morning-call-api/internal/usecase/relationship"
	userUC "github.com/ochamu/morning-call-api/internal/usecase/user"
)

func main() {
	// 設定の読み込み
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("設定の検証に失敗しました: %v", err)
	}

	// ログの初期化
	log.Printf("Morning Call API サーバーを起動します (ポート: %s)", cfg.Server.Port)

	// リポジトリの初期化（インメモリ実装）
	userRepo := memory.NewUserRepository()
	morningCallRepo := memory.NewMorningCallRepository()
	relationshipRepo := memory.NewRelationshipRepository()
	transactionManager := memory.NewTransactionManager()

	// リポジトリファクトリーの作成
	factory := &repositoryFactory{
		userRepo:           userRepo,
		morningCallRepo:    morningCallRepo,
		relationshipRepo:   relationshipRepo,
		transactionManager: transactionManager,
	}

	// パスワードサービスの初期化
	passwordService := auth.NewPasswordService()

	// セッションマネージャーの初期化
	sessionManager := auth.NewSessionManager(24 * time.Hour) // 24時間のセッションタイムアウト

	// ユースケースの初期化
	authUseCase := authUC.NewAuthUseCase(userRepo, passwordService)
	userUseCase := userUC.NewUserUseCase(userRepo, passwordService)

	// モーニングコールユースケースの初期化
	createMorningCallUC := morningCallUC.NewCreateUseCase(morningCallRepo, userRepo, relationshipRepo)
	updateMorningCallUC := morningCallUC.NewUpdateUseCase(morningCallRepo, userRepo)
	deleteMorningCallUC := morningCallUC.NewDeleteUseCase(morningCallRepo) // DeleteUseCaseは引数が1つのみ
	listMorningCallUC := morningCallUC.NewListUseCase(morningCallRepo, userRepo)
	confirmWakeUC := morningCallUC.NewConfirmWakeUseCase(morningCallRepo, userRepo)

	// 関係性ユースケースの初期化
	sendFriendRequestUC := relationshipUC.NewSendFriendRequestUseCase(relationshipRepo, userRepo)
	acceptFriendRequestUC := relationshipUC.NewAcceptFriendRequestUseCase(relationshipRepo, userRepo)
	rejectFriendRequestUC := relationshipUC.NewRejectFriendRequestUseCase(relationshipRepo, userRepo)
	blockUserUC := relationshipUC.NewBlockUserUseCase(relationshipRepo, userRepo)
	removeRelationshipUC := relationshipUC.NewRemoveRelationshipUseCase(relationshipRepo, userRepo)
	listFriendsUC := relationshipUC.NewListFriendsUseCase(relationshipRepo, userRepo)
	listFriendRequestsUC := relationshipUC.NewListFriendRequestsUseCase(relationshipRepo, userRepo)

	// ハンドラーの初期化
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
		removeRelationshipUC,
		listFriendsUC,
		listFriendRequestsUC,
		userUseCase,
		sessionManager,
	)

	// 認証ミドルウェアの初期化
	authMiddleware := middleware.NewAuthMiddleware(sessionManager, userRepo)

	// 依存性コンテナの作成
	deps := &server.Dependencies{
		Config:            cfg,
		RepositoryFactory: factory,
		PasswordService:   passwordService,
		SessionManager:    sessionManager,
		Handlers: server.Handlers{
			Auth:         authHandler,
			User:         userHandler,
			MorningCall:  morningCallHandler,
			Relationship: relationshipHandler,
		},
		AuthMiddleware: authMiddleware,
		UseCases: server.UseCases{
			Auth:                authUseCase,
			User:                userUseCase,
			CreateMorningCall:   createMorningCallUC,
			UpdateMorningCall:   updateMorningCallUC,
			DeleteMorningCall:   deleteMorningCallUC,
			ListMorningCalls:    listMorningCallUC,
			ConfirmWake:         confirmWakeUC,
			SendFriendRequest:   sendFriendRequestUC,
			AcceptFriendRequest: acceptFriendRequestUC,
			RejectFriendRequest: rejectFriendRequestUC,
			BlockUser:           blockUserUC,
			RemoveRelationship:  removeRelationshipUC,
			ListFriends:         listFriendsUC,
			ListFriendRequests:  listFriendRequestsUC,
		},
	}

	// HTTPサーバーの作成
	srv := server.NewHTTPServer(cfg, deps)

	// シグナルハンドリングの設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// サーバーの起動（ゴルーチン）
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Printf("HTTPサーバーを起動しました: http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("サーバーの起動に失敗しました: %v", err)
		}
	}()

	// シグナル待機
	sig := <-sigChan
	log.Printf("シグナルを受信しました: %v", sig)

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("サーバーのシャットダウンに失敗しました: %v", err)
	}

	log.Println("サーバーを正常に停止しました")
}

// repositoryFactory はリポジトリファクトリーの実装です
type repositoryFactory struct {
	userRepo           repository.UserRepository
	morningCallRepo    repository.MorningCallRepository
	relationshipRepo   repository.RelationshipRepository
	transactionManager repository.TransactionManager
}

// UserRepository はユーザーリポジトリを返します
func (f *repositoryFactory) UserRepository() repository.UserRepository {
	return f.userRepo
}

// MorningCallRepository はモーニングコールリポジトリを返します
func (f *repositoryFactory) MorningCallRepository() repository.MorningCallRepository {
	return f.morningCallRepo
}

// RelationshipRepository は関係性リポジトリを返します
func (f *repositoryFactory) RelationshipRepository() repository.RelationshipRepository {
	return f.relationshipRepo
}

// TransactionManager はトランザクションマネージャーを返します
func (f *repositoryFactory) TransactionManager() repository.TransactionManager {
	return f.transactionManager
}
