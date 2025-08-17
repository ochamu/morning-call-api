package repository

import "context"

// TransactionManager はトランザクション管理を担うインターフェース
type TransactionManager interface {
	// ExecuteInTransaction はトランザクション内で処理を実行する
	// 処理が成功した場合はコミット、エラーが発生した場合はロールバックする
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Transaction はトランザクションを表すインターフェース
type Transaction interface {
	// Commit はトランザクションをコミットする
	Commit() error

	// Rollback はトランザクションをロールバックする
	Rollback() error
}

// TransactionalRepository はトランザクション対応のリポジトリインターフェース
type TransactionalRepository interface {
	// BeginTransaction はトランザクションを開始する
	BeginTransaction(ctx context.Context) (Transaction, error)

	// WithTransaction は指定されたトランザクションでリポジトリを取得する
	WithTransaction(tx Transaction) interface{}
}
