# Apartment display

A local wall-display prototype built with Svelte 5, TypeScript, Vite, and a small Go REST service. All weather, sports, financial, and calendar information is **mock data**. The clock uses the browser’s real local time.

## Run on macOS

Requires **Node.js 22.18+** (24 LTS recommended), npm, and **Go 1.22+**. If Go is missing, install it with `brew install go` or the official Go installer.

```sh
make install
make dev
```

Open **http://127.0.0.1:5173**. The API listens on **127.0.0.1:8080**. Both servers bind only to loopback. Ctrl-C stops both processes. Ports are fixed so a busy port produces an error instead of silently starting elsewhere.

Alternatively, use two terminals:

```sh
make backend
make frontend
```

`GO=/absolute/path/to/go make dev` selects an alternate Go installation. During initial development, a verified official Go toolchain was downloaded to `/tmp/apt-display-toolchain/go/bin/go`; that temporary path is usable until macOS cleans it up. It is not a system installation.

Use the **•••** button to open display controls. Choose **Fullscreen** (or your browser’s fullscreen command), then close the controls. Prevent macOS from sleeping for a wall-display session using its normal settings or `caffeinate -d` in another terminal.

## Using the prototype

- **Home** displays time, weather, the next calendar event, next bill, and next game.
- **Weather, Sports, Finance** show larger, more detailed widgets. Sports uses a Dallas Mavericks vs. San Antonio Spurs mock; Finance shows bills only, without a budget or spending summary.
- **Automatic rotation** starts enabled. Home is the resting screen between other screens. Pause rotation to inspect a screen manually. Bill reminders receive longer visits as their due date approaches.
- **Keep visible** in the display controls pins individual widgets across rotating screens. Pins persist in this browser independently of layouts. Unpinning restores each screen’s configured arrangement. A fullscreen severe-weather alert temporarily takes precedence over pins.
- **Edit layout** pauses rotation and dynamic emphasis. Drag a widget’s top handle to move it; drag the bottom-right handle to resize it. Both handles support arrow keys. Widgets snap to a 12-column grid, cannot overlap, and have minimum sizes.
- Add or remove widgets in Edit Mode. Each screen has its own layout; edits across screens remain in one draft. **Save layout** stores all four screen layouts in this browser’s local storage. **Cancel** discards the draft. **Restore this screen** resets only the current draft screen; Save makes it persistent.
- The editor supports up to 12 rows. Display Mode fits the available viewport; opening controls does not extend the board. Portrait mode reflows the saved reading order into a full-width main widget and two-column supporting widgets. Narrow portrait displays use compact single-column widgets. Turning back to landscape restores the configured geometry. Edit Mode retains the canonical grid and may scroll.
- **Simulate** opens the development scenario panel. Try rain soon, severe weather, live basketball, a starting game, a recent final score, a bill due tomorrow, a financial alert, or an event in 20 minutes. Select Normal day to clear an event immediately. Scenarios expire automatically after two minutes.
- Scenario changes update data immediately. Automatic screen selection respects a 10-second minimum reading time. Paused rotation stays paused until resumed, including for alerts.

Layout storage is versioned and validated. Corrupt or incompatible layouts fall back to defaults with a notice in the controls. If saving is blocked, the editable draft stays available and an error is shown. Mock event dates remain stable between polls, anchored to backend startup; restarting the backend creates a fresh mock day.

## Architecture

```text
Go mock sources (weather / sports / finance / calendar)
    ↓
Go normalized domain state + REST endpoints
    ↓
Frontend polling coordinator (App.svelte + lib/api.ts)
    ↓
Bill date events (lib/bills.ts) + priority calculations (lib/priority.ts)
    ↓
Pure screen manager (lib/screens.ts)
    ↓
Temporary presentation projection (lib/layout.ts)
    ↓
Svelte grid + size-aware widgets
```

The backend owns normalized source data and event lifetimes. The frontend owns the current display session and local user layouts, because those are specific to the browser/display. Widgets receive data as props and never fetch APIs. The top-level Svelte component coordinates polling and rendering; priority, cooldown, screen selection, and layout rules remain in independent TypeScript functions.

**User Layout** and **Dynamic Presentation State** are separate. Presentation is a pure projection of the saved layout:

- `NORMAL`: configured positions and sizes.
- `EMPHASIZED`: an 8-column primary widget with compact supporting widgets.
- `FULLSCREEN`: one widget occupying the whole display grid (not the browser Fullscreen API).

A temporary event can show a widget even if the user removed it. It does not save that widget or change the user’s arrangement. Ending the event immediately restores the configured geometry; the screen manager returns an event-selected screen to Home once its minimum duration has elapsed.

### Priority and screen policy

Base priorities: Home/clock 10, weather 20, sports 20, finance 25, calendar 25. Calendar attention routes to Home, where the calendar widget can be emphasized.

Normalized attention events carry an ID, widget, kind, start, expiry, and optional linear decay. The priority module maps kinds to modifiers:

