# Compatibility Rules

This module is a shared client/server contract. The rules below are mandatory
for backward compatibility.

## Wire Evolution (`wire.go`)

- Existing JSON fields must not be removed or renamed in minor updates.
- New fields must be additive and optional (`omitempty`) where possible.
- Existing message type constants must not change semantic meaning.
- New protocol behavior should be introduced via new `Type*` values.

## Schema Versioning

- Incompatible changes in character data must bump `CharacterDataSchemaVersion`.
- Incompatible changes in content JSON must bump:
  - `CatalogSchemaVersion`
  - `ScenarioSchemaVersion`
- Incompatible snapshot format changes must bump `SnapshotSchemaVersion`.

## ECS Contract

- `components.go` types are part of public API.
- Breaking field/tag changes in ECS components require coordinated client and
  server release.

## Client Safety

- Clients should ignore unknown fields in payloads.
- Server should tolerate unknown payload fields when possible.
- Shared payload DTOs (for example `RejectPayload`) must live in `gamekit` so
  server and clients deserialize the same structure.

