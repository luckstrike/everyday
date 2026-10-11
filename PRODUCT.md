# Apartment display

## Register

product

## Users

The owner of a personal wall-mounted information display, initially developed on macOS and eventually intended for a small Raspberry Pi LCD. Information should be understandable at a glance from across the room. The owner is an experienced programmer who wants to understand and incrementally develop the system.

## Product Purpose

Keep time, weather, calendar events, upcoming bills, Dallas sports, and local event suggestions visible in a stable, configured layout. Support reliable continuous operation with a lightweight Svelte, TypeScript, and Go application.

## Brand Personality

Minimal, ambient, calm. Preserve the green direction the user likes with a quiet presentation based on the approved Paper prototypes.

## Anti-references

Automatically rotating screens, widgets that unexpectedly move, and added product scope beyond the personal display. Avoid budget or spending summaries; the financial view is upcoming bills.

## Design Principles

- Keep information in predictable locations.
- Use named boards with manual selection only; start with a single Everyday board.
- Keep every saved board within one display viewport, without scrolling or automatic movement.
- Preserve the user's configured layout independently of temporary attention states.
- Adapt each widget's content meaningfully to its available size.
- Favor glanceable information and restrained presentation.
- Develop incrementally with understandable components and clear boundaries between data, priority logic, and presentation.

## Accessibility & Inclusion

Prioritize distance legibility and clear contrast on a small wall display. The existing prototype includes keyboard controls and reduced-motion handling. No additional accessibility standard has been specified.

## Implementation Scope

Implement the approved Paper prototypes in the existing Svelte application. Keep the Go REST data boundary and pure priority calculations. Each board stores portrait and landscape arrangements independently, so editing and rotating the physical display preserve both compositions.

## Board and Editing Behavior

The board name replaces the house icon and opens a board picker. Additional boards are named, saved layouts, each with its own widget arrangement. Boards change only when the user explicitly selects one; priority events never choose or rotate boards. Ordinary priority events can emphasize a widget in place while preserving the user's layout.

Each board fits one screen. Neither manual scrolling nor automatic scrolling is part of the wall-display experience. Adding or resizing widgets must preserve a layout that fits the screen; a draft that does not fit should be corrected before saving. Smaller widgets show deliberately reduced content. Additional information can go on another manually selected board.

Every widget has a subtle rounded outline. A floating Add widget button opens the widget picker. Editing reveals drag handles and placement previews with Save and Cancel controls. Widget catalogs and editing controls may scroll when necessary, independently of the saved board. Keep the current board's identity visible while editing and protect unsaved drafts when navigating.

## Touch and Placement

Treat touch as the primary editing input. Use effective touch targets of at least 44 by 44 pixels. Selecting a widget reveals valid positions for its chosen Small, Medium, or Large presentation. Tap a highlighted position to place it; dragging is an optional alternative, never a requirement. There is no permanently reserved slot or widget-specific drop target before selection. Placement previews represent the selected widget's actual footprint and disappear when placement is canceled or completed.

Unused space may stay empty; do not stretch widgets or fill gaps automatically. Size changes and widget placement are explicit edits. If the chosen size cannot fit, offer a smaller supported size when possible. If even the smallest presentation cannot fit, explain that the board is full and let the user explicitly resize or remove a widget, choose another board, or cancel. Never overlap widgets, silently rearrange or shrink existing widgets, extend the board, or introduce scrolling to force a new widget onto it.
