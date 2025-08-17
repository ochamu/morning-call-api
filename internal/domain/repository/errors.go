package repository

import "errors"

// Common repository errors
var (
	// ErrNotFound はエンティティが見つからない場合のエラー
	ErrNotFound = errors.New("entity not found")

	// ErrAlreadyExists はエンティティが既に存在する場合のエラー
	ErrAlreadyExists = errors.New("entity already exists")

	// ErrInvalidArgument は不正な引数が渡された場合のエラー
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrUpdateConflict は更新競合が発生した場合のエラー
	ErrUpdateConflict = errors.New("update conflict")

	// ErrTransactionFailed はトランザクションが失敗した場合のエラー
	ErrTransactionFailed = errors.New("transaction failed")

	// ErrConnectionFailed は接続エラーが発生した場合のエラー
	ErrConnectionFailed = errors.New("connection failed")

	// ErrTimeout はタイムアウトが発生した場合のエラー
	ErrTimeout = errors.New("operation timeout")

	// ErrPermissionDenied は権限がない場合のエラー
	ErrPermissionDenied = errors.New("permission denied")
)
