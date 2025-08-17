package valueobject

// NGReason はドメイン検証結果を表す値オブジェクト
// 空文字列は成功(OK)、非空文字列はエラーメッセージ(NG)を表す
type NGReason string

// IsOK は検証が成功したかを判定する
func (r NGReason) IsOK() bool {
	return r == ""
}

// IsNG は検証が失敗したかを判定する
func (r NGReason) IsNG() bool {
	return r != ""
}

// Error はエラーメッセージを返す
func (r NGReason) Error() string {
	return string(r)
}

// OK は成功を表すNGReasonを返す
func OK() NGReason {
	return NGReason("")
}

// NG はエラーメッセージを持つNGReasonを返す
func NG(message string) NGReason {
	return NGReason(message)
}
