# Contributing to EFK Stack Operator

Thank you for your interest in contributing to the EFK Stack Operator! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please be respectful and constructive in all interactions.

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check if the issue has already been reported:
- Search existing [GitHub Issues](https://github.com/zlorgoncho1/efk-operator/issues)
- Check closed issues as well

When creating a bug report, please use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md) and include:
- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Environment details (Kubernetes version, operator version, etc.)
- Relevant logs and error messages

### Suggesting Features

Feature requests are welcome! Please use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md) and include:
- Clear description of the feature
- Use case and motivation
- Proposed implementation (if you have ideas)

### Pull Requests

1. **Fork the repository** and create a branch from `main`
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes** following the coding standards below

3. **Test your changes**
   ```bash
   make test-quick
   make test
   ```

4. **Update documentation** if needed

5. **Commit your changes** following the commit message format
   ```bash
   git commit -m "feat: add new feature"
   # or
   git commit -m "fix: resolve bug in controller"
   ```

6. **Push to your fork** and create a Pull Request
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Fill out the PR template** completely

## Development Setup

See [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md) for detailed setup instructions.

Quick setup:
```bash
# Build the development environment
make docker-build

# Start development environment
make dev-up

# Open a development shell
make dev-shell
```

## Coding Standards

### Go Code

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (run `make fmt`)
- Follow the existing code style and patterns
- Add comments for exported functions and types
- Keep functions focused and small

### Code Formatting

```bash
# Format code
make fmt

# Check formatting
make vet
```

### Testing

- Write tests for new features and bug fixes
- Aim for good test coverage
- Use table-driven tests when appropriate
- Test both success and error cases

```bash
# Run tests
make test

# Run quick tests
make test-quick
```

### Documentation

- Update relevant documentation when adding features
- Add code comments for complex logic
- Update examples if API changes
- Keep the README and user guide up to date

## Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples

```
feat(controller): add support for custom Elasticsearch config

Add ability to specify custom Elasticsearch configuration
through the EFKStack spec.

Closes #123
```

```
fix(helm): resolve storage class issue in StatefulSet

The StatefulSet was not using the correct storage class
from the spec. This fix ensures the storage class is
properly applied.

Fixes #456
```

## Pull Request Process

1. **Ensure tests pass**: All tests must pass before merging
2. **Update documentation**: Update relevant docs if needed
3. **Get reviews**: At least one approval is required
4. **Keep PRs focused**: One feature or fix per PR
5. **Update CHANGELOG**: Add an entry for user-facing changes

### PR Checklist

- [ ] Code follows the project's style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (if applicable)
- [ ] Commit messages follow the format
- [ ] PR description is clear and complete

## Review Process

- Maintainers will review PRs within a reasonable timeframe
- Address review comments promptly
- Be open to feedback and suggestions
- Keep discussions constructive and respectful

## Questions?

- Open an issue for questions or discussions
- Check existing documentation in `docs/`
- Review existing code for examples

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

