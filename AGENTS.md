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
fixed layout
      ↓
widgets / UI
```

A priority engine and screen manager sit between application state and the layout in the long-term design. They are deferred, not part of the MVP. See "Deferred (post-MVP)" below.

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

The MVP has one fixed landscape layout, defined in code.

Widgets are placed on a fixed 24 by 24 grid. Keep that placement format so a future admin panel can edit the same data.

The MVP has no layout editing, no multiple boards, and no portrait layout.

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

- fixed layouts staying inside the grid without overlapping
- month calendar date placement
- date and time boundaries
- mock data staying stable while polling

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

## Settings panel

A settings panel is in scope, by the owner's instruction. It edits data, not layout:

- Bills first, then calendars and location, alongside the existing theme choice.
- The header keeps one settings gear. It opens Settings as a full screen with a left rail of sections and a Done button, and hands the screen back after a quiet spell. Do not add tabs or a Home button to the dashboard header.
- It is built for touch on the wall display: large targets, steppers instead of typing where possible.
- It is reachable by address (`#/settings/...`), so it can be opened from another device later.

For now it is localhost-only and has no login. The endpoints that change things refuse requests that are not addressed to a loopback name, are not JSON, or name another origin. A login (Supabase has been mentioned) is deferred and needs explicit instruction; when it arrives it replaces those checks. Do not expose the server beyond localhost before then.

## Deferred (post-MVP)

These were built in the prototype and removed for the MVP. The git tag `prototype-v0` preserves them; restore code from the tag rather than rewriting it.

Do not reintroduce these without explicit instruction.

- Widget and layout editing: moving or resizing widgets. The settings panel (see above) does not do this.
- Multiple named boards, and portrait layouts.
- Sports and Around Dallas widgets. Planned for V1+ or V2.
- Priority engine and attention system. When it returns, keep it deterministic and testable, outside Svelte components, and support:
  - base priority
  - event modifiers
  - cooldown
  - minimum display duration
  - maximum display duration
  - priority decay
- Dynamic presentation states: NORMAL, EMPHASIZED, FULLSCREEN. They must never permanently modify the user's layout. Keep User Layout and Dynamic Presentation State separate.
- Screen manager.
