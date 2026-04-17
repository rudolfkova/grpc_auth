# Tick Invariants and Behavioral Contracts

This document defines behavior that must stay stable during refactoring.
If implementation changes, these invariants are the source of truth for tests
and review.

## Tick Loop

- The loop is driven by a fixed ticker configured by `tick_rate`.
- On each tick, the service drains all currently buffered actions from ingress.
- Action order inside one tick equals channel drain order.
- `save_world` actions are handled in the app layer and are not passed into ECS `ProcessTick`.
- All other actions from the drained batch are passed to `ProcessTick` in the same order.

## ECS Processing

- ECS simulation for one tick runs under one engine mutex.
- Movement intent capture and movement application are separated:
  - `move` updates the latest intent per player;
  - physical movement is applied in post-tick phase only.
- Per-action and post-tick system order is fixed by `SystemRegistry`.
- `movement_apply_every_n_ticks` controls only movement application frequency.
  State broadcast cadence still follows `tick_rate`.

## State Payload

- `state` is broadcast every tick.
- `players` and `tick_at` are present on every `state`.
- Tile sync behavior:
  - periodically send full snapshot in `tiles`;
  - otherwise send only `tile_updates`;
  - both can be empty on ticks without tile changes.
- `tile_updates.op` is constrained to `upsert` and `remove`.

## Queue and Delivery Semantics

- Ingress overload rejects new actions (`Submit` returns false).
- Outbound event channel is non-blocking:
  - if channel is full, event is dropped.
- WS per-connection outbound channels are also non-blocking; full channel drops message.

## Character Session Boundaries

- Player join path restores ECS state from opaque `character.data`.
- Persist to character-service happens when the last WS connection of user leaves.
- When character integration is disabled, join/persist paths become no-op.

