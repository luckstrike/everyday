# Apartment display

A local wall-display prototype built with Svelte 5, TypeScript, Vite, and a small Go REST service. Weather, sports, bills, calendar entries, and Dallas event suggestions are **mock data**. The clock uses the browser’s local time.

## Run on macOS

Requires Node.js 22.18+ (24 LTS recommended), npm, and Go 1.22+. If Go is missing, use `brew install go` or the official installer.

```sh
make install
make dev
```

Open **http://127.0.0.1:5173**. The API listens on **127.0.0.1:8080**. Both servers bind only to loopback. Ctrl-C stops both. Ports are fixed so a conflict produces an error instead of silently changing the address.

Alternatively, start `make backend` and `make frontend` in separate terminals. `GO=/absolute/path/to/go make dev` selects an alternate toolchain. The current development session uses the verified temporary installation at `/tmp/apt-display-toolchain/go/bin/go`; install Go normally for a lasting setup.

Use **••• → Fullscreen**, or your browser’s fullscreen command. Display controls close when entering fullscreen. For an extended session, prevent macOS display sleep with normal system settings or `caffeinate -d` in another terminal.

## A stable board

The board **does not rotate automatically**. Information stays in predictable positions. Home shows six widgets:

- Time and date.
- Weather, with manual **Hours / Days** controls and a four-day forecast.
- **Around Dallas**, showing illustrative event suggestions. **Next** cycles through samples only when clicked. These are invented suggestions, not real schedules, ticket prices, or personalized recommendations.
- Calendar: your next personal event.
- Bills: the next bill and, at a larger size, additional upcoming bills. No budget or spending summary.
- Sports: a Dallas Mavericks vs. San Antonio Spurs mock.

Manual Weather, Sports, and Finance layouts remain available in the display controls for a closer look. Automatic Smart Stacks are intentionally not implemented: each widget keeps its meaning and location, and optional in-widget controls change only its own content.

**Keep visible** pins individual widgets across these manual layouts. Pins persist separately from saved layouts. A severe-weather fullscreen alert temporarily takes precedence over pins.

The display uses the remaining viewport after the header, alerts, and footer. Controls float over the board. Portrait reflows Home into a clock, full-width weather, and two-column supporting widgets; narrow portrait uses compact single-column widgets. Content changes by available size rather than relying on internal scrolling. Landscape restores the configured grid geometry.

## Edit and save

**Edit layout** shows the canonical 12-column grid, independent of portrait reflow. Drag a top handle to move a widget or its lower-right handle to resize. Both handles also support arrow keys. Widgets cannot overlap and have minimum sizes.

Add/remove widgets using Edit Mode. Each manual layout has its own configuration. **Save layout** stores all layouts locally in this browser; **Cancel** discards the draft. **Restore this screen** resets the current draft only. The editor supports up to 12 rows and may scroll when editing a larger board.

The previous untouched five-widget Home default upgrades to the six-widget default. Customized saved boards remain unchanged; add Around Dallas through Edit Mode if needed. Saved layouts are validated; corrupt data falls back to defaults with a notice. Storage failures leave the draft available.

## Priorities and simulations

Go provides normalized attention events. The TypeScript priority engine remains independent of Svelte and uses an explicit time argument for deterministic calculations. The strongest event wins; modifiers do not accumulate.

| Source | Modifiers |
| --- | --- |
| Weather | Rain +15, severe +60 |
| Sports | Today +10, starting +20, live +40, recent final +20 |
| Bills | Within a week +5, within 3 days +15, tomorrow +25, today +35, overdue/alert +40 |
| Calendar | Today +10, within 2 hours +25, within 30 minutes +50 |

Bill events are derived from due dates using local calendar days. The mock list represents outstanding bills; there is no paid-state model yet. Other sources emit events for selected scenarios.

Ordinary priority events **highlight their widget in place**, with a short attention banner. They do not change the current screen or resize the board. A severe-weather event can temporarily take over the display, then restore the exact configured layout. User Layout and Dynamic Presentation State remain separate. Final-score priority decays linearly until expiration.