| Source | Modifiers |
| --- | --- |
| Weather | Rain +15, severe +60 |
| Sports | Today +10, starting +20, live +40, recent final +20 |
| Finance | Within a week +5, within 3 days +15, tomorrow +25, today +35, overdue/alert +40 |
| Calendar | Today +10, within two hours +25, within 30 minutes +50 |

The frontend also derives bill events from due dates using local calendar days, including midnight and DST boundaries. Bills are outstanding items in this prototype; there is no paid-state model yet, so past due dates remain overdue until the source removes them. Other sources still explicitly emit events for selected scenarios. A future source adapter will generate those normalized events. The strongest active event per screen wins; modifiers do not accumulate. Scores decay using the supplied `now` argument, so calculation is deterministic and testable. Final-score events decay over their two-minute lifetime.

- Minimum duration: **10 seconds** on each screen, even when an urgent event arrives.
- Maximum duration: **20 seconds** for Home, **25 seconds** for other normal screens. Finance visits last **45 seconds** when a bill is within 7 days, **60 seconds** within 3 days, **90 seconds** tomorrow, and **120 seconds** today or overdue. Higher-priority events can still interrupt after minimum duration.
- Non-Home cooldown: **65 seconds from exit**.
- Events with a modifier of at least 15 can preempt after minimum duration if their score exceeds the current screen’s score.
- Events with a modifier of at least 50 can bypass cooldown **once per event ID**. Maximum duration still forces a return to Home.
- Without significant events (modifier below 15), eligible screens rotate by least recently shown, then priority and ID. This prevents baseline finance priority and low-level week-ahead reminders from starving weather or sports.
- If no other screen is eligible, Home remains visible.

Timing belongs to screens rather than independent per-widget schedulers. A widget’s priority affects its owning screen and presentation; individual widgets do not rotate on separate timers. This keeps the first milestone deterministic and understandable.

### REST API

| Method | Path | Result |
| --- | --- | --- |
| GET | `/api/state` | Normalized mock state and attention events |
| GET | `/api/widgets` | Widget definitions, priorities, minimum/default sizes |
| GET | `/api/config` | Screen priorities/timing, polling interval, dev flag |
| POST | `/api/scenario` | Development only: `{"scenario":"severe"}` |

`make backend` and `make dev` set `DASHBOARD_DEV=1`. Without this flag, the scenario endpoint does not exist. The panel also requires Vite development mode, so it is excluded from production builds. The frontend polls every five seconds, times requests out after eight seconds, retains the last successful state during outages, and pauses rotation after 15 seconds without a successful response. Expired emphasis still clears locally during an outage. Polls do not overlap; timers are cleaned up when the app unmounts.

## Structure

```text
backend/
  cmd/server/main.go          HTTP lifecycle and graceful shutdown
  internal/domain/models.go  Normalized models and configuration
  internal/mock/             Source functions and scenario fixtures
  internal/api/              REST handlers and development simulation
frontend/
  src/App.svelte             Polling/session coordinator and controls
  src/lib/                   Contracts, API client, bill events, priority, screens, layout
  src/components/            Grid editor and widget switcher
  src/components/widgets/    Five modular, size-aware widgets
  tests/                     Deterministic priority, pin, and portrait tests
scripts/dev.sh               Starts/stops both local processes
Makefile                     Development, checks, and builds
```

Dependencies stay small: Svelte is the only frontend runtime dependency; Vite, TypeScript, Svelte checking, and Prettier are development tools. The backend uses only Go’s standard library. Tests use Go’s testing package and Node’s built-in test runner. No layout framework, icon library, state-management library, network fonts, database, or cloud services are required.

## Validation and production preview

```sh
make check   # Go race tests + vet, Svelte/TS checking, Node behavioral tests
make build   # Go binary in /tmp + frontend/dist
```

Frontend formatting: `cd frontend && npm run format`.

To inspect a production frontend, start the backend without `DASHBOARD_DEV`, build the frontend, and run `npm run preview` in `frontend/`. Open http://127.0.0.1:4173. Vite preview proxies `/api` to the same local backend. Preview is a local verification tool, not the eventual Pi deployment setup.

Tests cover priority scores, exact expiration, decay, minimum and maximum durations, cooldown and urgent bypass, event clearing, all-screen rotation over a simulated hour, presentation restoration, saved-layout validation, mock date stability, endpoint validation, and development gating. Browser checks cover 1280×800 and 1024×600 landscape, 800×1280 and 390×844 portrait, editing/persistence, pin persistence across screens, and alert presentation. A real one-hour fullscreen soak test is still recommended on the intended browser and display.

## Next milestone

Run a real one-hour session and tune typography, timing, and useful size states for your actual LCD dimensions. Then add one real read-only source (weather is a practical first choice), keeping its polling/cache and stale-data handling behind the existing normalized backend models. Keep financial data mocked until its intended scope is explicitly decided. Raspberry Pi deployment and process supervision come after this local behavior is settled.
