# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security seriously at FacePass. If you discover a security vulnerability, please follow responsible disclosure practices.

### How to Report

**DO NOT** report security vulnerabilities through public GitHub issues.

Instead, please send a detailed report to:

1. **Email**: Create a private security advisory on GitHub
2. **GitHub Security Advisories**: Use the "Security" tab → "Report a vulnerability"

### What to Include

Please include as much of the following information as possible:

- **Type of vulnerability** (e.g., authentication bypass, privilege escalation, data exposure)
- **Full paths of source files** related to the vulnerability
- **Step-by-step instructions** to reproduce the issue
- **Proof-of-concept or exploit code** (if available)
- **Impact assessment** of the vulnerability
- **Suggested fix** (if you have one)

### Response Timeline

- **Initial response**: Within 48 hours
- **Status update**: Within 7 days
- **Resolution target**: Within 90 days for critical issues

### What to Expect

1. **Acknowledgment**: We'll confirm receipt of your report
2. **Investigation**: We'll investigate and assess severity
3. **Updates**: We'll keep you informed of progress
4. **Credit**: We'll credit you in the security advisory (unless you prefer anonymity)
5. **Disclosure**: Coordinated disclosure after fix is released

---

## Security Architecture

### Face Data Protection

FacePass uses multiple layers of protection for face biometric data:

#### Encryption at Rest

```
Face Embeddings → NaCl SecretBox (XSalsa20 + Poly1305) → Encrypted Storage
```

- **Algorithm**: NaCl secretbox (XSalsa20 stream cipher + Poly1305 MAC)
- **Key derivation**: Machine-specific key using hardware identifiers
- **Nonce**: Unique random nonce per encryption operation

#### Storage Security

- Face data stored in user home directory (`~/.local/share/facepass/`)
- File permissions: `0700` (owner read/write/execute only)
- Encryption key tied to machine (data not portable)

### Authentication Security

#### Liveness Detection

FacePass implements multi-tier liveness detection to prevent spoofing:

| Level | Protections | Use Case |
|-------|-------------|----------|
| Basic | Blink detection, frame consistency | Low-security |
| Standard | + Movement detection | Normal desktop |
| Strict | + Challenge-response, IR analysis | Secure workstations |
| Paranoid | + Texture analysis, all checks | High-security |

#### Anti-Spoofing Measures

- **Photo attacks**: Blink detection, movement analysis
- **Video attacks**: Frame consistency, micro-movements
- **Screen replay**: Texture/moire pattern detection
- **IR reflection**: Analysis for IR camera environments

### PAM Integration Security

#### Privilege Separation

- PAM module runs with minimal required privileges
- Face recognition process isolated from authentication decisions
- Fallback to password authentication on any failure

#### Timeout Protection

- Default 10-second timeout prevents blocking
- Maximum 3 authentication attempts
- Automatic fallback to password on timeout

---

## Security Best Practices

### For Users

1. **Use IR cameras** when available for better security
2. **Enable strict liveness detection** for sensitive systems
3. **Keep PAM fallback enabled** to prevent lockouts
4. **Regularly update** FacePass for security patches
5. **Test thoroughly** before enabling system-wide PAM

### For Administrators

1. **Audit PAM configuration** before deployment
2. **Monitor authentication logs** (`/var/log/facepass.log`)
3. **Use strict mode** for multi-user systems
4. **Implement network segmentation** for critical systems
5. **Have recovery procedures** (root shell, live USB)

### PAM Configuration Warning

> **WARNING**: Incorrect PAM configuration can lock you out of your system!

Always:
- Keep a root terminal open when testing
- Test with `sudo` only before system-wide
- Have a recovery method ready (live USB, single-user mode)

---

## Known Security Considerations

### Current Limitations

1. **Single-factor authentication**: Face recognition should complement, not replace passwords for high-security needs
2. **Identical twins**: May have similar face embeddings
3. **Extreme lighting**: Can affect recognition accuracy
4. **Hardware dependency**: Security depends on camera quality

### Threat Model

FacePass is designed to protect against:

- ✅ Photo-based spoofing attempts
- ✅ Video replay attacks
- ✅ Screen-based attacks (with texture analysis)
- ✅ Casual impersonation attempts
- ✅ Data theft from storage (encryption)

FacePass is **NOT** designed to protect against:

- ❌ Sophisticated 3D mask attacks
- ❌ Nation-state level adversaries
- ❌ Physical coercion
- ❌ Compromised system (root access)

### Recommended Security Layers

For high-security environments, combine FacePass with:

1. Strong password (second factor)
2. Hardware security key (FIDO2)
3. Full disk encryption
4. Network-level access controls
5. Physical security measures

---

## Security Checklist

### Before Deployment

- [ ] Reviewed PAM configuration
- [ ] Tested authentication thoroughly
- [ ] Enabled appropriate liveness level
- [ ] Configured timeout and fallback
- [ ] Set up logging and monitoring
- [ ] Documented recovery procedures

### Ongoing

- [ ] Regular security updates
- [ ] Log monitoring
- [ ] Periodic re-enrollment
- [ ] Configuration audits

---

## Vulnerability Disclosure History

| Date | Version | Severity | Description | CVE |
|------|---------|----------|-------------|-----|
| - | - | - | No vulnerabilities reported yet | - |

---

## Contact

For security-related questions that are not vulnerabilities, please open a GitHub issue with the `security-question` label.

For vulnerability reports, use GitHub Security Advisories or contact maintainers directly.

---

## Acknowledgments

We thank all security researchers who responsibly disclose vulnerabilities. Contributors will be acknowledged here (with permission):

- *Your name could be here*

---

## License

This security policy is part of the FacePass project and is licensed under MIT License.
