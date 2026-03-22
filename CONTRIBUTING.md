# Contributing

Thanks for your interest in contributing to reseed.

## Setup

```bash
git clone https://github.com/nattergabriel/reseed.git
cd reseed
make setup
```

This requires Go 1.24+ and [golangci-lint](https://golangci-lint.run/).

`make setup` enables pre-commit hooks that run build, vet, and lint checks before each commit.

## Commands

```bash
make build    # build
make test     # run tests
make lint     # go vet + golangci-lint
```

## Guidelines

- Keep it simple. No unnecessary abstractions.
- Wrap errors with context: `fmt.Errorf("doing thing: %w", err)`
- All lint checks must pass before committing.
- No third-party dependencies unless strictly necessary.

## Reporting issues

Use [GitHub Issues](https://github.com/nattergabriel/reseed/issues) for bug reports and feature requests.
