# Docker Runbook

## Quick Commands

- Start/upgrade stack: `make docker-up`
- Stop stack (keep DB/Redis data): `make docker-down`
- Stream logs: `make docker-logs`
- Soft reset (recreate containers, keep volumes): `make docker-reset`
- Hard reset (remove volumes and data, then recreate): `make docker-reset-wipe`

## Data Safety

- `docker-down` and `docker-reset` keep `postgres-data` and `redis-data`.
- `docker-reset-wipe` removes both volumes and all local data in them.

## Typical Flows

### Clean start for daily development

1. `make docker-up`
2. Wait for services to become healthy.
3. Open Gateway at `http://localhost:8080`.

### Restart stack without data loss

1. `make docker-reset`
2. Check logs: `make docker-logs`

### Full clean environment (schema/data reset)

1. Ensure you do not need current local data.
2. `make docker-reset-wipe`
3. Verify migrations finished in logs.

## Migration Commands (local/manual)

- Auth: `make migrate-auth-up DB_DSN_AUTH="postgres://.../auth_db?sslmode=disable"`
- Chat: `make migrate-chat-up DB_DSN_CHAT="postgres://.../chat_db?sslmode=disable"`
- World: `make migrate-world-up DB_DSN_WORLD="postgres://.../world_db?sslmode=disable"`
- Character: `make migrate-character-up DB_DSN_CHARACTER="postgres://.../character_db?sslmode=disable"`

## Troubleshooting

- If startup hangs on dependency checks, inspect health:
  - `docker compose ps`
  - `docker compose logs <service>`
- If network policies on host break Docker bridge traffic, run:
  - `make docker-fix-iptables`
  - then `make docker-up`
