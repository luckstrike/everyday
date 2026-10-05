# AGENTS.md

## Project goal

This repository contains a personal wall-mounted information dashboard.

It will initially run on macOS for development and eventually run on a Raspberry Pi 4 attached to a small LCD display.

The application should remain lightweight, understandable, and suitable for running continuously as a home appliance.

## Stack

- Svelte
- TypeScript
- Go
- REST between frontend and backend

Do not replace these technologies without explicit instruction.

## Architecture

Maintain clear separation between:

```text
external/mock data
      ↓
backend services
      ↓
normalized application state
      ↓
priority engine
      ↓
screen manager
      ↓
widgets / UI
```

Do not move business logic into UI components unless it is strictly presentation logic.

## Widgets

Widgets should be modular and reusable.

Widgets should be capable of adapting their presentation based on available size.

Prefer meaningful size states over simply scaling content.

For example:

```text
small
medium
large
```

Widgets should not directly fetch external APIs.

## Layout

The user has a persistent configured layout.

Dynamic priority/emphasis behavior must never permanently modify that layout.

Keep these concepts separate:

```text
User Layout
Dynamic Presentation State
```

Dynamic presentation states should initially be limited to:

```text
NORMAL
EMPHASIZED
FULLSCREEN
```

## Priority system

Keep the priority engine deterministic and testable.

Priority calculations should not live inside Svelte components.

Support:

- base priority
- event modifiers
- cooldown
- minimum display duration
- maximum display duration
- priority decay

Avoid adding complicated scheduling logic unless the current requirements need it.

## Dependencies

Do not add dependencies casually.

Before adding a significant third-party package:

1. Explain what problem it solves.
2. Check whether existing project tools can solve the problem simply.
3. Prefer actively maintained, focused packages.
4. Avoid large frameworks for small problems.

## Scope

Do not add features simply because they may be useful eventually.

In particular, do not introduce without explicit instruction:

- authentication
- cloud infrastructure
- Docker orchestration
- databases
- message brokers
- microservices
- plugin frameworks
- external financial integrations
- hardware/GPIO support

The project should grow incrementally.

## Coding style

Prefer:

- simple functions
- small components
- explicit types
- readable naming
- clear interfaces
- straightforward Go
- straightforward Svelte

Avoid:

- premature abstraction
- unnecessary design patterns
- giant generic systems
- clever code that obscures behavior

## Testing

Prioritize tests for behavior rather than superficial implementation details.

Important areas include:

- priority calculations
- screen selection
- cooldowns
- event expiration
- priority decay
- layout restoration

## Agent behavior

The user has professional programming experience and wants to understand the system.

Do not treat this as a request to blindly generate an entire product.

When making important architectural decisions:

- explain the decision briefly
- state relevant tradeoffs
- keep implementation understandable

For major changes, inspect existing architecture before modifying it.

Preserve existing design decisions unless there is a concrete reason to change them.

Do not rewrite large parts of the project merely to match personal stylistic preferences.

## Current development target

The initial target is a local macOS prototype.

Raspberry Pi deployment comes later.

Optimize first for:

1. correctness
2. clarity
3. quick iteration
4. reliable continuous operation
5. eventual Raspberry Pi compatibility
