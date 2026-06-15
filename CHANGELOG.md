# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project aims to
follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned

- Mobile client (Expo / React Native): decision capture, review, voice notes, push.
- File/attachment evidence in addition to URLs and free text.
- Email-based team invitations (currently members are added by user ID).
- Calibration-over-time view and a calibration training mode.

## [0.1.0] - 2026-06-15

First working end-to-end cut: a GraphQL-only Go backend and a React web client.

### Added

- **Decision lifecycle** - create decisions, record assumptions (with 0-1 confidence),
  attach evidence, make probabilistic predictions, and record outcomes, with an explicit
  status machine (draft → active → decided → archived).
- **Decision-quality analytics** computed from source data: Brier score, a calibration
  curve, decision success rate, and per-team accuracy.
- **Auth** - registration/login, argon2id password hashing, short-lived Ed25519 JWT access
  tokens, and rotating, reuse-detecting refresh tokens delivered as `httpOnly` cookies with
  CSRF double-submit protection.
- **Team-scoped RBAC** (Admin / Member / Viewer) enforced server-side.
- **Append-only event log** acting as both an audit trail and a transactional outbox; GraphQL
  subscriptions over Postgres `LISTEN/NOTIFY`.
- **Optional AI assistance** (disabled by default) to summarize evidence, critique assumptions,
  and flag bias - never on the critical path.
- **Web client** - decisions list, decision detail with the full lifecycle, analytics dashboard
  with charts, and team management; light/dark themes; responsive layout.
- **Infra & CI** - Docker Compose stack (Postgres 18, backend, Caddy), and GitHub Actions for
  lint, test (race + coverage gate), build, and a security pass (`govulncheck`, `gosec`, `trivy`,
  `gitleaks`).

### Known limitations

- Evidence supports URLs and text only (no file uploads).
- Predictions are editable but not deletable.
- Calibration is point-in-time.
- The mobile client is scaffolded but not implemented.
