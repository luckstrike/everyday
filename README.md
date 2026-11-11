# Everyday

A personal wall dashboard built with Svelte 5, TypeScript, Vite, and a small Go REST service. The ambient evergreen UI follows the approved Paper landscape prototype. Weather is **live** (Open-Meteo, no API key). Calendar and bills have no source yet, so those widgets show as empty; set `EVERYDAY_MOCK=1` to fill the display with made-up data instead.

## Run on macOS

Requires Node.js 22.18+ (24 LTS recommended), npm, and Go 1.22+.

```sh
make install
make dev
```

Open **http://127.0.0.1:5173**. The API binds to **127.0.0.1:8080**. Both processes stay on loopback; Ctrl-C stops both. Ports are fixed so conflicts fail visibly. `GO=/absolute/path/to/go make dev` selects another Go installation.

Use the browser's fullscreen command to fill the display. For a wall session, prevent macOS display sleep through system settings or `caffeinate -d` in another terminal.

## Configure

The backend reads environment variables, for example `EVERYDAY_LOCATION="Austin, TX" EVERYDAY_LAT=30.2672 EVERYDAY_LON=-97.7431 make dev`.

| Variable | Meaning |
| --- | --- |
| `EVERYDAY_LOCATION`, `EVERYDAY_LAT`, `EVERYDAY_LON` | Weather location: a display name and coordinates. Set all three or none; the default is Dallas, TX. |
| `EVERYDAY_MOCK` | `1` or `true` serves mock data for every widget instead of live sources. Off by default. |

Invalid settings stop the server at startup with a message.

## What it shows

The MVP is a landscape-only display with one fixed layout and five widgets:

| Widget | Shows |
| --- | --- |
| Clock | Time, weekday, and date |
| Month calendar | The current month, or only the current week when the widget is short |
| Weather | Current conditions with a daily or hourly forecast (the **Days / Hours** button switches it in place) and any alert |
| Up next | Your next calendar event |
| Bills | The next bill, with a due date and autopay status |

Each widget adapts to the space it receives. Smaller sizes show less detail instead of scaling content down. Use the settings icon beside **Dallas, TX** to choose **Evergreen** or **Turquoise**. The choice applies immediately and persists in this browser. Notices use a shared toast that dismisses after 4.5 seconds.

## Architecture

```text
Go sources (live weather, optional mock) → normalized REST state
    → frontend polling → fixed layout → size-aware Svelte widgets
```

Widgets receive props and never fetch APIs. Go owns the mock source data. The frontend owns the fixed layout, which is a list of placements on a 24 by 24 grid in `frontend/src/lib/layout.ts`.

| File or directory | Responsibility |
| --- | --- |
| `backend/internal/domain` | Normalized data models |
| `backend/internal/weather` | Open-Meteo client and the cache that keeps the last good forecast |
| `backend/internal/mock` | Independent mock sources, used only with `EVERYDAY_MOCK` |
| `backend/internal/config` | Environment settings |
| `backend/internal/api` | REST handlers |
| `frontend/src/App.svelte` | Polling lifecycle, theme, header and footer |
| `frontend/src/lib/layout.ts` | The fixed layout and grid helpers |
| `frontend/src/lib/month.ts` | Month grid and current-week calculation |
| `frontend/src/components` | The grid, widgets, theme settings, toast |

| Method | Path | Result |
| --- | --- | --- |
| GET | `/api/state` | Normalized state, with a `sources` status (`live`, `mock`, `unavailable`, `unconfigured`) for each section |
| GET | `/api/config` | Polling interval |

Polling uses a five-second interval with an eight-second timeout and never overlaps requests. Timers clean up on unmount. The last successful state remains visible during outages; the footer reports delay or connection failure. The backend contacts Open-Meteo every 15 minutes (retrying after a minute on failure), not on each poll. A failed refresh keeps the last forecast, which is dropped after an hour so stale weather is never shown as current. In mock mode, dates are anchored to backend startup rather than advancing on each poll.

Svelte is the only frontend runtime dependency. Go uses its standard library. No layout libraries, network fonts, databases, or cloud services are used.

## Verification

```sh
make check   # Go race tests + vet, Svelte/TS checks, Node behavioral tests
make build   # Go binary in /tmp, frontend build in frontend/dist
```

Format with `cd frontend && npm run format`.

Tests cover that the fixed layout stays inside the grid without overlapping, month dates and the current-week row, that mock dates stay stable while polling, Open-Meteo parsing, and that the weather cache keeps the last good forecast through failures. The intended LCD still needs a physical legibility and long-running check.

For a local production preview, run the backend, build the frontend, and run `npm run preview` in `frontend/` at http://127.0.0.1:4173.

## After the MVP

Removed from the prototype and preserved in the git tag `prototype-v0`:

- Widget and layout editing, planned as an admin panel
- Multiple named boards and portrait layouts
- Sports and Around Dallas widgets (V1+ or V2)
- The priority and attention system

Live calendar and bills data, empty states for widgets without a source, and Raspberry Pi deployment also remain future work.