**••• → Simulate** opens the development panel. Try rain, severe weather, live basketball, a starting game, a final score, a bill due tomorrow, a financial alert, or a calendar event in 20 minutes. Select **Normal day** to clear the simulated event. Simulations expire after two minutes. The normal mock bill may still have a low-level due-date priority.

The earlier screen scheduling module (`lib/screens.ts`) and its deterministic tests are retained as a reference, but the app no longer calls it. Its cooldowns and display durations have no effect on the stable board.

## Architecture

```text
Go mock source functions
    ↓
Normalized Go domain models + REST endpoints
    ↓
Frontend polling coordinator
    ↓
Bill date events + pure priority calculations
    ↓
Manual layout selection + pure display projection
    ↓
Size-aware Svelte widgets
```

Widgets receive props and never fetch APIs. Source data and scenario lifetimes live in Go. The frontend owns browser-specific layouts, pins, orientation, and manual selection. Priority and geometry rules live in independent TypeScript modules. Selecting Hours, Days, or the next event is presentation state inside that widget.

| Method | Path | Result |
| --- | --- | --- |
| GET | `/api/state` | Normalized mock state, daily/hourly weather, sample local events, attention events |
| GET | `/api/widgets` | Widget definitions and minimum/default sizes |
| GET | `/api/config` | Manual layouts, retained timing metadata, polling interval, development flag |
| POST | `/api/scenario` | Development only, e.g. `{"scenario":"severe"}` |

`make dev` and `make backend` enable `DASHBOARD_DEV=1`. Without it, the simulation endpoint does not exist. The panel also requires Vite development mode and is absent from production builds.

Polling runs every five seconds with an eight-second timeout. Requests do not overlap; timers are cleaned up on unmount. The last successful data remains visible during outages with a stale-data notice. Attention expiration still runs locally. Mock dates are anchored to backend startup so polling does not push events into the future; restarting the backend creates a fresh mock day.

## Structure

```text
backend/
  cmd/server/                HTTP lifecycle and graceful shutdown
  internal/domain/           Normalized models and configuration
  internal/mock/             Separate mock source functions and scenarios
  internal/api/              REST handlers and development simulation
frontend/
  src/App.svelte             Polling, manual layout selection, and controls
  src/lib/                   API client, bills, priorities, layouts, formatting
  src/components/            Grid editor and widget switcher
  src/components/widgets/    Six reusable widgets
  tests/                     Priority, geometry, persistence-validation tests
scripts/dev.sh               Starts/stops both local processes
Makefile                     Development, checks, and builds
```

Svelte is the only frontend runtime dependency. Vite, TypeScript, Svelte checking, and Prettier are development tools. Go uses its standard library. Tests use Go’s testing package and Node’s built-in runner. There are no external APIs, network fonts, databases, or cloud services.

## Checks and preview

```sh
make check   # Go race tests + vet, Svelte/TS checks, Node tests
make build   # Go binary in /tmp and frontend/dist
```

Format with `cd frontend && npm run format`.

For production preview, start the backend without `DASHBOARD_DEV`, build the frontend, then run `npm run preview` in `frontend/`. Open http://127.0.0.1:4173. This is local verification, not a Pi deployment setup.

Behavioral tests cover priority scores, expiration/decay, bill-date boundaries, layout validation, portrait and pin projections, stable board geometry, safe default migration, mock date stability, API validation, and development gating. Historical scheduler tests are also retained. Browser checks exercise portrait/landscape sizing, forecast selection, recommendation paging, and the bills layout. A real extended session on the intended display is still recommended.

## Next milestone

Tune the stable board on the actual portrait LCD before introducing more automatic behavior. A useful next step is a live read-only source for weather or local events, with cached normalized data, source links, filtering preferences, and stale-data handling. Event discovery should distinguish verified listings from personalized ranking. Keep financial integrations and Raspberry Pi deployment out of scope until explicitly requested.
