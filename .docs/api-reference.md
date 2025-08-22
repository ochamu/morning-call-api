# Morning Call API リファレンス

## 概要

Morning Call APIは、友達にアラーム（モーニングコール）を設定できるRESTful APIサービスです。ユーザー間の友達関係を管理し、相手のために起床アラームを設定・管理する機能を提供します。

## ベースURL

```
http://localhost:8080/api/v1
```

## 認証

APIは基本的なセッションベース認証を使用します。ログイン後、セッションIDがCookieに設定され、認証が必要なエンドポイントへのアクセスに使用されます。

### セッション管理

- セッションの有効期限: 24時間
- セッションはCookie（`session_id`）で管理
- HttpOnly属性付きで安全に管理

## 共通レスポンス形式

### 成功レスポンス

```json
{
  "success": true,
  "data": {...},
  "message": "操作が成功しました"
}
```

### エラーレスポンス

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "エラーメッセージ",
    "details": [
      {
        "field": "フィールド名",
        "message": "詳細なエラーメッセージ"
      }
    ]
  }
}
```

## エラーコード一覧

| コード | HTTPステータス | 説明 |
|--------|--------------|------|
| `VALIDATION_ERROR` | 400 | 入力値が不正です |
| `INVALID_REQUEST` | 400 | リクエストの形式が不正です |
| `AUTHENTICATION_ERROR` | 401 | 認証が必要です |
| `INVALID_CREDENTIALS` | 401 | ユーザー名またはパスワードが間違っています |
| `FORBIDDEN` | 403 | この操作を実行する権限がありません |
| `NOT_FOUND` | 404 | リソースが見つかりません |
| `METHOD_NOT_ALLOWED` | 405 | HTTPメソッドが許可されていません |
| `ALREADY_EXISTS` | 409 | リソースが既に存在します |
| `INTERNAL_SERVER_ERROR` | 500 | サーバーエラーが発生しました |

## API エンドポイント

### ヘルスチェック

#### GET /health

サービスの稼働状況を確認します。

**認証**: 不要

**レスポンス**

```json
{
  "status": "healthy",
  "timestamp": 1234567890,
  "service": "morning-call-api",
  "version": "1.0.0"
}
```

**例**

```bash
curl http://localhost:8080/health
```

---

### API情報

#### GET /api/v1

API全体の情報を取得します。

**認証**: 不要

**レスポンス**

```json
{
  "name": "Morning Call API",
  "version": "1.0.0",
  "description": "友達にアラームを設定できるAPIサービス",
  "endpoints": {
    "health": "/health",
    "auth": "/api/v1/auth",
    "users": "/api/v1/users",
    "relationships": "/api/v1/relationships",
    "morning_calls": "/api/v1/morning-calls"
  }
}
```

**例**

```bash
curl http://localhost:8080/api/v1
```

---

## 認証 (Authentication)

### ログイン

#### POST /api/v1/auth/login

ユーザー名とパスワードでログインします。

**認証**: 不要

**リクエストボディ**

```json
{
  "username": "string",
  "password": "string"
}
```

**フィールド説明**

| フィールド | 型 | 必須 | 説明 | 制約 |
|-----------|-----|------|------|------|
| `username` | string | ✓ | ユーザー名 | 3文字以上50文字以内 |
| `password` | string | ✓ | パスワード | 8文字以上 |

**レスポンス** (200 OK)

```json
{
  "session_id": "session_abc123...",
  "user": {
    "id": "user_123",
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "expires_at": "2024-01-02T00:00:00Z"
}
```

**エラーレスポンス**

- 401 Unauthorized: ユーザー名またはパスワードが間違っています
- 400 Bad Request: バリデーションエラー

**例**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "Password123!"}'
```

---

### ログアウト

#### POST /api/v1/auth/logout

現在のセッションを終了します。

**認証**: 必要

**レスポンス** (200 OK)

```json
{
  "success": true,
  "message": "ログアウトしました"
}
```

**例**

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Cookie: session_id=session_abc123..."
```

---

### 現在のユーザー情報取得

#### GET /api/v1/auth/me

ログイン中のユーザー情報を取得します。

**認証**: 必要

**レスポンス** (200 OK)

```json
{
  "user": {
    "id": "user_123",
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**例**

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Cookie: session_id=session_abc123..."
```

---

### セッション検証

#### GET /api/v1/auth/validate

セッションの有効性を確認します。

**認証**: 不要（セッションIDの検証のみ）

**レスポンス** (200 OK)

```json
{
  "valid": true
}
```

**例**

```bash
curl http://localhost:8080/api/v1/auth/validate \
  -H "Cookie: session_id=session_abc123..."
```

---

### セッション更新

#### POST /api/v1/auth/refresh

セッションの有効期限を延長します。

**認証**: 必要

**レスポンス** (200 OK)

```json
{
  "success": true,
  "expires_at": "2024-01-03T00:00:00Z",
  "message": "セッションの有効期限を延長しました"
}
```

**例**

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Cookie: session_id=session_abc123..."
```

---

## ユーザー (Users)

### ユーザー登録

#### POST /api/v1/users/register

新しいユーザーアカウントを作成します。

**認証**: 不要

**リクエストボディ**

```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

**フィールド説明**

| フィールド | 型 | 必須 | 説明 | 制約 |
|-----------|-----|------|------|------|
| `username` | string | ✓ | ユーザー名 | 3-30文字、英数字・アンダースコア・ハイフンのみ |
| `email` | string | ✓ | メールアドレス | 有効なメール形式、最大255文字 |
| `password` | string | ✓ | パスワード | 8-100文字、大文字・小文字・数字・特殊文字を各1文字以上含む |

**レスポンス** (201 Created)

```json
{
  "success": true,
  "user": {
    "id": "user_123",
    "username": "newuser",
    "email": "new@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "message": "ユーザー登録が完了しました"
}
```

**エラーレスポンス**

- 409 Conflict: ユーザー名またはメールアドレスが既に使用されています
- 400 Bad Request: バリデーションエラー

**例**

```bash
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "new@example.com",
    "password": "SecurePass123!"
  }'
```

---

### プロフィール取得

#### GET /api/v1/users/profile

ログイン中のユーザーのプロフィール情報を取得します。

**認証**: 必要

**レスポンス** (200 OK)

```json
{
  "id": "user_123",
  "username": "testuser",
  "email": "test@example.com",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

**例**

```bash
curl http://localhost:8080/api/v1/users/profile \
  -H "Cookie: session_id=session_abc123..."
```

---

### ユーザー検索

#### GET /api/v1/users/search

ユーザー名やメールアドレスでユーザーを検索します。

**認証**: 必要

**クエリパラメータ**

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `query` | string | ✓ | 検索文字列 |

**レスポンス** (200 OK)

```json
{
  "users": [],
  "count": 0,
  "message": "検索機能は現在実装中です"
}
```

**注**: この機能は現在開発中です。

**例**

```bash
curl "http://localhost:8080/api/v1/users/search?query=test" \
  -H "Cookie: session_id=session_abc123..."
```

---

### ユーザー情報取得（ID指定）

#### GET /api/v1/users/{id}

指定したIDのユーザー情報を取得します。

**認証**: 必要

**パスパラメータ**

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `id` | string | ✓ | ユーザーID |

**レスポンス** (200 OK)

```json
{
  "id": "user_456",
  "username": "otheruser",
  "email": "other@example.com",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

**エラーレスポンス**

- 404 Not Found: ユーザーが見つかりません

**例**

```bash
curl http://localhost:8080/api/v1/users/user_456 \
  -H "Cookie: session_id=session_abc123..."
```

---

## 友達関係 (Relationships) - 計画中

以下のエンドポイントは実装予定です。

### 友達リクエスト送信

#### POST /api/v1/relationships/request

友達リクエストを送信します。

**認証**: 必要

**リクエストボディ**

```json
{
  "receiver_id": "string"
}
```

---

### 友達リクエスト承認

#### PUT /api/v1/relationships/{id}/accept

友達リクエストを承認します。

**認証**: 必要

---

### 友達リクエスト拒否

#### PUT /api/v1/relationships/{id}/reject

友達リクエストを拒否します。

**認証**: 必要

---

### ユーザーブロック

#### PUT /api/v1/relationships/{id}/block

ユーザーをブロックします。

**認証**: 必要

---

### 関係削除

#### DELETE /api/v1/relationships/{id}

友達関係を削除します。

**認証**: 必要

---

### 友達一覧取得

#### GET /api/v1/relationships/friends

友達一覧を取得します。

**認証**: 必要

---

### 友達リクエスト一覧取得

#### GET /api/v1/relationships/requests

受信した友達リクエスト一覧を取得します。

**認証**: 必要

---

## モーニングコール (Morning Calls) - 計画中

以下のエンドポイントは実装予定です。

### モーニングコール作成

#### POST /api/v1/morning-calls

新しいモーニングコールを作成します。

**認証**: 必要

**リクエストボディ**

```json
{
  "receiver_id": "string",
  "scheduled_time": "2024-01-02T07:00:00Z",
  "message": "string"
}
```

**フィールド説明**

| フィールド | 型 | 必須 | 説明 | 制約 |
|-----------|-----|------|------|------|
| `receiver_id` | string | ✓ | 受信者のユーザーID | 友達関係にあるユーザーのみ |
| `scheduled_time` | string | ✓ | アラーム時刻（ISO 8601形式） | 現在時刻より後、30日以内 |
| `message` | string | × | メッセージ | 最大500文字 |

---

### モーニングコール詳細取得

#### GET /api/v1/morning-calls/{id}

指定したモーニングコールの詳細を取得します。

**認証**: 必要

---

### モーニングコール更新

#### PUT /api/v1/morning-calls/{id}

モーニングコールを更新します（スケジュール済みのもののみ）。

**認証**: 必要

---

### モーニングコール削除

#### DELETE /api/v1/morning-calls/{id}

モーニングコールを削除します。

**認証**: 必要

---

### 送信済みモーニングコール一覧

#### GET /api/v1/morning-calls/sent

自分が送信したモーニングコール一覧を取得します。

**認証**: 必要

---

### 受信済みモーニングコール一覧

#### GET /api/v1/morning-calls/received

自分が受信したモーニングコール一覧を取得します。

**認証**: 必要

---

### 起床確認

#### PUT /api/v1/morning-calls/{id}/confirm

起床を確認します。

**認証**: 必要

---

## ドメインモデル

### User エンティティ

ユーザーを表すエンティティです。

```go
type User struct {
    ID           string    // ユーザーID（UUID）
    Username     string    // ユーザー名（3-30文字、英数字・アンダースコア・ハイフン）
    Email        string    // メールアドレス（最大255文字）
    PasswordHash string    // ハッシュ化されたパスワード
    CreatedAt    time.Time // 作成日時
    UpdatedAt    time.Time // 更新日時
}
```

**バリデーションルール**

- `Username`: 3-30文字、英数字・アンダースコア・ハイフンのみ使用可能
- `Email`: 有効なメールアドレス形式、最大255文字
- `Password`: 8-100文字、大文字・小文字・数字・特殊文字を各1文字以上含む

### MorningCall エンティティ

モーニングコールを表すエンティティです。

```go
type MorningCall struct {
    ID            string                // モーニングコールID（UUID）
    SenderID      string                // 送信者のユーザーID
    ReceiverID    string                // 受信者のユーザーID
    ScheduledTime time.Time             // アラーム時刻
    Message       string                // メッセージ（最大500文字）
    Status        MorningCallStatus     // ステータス
    CreatedAt     time.Time             // 作成日時
    UpdatedAt     time.Time             // 更新日時
}
```

**ステータス**

- `scheduled`: スケジュール済み
- `delivered`: 配信済み
- `confirmed`: 起床確認済み
- `cancelled`: キャンセル済み
- `expired`: 期限切れ

**バリデーションルール**

- `ScheduledTime`: 現在時刻より後、30日以内
- `Message`: 最大500文字（日本語対応）
- 自分自身へのモーニングコール設定は不可

### Relationship エンティティ

ユーザー間の友達関係を表すエンティティです。

```go
type Relationship struct {
    ID          string              // 関係ID（UUID）
    RequesterID string              // リクエスト送信者のユーザーID
    ReceiverID  string              // リクエスト受信者のユーザーID
    Status      RelationshipStatus  // ステータス
    CreatedAt   time.Time           // 作成日時
    UpdatedAt   time.Time           // 更新日時
}
```

**ステータス**

- `pending`: 承認待ち
- `accepted`: 承認済み（友達）
- `blocked`: ブロック済み
- `rejected`: 拒否済み

**バリデーションルール**

- 自分自身への友達リクエストは不可
- 重複したリクエストは不可

## セキュリティ

### 認証フロー

1. ユーザーは `/api/v1/auth/login` でログイン
2. サーバーはセッションIDを生成し、Cookieに設定
3. 以降のリクエストでCookieのセッションIDを使用して認証
4. セッションは24時間有効、`/api/v1/auth/refresh` で延長可能

### パスワードセキュリティ

- bcryptを使用してハッシュ化（コスト: 10）
- 最小8文字、大文字・小文字・数字・特殊文字を各1文字以上含む
- 平文パスワードは保存されない

### セッション管理

- セッションIDはUUIDv4で生成
- HttpOnly Cookieで管理
- 24時間で自動失効
- ログアウト時に即座に無効化

## レート制限

現在のバージョンではレート制限は実装されていません。

## CORSポリシー

以下のCORSヘッダーが設定されています：

- `Access-Control-Allow-Origin`: `*`
- `Access-Control-Allow-Methods`: `GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers`: `Content-Type, Authorization`

## HTTPステータスコード

| コード | 説明 | 使用場面 |
|--------|------|----------|
| 200 | OK | リクエスト成功 |
| 201 | Created | リソース作成成功 |
| 204 | No Content | 成功（レスポンスボディなし） |
| 400 | Bad Request | 不正なリクエスト、バリデーションエラー |
| 401 | Unauthorized | 認証が必要、認証失敗 |
| 403 | Forbidden | アクセス権限なし |
| 404 | Not Found | リソースが見つからない |
| 405 | Method Not Allowed | HTTPメソッドが許可されていない |
| 409 | Conflict | リソースの競合（既に存在等） |
| 500 | Internal Server Error | サーバーエラー |

## 開発環境での使用例

### cURLを使った基本的な使用フロー

```bash
# 1. ユーザー登録
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'

# 2. ログイン（セッションIDをCookieに保存）
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "SecurePass123!"}' \
  -c cookie.txt

# 3. プロフィール取得（保存したCookieを使用）
curl http://localhost:8080/api/v1/users/profile \
  -b cookie.txt

# 4. ログアウト
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -b cookie.txt
```

### HTTPieを使った例

```bash
# ユーザー登録
http POST localhost:8080/api/v1/users/register \
  username=testuser \
  email=test@example.com \
  password=SecurePass123!

# ログイン（セッション開始）
http --session=./session.json POST localhost:8080/api/v1/auth/login \
  username=testuser \
  password=SecurePass123!

# プロフィール取得
http --session=./session.json GET localhost:8080/api/v1/users/profile
```

## 注意事項

1. **開発中の機能**: 友達関係管理とモーニングコール機能は現在開発中です
2. **データ永続化**: 現在はインメモリストレージを使用しているため、サーバー再起動時にデータが失われます
3. **HTTPS**: 本番環境では必ずHTTPSを使用してください
4. **タイムゾーン**: すべての日時はUTCで扱われます

## バージョン履歴

### v1.0.0 (現在)
- 基本的な認証機能
- ユーザー登録・管理
- セッション管理
- ヘルスチェック

### v1.1.0 (計画中)
- 友達関係管理機能
- モーニングコール機能
- MySQL対応

## サポート

問題が発生した場合は、GitHubのIssueトラッカーで報告してください。

---

*最終更新日: 2025年1月*