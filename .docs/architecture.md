# Morning Call API - 技術アーキテクチャドキュメント

## 目次

1. [エグゼクティブサマリー](#エグゼクティブサマリー)
2. [システムアーキテクチャ概要](#システムアーキテクチャ概要)
3. [ドメイン駆動設計の実装](#ドメイン駆動設計の実装)
4. [レイヤー設計と責務](#レイヤー設計と責務)
5. [エラーハンドリング戦略](#エラーハンドリング戦略)
6. [認証・認可システム](#認証認可システム)
7. [データモデルと関係性](#データモデルと関係性)
8. [API設計とRESTful実装](#api設計とrestful実装)
9. [実装状況と今後の拡張ポイント](#実装状況と今後の拡張ポイント)
10. [パフォーマンス設計](#パフォーマンス設計)
11. [セキュリティ設計](#セキュリティ設計)

---

## エグゼクティブサマリー

Morning Call APIは、ユーザー間で「モーニングコール（アラーム）」を設定できるソーシャル機能を持つRESTful APIサービスです。友達になったユーザー同士が、相手のためにアラームを設定し、起床体験を共有することで、新しいコミュニケーションの形を提供します。

### プロジェクトの特徴

- **Pure Go実装**: 外部フレームワークに依存せず、Go標準ライブラリのみで実装
- **Clean Architecture**: ドメイン駆動設計（DDD）とクリーンアーキテクチャの原則に従った設計
- **NGReasonパターン**: 独自の軽量なエラーハンドリングパターンによる高速な検証処理
- **インメモリ先行開発**: 初期実装はインメモリストレージで行い、将来的にMySQLへ移行可能な設計

### 技術的判断の背景

1. **標準ライブラリのみ使用**: 依存関係を最小限に抑え、保守性とパフォーマンスを優先
2. **NGReasonパターン採用**: Goのerror型よりも軽量で、ドメイン検証に特化した設計
3. **インメモリファースト**: 高速なプロトタイピングと、データ永続化層の抽象化を両立

---

## システムアーキテクチャ概要

### アーキテクチャダイアグラム

```
┌─────────────────────────────────────────────────────────────┐
│                         Client Layer                         │
│                    (Web/Mobile Applications)                 │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      Presentation Layer                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                    HTTP Handlers                     │   │
│  │  (AuthHandler, UserHandler, MorningCallHandler...)   │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                     Middleware                       │   │
│  │    (Auth, Logging, CORS, Recovery, RateLimit)       │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                      Use Cases                       │   │
│  │   (CreateMorningCall, SendFriendRequest, Login...)   │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                       Domain Layer                           │
│  ┌──────────────┬──────────────┬──────────────────────┐    │
│  │   Entities   │ Value Objects │   Domain Services    │    │
│  │  ・User      │  ・NGReason   │  ・PasswordService  │    │
│  │  ・MorningCall│  ・Status     │                     │    │
│  │  ・Relationship              │                     │    │
│  └──────────────┴──────────────┴──────────────────────┘    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Repository Interfaces                   │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                       │
│  ┌──────────────┬──────────────┬──────────────────────┐    │
│  │   Memory     │     Auth     │      Server          │    │
│  │ Repositories │   Services   │   Configuration      │    │
│  └──────────────┴──────────────┴──────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### レイヤー間の依存関係

```
Handler → UseCase → Domain ← Infrastructure
    ↓        ↓         ↑          ↑
    └────────┴─────────┴──────────┘
         (依存性の逆転原則)
```

---

## ドメイン駆動設計の実装

### ドメインエンティティ

#### 1. User エンティティ

```go
type User struct {
    ID           string
    Username     string
    Email        string
    PasswordHash string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**ビジネスルール:**
- ユーザー名は3-30文字の英数字、アンダースコア、ハイフンのみ
- メールアドレスは正規表現による検証と255文字制限
- パスワードは8-100文字で、大文字・小文字・数字・特殊文字を各1文字以上含む
- 自分自身にはモーニングコールを設定できない

#### 2. MorningCall エンティティ

```go
type MorningCall struct {
    ID            string
    SenderID      string
    ReceiverID    string
    ScheduledTime time.Time
    Message       string
    Status        MorningCallStatus
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

**ビジネスルール:**
- アラーム時刻は現在時刻より後で、30日以内
- メッセージは500文字以内（Unicode対応）
- ステータス遷移は厳密に制御
- 同一ユーザーペアで1分以内の重複アラームは禁止

#### 3. Relationship エンティティ

```go
type Relationship struct {
    ID          string
    RequesterID string
    ReceiverID  string
    Status      RelationshipStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**ビジネスルール:**
- 自分自身への友達リクエストは不可
- ブロック状態からの復帰は新規リクエストとして処理
- 拒否されたリクエストは24時間後に再送信可能

### 値オブジェクト（Value Objects）

#### NGReason - 革新的な検証パターン

```go
type NGReason string

// 空文字列 = OK（成功）
// 非空文字列 = NG（失敗）+ エラーメッセージ
```

**設計思想:**
- Goの`error`型よりも軽量（文字列比較のみ）
- ドメイン検証に特化し、ビジネスロジックを明確に表現
- 日本語メッセージによる直感的なエラー表現

**実装例:**
```go
func (u *User) ValidateUsername() NGReason {
    if u.Username == "" {
        return NG("ユーザー名は必須です")
    }
    if len(u.Username) < 3 {
        return NG("ユーザー名は3文字以上である必要があります")
    }
    return OK() // 空文字列 = 成功
}
```

### ステータス管理

#### MorningCallStatus

```
scheduled → delivered → confirmed
    ↓          ↓
cancelled   expired
```

- **scheduled**: スケジュール済み（初期状態）
- **delivered**: 配信済み（アラーム時刻到達）
- **confirmed**: 起床確認済み（最終状態）
- **cancelled**: キャンセル済み（最終状態）
- **expired**: 期限切れ（最終状態）

#### RelationshipStatus

```
pending → accepted
   ↓         ↓
rejected  blocked
   ↓
pending（再送信）
```

---

## レイヤー設計と責務

### Domain Layer（ドメイン層）

**責務:**
- ビジネスルールの実装と保護
- エンティティの不変条件の維持
- ドメインサービスの提供

**主要コンポーネント:**
```
internal/domain/
├── entity/           # ドメインエンティティ
├── valueobject/      # 値オブジェクト
├── repository/       # リポジトリインターフェース
└── service/          # ドメインサービス
```

### UseCase Layer（ユースケース層）

**責務:**
- アプリケーションのビジネスロジック実装
- トランザクション境界の管理
- 複数エンティティの協調動作

**実装パターン:**
```go
type CreateUseCase struct {
    morningCallRepo  repository.MorningCallRepository
    userRepo         repository.UserRepository
    relationshipRepo repository.RelationshipRepository
}

func (uc *CreateUseCase) Execute(ctx context.Context, input CreateInput) (*CreateOutput, error) {
    // 1. 入力検証
    // 2. ドメインルールの確認
    // 3. エンティティの生成
    // 4. 永続化
    // 5. 結果の返却
}
```

### Handler Layer（ハンドラー層）

**責務:**
- HTTPリクエスト/レスポンスの処理
- 入力データのバリデーション
- エラーレスポンスの生成

**BaseHandler による共通処理:**
```go
type BaseHandler struct {
    // 共通メソッド群
    SendJSON()
    SendError()
    ParseJSON()
    RequireAuth()
}
```

### Infrastructure Layer（インフラストラクチャ層）

**責務:**
- 外部システムとの連携
- データ永続化の実装
- 技術的詳細の隠蔽

**インメモリリポジトリの特徴:**
```go
type UserRepository struct {
    users         map[string]*entity.User
    usernameIndex map[string]string  // 高速検索用インデックス
    emailIndex    map[string]string
    mu            sync.RWMutex       // 並行アクセス制御
}
```

---

## エラーハンドリング戦略

### 3層エラーハンドリングアーキテクチャ

#### 1. ドメイン層 - NGReason

```go
// ビジネスルール違反の表現
reason := user.ValidateEmail()
if reason.IsNG() {
    return nil, reason  // "メールアドレスの形式が正しくありません"
}
```

#### 2. ユースケース層 - error型

```go
// システムエラーとビジネスエラーの統合
if !areFriends {
    return nil, fmt.Errorf("友達関係にないユーザーにはモーニングコールを設定できません")
}
```

#### 3. ハンドラー層 - HTTPステータスコード

```go
// HTTPレスポンスへのマッピング
type ErrorResponse struct {
    Error struct {
        Code    string            `json:"code"`
        Message string            `json:"message"`
        Details []ValidationError `json:"details,omitempty"`
    } `json:"error"`
}
```

### エラーコード体系

```
VALIDATION_ERROR     - 400 Bad Request
AUTHENTICATION_ERROR - 401 Unauthorized
FORBIDDEN           - 403 Forbidden
NOT_FOUND          - 404 Not Found
INTERNAL_ERROR     - 500 Internal Server Error
```

---

## 認証・認可システム

### セッション管理アーキテクチャ

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Client    │───▶│ AuthHandler  │───▶│SessionManager│
└─────────────┘    └──────────────┘    └─────────────┘
                            │                    │
                            ▼                    ▼
                    ┌──────────────┐    ┌─────────────┐
                    │  AuthUseCase │    │   Sessions  │
                    └──────────────┘    │  (InMemory) │
                            │           └─────────────┘
                            ▼
                    ┌──────────────┐
                    │UserRepository│
                    └──────────────┘
```

### 認証フロー

#### 1. ログイン処理

```go
1. クライアント → POST /api/v1/auth/login
2. 入力検証（ユーザー名、パスワード）
3. ユーザー検索（UserRepository）
4. パスワード検証（bcrypt）
5. セッション生成（64バイトのランダムトークン）
6. Cookie設定 + レスポンス返却
```

#### 2. セッション検証

```go
1. リクエストからセッションID取得
   - Authorizationヘッダー（Bearer/Session）
   - X-Session-IDヘッダー
   - Cookieのsession_id
2. SessionManagerで有効性確認
3. 有効期限チェック
4. ユーザー情報をContextに格納
```

### セキュリティ機能

- **パスワードハッシュ化**: bcrypt（コスト10）
- **セッションタイムアウト**: デフォルト24時間
- **自動クリーンアップ**: 5分ごとに期限切れセッション削除
- **セッション延長**: リクエストごとに自動延長可能

---

## データモデルと関係性

### エンティティ関係図

```
┌─────────────┐        ┌──────────────┐        ┌─────────────┐
│    User     │───────▶│ Relationship │◀───────│    User     │
└─────────────┘  1:N   └──────────────┘   N:1  └─────────────┘
       │                                               │
       │ 1:N                                       N:1 │
       ▼                                               ▼
┌──────────────────────────────────────────────────────────┐
│                      MorningCall                          │
└──────────────────────────────────────────────────────────┘
```

### インデックス設計（インメモリ実装）

#### UserRepository
- **Primary Index**: ID → User
- **Secondary Indexes**:
  - username（小文字正規化） → ID
  - email（小文字正規化） → ID

#### MorningCallRepository
- **Primary Index**: ID → MorningCall
- **Secondary Indexes**:
  - senderID → []ID
  - receiverID → []ID
  - userPair（sender+receiver） → []ID

#### RelationshipRepository
- **Primary Index**: ID → Relationship
- **Secondary Indexes**:
  - userPair（正規化済み） → ID
  - userID → []ID（関連する全関係）

### データ整合性の保証

```go
// トランザクション管理（将来のDB実装用）
type TransactionManager interface {
    ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
```

---

## API設計とRESTful実装

### APIエンドポイント一覧

#### 認証系API

| メソッド | パス | 説明 | 認証 |
|---------|------|------|------|
| POST | /api/v1/auth/login | ログイン | 不要 |
| POST | /api/v1/auth/logout | ログアウト | 必要 |
| GET | /api/v1/auth/me | 現在のユーザー情報 | 必要 |
| GET | /api/v1/auth/validate | セッション検証 | 不要 |
| POST | /api/v1/auth/refresh | セッション延長 | 必要 |

#### ユーザー系API

| メソッド | パス | 説明 | 認証 |
|---------|------|------|------|
| POST | /api/v1/users/register | ユーザー登録 | 不要 |
| GET | /api/v1/users/profile | プロフィール取得 | 必要 |
| GET | /api/v1/users/search | ユーザー検索 | 必要 |
| GET | /api/v1/users/{id} | ユーザー詳細 | 必要 |

#### 関係性系API

| メソッド | パス | 説明 | 認証 |
|---------|------|------|------|
| POST | /api/v1/relationships/request | 友達リクエスト送信 | 必要 |
| PUT | /api/v1/relationships/{id}/accept | リクエスト承認 | 必要 |
| PUT | /api/v1/relationships/{id}/reject | リクエスト拒否 | 必要 |
| PUT | /api/v1/relationships/{id}/block | ユーザーブロック | 必要 |
| DELETE | /api/v1/relationships/{id} | 関係削除 | 必要 |
| GET | /api/v1/relationships/friends | 友達一覧 | 必要 |
| GET | /api/v1/relationships/requests | リクエスト一覧 | 必要 |

#### モーニングコール系API

| メソッド | パス | 説明 | 認証 |
|---------|------|------|------|
| POST | /api/v1/morning-calls | モーニングコール作成 | 必要 |
| GET | /api/v1/morning-calls/{id} | 詳細取得 | 必要 |
| PUT | /api/v1/morning-calls/{id} | 更新 | 必要 |
| DELETE | /api/v1/morning-calls/{id} | 削除 | 必要 |
| GET | /api/v1/morning-calls/sent | 送信済み一覧 | 必要 |
| GET | /api/v1/morning-calls/received | 受信一覧 | 必要 |
| PUT | /api/v1/morning-calls/{id}/confirm | 起床確認 | 必要 |

### リクエスト/レスポンス設計

#### 成功レスポンス

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "morning_user",
      "email": "user@example.com",
      "created_at": "2024-01-01T09:00:00Z",
      "updated_at": "2024-01-01T09:00:00Z"
    }
  },
  "message": "ユーザー情報を取得しました"
}
```

#### エラーレスポンス

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "入力値が不正です",
    "details": [
      {
        "field": "scheduled_time",
        "message": "アラーム時刻は現在時刻より後である必要があります"
      }
    ]
  }
}
```

### ミドルウェアスタック

```go
1. Recovery（パニックリカバリー）
   ↓
2. Logging（アクセスログ）
   ↓
3. CORS（クロスオリジン対応）
   ↓
4. RateLimit（レート制限）※未実装
   ↓
5. Authentication（認証）
   ↓
6. Authorization（認可）※未実装
   ↓
7. Handler（ビジネスロジック）
```

---

## 実装状況と今後の拡張ポイント

### 現在の実装状況

#### ✅ 実装済み

- **ドメイン層**
  - 全エンティティ（User, MorningCall, Relationship）
  - 値オブジェクト（NGReason, Status）
  - リポジトリインターフェース

- **インフラ層**
  - インメモリリポジトリ（全エンティティ対応）
  - セッション管理
  - パスワードサービス
  - HTTPサーバー基盤

- **ユースケース層**
  - 認証系（Login, Logout）
  - ユーザー系（Register, GetProfile）
  - 関係性系（全機能実装）
  - モーニングコール系（基本CRUD）

- **ハンドラー層**
  - 認証ハンドラー
  - ユーザーハンドラー
  - ベースハンドラー（共通処理）

#### 🚧 実装中/部分実装

- モーニングコールハンドラー（UseCase実装済み、Handler未実装）
- 関係性ハンドラー（UseCase実装済み、Handler未実装）

#### 📋 未実装（計画中）

- **永続化層**
  - MySQL実装
  - トランザクション管理
  - マイグレーション

- **高度な機能**
  - WebSocket（リアルタイム通知）
  - プッシュ通知
  - 繰り返しアラーム
  - スヌーズ機能

- **運用機能**
  - 管理者機能
  - メトリクス収集
  - ログ集約

### 拡張ポイント

#### 1. データベース移行

```go
// 現在のインメモリ実装
type UserRepository struct {
    users map[string]*entity.User
    // ...
}

// MySQL実装への移行例
type MySQLUserRepository struct {
    db *sql.DB
}

func (r *MySQLUserRepository) Create(ctx context.Context, user *entity.User) error {
    query := `INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?)`
    // ...
}
```

#### 2. キャッシュ層の追加

```go
type CachedUserRepository struct {
    cache Cache
    repo  repository.UserRepository
}

func (r *CachedUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
    // キャッシュチェック
    if user, found := r.cache.Get(id); found {
        return user.(*entity.User), nil
    }
    // リポジトリから取得
    user, err := r.repo.FindByID(ctx, id)
    if err == nil {
        r.cache.Set(id, user, 5*time.Minute)
    }
    return user, err
}
```

#### 3. イベント駆動アーキテクチャ

```go
type EventBus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventType string, handler EventHandler) error
}

// モーニングコール作成イベント
type MorningCallCreatedEvent struct {
    MorningCallID string
    SenderID      string
    ReceiverID    string
    ScheduledTime time.Time
}
```

---

## パフォーマンス設計

### インメモリ実装の最適化

#### インデックス戦略

```go
// O(1)検索を実現
usernameIndex map[string]string  // username → ID
emailIndex    map[string]string  // email → ID

// 複合インデックス
userPairIndex map[string]string  // "userA:userB" → relationshipID
```

#### 並行処理制御

```go
type UserRepository struct {
    mu sync.RWMutex  // 読み込み優先のロック
}

// 読み込み処理（並行実行可能）
func (r *UserRepository) FindByID() {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // ...
}

// 書き込み処理（排他制御）
func (r *UserRepository) Create() {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ...
}
```

### サーバー設定の最適化

```go
type ServerConfig struct {
    ReadTimeout     15 * time.Second  // リクエスト読み込みタイムアウト
    WriteTimeout    15 * time.Second  // レスポンス書き込みタイムアウト
    IdleTimeout     60 * time.Second  // アイドル接続タイムアウト
    MaxHeaderBytes  1 << 20           // 1MB
}
```

### ベンチマーク指標（目標値）

| 操作 | 目標レスポンスタイム | 備考 |
|------|---------------------|------|
| ユーザー認証 | < 100ms | bcrypt検証含む |
| ユーザー検索 | < 10ms | インデックス使用 |
| モーニングコール作成 | < 50ms | 関係性確認含む |
| 友達一覧取得 | < 30ms | 100件まで |

---

## セキュリティ設計

### 認証・認可

#### パスワードポリシー

```go
func ValidatePassword(password string) NGReason {
    // 長さ: 8-100文字
    // 必須: 大文字、小文字、数字、特殊文字
    // 強度チェック実装済み
}
```

#### セッション管理

- **セッションID**: 64文字の16進数（256ビット）
- **有効期限**: デフォルト24時間
- **自動クリーンアップ**: 5分間隔
- **同時セッション**: ユーザーごとに複数可能

### 入力検証

#### SQLインジェクション対策

```go
// プリペアドステートメント使用（将来のDB実装）
query := "SELECT * FROM users WHERE id = ?"
rows, err := db.Query(query, userID)
```

#### XSS対策

```go
// HTMLエスケープ（必要に応じて）
func escapeHTML(s string) string {
    return html.EscapeString(s)
}
```

### アクセス制御

```go
// リソースアクセス制御
func (mc *MorningCall) CanBeUpdatedBy(userID string) bool {
    return mc.SenderID == userID && mc.Status == StatusScheduled
}

func (r *Relationship) CanBeAcceptedBy(userID string) bool {
    return r.ReceiverID == userID && r.Status == StatusPending
}
```

### 監査ログ

```go
// アクセスログフォーマット
[POST] /api/v1/auth/login 192.168.1.1 200 125ms
[GET] /api/v1/users/profile 192.168.1.1 401 2ms
```

---

## 開発ワークフロー

### ブランチ戦略

```bash
main
  ├── feature/domain-entities
  ├── feature/usecase-morning-call
  ├── feature/handler-auth
  └── fix/session-timeout
```

### コミット規約

```
feat: 新機能追加
fix: バグ修正
refactor: リファクタリング
test: テスト追加・修正
docs: ドキュメント更新
chore: ビルド・ツール関連
```

### テスト戦略

#### テストピラミッド

```
         /\
        /E2E\      （5%）
       /──────\
      /Integration\ （25%）
     /──────────────\
    /  Unit Tests   \ （70%）
   /──────────────────\
```

#### テストカバレッジ目標

- ドメイン層: 90%以上
- ユースケース層: 80%以上
- ハンドラー層: 70%以上
- 全体: 75%以上

### 品質チェックリスト

- [ ] `gofmt -w .` でコードフォーマット
- [ ] `go build ./...` でビルド確認
- [ ] `go test ./...` でテスト実行
- [ ] `go vet ./...` で静的解析
- [ ] `staticcheck ./...` で追加チェック

---

## 付録

### 用語集

| 用語 | 説明 |
|------|------|
| NGReason | ドメイン検証結果を表す値オブジェクト |
| モーニングコール | ユーザー間で設定するアラーム |
| 関係性（Relationship） | ユーザー間の友達関係 |
| セッション | ログイン状態を保持する仕組み |
| リポジトリ | データ永続化の抽象化層 |
| ユースケース | アプリケーションのビジネスロジック |

### 参考資料

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://martinfowler.com/tags/domain%20driven%20design.html)
- [Go Standard Library](https://pkg.go.dev/std)
- [RESTful API Design](https://restfulapi.net/)

### トラブルシューティング

#### Q: セッションが期限切れになる
A: デフォルトの有効期限は24時間です。環境変数`AUTH_SESSION_TIMEOUT`で調整可能。

#### Q: 友達リクエストが送信できない
A: 既存の関係性（承認待ち、承認済み、ブロック）を確認してください。

#### Q: モーニングコールが作成できない
A: 友達関係の確認、時刻の妥当性（現在〜30日以内）を確認してください。

---

## 変更履歴


---

*このドキュメントは、Morning Call APIの技術仕様を包括的に記述したものです。実装の進捗に応じて継続的に更新されます。*