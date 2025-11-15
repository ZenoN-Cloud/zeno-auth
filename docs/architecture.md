# Zeno Auth – Architecture

Core authentication and identity service for the **ZenoN-Cloud** platform.

This document describes the architecture, data model, token model and integration patterns for the `zeno-auth` service. It is the single source of truth for how authentication and identity work across the ZenoN-Cloud ecosystem.

---

## 1. Role of Zeno Auth in the platform

### 1.1 Responsibilities

`zeno-auth` is responsible for:

- User accounts:
    - registration and login
    - password management
    - account activation/deactivation
- Organizations (tenants):
    - creation of organizations
    - ownership and membership
    - role-based access at organization level
- Authentication:
    - issuing access and refresh tokens
    - validating credentials
- Identity boundary:
    - providing a single source of truth for user identity
    - exposing identity information to other services in a controlled way

### 1.2 Non-responsibilities

`zeno-auth` explicitly **does not**:

- store or process business documents (bank statements, invoices, identity documents)
- implement business rules of `zeno-docs` or `zeno-id-docs`
- manage billing or subscription plans
- serve as an API gateway

Those concerns belong to other services:

- Financial documents: [Zeno Docs](https://github.com/ZenoN-Cloud/zeno-docs)
- Identity documents: [Zeno ID Docs](https://github.com/ZenoN-Cloud/zeno-id-docs)
- Infrastructure: [Zeno Infra](https://github.com/ZenoN-Cloud/zeno-infra)

---

## 2. High-level overview

ZenoN-Cloud is a multi-tenant SaaS platform. Each tenant is an **organization** (customer). Users can belong to one or more organizations with different roles.

`zeno-auth` sits at the center of the platform:

- It owns the **global user identity** (`users` table).
- It owns **organizations** and **memberships**.
- It issues **JWT access tokens** that other services validate.
- It issues and stores **refresh tokens** for session management.

Other services (e.g. `zeno-docs`, `zeno-id-docs`) trust `zeno-auth` via:

- public signing keys (for validating JWTs)
- optional service-to-service calls (e.g. `/me`, `/introspect`) if needed later

---

## 3. Data model

### 3.1 Users

Represents a human user of the platform.

**Table: `users`**

Required fields:

- `id` (UUID, PK)
- `email` (unique, indexed, lower-cased)
- `password_hash` (Argon2id / bcrypt hash, never store raw password)
- `full_name` (optional)
- `is_active` (bool, default: true)
- `created_at` (timestamp with time zone)
- `updated_at` (timestamp with time zone)

Constraints & notes:

- `email` must be unique.
- Users exist independently of organizations (a user can join multiple orgs).
- Deactivating a user (`is_active = false`) prevents login and token issuing.

---

### 3.2 Organizations

Represents a customer account / tenant (company, business, team).

**Table: `organizations`**

Fields:

- `id` (UUID, PK)
- `name` (string, required)
- `owner_user_id` (UUID, FK → `users.id`)
- `status` (enum: `active`, `trial`, `suspended`; default: `active`)
- `created_at` (timestamp with time zone)
- `updated_at` (timestamp with time zone)

Notes:

- `owner_user_id` is the initial owner when the organization is created.
- Ownership is also reflected via `org_memberships` with role `OWNER`.
- The `status` field can later be used for billing and account control.

---

### 3.3 Organization memberships

Many-to-many relation between users and organizations with an explicit role.

**Table: `org_memberships`**

Fields:

- `id` (UUID, PK)
- `user_id` (UUID, FK → `users.id`)
- `org_id` (UUID, FK → `organizations.id`)
- `role` (enum: `OWNER`, `ADMIN`, `MEMBER`, `VIEWER`)
- `is_active` (bool, default: true)
- `created_at` (timestamp with time zone)

Constraints:

- Unique `(user_id, org_id)` pair (a user can have only one membership per org).
- At least one `OWNER` per organization is required (enforced at business level).

Semantics:

- `OWNER` – full control, can manage org settings and members.
- `ADMIN` – manage members and configuration, but not billing/legal (future).
- `MEMBER` – regular user with standard access.
- `VIEWER` – read-only access (analytics, reports, dashboards).

---

### 3.4 Refresh tokens

Refresh tokens allow long-lived sessions without keeping access tokens alive for too long.

**Table: `refresh_tokens`**

Fields:

- `id` (UUID, PK)
- `user_id` (UUID, FK → `users.id`)
- `org_id` (UUID, FK → `organizations.id`, represents active org for this session)
- `token_hash` (string, hash of the refresh token)
- `user_agent` (string, optional – for UX and security logs)
- `ip_address` (string, optional)
- `created_at` (timestamp with time zone)
- `expires_at` (timestamp with time zone)
- `revoked_at` (timestamp with time zone, nullable)

Important rules:

- The raw refresh token **is never stored**. Only `token_hash` is persisted.
- On refresh, we:
    - hash the provided token,
    - look up by `token_hash`,
    - check `expires_at` and `revoked_at`.
- Revoking all sessions for a user = setting `revoked_at` for all refresh tokens of that user.

---

### 3.5 Future entities (not implemented yet)

Planned but not required for initial MVP:

- `api_keys` – organization-level access keys for external integrations.
- `audit_log` – security/audit events (may live in a separate `audit` service).

---

## 4. Token model

`zeno-auth` uses a **dual-token** approach:

- Short-lived **access tokens** (JWT).
- Long-lived **refresh tokens** (opaque strings).

### 4.1 Access tokens

Format: **JWT (JSON Web Token)**

- Signed with an asymmetric key pair (e.g. RSA / ES256).
- Private key is stored and used **only** by `zeno-auth`.
- Public key is exposed to other services (e.g. via `/jwks` endpoint).

Typical lifetime: `15–30 minutes`.

**Standard claims:**

- `iss` – issuer, e.g. `zeno-auth`
- `sub` – user id (UUID from `users.id`)
- `aud` – intended audience (`zeno-docs`, `zeno-id-docs`, or generic `zenon-cloud`)
- `iat` – issued at
- `exp` – expiration time
- `jti` – unique token id

**Custom claims:**

- `org` – active organization id (UUID from `organizations.id`)
- `roles` – list of roles for this user in this organization (e.g. `["OWNER"]`)

Other services (`zeno-docs`, `zeno-id-docs`) use the token as follows:

1. Validate signature using the public key.
2. Verify `exp` and optionally `aud`.
3. Read `sub` (user id), `org` (organization), `roles` (authorization).
4. Decide whether the requested operation is allowed.

`zeno-auth` is **not called** on every request – JWT validation is local to each service.

---

### 4.2 Refresh tokens

Refresh tokens:

- are opaque, random strings (e.g. 256-bit random).
- are stored only on the client side (e.g. HTTP-only cookie or secure storage).
- are hashed and stored in the `refresh_tokens` table.

Typical lifetime: `7–30 days` (configurable per environment).

Flow:

1. Client sends `refresh_token` to `/auth/refresh`.
2. `zeno-auth` hashes it and looks up `refresh_tokens.token_hash`.
3. If found and:
    - not expired,
    - not revoked,
    - user and membership are still active,
      then:
    - issues a new access token,
    - optionally rotates refresh token (issues a new one and revokes the old).

This allows:

- logout from a single device (delete one refresh token),
- logout from everywhere (delete all refresh tokens for the user).

---

## 5. Core flows

### 5.1 Registration flow

1. User submits email + password (+ optional organization name) to `/auth/register`.
2. `zeno-auth`:
    - creates `users` row,
    - creates `organizations` row (if this is the first org),
    - creates `org_memberships` with role `OWNER`,
    - creates a refresh token entry,
    - issues an access token + refresh token.
3. Client stores the refresh token securely and uses the access token for API calls.

### 5.2 Login flow

1. User submits `email + password` to `/auth/login`.
2. `zeno-auth`:
    - finds the user by email,
    - verifies `password_hash`,
    - selects default / active organization (if user has multiple orgs),
    - creates a new refresh token,
    - issues access token + refresh token pair.
3. Client continues as in the registration flow.

### 5.3 Accessing other services (`zeno-docs`, `zeno-id-docs`)

1. Frontend includes header:  
   `Authorization: Bearer <access_token>`
2. The target service:
    - validates the JWT,
    - extracts `sub`, `org`, `roles`,
    - applies its own authorization rules.
3. No call to `zeno-auth` is needed in the happy path.

---

## 6. Integration with other services

### 6.1 Zeno Docs

`zeno-docs` consumes:

- user id (`sub`)
- organization id (`org`)
- roles (`roles`) to decide:
    - who can upload documents,
    - who can manage integrations and mappings.

`zeno-docs` never stores passwords or handles login. It only trusts JWTs issued by `zeno-auth`.

### 6.2 Zeno ID Docs

`zeno-id-docs` consumes:

- `sub`, `org`, `roles` from the JWT,
- uses them to control who can:
    - upload identity documents,
    - view verification results.

It never stores credential data – only document verification state and minimal identity attributes needed for business logic.

---

## 7. Security considerations

- Passwords are stored only as strong hashes (no plaintext, no reversible encryption).
- Refresh tokens are stored only as hashes.
- Access tokens are short-lived and signed with an asymmetric key.
- Services never accept tokens that fail signature or expiry validation.
- Sensitive operations (e.g. managing organizations, members, security settings) require elevated roles (`OWNER` or `ADMIN`).

---

## 8. Future extensions

Planned future work for `zeno-auth`:

- Organization-level API keys with scoped permissions.
- Email verification and password reset flows.
- Audit trail of security-sensitive events (possibly via a dedicated `audit` service).
- Organization invitations (inviting users by email to join an existing organization).
