# Security Policy

## Supported Versions

We actively support the following versions of Sortify with security updates:

| Version | Supported          |
| ------- | ------------------ |
| main    | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of Sortify seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report a Security Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **GitHub Security Advisory** (Preferred)
   - Go to the [Security tab](https://github.com/Steven-harris/sortify/security) in this repository
   - Click "Report a vulnerability"
   - Fill out the security advisory form

2. **Email**
   - Send an email to the repository owner
   - Include the word "SECURITY" in the subject line
   - Provide as much information as possible about the vulnerability

### What to Include in Your Report

Please include the following information in your security report:

- **Description** of the vulnerability
- **Steps to reproduce** the issue
- **Potential impact** of the vulnerability
- **Suggested fix** (if you have one)
- **Your contact information** for follow-up questions

### Response Timeline

You can expect the following response times:

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution**: Varies based on complexity, but we aim for 30 days or less

### Security Measures

Sortify implements several security measures:

#### Code Security
- **Static Analysis**: Automated security scanning with gosec
- **Dependency Scanning**: Regular dependency vulnerability checks with Trivy
- **Container Security**: Docker image security scanning
- **Path Validation**: Comprehensive input validation to prevent directory traversal
- **Secure File Operations**: All file operations use validated paths

#### Infrastructure Security
- **Minimal Docker Images**: Using scratch-based containers to reduce attack surface
- **Non-root Execution**: Application runs as non-root user
- **Secure Defaults**: Conservative file permissions (0600 for files, 0750 for directories)

#### Development Security
- **Automated Testing**: Comprehensive test suite with security-focused tests
- **Code Review**: All changes require review before merging
- **Continuous Integration**: Automated security checks on all commits
- **Dependency Management**: Regular updates and vulnerability monitoring

### Scope

This security policy applies to:

- The main Sortify application (backend and frontend)
- Docker containers and configuration
- CI/CD pipelines and workflows
- Documentation and examples

### Out of Scope

The following are generally considered out of scope:

- Vulnerabilities in third-party dependencies (please report to the respective maintainers)
- Issues in development/testing environments
- Social engineering attacks
- Physical security issues

### Safe Harbor

We support responsible disclosure and will not pursue legal action against researchers who:

- Make a good faith effort to avoid privacy violations and disruption
- Only interact with accounts you own or with explicit permission
- Do not access or modify data belonging to others
- Contact us before making any public disclosure
- Give us reasonable time to respond before disclosure

### Recognition

We appreciate the security community's efforts to improve the security of Sortify. Researchers who responsibly disclose vulnerabilities will be:

- Credited in our security advisories (unless they prefer to remain anonymous)
- Mentioned in our release notes for security fixes
- Added to our security contributors list

### Additional Resources

- [GitHub Security Features](https://docs.github.com/en/code-security)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE/SANS Top 25](https://cwe.mitre.org/top25/)

---

**Thank you for helping keep Sortify and our users safe!**
