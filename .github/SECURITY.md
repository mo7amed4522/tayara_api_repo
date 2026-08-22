# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 2.x.x   | :white_check_mark: |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of our software seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **[security@your-domain.com]**

You should receive a response within 48 hours. If for some reason you do not, please follow up via email to ensure we received your original message.

Please include the following information:
- Type of issue (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### What to Expect

- We will acknowledge receipt of your vulnerability report within 48 hours
- We will send you regular updates about our progress
- We will notify you when the vulnerability is fixed
- We may ask for additional information or guidance

### Security Update Process

1. Security report received and assigned a primary handler
2. The problem is confirmed and affected versions are determined
3. Code is audited to find any similar problems
4. Fixes are prepared for all supported releases
5. New releases are issued and announcements are made

## Preferred Languages

We prefer all communications to be in English or Arabic.

## Security Best Practices

When deploying this application:

- Always use HTTPS in production
- Keep dependencies up to date
- Use strong, unique passwords for database connections
- Enable JWT token expiration
- Implement rate limiting
- Keep your Go version updated
- Review and rotate API keys regularly
- Use environment variables for sensitive configuration
- Enable database connection encryption
- Implement proper CORS policies
- Use prepared statements for all database queries
- Validate and sanitize all user inputs
- Implement proper authentication and authorization checks
- Enable security headers in your web server
- Regular security audits and dependency scans

## Security Headers

Ensure the following security headers are enabled:
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
```

## Third-Party Security Tools

We recommend using the following tools:
- **Dependabot**: Automated dependency updates
- **CodeQL**: Semantic code analysis
- **Snyk**: Dependency vulnerability scanning
- **gosec**: Go security checker

## Attribution

We believe in responsible disclosure and will credit researchers who report vulnerabilities to us (unless they prefer to remain anonymous).
