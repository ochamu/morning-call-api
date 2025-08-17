# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Morning Call API - A Go-based REST API server that enables users to set alarms (morning calls) for their friends, providing wake-up experiences between connected users.

## Architecture

### Domain-Driven Design Structure
The project follows a clean architecture pattern with clear separation of concerns:
- **Domain Layer**: Core business entities and rules (User, MorningCall, Relationship)
- **UseCase Layer**: Application business logic
- **Handler Layer**: HTTP request/response handling
- **Repository Layer**: Data persistence (initially in-memory, planned MySQL migration)

### Core Domain Models

1. **MorningCall**: Represents an alarm set by one user for another
   - Fields: ID, SenderID, ReceiverID, Time, Message, Status
   
2. **User**: Represents a system user
   - Fields: ID, Username, Email, CreatedAt, UpdatedAt
   
3. **Relationship**: Manages friend connections between users
   - Fields: ID, RequesterID, ReceiverID, Status, CreatedAt, UpdatedAt

### Error Handling Pattern - NGReason
The project uses a custom `NGReason` type for domain validation:
- Empty string represents success (OK)
- Non-empty string contains the error message (NG)
- Domain models implement validation methods returning NGReason
- UseCase layer checks NGReason and converts to appropriate errors
- Handler layer maps errors to HTTP status codes

## Development Commands

### Build & Run
```bash
# Build the application
go build -o morning-call-api

# Run the application
go run main.go

# Run with race detector during development
go run -race main.go
```

### Code Quality
```bash
# Format code
gofmt -w .

# Static analysis
staticcheck ./...

# Vet code for suspicious constructs
go vet ./...
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestFunctionName ./...

# Run tests with verbose output
go test -v ./...
```

### Module Management
```bash
# Initialize module (already done)
go mod init github.com/ochamu/morning-call-api

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Key Implementation Notes

1. **Standard Library Only**: The initial implementation uses only Go's standard library (no external frameworks)

2. **Authentication**: Basic authentication planned for initial implementation

3. **Data Storage**: In-memory storage initially, with planned migration to MySQL

4. **API Features**:
   - Friend alarm management (CRUD operations)
   - Sent alarms tracking
   - Wake confirmation functionality
   - Friend management (add/block)
   - Received alarms viewing

5. **Out of Scope**:
   - Self-alarms are not supported
   - Alarm approval/rejection mechanism not implemented

## Development Workflow Policy

### Small & Iterative Implementation Approach

Development must follow a **small, iterative implementation** approach. Adhere to the following workflow:

1. **Create Feature Branches for Small Units**
   ```bash
   # Example: Domain entity implementation
   git checkout -b feature/domain-entities
   
   # Example: Specific use case implementation
   git checkout -b feature/usecase-create-morning-call
   ```

2. **Implementation Cycle (Repeat for Each Small Feature)**
   ```bash
   # Step 1: Implement one file or one feature
   # Step 2: Format code
   gofmt -w .
   
   # Step 3: Verify implementation
   go build ./...          # Build check
   go test ./...           # Run tests
   go vet ./...            # Static analysis
   
   # Step 4: Fix any issues
   # Step 5: Request user confirmation before commit
   # IMPORTANT: Always ask "Ready to commit [feature description]?" 
   # and wait for user approval
   
   # Step 6: Commit (only after user confirmation)
   git add .
   git commit -m "feat: implement User entity with validation"
   
   # Step 7: Push
   git push -u origin feature/domain-entities
   ```

3. **Commit Granularity Guidelines**
   - 1 entity = 1 commit
   - 1 use case = 1 commit
   - 1 endpoint = 1 commit
   - 1 bug fix = 1 commit

4. **Commit Message Convention**
   ```
   feat: Add new feature
   fix: Fix bug
   refactor: Refactor code
   test: Add or modify tests
   docs: Update documentation
   chore: Build/tool related changes
   ```

### Implementation Priority Order

To ensure each unit can be verified independently:

1. **Minimum Verifiable Units**
   - Domain entity → Test → Commit
   - Repository interface → In-memory implementation → Test → Commit
   - Single use case → Test → Commit
   - Single HTTP handler → Verify → Commit

2. **Gradual Integration**
   - Keep each layer independently verifiable
   - Minimize dependencies
   - Use mocks for layer-isolated testing

### Quality Checklist (Required Before Each Commit)

- [ ] Code formatted with `gofmt -w .`
- [ ] Build passes with `go build ./...`
- [ ] Tests pass with `go test ./...`
- [ ] No warnings from `go vet ./...`
- [ ] Tests added for new code
- [ ] Proper error handling implemented
- [ ] Appropriate logging added

### Key Principles

- **Never commit large changes** - Break down into smaller, logical commits
- **Each commit should be deployable** - No broken builds
- **Test as you go** - Don't accumulate untested code
- **Format consistently** - Always run gofmt before committing

### User Confirmation Policy

**IMPORTANT: User confirmation is REQUIRED before every commit**

- **Never auto-commit** - Always request explicit user approval
- **Present implementation summary** - Show what was implemented before asking for confirmation
- **Wait for user response** - Do not proceed with commit until user explicitly approves
- **Confirmation format**: "Ready to commit: [brief description of changes]? (yes/no)"

Example workflow:
```
1. Implementation complete
2. Tests passing
3. Code formatted
4. Present to user: "Ready to commit: User entity with validation logic? (yes/no)"
5. Wait for user confirmation
6. Only proceed with git commit after receiving "yes"
```

## Go-Specific Guidelines

- Use Go 1.25.0 or later
- Always handle errors explicitly - no silent failures with `_`
- Keep interfaces minimal and focused (ISP)
- Use meaningful variable and function names
- Follow Go naming conventions (exported vs unexported)
- Use context.Context for request-scoped values and cancellation
- Implement graceful shutdown for the HTTP server
- Use proper HTTP status codes and RESTful conventions

## Language Usage Guidelines

### Japanese for Business Logic

**Use Japanese for:**
- Business logic error messages
- Domain validation messages
- Code comments explaining business rules
- NGReason validation messages

**Examples:**
```go
// ValidateTime はアラーム時刻の妥当性を検証します
func (m *MorningCall) ValidateTime() NGReason {
    if m.Time.Before(time.Now()) {
        return NGReason("アラーム時刻は現在時刻より後である必要があります")
    }
    if m.Time.After(time.Now().Add(30 * 24 * time.Hour)) {
        return NGReason("アラーム時刻は30日以内で設定してください")
    }
    return NGReason("")
}

