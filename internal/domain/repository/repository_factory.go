package repository

// RepositoryFactory はすべてのリポジトリを提供するファクトリインターフェース
type RepositoryFactory interface {
	// UserRepository はユーザーリポジトリを取得する
	UserRepository() UserRepository

	// MorningCallRepository はモーニングコールリポジトリを取得する
	MorningCallRepository() MorningCallRepository

	// RelationshipRepository は友達関係リポジトリを取得する
	RelationshipRepository() RelationshipRepository

	// TransactionManager はトランザクションマネージャーを取得する
	TransactionManager() TransactionManager
}

// Repositories はリポジトリの集合を表す構造体
type Repositories struct {
	User         UserRepository
	MorningCall  MorningCallRepository
	Relationship RelationshipRepository
	TxManager    TransactionManager
}
