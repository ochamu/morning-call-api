package memory

import "fmt"

// ===== テスト用ヘルパー関数 =====

// generateTestID はテスト用のIDを生成する
func generateTestID(prefix string, index int) string {
	return fmt.Sprintf("%s%d", prefix, index)
}

// generateTestUserID はテスト用のユーザーIDを生成する
func generateTestUserID(index int) string {
	return generateTestID("user", index)
}

// generateTestRelationshipID はテスト用の関係IDを生成する
func generateTestRelationshipID(index int) string {
	return generateTestID("rel", index)
}

// generateTestMorningCallID はテスト用のモーニングコールIDを生成する
func generateTestMorningCallID(index int) string {
	return generateTestID("mc", index)
}
