# Password Policy

## Overview

Zeno Auth enforces a strong password policy to protect user accounts from unauthorized access and brute-force attacks.

## Requirements

All passwords must meet the following criteria:

### 1. Minimum Length
- **Requirement:** At least 8 characters
- **Rationale:** Provides sufficient entropy against brute-force attacks
- **Error:** `password must be at least 8 characters long`

### 2. Character Composition

#### Uppercase Letters (Required)
- **Requirement:** At least 1 uppercase letter (A-Z)
- **Rationale:** Increases password complexity
- **Error:** `password must contain at least one uppercase letter`

#### Lowercase Letters (Required)
- **Requirement:** At least 1 lowercase letter (a-z)
- **Rationale:** Increases password complexity
- **Error:** `password must contain at least one lowercase letter`

#### Digits (Required)
- **Requirement:** At least 1 digit (0-9)
- **Rationale:** Adds numeric complexity
- **Error:** `password must contain at least one digit`

#### Special Characters (Optional)
- **Requirement:** Not enforced by default
- **Note:** Can be enabled via configuration
- **Examples:** `!@#$%^&*()_+-=[]{}|;:,.<>?`

### 3. Common Password Check
- **Requirement:** Password must not be in the list of top 100 most common passwords
- **Rationale:** Prevents use of easily guessable passwords
- **Error:** `password is too common, please choose a stronger password`
- **Blocked passwords include:** `password`, `123456`, `qwerty`, `admin`, etc.

## Examples

### ✅ Valid Password Patterns
```
[Word][Word][Numbers]     # Example: SecurePass123
[Word][@][Word][Digit]    # Example: MyP@ssw0rd  
[Word][Year][Symbol]      # Example: Welcome2024!
[Role][Numbers][Word]     # Example: Admin123Pass
```

### ❌ Invalid Passwords
```
password        # Too common
12345678        # No uppercase, no lowercase
abcdefgh        # No uppercase, no digit
ABCDEFGH        # No lowercase, no digit
Password        # No digit
```

## Implementation

The password policy is implemented in `internal/validator/password.go` using the `PasswordValidator` struct.

### Default Configuration

```go
PasswordValidator{
    MinLength:        8,
    RequireUppercase: true,
    RequireLowercase: true,
    RequireDigit:     true,
    RequireSpecial:   false,  // Optional
    CheckCommon:      true,
}
```

### Usage

```go
validator := validator.NewPasswordValidator()
err := validator.Validate(password)
if err != nil {
    // Handle validation error
}
```

## Security Features

### 1. Argon2id Hashing
- All passwords are hashed using Argon2id algorithm
- Memory-hard function resistant to GPU attacks
- Configurable time, memory, and parallelism parameters

### 2. Account Lockout
- After 5 failed login attempts, account is locked for 30 minutes
- Prevents brute-force attacks
- User receives email notification

### 3. Password Reset
- Secure token-based password reset flow
- Tokens expire after 1 hour
- One-time use tokens

### 4. Session Management
- All sessions are revoked when password is changed
- Prevents unauthorized access after password change

## Best Practices

### For Users
1. Use a unique password for each service
2. Consider using a password manager
3. Enable 2FA when available (coming soon)
4. Change password if you suspect compromise

### For Administrators
1. Regularly review failed login attempts
2. Monitor account lockout events
3. Educate users about password security
4. Consider implementing password expiration (optional)

## Compliance

This password policy helps meet:
- **GDPR Article 32:** Security of processing
- **NIST SP 800-63B:** Digital Identity Guidelines
- **OWASP:** Password Storage Cheat Sheet

## Future Enhancements

- [ ] Password expiration (configurable)
- [ ] Password history (prevent reuse)
- [ ] Breach detection (HaveIBeenPwned API)
- [ ] Multi-factor authentication (MFA/2FA)
- [ ] Passkey/WebAuthn support

## Configuration

Password policy can be customized via environment variables (future):

```env
PASSWORD_MIN_LENGTH=8
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_DIGIT=true
PASSWORD_REQUIRE_SPECIAL=false
PASSWORD_CHECK_COMMON=true
```

## Testing

> **⚠️ Security Warning:** Never use hardcoded credentials in production code or documentation. Always use environment variables or secure configuration management.

Test password validation:

```bash
# Run password validator tests
go test ./internal/validator -v -run TestPasswordValidator

# Set up test environment variables
export TEST_EMAIL="user@example.com"
export TEST_WEAK_PASSWORD="weak"
export TEST_USER_NAME="Test User"
export TEST_ORG_NAME="Test Organization"

# Test registration with weak password
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "'${TEST_EMAIL}'",
    "password": "'${TEST_WEAK_PASSWORD}'",
    "full_name": "'${TEST_USER_NAME}'",
    "organization_name": "'${TEST_ORG_NAME}'"
  }'
```

## References

- [NIST Password Guidelines](https://pages.nist.gov/800-63-3/sp800-63b.html)
- [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [Argon2 Specification](https://github.com/P-H-C/phc-winner-argon2)

---

**Last Updated:** 2024  
**Version:** 1.0.0
