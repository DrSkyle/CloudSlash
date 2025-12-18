# Contributing to CloudSlash

This guide outlines the development standards and contribution workflow for CloudSlash.

## Open Core Model

CloudSlash is Open Source software licensed under the **GNU Affero General Public License v3.0 (AGPLv3)**.

- **Core Engine**: The scanning logic, heuristics, and TUI are open source. You are encouraged to audit, improve, and extend them.
- **Enterprise Features**: Automated reporting and remediation generation are reserved for licensed users.

### License Enforcement

You may not modify the source code to bypass or remove license checks for the purpose of redistributing the software as a commercial product or SaaS service. This is a violation of the AGPLv3.

## Workflow

### Reporting Issues

1. Search existing issues to avoid duplicates.
2. Include the output of `cloudslash --version` (or the commit hash).
3. Provide a minimal reproduction or a sanitized log output.

### Pull Requests

1. Fork the repository.
2. Create a feature branch (`git checkout -b feat/my-feature`).
3. Ensure your code compiles: `go build ./cmd/cloudslash`.
4. Format your code: `go fmt ./...`.
5. Submit a Pull Request.

## Development

**Prerequisites**: Go 1.25+

1. **Clone**:

   ```bash
   git clone https://github.com/DrSkyle/CloudSlash.git
   cd CloudSlash
   ```

2. **Run**:

   ```bash
   go run ./cmd/cloudslash
   ```

## Support

If this tool saved you time or money, you can support development via the [Support Section](https://cloudslash.pages.dev/#support) on our website.

3. **Mock Mode** (Run without AWS credentials):
   ```bash
   go run ./cmd/cloudslash --mock
   ```

## Coding Standards

- **Error Handling**: Never ignore errors. Wrap them with context or handle them explicitly.
- **AWS API**: Always verify identity and region before making calls. Assume credentials may be missing or expired.
- **UI**: Maintain the existing TUI aesthetic defined in `internal/ui`. Do not introduce new libraries without prior discussion.

## Legal

By contributing to this repository, you agree that your contributions will be licensed under the AGPLv3.
