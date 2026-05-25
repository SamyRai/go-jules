# Security Policy

## Supported Versions

go-jules is pre-1.0. Security fixes are handled on `main` unless a maintained release branch exists.

## Report A Vulnerability

Do not open a public issue for vulnerabilities.

Email: <security@glpx.pro>

Include:

- affected version or commit
- description and impact
- reproduction steps
- relevant logs or proof of concept
- suggested fix, if available

Expected response:

- initial response within 48 hours
- status update within 7 days
- fix timeline based on severity and scope

## Secrets

- Do not commit API keys, tokens, passwords, private keys, or local config containing credentials.
- Prefer environment variables for setup inputs and untracked local config files
  for stored credentials.
- Keep local config files untracked.
- Redact credentials in logs, issue reports, and screenshots.

## Runtime Considerations

- Pass Jules API keys through environment-backed application configuration or a
  secret manager; the SDK does not load config files directly.
- Use least-privilege credentials in applications that wrap this SDK.
- Keep dependencies current and review security scan output in CI.
