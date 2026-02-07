# Routing Profiles Strict Checklist

## Completed (with commit refs)

- [x] Add core routing profile engine with virtual provider/model alias resolution.
- [x] Add routing profile CRUD endpoints.
- [x] Add routing profile detail endpoint.
- [x] Add routing profile list sorting and virtual provider filtering.
- [x] Add routing profile export endpoint and UI export panel.
- [x] Add routing profile simulation endpoint and UI simulation panel.
- [x] Add routing profile import endpoint and UI import panel.
- [x] Add in-memory reload after routing profile CRUD/import writes (`e6359983`).
- [x] Add configstore table/migration/CRUD groundwork for routing profiles (`463ae744`).
- [x] Add table backfill from plugin config when table path is available (`7f5674fc`).
- [x] Add routing profile hash generation + persistence writes (`8dbc2374`).
- [x] Harden validation:
  - unique virtual providers and profile names
  - unique profile IDs
  - virtual provider conflicts with real providers
  - virtual model alias rules and strategy validation
- [x] Add capability-aware target matching (`6117e1ad`).
- [x] Add simulation utility in governance package for reusable decision tests (`0dc5fa46`).
- [x] Add decision observability context attributes (`484ce351`).
- [x] Add/expand tests for alias validation and simulation paths (`152354e7`, `f84033b7`, `6a40cf1c`, `572a7b45`).
- [x] Reduce intrusive enterprise fallback messaging in OSS build (`35b7b192`).
- [x] Replace reflection bridge in transport handler with typed routing-profile table integration (`91eef37a`).
- [x] Add dedicated handler-level tests for routing profile CRUD/import/export/simulate HTTP routes (`72e2f9dd`).
- [x] Move routing profile page from JSON-centric editing to structured form components (`d722ad68`).
- [x] Add explicit rejection analytics emission for routing profile candidate evaluation (`b1e4c7b8`).
- [x] Add parity test coverage between simulation output and pre-hook routing decisions (`792f1641`).

## Remaining

- [x] Optional: add frontend component tests for structured routing profile target-row interactions (`84f6fda0`).
- [x] Optional: emit dedicated telemetry counters into the metrics pipeline (`cfcbc755`).
