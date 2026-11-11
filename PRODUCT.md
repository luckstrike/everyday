# Everyday

## Register

product

## Users

The owner of a personal wall-mounted information display, initially developed on macOS and eventually intended for a small Raspberry Pi LCD. Information should be understandable at a glance from across the room. The owner is an experienced programmer who wants to understand and incrementally develop the system.

## Product Purpose

Keep time, a month calendar, weather, the next event, and upcoming bills visible in one stable, fixed layout. Dallas sports and local event suggestions are planned for V1+ or V2. Support reliable continuous operation with a lightweight Svelte, TypeScript, and Go application.

## Brand Personality

Minimal, ambient, calm. Preserve the green direction the user likes with a quiet presentation based on the approved Paper prototypes.

## Anti-references

Automatically rotating screens, widgets that unexpectedly move, and added product scope beyond the personal display. Avoid budget or spending summaries; the financial view is upcoming bills.

## Design Principles

- Keep information in predictable locations.
- Show one fixed landscape layout. Widgets never move, rotate, or take over the screen.
- Keep the layout within one display viewport, without scrolling.
- Adapt each widget's content meaningfully to its available size.
- Favor glanceable information and restrained presentation.
- Develop incrementally with understandable components and clear boundaries between data and presentation.

## Accessibility & Inclusion

Prioritize distance legibility and clear contrast on a small wall display. The theme settings control is keyboard accessible. No additional accessibility standard has been specified.

## Implementation Scope

The MVP is a landscape-only display with five widgets: Clock, Month calendar, Weather, Up next, and Bills. Weather is live from Open-Meteo and served, with everything else, by the Go REST API. Calendar and bills have no live source yet; mock data is available as an opt-in for development.

Deferred to later versions, and preserved in the git tag `prototype-v0`:

- Widget and layout editing, planned as an admin panel rather than part of the wall display.
- Multiple named boards and portrait layouts.
- Sports and Around Dallas widgets.
- The priority and attention system, including emphasis and fullscreen takeovers.
- Live calendar and bills data, and empty states for widgets without a source.
- Raspberry Pi deployment.
