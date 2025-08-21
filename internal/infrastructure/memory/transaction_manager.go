package memory

import (
	"context"

	"github.com/ochamu/morning-call-api/internal/domain/repository"
)

// TransactionManager はインメモリ実装のトランザクションマネージャー
// インメモリ実装ではトランザクションは必要ないため、何もしない実装を提供
type TransactionManager struct{}

// NewTransactionManager は新しいTransactionManagerを作成する
func NewTransactionManager() *TransactionManager {
	return &TransactionManager{}
}

// ExecuteInTransaction はトランザクション内で関数を実行する
// インメモリ実装では単に関数を実行するだけ
func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// インターフェースの実装を保証
var _ repository.TransactionManager = (*TransactionManager)(nil)
