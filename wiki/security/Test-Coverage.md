# Test Coverage

Testing documentation and coverage statistics for TrueRepublic.

## Overall Statistics

| Module | Tests | Lines of Test Code | Coverage |
|--------|-------|-------------------|----------|
| x/truedemocracy | 116 | ~1,800 | 45.5% |
| x/dex | 24 | ~400 | 26.2% |
| treasury/keeper | 31 | ~500 | 97.0% |
| **Total Backend** | **182** | **~2,705** | **52.1%** |

**Goal:** 80% coverage across all codebases

---

## Test Types

### Unit Tests (182 tests)

**Running Tests:**

```bash
# Run all tests
make test

# Run with coverage
go test ./... -cover

# Run with race detector
go test ./... -race -cover

# Specific module
go test ./x/truedemocracy/... -v

# Specific test
go test ./x/truedemocracy/... -run TestCreateDomain -v
```

**Test Structure:**

```go
func TestCreateDomain(t *testing.T) {
    // Setup
    keeper, ctx := setupKeeper(t)

    // Execute
    err := keeper.CreateDomain(ctx, "test-domain", coins)

    // Assert
    require.NoError(t, err)

    domain := keeper.GetDomain(ctx, "test-domain")
    require.Equal(t, "test-domain", domain.Name)
}
```

### Test Breakdown by Module

**x/truedemocracy (116 tests):**

| Test File | Tests | What It Covers |
|-----------|-------|---------------|
| stones_test.go | 20 | Stone placement, VoteToEarn, list sorting |
| lifecycle_test.go | 22 | Green/yellow/red zones, auto-delete, fast-delete |
| governance_test.go | 27 | Admin election, member exclusion, inactivity cleanup |
| anonymity_test.go | 15 | Permission register, domain key pairs, purge |
| validator_test.go | 26 | PoD registration, staking, transfer limits |
| slashing_test.go | 6 | Double-sign (5%), downtime (1%), jailing |

**x/dex (24 tests):**

| Test File | Tests | What It Covers |
|-----------|-------|---------------|
| keeper_test.go | 24 | Pool creation, swaps, liquidity, fees, burn |

**treasury/keeper (31 tests):**

| Test File | Tests | What It Covers |
|-----------|-------|---------------|
| rewards_test.go | 31 | Equations 1-5, domain interest, staking rewards, decay |

---

## Coverage by File

### x/truedemocracy

```
keeper.go         78%  - Domain CRUD, proposals, ratings
stones.go         72%  - VoteToEarn, stone voting, list sorting
lifecycle.go      68%  - Green/yellow/red zones, auto-delete
governance.go     65%  - Admin election, exclusion, cleanup
anonymity.go      61%  - Permission register, domain keys
validator.go      23%  ⚠️ - PoD registration, staking, transfer limits
slashing.go       18%  ⚠️ - Double-sign, downtime, jailing
msgs.go           85%  - Message validation
msg_server.go     42%  - gRPC message handlers
querier.go        91%  - ABCI query routes
query_server.go   88%  - gRPC query handlers
types.go          95%  - Data structures
module.go         35%  - Module wiring, EndBlock
```

### x/dex

```
keeper.go         45%  ⚠️ - CreatePool, Swap, Add/RemoveLiquidity
msgs.go           82%  - Message validation
msg_server.go     38%  ⚠️ - gRPC message handlers
querier.go        82%  - ABCI query routes
query_server.go   85%  - gRPC query handlers
types.go          92%  - Data structures
module.go         30%  - Module wiring
```

### treasury/keeper

```
rewards.go        97%  - All 5 equations fully tested
```

---

## Frontend Tests

### Component Tests

**React Testing Library:**

```javascript
test('DomainList displays domains', () => {
  const domains = [
    { name: 'tech', members: 42 },
    { name: 'climate', members: 18 }
  ];

  render(<DomainList domains={domains} />);

  expect(screen.getByText('tech')).toBeInTheDocument();
  expect(screen.getByText('42 members')).toBeInTheDocument();
});
```

**Running Tests:**

```bash
cd web-wallet
npm test

# With coverage
npm test -- --coverage
```

### Smart Contract Tests

**Rust Tests:**

```rust
#[test]
fn test_submit_proposal() {
    let mut deps = mock_dependencies();

    // Instantiate contract
    let info = mock_info("creator", &[]);
    instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

    // Submit proposal
    let msg = ExecuteMsg::SubmitProposal {
        issue: "Test".to_string(),
        suggestion: "Solution".to_string(),
    };

    let res = execute(deps.as_mut(), mock_env(), info, msg).unwrap();
    assert_eq!(res.attributes[0].value, "proposal_created");
}
```

**Running:**

```bash
cd contracts/governance
cargo test
```

---

## CI/CD Testing

### GitHub Actions

**Go CI (.github/workflows/go-ci.yml):**

```yaml
name: Go CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.23
      - run: make test
      - run: go test ./... -cover -coverprofile=coverage.out
```

**React CI (.github/workflows/react-ci.yml):**

```yaml
name: React CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-node@v2
        with:
          node-version: 18
      - run: cd web-wallet && npm ci
      - run: npm test -- --coverage
```

**Rust CI (.github/workflows/rust-ci.yml):**

```yaml
name: Rust CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
      - run: cd contracts && cargo test
```

---

## Test Improvements Needed

### High Priority

| File | Current | Target | Action Items |
|------|---------|--------|-------------|
| x/dex/keeper.go | 45% | 80% | Edge cases, AMM precision, slippage |
| x/truedemocracy/validator.go | 23% | 80% | PoD validation, stake provenance |
| x/truedemocracy/slashing.go | 18% | 80% | Slashing conditions, unjail flow |
| x/dex/msg_server.go | 38% | 80% | Handler error paths |

### Medium Priority

| File | Current | Target | Action Items |
|------|---------|--------|-------------|
| x/truedemocracy/msg_server.go | 42% | 70% | Handler edge cases |
| x/truedemocracy/module.go | 35% | 70% | EndBlock processing |
| x/dex/module.go | 30% | 70% | Module wiring |

### Low Priority

| File | Current | Target | Action Items |
|------|---------|--------|-------------|
| Frontend hooks | 54% | 70% | Connection errors, auto-refresh |
| Frontend components | 72% | 80% | Loading states, edge cases |

---

## Manual Testing Checklist

### Before Each Release

**Backend:**
- [ ] All unit tests pass (`make test`)
- [ ] No race conditions (`go test -race ./...`)
- [ ] Coverage threshold met (>50%)
- [ ] Linting passes (`go vet ./...`)

**Frontend:**
- [ ] Component tests pass (`npm test`)
- [ ] Manual smoke test (create domain, submit proposal, vote)
- [ ] Mobile responsive check
- [ ] Browser compatibility (Chrome, Firefox, Brave)

**Smart Contracts:**
- [ ] All Rust tests pass (`cargo test`)
- [ ] Clippy warnings resolved
- [ ] Contract builds cleanly

**Security:**
- [ ] Dependency scan (`npm audit`, `go mod tidy`)
- [ ] Static analysis (`gosec`, ESLint)
- [ ] Manual security review of changes

---

## Next Steps

- [Known Issues](Known-Issues) -- Current bugs
- [Audit Reports](Audit-Reports) -- Security audits
- [Best Practices](Best-Practices) -- Development guidelines
