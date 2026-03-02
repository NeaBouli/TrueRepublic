# Contributing to TrueRepublic

Thank you for your interest in contributing to TrueRepublic!

## Getting Started

### Prerequisites

- Go 1.24+
- Rust 1.75+ (for contracts)
- Node.js 18+ (for frontend)
- Git

### Development Setup

```bash
# Clone repository
git clone https://github.com/NeaBouli/TrueRepublic.git
cd TrueRepublic

# Build (CGO required for wasmvm)
CGO_ENABLED=1 make build

# Run tests
make test
```

---

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

**Branch naming:**
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation
- `refactor/` - Code refactoring
- `test/` - Test improvements

### 2. Make Changes

**Code Style:**
- Go: Follow standard `gofmt` formatting
- Rust: Use `cargo fmt`
- JavaScript: Prettier with 2-space indent

**Testing:**
- Add tests for new features
- Maintain 100% test pass rate
- Run `make test` before committing

### 3. Commit Messages

Follow conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `test`: Tests
- `refactor`: Code refactoring
- `chore`: Maintenance

**Examples:**

```
feat(dex): add multi-hop swap routing

Implements BFS algorithm to find optimal swap paths
through multiple pools. Supports up to 5 hops with
atomic execution.

Closes #123
```

```
fix(zkp): prevent nullifier reuse attack

Added nullifier existence check before proof verification.

Fixes #456
```

### 4. Push and Create PR

```bash
# Push branch
git push origin feature/your-feature-name

# Create pull request on GitHub
# - Clear title and description
# - Link related issues
# - Add screenshots if UI changes
```

---

## Testing Guidelines

### Go Tests

```bash
# Run all tests
go test ./... -timeout=600s

# Run specific module
go test ./x/dex/...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race -cover -count=1 -timeout=600s ./...
```

Writing tests:

```go
func TestFeatureName(t *testing.T) {
    // Setup
    keeper, ctx := setupKeeper(t)

    // Execute
    result, err := keeper.DoSomething(ctx, input)

    // Assert
    require.NoError(t, err)
    require.Equal(t, expected, result)
}
```

### Rust Tests

```bash
# Run all contract tests
cd contracts && cargo test --workspace

# Run specific contract
cargo test -p governance-dao

# With output
cargo test -- --nocapture
```

### Frontend Tests

```bash
cd web-wallet

# Run tests
npm test

# Run with coverage
npm test -- --coverage
```

---

## Code Review Process

1. **Automated Checks:**
   - CI must pass (tests, lint, build)
   - No merge conflicts

2. **Manual Review:**
   - Code quality and style
   - Test coverage
   - Documentation updated
   - Breaking changes noted

3. **Approval:**
   - Requires 1 maintainer approval
   - Address review comments
   - Re-request review after changes

4. **Merge:**
   - Squash commits for clean history
   - Update changelog if needed

---

## Areas for Contribution

### High Priority

- **ZKP Circuit Optimization:** Reduce proof generation time
- **DEX Analytics:** More detailed pool statistics
- **IBC Asset Support:** Add more chain integrations
- **Testing:** Increase coverage in edge cases

### Good First Issues

Look for issues labeled `good-first-issue`:
- Documentation improvements
- Test additions
- CLI help text
- Error message clarity

### Feature Requests

Check existing issues or create new ones:
- Describe use case
- Provide examples
- Discuss trade-offs

---

## Community Guidelines

- Be respectful and inclusive
- Help others learn
- Give constructive feedback
- Credit contributors
- Follow Code of Conduct

---

## Release Process

**Versioning:** Semantic versioning (MAJOR.MINOR.PATCH)

**Release Checklist:**
1. Update version in code
2. Update CHANGELOG.md
3. Run full test suite
4. Create release tag
5. Build binaries
6. Publish release notes

---

## Getting Help

- **GitHub Issues:** Bug reports and features
- **GitHub Discussions:** Questions and ideas
- **Documentation:** `docs/` directory
- **Telegram:** [t.me/truerepublic](https://t.me/truerepublic)
- **Email:** p.cypher@protonmail.com

---

## License

By contributing, you agree that your contributions will be licensed
under the same license as the project (Apache 2.0).
