# Prompt & Coding Instructions
This is the prompot instructions for AI agent to follow for `go-top` project.

## Documentation location
*Requirement Doc* refers to `doc/prd/requirements.md` and `doc/prd/prompt.md` for Product requirement and prompt instructions.

*Design Doc* refers to `doc/design.md` is the main piece of technical documents.

*BDD Doc* refers to `doc/bdd/feature.md` illustrates BDD-style (i.e. Given-When-Then) feature definition.

*Operational Doc* refers to `doc/troubleshooting.md` for technical troubleshooting and release management.

## 1. Design Practice
**DO NOT** proceed to implementation phase until obtaining approval from user.

For every new feature, 
1. Requirement doc
    - Review *Requirement Doc*
2. Design doc
    - Update *Design Doc* first with high level concept.  
    - Here is the top-down structure to follow   
       1. Overview flowchart showing how components work together
       2. Detailed component sections
       3. Include simplified Pseudo code and flowcharts for each component
       4. Edge cases
3. Behavior Driven Doc
    - Update a BDD-style feature definition in *BDD Doc*
4. Cross reference all docs with proper links.
   - feature requirement in *Requirement Doc*.
   - high level design doc at *Design Doc*
   - BDD-style definition in *BDD Doc* 
   - Maintain those links **at all time**.

## 2. Implementation Practice
Follow this practices in priorities:

