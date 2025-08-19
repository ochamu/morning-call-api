package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/ochamu/morning-call-api/internal/domain/entity"
	"github.com/ochamu/morning-call-api/internal/domain/repository"
	"github.com/ochamu/morning-call-api/internal/domain/service"
	"github.com/ochamu/morning-call-api/pkg/utils"
)

// UserUseCase はユーザー関連のユースケースを実装する
type UserUseCase struct {
	userRepo        repository.UserRepository
	passwordService service.PasswordService
}

// NewUserUseCase は新しいUserUseCaseを作成する
func NewUserUseCase(userRepo repository.UserRepository, passwordService service.PasswordService) *UserUseCase {
	return &UserUseCase{
		userRepo:        userRepo,
		passwordService: passwordService,
	}
}

// RegisterInput はユーザー登録の入力パラメータ
type RegisterInput struct {
	Username string
	Email    string
	Password string
}

// RegisterOutput はユーザー登録の出力結果
type RegisterOutput struct {
	User *entity.User
}

// Register は新しいユーザーを登録する
func (uc *UserUseCase) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	// 入力検証
	if input.Username == "" || input.Email == "" || input.Password == "" {
		return nil, fmt.Errorf("ユーザー名、メールアドレス、パスワードは必須です")
	}

	// パスワードの妥当性検証
	if reason := entity.ValidatePassword(input.Password); reason.IsNG() {
		return nil, fmt.Errorf("%s", reason)
	}

	// ユーザー名の重複チェック
	exists, err := uc.userRepo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("ユーザー名 '%s' は既に使用されています", input.Username)
	}

	// メールアドレスの重複チェック
	exists, err = uc.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("メールアドレス '%s' は既に登録されています", input.Email)
	}

	// パスワードのハッシュ化
	passwordHash, err := uc.passwordService.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// UUIDの生成
	userID, err := utils.GenerateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	// ユーザーエンティティの作成
	user, reason := entity.NewUser(userID, input.Username, input.Email, passwordHash)
	if reason.IsNG() {
		return nil, fmt.Errorf("%s", reason)
	}

	// リポジトリに保存
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &RegisterOutput{
		User: user,
	}, nil
}

// GetByID はIDでユーザーを取得する
func (uc *UserUseCase) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("ユーザーIDは必須です")
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// GetByUsername はユーザー名でユーザーを取得する
func (uc *UserUseCase) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	if username == "" {
		return nil, fmt.Errorf("ユーザー名は必須です")
	}

	user, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// VerifyPassword はユーザーのパスワードを検証する
func (uc *UserUseCase) VerifyPassword(ctx context.Context, username, password string) (*entity.User, error) {
	// ユーザーを取得
	user, err := uc.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// パスワードを検証
	valid, err := uc.passwordService.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}

	if !valid {
		return nil, fmt.Errorf("パスワードが正しくありません")
	}

	return user, nil
}
