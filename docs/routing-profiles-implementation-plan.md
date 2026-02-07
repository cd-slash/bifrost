# Routing Profiles Implementation Plan

## Goal

Add first-class routing profiles so requests can target virtual providers (for example `fast/glm-4.7` or `low-cost/glm-4.7`) instead of hard-coding concrete providers in every API request.

## Product Outcomes

1. Users can define virtual providers and map them to one or more concrete provider/model targets.
2. Routing can enforce target-level rate limits and apply ordered/weighted preference with failover.
3. Routing can choose targets by request capability (request type and modality tags).
4. UI supports profile CRUD and route simulation, while preserving existing routing-rules workflows.

## Current State (Validated in Repo)

- Routing rules + CEL builder already exist in OSS code paths.
- Governance plugin already mutates request `model` and `fallbacks` in HTTP transport pre-hook.
- Core request handling already supports ordered fallback chains.
- Governance store already tracks provider/model rate-limit and budget usage percentages.

## Architecture Direction

Add a new routing profile engine in governance, then resolve virtual-provider models in HTTP pre-hook before normal request execution.

Resolution order:

1. Explicit routing rule override (existing behavior)
2. Routing profile virtual-provider resolution
3. Existing default provider/model path

## Data Model Additions

1. `routing_profiles`
   - `id`, `name`, `description`, `virtual_provider`, `strategy`, `enabled`, timestamps
2. `routing_profile_targets`
   - `id`, `profile_id`, `provider`, `model`, `priority`, `weight`, `enabled`
   - optional target constraints: `request_types`, `capabilities`
   - optional target limits: request/token windows
3. `routing_profile_rate_limits` (optional extraction if needed)

## Backend Implementation Phases

### Phase 1 (MVP Vertical Slice)

1. Add governance config types and in-memory store support for routing profiles.
2. Implement virtual-provider resolution engine:
   - parse model `virtual_provider/base_model`
   - select candidate targets by profile strategy
   - filter by enabled + request type/capability
   - apply target-level rate limit checks
   - output primary + fallback chain
3. Integrate in governance HTTP pre-hook (same place routing rules are applied).
4. Add tracing/log context fields for selected profile/target.
5. Add unit tests for profile selection, failover ordering, and throttling behavior.

### Phase 2 (Persistence + APIs)

1. Add configstore tables + migrations.
2. Add ConfigStore CRUD methods.
3. Add governance HTTP handlers for profile CRUD.
4. Add in-memory reload hooks from server manager.
5. Add namespace hardening and uniqueness constraints:
   - prevent virtual provider overlap with configured real provider names
   - enforce unique virtual provider names across profiles (case-insensitive)
   - enforce unique virtual model names within each virtual provider scope (case-insensitive)
   - reserve wildcard tokens (for example `*`) to avoid ambiguous model alias semantics

### Phase 3 (UI)

1. Add Routing Profiles page with table + create/edit sheet.
2. Add target management UX (provider/model, priority/weight, limits, capability tags).
3. Add route simulation panel.
4. Wire RTK query endpoints.

### Phase 4 (Hardening)

1. End-to-end tests for OpenAI-compatible routes and fallback behavior.
2. Metrics + logs for profile hit-rate and target rejection reasons.
3. Backward-compatibility tests with existing routing rules.

## UI Cleanup Track

1. Remove intrusive premium/enterprise marketing copy from OSS fallback pages.
2. Replace upsell-oriented messages with neutral feature-availability text.
3. Ensure unavailable enterprise sections fail quietly without disruptive banners.

## Non-Goals for Initial MVP

1. Latency-adaptive target scoring from historical logs.
2. Complex multi-objective cost+latency optimization.
3. Replacing existing routing rules; profiles are additive.

## Immediate Execution Checklist

1. [x] Add core routing profile structs and strategy enum.
2. [x] Add governance plugin profile engine and integrate in HTTP pre-hook.
3. [x] Add tests for alias resolution (`fast/model`) and failover chain generation.
4. [x] Add API/types scaffolding and routing profile CRUD endpoints.
5. [x] Start UI slice with routing profile page and basic CRUD JSON workflow.

## Current Status

- Phase 1 routing-profile resolution is active in governance HTTP pre-hook.
- Phase 1 CRUD is available via `/api/governance/routing-profiles` endpoints.
- Phase 2 hardening has started with virtual namespace conflict checks and alias constraints.
- Phase 2 table persistence groundwork exists in configstore (table + migration + CRUD methods).

## Phase 1 Config Shape (Current)

```json
{
  "name": "governance",
  "enabled": true,
  "config": {
    "is_vk_mandatory": false,
    "routing_profiles": [
      {
        "name": "Fast Provider Alias",
        "virtual_provider": "fast",
        "enabled": true,
        "strategy": "ordered_failover",
        "targets": [
          { "provider": "cerebras", "priority": 1, "enabled": true },
          { "provider": "openai", "priority": 2, "enabled": true }
        ]
      },
      {
        "name": "Light Model Alias",
        "virtual_provider": "light",
        "enabled": true,
        "strategy": "ordered_failover",
        "targets": [
          { "provider": "cerebras", "virtual_model": "light", "model": "glm-4.7-flash", "priority": 1, "enabled": true },
          { "provider": "anthropic", "virtual_model": "light", "model": "claude-3-5-haiku-latest", "priority": 2, "enabled": true }
        ]
      }
    ]
  }
}
```

Behavior:

- `fast/glm-4.7` routes by virtual provider and preserves model unless target overrides model.
- `light/light` routes by virtual provider (`light`) and virtual model (`light`) to concrete models per target.