1. Follow idiomatic Go and [Effective Go](https://go.dev/doc/effective_go) best practices.
2. **Progressive Escalation for Persistent Errors**:
   - If we continuously made the same error more than two times, roll back and think harder again on how should approach better
   - If we continuously made the same error more than two additional times, roll back and ultra think again on how should approach better 
3. Use idiomatic Go idioms and patterns.
4. Before implementation, check interface / struct definition:
    1. For common or shared types, add into `common/type.go` to **avoid import cycle**.
    2. For package specific types, add into `type.go`.
    3. Similar idea applies to constants into `constant.go` and mocks into `mock.go`.
5. Before implementation, **ASK the user** for backward compatiblility requirement. If not required, completely replace functions with old signature with new one.
6. Practice Test Based Development: Test → Implement → Refactor
    1. Create test cases FIRST before any implementation
    2. Write incremental implementation and new tests while maintaining tests. Make sure **tests do NOT fail** for each iteration.
    3. This helps to validate design (detect design flaw early) to prevent regression.
    4. For any unimplemented code block, put a comment starting with "//TODO:<comment about task>" for future reference. 
    5. Constantly review *Design Doc* to make sure implementation is align with it.
7. If the same code change is needed across multiple files, Make sure to examine for syntax error after making updates and before proceeding to the next.     
8. Source code is placed under `pkg/` directory. Organize code by modules (e.g., `config/`, `common/`, `analyzer/`, `formatter/`, `jvm/`, `memory/`, `proc/`).
9. Use interfaces for decoupling and testability.
    - Use interfaces and struct to decouple components and enable testing. 
    - Add interfaces to `type.go` under each package. 
    - Add constants to `constant.go` under each package.
10. Use structured logging for debugging and monitoring.
    - **Logging Framework**: All logging uses `common/logging/StructuredLogger` with `LoggerRegistry` pattern for consistent Component/Subcomponent/Labels hierarchy
    - **Two Implementation Patterns** (struct-based has higher priority):

      **A. Struct-Based Logger Injection (PREFERRED)**:
      ```go
      func NewRiskManager(accountData *client.AccountDetailsData, loggerRegistry *logging.LoggerRegistry) RiskManager {
          logger := loggerRegistry.GetLogger(logging.ModuleTrade, logging.FuncRiskManager)
          return &riskManager{
              accountData: accountData,
              logger:      logger,
          }
      }
      ```

      **B. Package-Scoped Logger (for static functions)**:
      ```go
      // In type.go
      var (
          clientLogger *logging.StructuredLogger
          initOnce sync.Once
      )

      func InitializePackageLogger(registry *logging.LoggerRegistry) {
          initOnce.Do(func() {
              clientLogger = registry.GetLogger(logging.ModuleClient, "static")
          })
      }
      ```

    - **Structured Logging Pattern**:
      ```go
      logger.WithFields(map[string]interface{}{
          "symbol":       "AAPL",
          "strategy":     "collar",
          "request_id":   "req-123",  // 🔥 MANDATORY for trade-related logs
          "action":       "buy_to_open",
      }).Info("Strategy execution completed")

      // ERROR level logs MUST include error_code
      logger.WithFields(map[string]interface{}{
          "request_id": "req-123",
          "error_code": logging.ErrorCodeRiskValidationFailed,  // 🔥 MANDATORY
          "violation_type": "buying_power_insufficient",
      }).Error("Risk validation failed", logging.ErrorCodeRiskValidationFailed)
      ```

    - **Mandatory Fields**:
      - `request_id`: Required for ALL trade-related logs for transaction traceability
      - `error_code`: Required for ALL ERROR level logs for systematic error classification
      - `duration_ms`: Required for API performance logging (endTime - startTime)
      - `session_id`: Required for E2E test logs when session ID is available

    - **Constants Management**: If a label is used 5+ times, create constant in `common/logging/constant.go`

    - **Log Level Classification**:
      - **FATAL**: System cannot continue (startup failures, critical config missing)
      - **ERROR**: Primary feature malfunction (trade failures, API errors) - MUST include error_code
      - **WARN**: Unexpected condition but operation continues (fallbacks, retries)
      - **INFO**: Normal operational events (lifecycle, successful operations)
      - **DEBUG**: Detailed troubleshooting information (internal state, decisions)
      - **TRACE**: Very detailed execution flow (function entry/exit, HTTP details)

    - **No Null Checks**: Never check `if logger != nil` - struct-based loggers are guaranteed to be initialized by LoggerRegistry
11. Always implement validation (i.e. parameter validation) at entry points.
12. Always check and handle errors; return wrapped errors with context.
13. Try to use bulk edit tools (i.e. sed) if possible to improve efficiency. Always verify patterns with sed/grep before replacing to prevent unintended changes.
14. **Code Formatting**: Once coding is complete, ALWAYS run `go fmt ./...` to format all Go files according to Go standards before committing or running tests.

## 3. Testing Practice

### Test Standards Overview
- Use Go build tags for selective test execution

### Build Tags (REQUIRED)
All test files MUST have build tags on the first line with blank line after:

```go
//go:build unit

package client_test
```

**Tag Categories:**
- `unit` - Fast unit tests (< 1s per test)
- `integration` - Integration tests requiring external services
- `benchmark` - Performance benchmark tests
- `core` - Core system tests (main-level, critical flows)
- `e2e` - End-to-end tests (`cmd/e2e/`)
- `tool` - Tool/utility tests (`cmd/tool/validate/`, etc.)

**Tag Combinations:**
- `//go:build unit && core` - Core system unit tests (main_test.go)
- `//go:build integration && core` - Core integration tests (main_integration_test.go)
- `//go:build unit && e2e` - E2E framework unit tests
- `//go:build unit && tool` - Tool unit tests

### File Naming Conventions
- **Unit tests**: `<name>_test.go` (e.g., `alpaca_test.go`)
- **Integration tests**: `<name>_integration_test.go` (e.g., `alpaca_integration_test.go`)
- **Benchmark tests**: `<name>_benchmark_test.go` (e.g., `alpaca_benchmark_test.go`)
- **Mock definitions**: `mock.go` (testify mocks for method tracking)
- **Stub definitions**: `stub.go` (hard-coded test doubles)

### Function Naming Conventions
- **Unit tests**: `Test_<FunctionOrFeature>` (e.g., `Test_GetAccountDetails`)
- **Integration tests**: `TestIntegration_<FunctionOrFeature>` (e.g., `TestIntegration_AlpacaOrderPlacement`)
- **Benchmark tests**: `Benchmark_<FunctionOrFeature>` (e.g., `Benchmark_StrategyEvaluation`)

### Mock and Stub Naming
- **Testify mocks**: `MockXXXX` in `mock.go` (e.g., `MockAccountDataProvider`)
- **Stub implementations**: `StubXXXX` in `stub.go` (e.g., `StubHttpClient`)

### Black-box Testing (Optional)
- **Preferred**: Use `package xxx_test` for exported API testing
- **Allowed**: Use `package xxx` (same package) for complex internal logic
- Document rationale when using white-box testing

### Test Execution Commands
```bash
# Unit tests (fast, runs in CI gate)
go test -tags=unit ./...
go test -tags="unit && core" ./...

# Integration tests (slower, requires credentials)
go test -tags=integration ./...
go test -tags="integration && core" ./...

# E2E and tool tests
go test -tags="tool || e2e" ./...

# Benchmark tests
go test -tags=benchmark -bench=. -benchmem ./...

# Run with race detector
go test -race -tags=unit ./...
```

### Centralized Test Setup Pattern (REQUIRED)

Every package MUST include a `test_setup.go` file with standardized setup/teardown functions:

```go
// Package-level setup/teardown (runs once for entire test suite)
func setupSuite(tb testing.TB) func(tb testing.TB) {
    // Suite-level setup: Database connections, mock servers, logger configuration
    return func(tb testing.TB) {
        // Suite-level teardown: Cleanup resources
    }
}

// Test-level setup/teardown (runs before/after each test case)
func setupTest(tb testing.TB) func(tb testing.TB) {
    // Test-level setup: Per-test resource initialization
    return func(tb testing.TB) {
        // Test-level teardown: Cleanup per-test resources
    }
}
```

**Usage Pattern:**
```go
func TestExample(t *testing.T) {
    teardownSuite := setupSuite(t)
    defer teardownSuite(t)

    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "expected1"},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            teardownTest := setupTest(t)
            defer teardownTest(t)

            // Test logic here
        })
    }
}
```

**File Structure:**
```
client/
├── alpaca.go          # Production code
├── alpaca_test.go     # Unit tests (//go:build unit)
├── test_setup.go      # Centralized test setup/teardown & shared resources
├── mock.go            # Test mocks
└── stub.go            # Test stubs
```

**Beyond Basic Setup/Teardown:**

`test_setup.go` can centralize common test infrastructure including:
- **Test Object Constructors**: Factory functions for creating test fixtures
- **Mock Behavior Configuration**: Reusable mock setups and default expectations
- **Test Data Builders**: Fluent interfaces for building test scenarios
- **Helper Functions**: Common test utilities and assertions

**Example - Centralized Mock Setup:**
```go
// TestSetup provides centralized mock configuration
type TestSetup struct {
    MockDataProvider *client.MockClient
    MockTradeManager *trade.MockTradeManager
    AppConfig        *common.AppConfig
}

// NewTestSetup creates test setup with common mocks initialized
func NewTestSetup(t *testing.T) *TestSetup {
    mockDataProvider := &client.MockClient{}
    mockTradeManager := &trade.MockTradeManager{}

    // Default mock behavior setup
    mockTradeManager.On("ValidateOrder", mock.Anything).Return(nil).Maybe()

    setup := &TestSetup{
        MockDataProvider: mockDataProvider,
        MockTradeManager: mockTradeManager,
        AppConfig:        createDefaultTestConfig(),
    }

    // Auto-verify mock expectations on cleanup
    t.Cleanup(func() {
        mockDataProvider.AssertExpectations(t)
        mockTradeManager.AssertExpectations(t)
    })

    return setup
}

// Fluent builder methods for common scenarios
func (s *TestSetup) WithExistingPosition(symbol string, qty int) *TestSetup {
    s.MockDataProvider.On("GetAccountPositionData", symbol).Return(
        client.AccountPositionData{Qty: qty}, nil,
    )
    return s
}
```

**Benefits of Centralized Shared Resources:**
- Eliminate duplicate mock setup across test files
- Fluent builders make test intent clear
- Consistent mock behaviors across all tests
- Mock changes in one location
- Type-safe test setup with compile-time verification

**Reference**: Complete pattern documentation with advanced examples in `doc/design.md` section 9.1.4.

### Test Development Workflow
- Use table-driven approach to construct tests where possible
- Replace struct-based mocks with `testify` mocks in `mock.go` under each package
- **ALWAYS** implement centralized test setup pattern in `test_setup.go`
- Upon adding new test case, add **one test case at a time** and verify it runs correctly
- **ALWAYS** run `go vet ./...` before running tests to catch code issues early
- Ensure to run `go test -tags=unit -cover -v ./...` for every refactoring round
- Ensure to run `go test -tags=integration ./...` for final validation
- Ensure to build and run entry applications:
    - `main.go`
    - `cmd/tool/validate/main/validate_trade_plans.go`

### CI/CD Integration
- All workflows run `go vet` before tests to catch code issues early
- All workflows support manual execution with branch selection
- CI/CD configuration files located under `.github/workflows/`
- Complete documentation: `doc/design.md` section 10.1 (CI/CD Pipeline)

### Race Detection
- Race detector enabled by default in CI (`-race` flag)
- Run locally: `go test -race -tags=unit ./...`
- Requires CGO and C compiler (pre-installed on CI runners)

**Reference**: Shared test conventions and suite guidance in `doc/design.md` section 9 (see 9.1-9.5)


## 4. Documentation Practice
- Update all docs, *RequirementDoc*, *Design Doc*, *BDD Doc*, *Operational Doc* after **any change** is made.

## 5. Version Control Practice
- Use `git log`, `git add .` and `git diff --cached` to review latest changes.
- Commit message process:
    1. Write commmit message to a file `commit_message.txt`
    2. User review (and update) the commit message.
    3. Commit with `git commit -F commit_message.txt`
    4. remove file `commit_message.txt`
    5. Check status in case anything is not missed: `git status`
- User will review the commited changes then push to remote manually.
- In case user forget, alwaysask for ticket number, i.e. #<ticket number> to be used as commit message prefix. There are exceptions, howeer;
    1. If it is general documentation update, we may use `DOC:` as commit message prefix
    2. If it is general testing update, not tied with specific feature, we may use `TEST:` as commit message prefix
    3. If it is general update to operational scripts, we may use `SCRIPT:` as commit message prefix.
    4. If it is release management related, we may use `RELEASE` as commit message prefix.
    5. If it is none of above, we may use `MAINT:` as commit message prefix.



## 5. Release Management Practice
- Follow release process.
    - **ALWAYS** ask user for version number in case user forgets
    - **ALWAYS** ask user permission to go through each step in the release process.

**Release Process**

    1. Perform Phase 1 to create release branch which will trigger automated deployment to production.
    2. Perform **Post Release Validation** (TBD)
    3. Check validation result
       1. If validation failed:
          1. **Document the specific failure** (error logs, failed commands)
          2. **Assess severity**: Critical (requires rollback) vs. Minor (can be hotfixed)
          3. **Execute rollback procedure** if critical issues detected (see deployment.md "Rollback Procedures")
          4. **Create hotfix branch** for minor issues and apply fix
          5. **Push fix to release branch** to trigger automated deployment
          6. **Re-run validation** after fixes applied (repeat step 2)
    4. If validation passed:
       1. **Perform Phase 2** to create tag for GitHub release:
          ```bash
          # Create version tag (after release is stable)
          git checkout release/vX.Y.Z
          git tag vX.Y.Z
          git push origin vX.Y.Z  # Creates GitHub Release
          ```
       2. **Merge changes back to master**:
          ```bash
          git checkout master
          git merge release/vX.Y.Z
          git push origin master
          ```
       3. **Optional cleanup**:
          ```bash
          git branch -d release/vX.Y.Z  # Delete local branch
          git push origin --delete release/vX.Y.Z  # Delete remote branch
          ```

**Post Release Validation Success Criteria** (all must be met):
- All validation commands execute without errors or timeouts
- No FATAL or ERROR level logs in the last 15 minutes
- Warning count < 5 warnings per 15-minute window  
- All required components show successful initialization (config, trading plans, market data)
- No trading system component errors detected

