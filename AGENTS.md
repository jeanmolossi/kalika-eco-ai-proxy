# Agent Instructions

## Tooling Requirements
- Use **GolangCI-Lint >= 2.6.2** and **Go >= 1.25**. Any downgrade of these tools is forbidden unless explicitly requested.

## GolangCI-Lint 2.6.2+ Installation Guide
1. Install Go 1.25+ following https://go.dev/doc/install if not already available.
2. Download GolangCI-Lint 2.6.2 (or newer 2.x release) via the official install script:
   ```bash
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
     | sh -s -- -b /usr/local/bin v2.6.2
   ```
3. Verify installation:
   ```bash
   golangci-lint --version
   ```
4. Run `golangci-lint run` from the repository root to lint the project.
