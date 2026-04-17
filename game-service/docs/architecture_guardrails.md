# Architecture Guardrails

This document defines mandatory constraints for future changes in `game-service`.

## Dependency Direction

- `internal/domain` must not import `internal/app`, `internal/ports`, `internal/infrastructure`.
- `internal/app` may depend on:
  - `internal/domain/models`
  - `internal/domain/ports`
  - shared contracts from `pkg/gamekit`
- `internal/app` must not import concrete gRPC clients directly.
- `internal/ports/ws` may depend on `internal/app` and shared contracts.
- `internal/infrastructure` implements ports and can depend on third-party SDKs.

## Shared Contract Rules (`pkg/gamekit`)

- `game-service` and clients must use the same DTOs from `pkg/gamekit`.
- Reject/error payloads are shared types from `gamekit`, no local duplicates.
- Contract evolution follows additive-first policy and schema versioning rules in
  `pkg/gamekit/COMPATIBILITY.md`.

## Tick and ECS Rules

- Tick invariants in `docs/tick_invariants.md` are behavioral contract.
- System execution order is explicit in `SystemRegistry` pipeline names.
- Any reordering of systems requires:
  - a dedicated changelog note;
  - updated pipeline tests;
  - explicit regression validation.

## Observability Rules

- Keep labels low-cardinality (`service`, `method`, stable `reason`/`code`).
- Never use `user_id`, email, world name, or free-form text as metric labels.
- Outbound queue drops must be observable (`game_events_dropped_total`).
- gRPC client calls must emit both:
  - request counters by `grpc_code`;
  - duration histograms by `service` and `method`.