// CanSendMorningCall は指定したユーザーにモーニングコールを送信可能か検証します
func (u *User) CanSendMorningCall(receiverID string) error {
    if u.ID == receiverID {
        return fmt.Errorf("自分自身にモーニングコールを設定することはできません")
    }
    // 友達関係の確認
    if !u.IsFriendWith(receiverID) {
        return fmt.Errorf("友達関係にないユーザーにはモーニングコールを設定できません")
    }
    return nil
}
```

### English for Technical Elements

**Use English for:**
- Variable names and function names
- Technical error messages (system errors, not business errors)
- Infrastructure layer comments
- API response field names
- Log messages for debugging

**Examples:**
```go
// NewUserRepository creates a new in-memory user repository
func NewUserRepository() *UserRepository {
    return &UserRepository{
        users: make(map[string]*User),
        mutex: &sync.RWMutex{},
    }
}

// System error (English)
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    return fmt.Errorf("failed to decode request body: %w", err)
}
```

## Testing Strategy

- Unit tests for domain logic and validation
- Integration tests for use cases
- HTTP handler tests using httptest package
- Table-driven tests for comprehensive coverage
- Mock repositories for testing business logic in isolation

## Implementation Plan

### Phase 1: Project Foundation (Current)
1. **Directory Structure Setup**
   - Create clean architecture layers
   - Establish package boundaries
   - Set up test structure

2. **Domain Layer Implementation**
   - Core entities (User, MorningCall, Relationship)
   - Value objects (Status types, NGReason)
   - Repository interfaces
   - Domain validation logic

3. **Infrastructure Layer - In-Memory Storage**
   - In-memory repository implementations with indexing
   - Basic authentication service
   - Configuration management

### Phase 2: Core Business Logic
1. **UseCase Layer**
   - User registration and authentication
   - Friend relationship management
   - Morning call CRUD operations
   - Wake confirmation logic

2. **Handler Layer**
   - HTTP handlers for each resource
   - Request/Response DTOs
   - Middleware (auth, logging, CORS)
   - Error response formatting

### Phase 3: API Integration
1. **Router Setup**
   - RESTful endpoint configuration
   - Middleware chain setup
   - Request validation

2. **Testing**
   - Unit tests for domain logic
   - Integration tests for use cases
   - HTTP handler tests
   - Table-driven test patterns

### Phase 4: Future Enhancements
1. **MySQL Migration** (Future)
   - Repository implementations for MySQL
   - Transaction management
   - Database migrations

2. **Advanced Features** (Future)
   - JWT authentication
   - WebSocket for real-time notifications
   - Recurring alarms
   - Push notifications

## Directory Structure (Final)

```
morning-call-api/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── user.go
│   │   │   ├── morning_call.go
│   │   │   └── relationship.go
│   │   ├── valueobject/
│   │   │   ├── status.go
│   │   │   └── ng_reason.go
│   │   └── repository/
│   │       └── interfaces.go    # Repository interfaces
│   ├── usecase/
│   │   ├── auth/
│   │   │   └── auth_usecase.go
│   │   ├── user/
│   │   │   └── user_usecase.go
│   │   ├── morning_call/
│   │   │   ├── create.go
│   │   │   ├── update.go
│   │   │   ├── delete.go
│   │   │   ├── list.go
│   │   │   └── confirm_wake.go
│   │   └── relationship/
│   │       ├── add_friend.go
│   │       └── block_friend.go
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── morning_call_handler.go
│   │   ├── relationship_handler.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── logging.go
│   │   │   └── cors.go
│   │   └── dto/
│   │       ├── request/
│   │       └── response/
│   ├── infrastructure/
│   │   ├── memory/
│   │   │   ├── user_repository.go
│   │   │   ├── morning_call_repository.go
│   │   │   └── relationship_repository.go
│   │   ├── auth/
│   │   │   └── basic_auth.go
│   │   └── server/
│   │       └── http.go
│   └── config/
│       └── config.go
├── pkg/
│   ├── errors/
│   │   └── errors.go
│   └── utils/
│       ├── uuid.go
│       └── time.go
├── tests/
│   ├── integration/
│   └── fixtures/
├── go.mod
├── go.sum
├── Makefile
├── CLAUDE.md
└── README.md
```

## API Endpoints Summary

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout

### Users
- `POST /api/v1/users/register` - Register new user
- `GET /api/v1/users/me` - Get current user profile
- `GET /api/v1/users/search` - Search users

### Relationships
- `POST /api/v1/relationships/request` - Send friend request
- `PUT /api/v1/relationships/{id}/accept` - Accept friend request
- `PUT /api/v1/relationships/{id}/block` - Block user
- `DELETE /api/v1/relationships/{id}` - Remove relationship
- `GET /api/v1/relationships/friends` - List friends
- `GET /api/v1/relationships/requests` - List friend requests

### Morning Calls
- `POST /api/v1/morning-calls` - Create morning call
- `GET /api/v1/morning-calls/{id}` - Get morning call details
- `PUT /api/v1/morning-calls/{id}` - Update morning call
- `DELETE /api/v1/morning-calls/{id}` - Delete morning call
- `GET /api/v1/morning-calls/sent` - List sent calls
- `GET /api/v1/morning-calls/received` - List received calls
- `PUT /api/v1/morning-calls/{id}/confirm` - Confirm wake up

## Enhanced Error Handling Strategy

### Domain Validation
```go
type DomainError struct {
    Code    string
    Message string
    Field   string
}

type ValidationResult struct {
    Errors []DomainError
}
```

### HTTP Error Response Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input provided",
    "details": [
      {
        "field": "scheduled_time",
        "message": "Must be in the future"
      }
    ]
  },
  "request_id": "uuid-v4"
}
```

## Implementation Priority

### Sprint 1 (Foundation)
1. Domain entities and value objects
2. Repository interfaces
3. In-memory repository implementations
4. Basic configuration

### Sprint 2 (Core Features)
1. User registration and authentication
2. Friend relationship management
3. Morning call creation and listing

### Sprint 3 (Advanced Features)
1. Wake confirmation
2. Comprehensive error handling
3. Middleware implementation
4. Integration tests

### Sprint 4 (Polish)
1. Performance optimization
2. Logging and monitoring
3. Documentation
4. Deployment preparation
