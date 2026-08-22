# Contributing to Tayara API

First off, thank you for considering contributing to Tayara API! It's people like you that make this project better for everyone.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples to demonstrate the steps**
- **Describe the behavior you observed and what you expected**
- **Include logs, error messages, and screenshots if applicable**
- **Specify your environment** (OS, Go version, database version)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description of the suggested enhancement**
- **Explain why this enhancement would be useful**
- **Include examples of how it would work**

### Your First Code Contribution

Unsure where to begin? Look for issues labeled:
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention needed
- `documentation` - Documentation improvements

## Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/tayara_api_repo.git
   cd tayara_api_repo
   ```

2. **Install Go 1.25 or later**
   ```bash
   go version
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Install swag for API documentation**
   ```bash
   go install github.com/swaggo/swag/cmd/swag@v1.8.12
   ```

5. **Set up your database**
   - Create a database for development
   - Copy `config.yaml.example` to `config.yaml`
   - Update database credentials

6. **Generate Swagger docs**
   ```bash
   swag init -g main.go -o docs
   ```

7. **Run the application**
   ```bash
   go run main.go
   ```

8. **Run tests**
   ```bash
   go test ./... -v
   ```

## Coding Standards

### Go Code Style

- Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Run `go vet` to catch common mistakes
- Use meaningful variable and function names
- Keep functions small and focused
- Add comments for exported functions and types

### Format Your Code

Before committing:
```bash
# Format code
go fmt ./...

# Run linter
go vet ./...

# Run static analysis (if you have golangci-lint)
golangci-lint run
```

### Project Structure

```
.
├── source/           # Application source code
│   ├── api/         # API handlers
│   ├── service/     # Business logic
│   ├── model/       # Data models
│   ├── router/      # Route definitions
│   └── middleware/  # Middleware functions
├── docs/            # Swagger documentation
├── config.yaml      # Configuration file
└── main.go          # Application entry point
```

## Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes
- `ci`: CI/CD changes

### Examples

```
feat(auth): add JWT token refresh endpoint

Implement token refresh mechanism to allow users to obtain new tokens
without re-authenticating.

Closes #123
```

```
fix(swagger): regenerate docs with swag v1.8.12

Binary literals in API responses were causing parsing errors.
Regenerated with the correct swag version.
```

```
docs(readme): update installation instructions

Added missing step for database migration and clarified
environment variable configuration.
```

## Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feat/your-feature-name
   ```

2. **Make your changes**
   - Write clean, readable code
   - Add tests for new features
   - Update documentation as needed
   - Regenerate Swagger docs if API changed

3. **Update Swagger docs** (if API changed)
   ```bash
   swag init -g main.go -o docs
   ```

4. **Run tests and verify**
   ```bash
   go test ./... -count=1
   go vet ./...
   go fmt ./...
   ```

5. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat(scope): your meaningful commit message"
   ```

6. **Push to your fork**
   ```bash
   git push origin feat/your-feature-name
   ```

7. **Create a Pull Request**
   - Use a clear title describing the change
   - Fill out the PR template completely
   - Link related issues
   - Request review from maintainers

### PR Requirements

- ✅ All tests pass
- ✅ Code is formatted with `gofmt`
- ✅ No `go vet` warnings
- ✅ Swagger docs are up to date (if API changed)
- ✅ New features have tests
- ✅ Documentation is updated
- ✅ Commit messages follow conventions
- ✅ No merge conflicts

### PR Review Process

- Maintainers will review your PR within 2-3 business days
- Address any requested changes
- Once approved, a maintainer will merge your PR
- Your contribution will be included in the next release

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests with verbose output
go test ./... -v

# Run specific test
go test ./source/service -run TestUserService
```

### Writing Tests

- Write tests for all new features
- Aim for at least 80% code coverage
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test both success and error cases

Example test structure:
```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   User
        want    User
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Documentation

### API Documentation

- Add Swagger annotations for all API endpoints
- Regenerate docs after API changes: `swag init -g main.go -o docs`
- Test the documentation at `/swagger/index.html`

### Code Documentation

- Document all exported functions, types, and constants
- Use godoc style comments
- Provide examples for complex functionality

### README Updates

- Update README.md if you add new features
- Keep installation and setup instructions current
- Add examples for new functionality

## Getting Help

- 💬 **Discussions**: Use GitHub Discussions for questions
- 🐛 **Issues**: Report bugs via GitHub Issues
- 📧 **Email**: Contact maintainers directly for sensitive matters

## Recognition

Contributors will be recognized in:
- Release notes
- CONTRIBUTORS.md file
- GitHub contributors page

Thank you for contributing to Tayara API! 🎉
