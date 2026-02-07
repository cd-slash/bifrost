# Routing Profiles API (WIP)

Routing profiles let you map virtual provider/model aliases (for example `light/light`) to concrete provider/model targets with failover.

## Endpoints

- `GET /api/governance/routing-profiles`
  - Optional query param: `virtual_provider`
  - Returns `{ profiles, count }`

- `GET /api/governance/routing-profiles/{profile_id}`
  - Returns `{ profile }`

- `POST /api/governance/routing-profiles`
  - Creates a routing profile
  - Request body is the profile object

- `PUT /api/governance/routing-profiles/{profile_id}`
  - Updates an existing profile

- `DELETE /api/governance/routing-profiles/{profile_id}`
  - Deletes a profile

- `GET /api/governance/routing-profiles/export`
  - Returns plugin config snippet:
    - `{ plugin: { name, enabled, config: { routing_profiles } } }`

- `POST /api/governance/routing-profiles/simulate`
  - Request body example:
    - `{ "model": "light/light", "request_type": "chat", "capabilities": ["vision"] }`
  - Returns resolved candidate list and chosen primary/fallback chain.

- `POST /api/governance/routing-profiles/import`
  - Accepts either:
    - `{ "routing_profiles": [...] }`
    - `{ "plugin": { "config": { "routing_profiles": [...] } } }`
  - Replaces current routing profile set with imported profiles.

## Validation Rules (Current)

- Profile `id`, `name`, and `virtual_provider` are required.
- `virtual_provider` must be unique (case-insensitive).
- `virtual_provider` cannot overlap configured real provider names.
- Profile `name` must be unique (case-insensitive).
- `strategy` must be `ordered_failover` or `weighted`.
- Each profile must contain at least one target.
- Each target must include `provider`.
- Target `capabilities` can be used for capability-aware selection (currently includes `text` and request-detected `vision`).
- If `virtual_model` is set, `model` must be set.
- `virtual_model` aliases must be unique per profile (case-insensitive).
- Mixing wildcard `virtual_model: "*"` with named aliases in one profile is rejected.
