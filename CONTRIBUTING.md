# Contributing to NexusCRM

We welcome contributions to NexusCRM! This document outlines the standards and process for contributing.

## Development Setup

1. **Prerequisites**:
   - Go 1.24+
   - Node.js 20+
   - TiDB Cloud (or MySQL 8.0+)

2. **Installation**:
   ```bash
   npm install
   cp .env.example .env            # Root config (TiDB, JWT)
   cp .env.example frontend/.env   # Frontend config (API URL)
   # Update .env with your TiDB credentials
   ```

3. **Running Locally**:
   ```bash
   npm run dev:full
   ```

## Code Standards

### Backend (Go)
- Follow **Clean Architecture** (handlers -> services -> infrastructure).
- Use `errcheck` to ensure all errors are handled.
- Run `go fmt ./...` before committing.

### Frontend (React)
- Use **Functional Components** and Hooks.
- Avoid `any` types; strictly type all props and state.
- Use `lucide-react` for icons.

## Submitting a Pull Request

1. Create a feature branch (`feature/my-feature`).
2. Ensure all tests pass:
   ```bash
   npm run test
   ```
3. Submit PR with a clear description of changes.

## License

By contributing, you agree that your contributions will be licensed under the project's license.
