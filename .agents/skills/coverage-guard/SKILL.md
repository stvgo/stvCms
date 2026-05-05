# Coverage Guard Skill

Enforces test coverage thresholds for the stvCms Go backend.

## Rules

- **Minimum coverage: 90%** for `./internal/services/...` and `./internal/repository/...`
- CI pipeline enforces this threshold — below 90% = build failure
- Always run coverage *before* pushing to catch failures early

## Workflow

1. After any code change to `internal/services/` or `internal/repository/`, run:
   ```bash
   go test ./internal/services/... ./internal/repository/... -coverprofile=/tmp/cov.out -count=1
   go tool cover -func=/tmp/cov.out | tail -1
   ```
2. If combined coverage < 90%, check per-function coverage:
   ```bash
   go tool cover -func=/tmp/cov.out | awk '$NF+0 < 80' 
   ```
3. Write tests for uncovered functions until coverage ≥ 90%
4. Verify with a final full run before committing

## Uncovered Function Patterns

Common reasons for low coverage:
- **Error paths** (Redis/DB errors) — mock the interface to return errors
- **New functions without tests** — add table-driven tests immediately
- **Constructors** (`New*Service`) — add a simple `NotNil` test
- **Branch conditions** (admin vs non-admin) — test both paths

## Test File Conventions

- Tests co-located with source: `Post_test.go`, `Post_approval_test.go`
- Use `gomock` for interface mocks in `internal/mocks/`
- Use real SQLite DB (`gorm.Open(sqlite.Open(":memory:"))`) for repo tests
- Skip Redis-dependent tests if Redis unavailable: `t.Skip("Redis not available")`
- Regenerate mocks after interface changes:
  ```bash
  ~/go/bin/mockgen -destination=internal/mocks/mock_X.go -package=mocks stvCms/internal/package InterfaceName
  ```

## Key Files

- Backend repo: `~/Documents/stvCms`
- Mocks: `internal/mocks/`
- Models: `internal/models/`
- Repositories: `internal/repository/`
- Services: `internal/services/`