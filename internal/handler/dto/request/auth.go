package request

// LoginRequest はログインリクエストのDTO
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Validate はログインリクエストのバリデーションを行う
func (r *LoginRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if r.Username == "" {
		errors["username"] = "ユーザー名は必須です"
	} else if len(r.Username) < 3 {
		errors["username"] = "ユーザー名は3文字以上である必要があります"
	} else if len(r.Username) > 50 {
		errors["username"] = "ユーザー名は50文字以内である必要があります"
	}

	if r.Password == "" {
		errors["password"] = "パスワードは必須です"
	} else if len(r.Password) < 8 {
		errors["password"] = "パスワードは8文字以上である必要があります"
	}

	return errors
}

// RegisterRequest はユーザー登録リクエストのDTO
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate はユーザー登録リクエストのバリデーションを行う
func (r *RegisterRequest) Validate() map[string]string {
	errors := make(map[string]string)

	// ユーザー名のバリデーション
	if r.Username == "" {
		errors["username"] = "ユーザー名は必須です"
	} else if len(r.Username) < 3 {
		errors["username"] = "ユーザー名は3文字以上である必要があります"
	} else if len(r.Username) > 50 {
		errors["username"] = "ユーザー名は50文字以内である必要があります"
	}

	// メールアドレスのバリデーション
	if r.Email == "" {
		errors["email"] = "メールアドレスは必須です"
	} else if !isValidEmail(r.Email) {
		errors["email"] = "有効なメールアドレスを入力してください"
	}

	// パスワードのバリデーション
	if r.Password == "" {
		errors["password"] = "パスワードは必須です"
	} else if len(r.Password) < 8 {
		errors["password"] = "パスワードは8文字以上である必要があります"
	} else if len(r.Password) > 100 {
		errors["password"] = "パスワードは100文字以内である必要があります"
	}

	return errors
}

// isValidEmail はメールアドレスの形式を簡易的にチェックする
func isValidEmail(email string) bool {
	// 簡易的なチェック（@と.が含まれているか）
	atIndex := -1
	dotIndex := -1

	for i, ch := range email {
		if ch == '@' {
			if atIndex != -1 {
				// @が複数ある
				return false
			}
			atIndex = i
		} else if ch == '.' {
			dotIndex = i
		}
	}

	// @が存在し、@の後に.があるかチェック
	return atIndex > 0 && dotIndex > atIndex+1 && dotIndex < len(email)-1
}
