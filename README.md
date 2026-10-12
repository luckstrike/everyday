# Apartment display

A personal wall dashboard built with Svelte 5, TypeScript, Vite, and a small Go REST service. The ambient evergreen UI follows the approved Paper portrait, landscape, board picker, and touch editor prototypes. All information except the browser’s local clock is **mock data**.

## Run on macOS

Requires Node.js 22.18+ (24 LTS recommended), npm, and Go 1.22+.

```sh
make install
make dev
```

Open **http://127.0.0.1:5173**. The API binds to **127.0.0.1:8080**. Both processes stay on loopback; Ctrl-C stops both. Ports are fixed so conflicts fail visibly. `GO=/absolute/path/to/go make dev` selects another Go installation. This development session uses `/tmp/apt-display-toolchain/go/bin/go`; install Go normally for a lasting setup.

Choose **Everyday → Fullscreen**, or use the browser’s fullscreen command. For a wall session, prevent macOS display sleep through system settings or `caffeinate -d` in another terminal.

## Your boards

The initial **Everyday** board includes time/date, weather, the next calendar event, upcoming bills, Around Dallas discoveries, and Dallas basketball. Select the board name to switch boards or create a named empty board. Board selection is always manual. The app never rotates boards or scrolls them automatically.

Every board fits one screen. Portrait and landscape have **independent saved arrangements**. Editing one orientation preserves the other. A new board starts empty in both orientations. Spare space stays empty, and existing widgets are never automatically moved or stretched.

Weather’s **Days / Hours** button changes its forecast in place. The arrow on Around Dallas manually advances the sample recommendation. Suggestions, dates, venues, and prices are illustrative, not verified listings. Bills show outstanding payments, with more detail at larger sizes; there is no budget or spending summary.

Use the settings icon beside **Dallas, TX** to choose **Evergreen** or **Turquoise**. The choice applies immediately and persists in this browser. Layout and widget notices use a shared toast that dismisses after 4.5 seconds, pausing while hovered or focused.

## Touch editing

**Add widget** opens a draft editor and the catalog. **Edit layout** beside it opens the editor without the catalog.

1. Select a widget, or choose **Edit** on a widget already on the board. Nothing is reserved or highlighted before selection.
2. Choose Small, Medium, or Large. Existing custom dimensions have a **Current** option, so opening the editor never changes their size. Each preset reports whether it fits.
3. **Choose position** clears the picker away. Tap a highlighted position to place it; **Widget options** reopens the size controls. Occupied space is protected.
4. Tap a widget’s upper-right handle to move it, choose another size, or remove it. You can also drag its handle, or focus it and use arrow keys for precise moves.
5. **Save layout** persists this orientation in this browser. **Cancel** restores the saved layout.

When a size cannot fit, choose a smaller size or go back to edit another widget and make room. The editor never overlaps tiles or extends the board to force placement. Catalogs and menus may scroll independently. Touch controls are at least 44px. Escape closes placement/picker/menu controls. Cancel confirms before discarding changed layouts; browser navigation warns only for actual unsaved changes.

Saved boards use `apartment-display.boards.v2` in localStorage. Old v1 layouts migrate into named boards; custom landscape geometry is preserved. Exact untouched Home defaults adopt the new Everyday composition. Original v1 layout and pin keys are retained for recovery, but pins no longer change presentation. Invalid saved data falls back to Everyday with a notice and does not erase the original value. A failed save keeps the draft open.

## Attention and simulations

Priority calculations remain pure TypeScript with an explicit time argument. The strongest event for each widget wins; modifiers do not accumulate. Bill reminders use local calendar days.

| Source | Modifiers |
| --- | --- |
| Weather | Rain +15, severe +60 |
| Sports | Today +10, starting +20, live +40, final +20 |
| Bills | Within a week +5, within 3 days +15, tomorrow +25, today +35, overdue/alert +40 |
| Calendar | Today +10, upcoming +25, soon +50 |

All attention, including severe weather, outlines a visible widget without changing its geometry or selecting a board. The border pulses three times and then stays accented until attention ends; reduced motion uses a steady border. Compact weather prioritizes alert text over the forecast. An absent Weather widget is never inserted automatically. Final-score priority decays until expiration. The presentation engine retains tested fullscreen timing (10–25 seconds, 65-second cooldown) behind an explicit policy opt-in, disabled for this dashboard. Timing and layout projections are independent of Svelte and never modify User Layout.

**Board name → Simulate** opens the development scenario panel. Try rain, severe weather, live basketball, a starting game, a final score, bill reminders, or an imminent calendar event. **Normal day** clears the simulation. Mock scenarios expire after two minutes. Simulation requires both Vite development mode and `DASHBOARD_DEV=1`; production builds omit the panel and the production API omits the endpoint.

## Architecture

```text
Go mock services → normalized REST state
    → frontend polling → bill events + widget priority engine
    → manual board selection + temporary presentation manager
    → bounded layout editor → size-aware Svelte widgets
```

Widgets receive props and never fetch APIs. Go owns mock source data and scenario lifetimes. The frontend owns local board storage, orientation, drafts, and manual selection. A fixed 24 by 24 grid provides bounded placement and collision validation. The old screen-rotation scheduler and pin projections have been removed.

| File or directory | Responsibility |
| --- | --- |
| `backend/internal/domain` | Normalized data and widget definitions |
| `backend/internal/mock` | Independent mock sources and scenarios |
| `backend/internal/api` | REST and development gating |
| `frontend/src/App.svelte` | Polling lifecycle, board state, controls |
| `frontend/src/lib/boards.ts` | Geometry, presets, validation, migration |
| `frontend/src/lib/priority.ts` | Deterministic widget ranking, decay, expiration |
| `frontend/src/lib/presentation.ts` | Bounded fullscreen timing, cooldown, restoration |
| `frontend/src/components` | Touch grid editor, picker, widget views |

| Method | Path | Result |
| --- | --- | --- |
| GET | `/api/state` | Normalized mock data and attention events |
| GET | `/api/widgets` | Widget definitions and minimum/default geometry metadata |
| GET | `/api/config` | Polling interval and development flag |
| POST | `/api/scenario` | Development only, e.g. `{"scenario":"severe"}` |

Polling uses a five-second interval with an eight-second timeout and never overlaps requests. Timers clean up on unmount. The last successful state remains visible during outages; the footer reports delay or connection failure. Attention still expires locally. Mock dates are anchored to backend startup rather than advancing on each poll.

Svelte is the only frontend runtime dependency. Go uses its standard library. No layout or drag libraries, network fonts, databases, or cloud services were added.

## Verification

```sh
make check   # Go race tests + vet, Svelte/TS checks, Node behavioral tests
make build   # Go binary in /tmp, frontend build in frontend/dist
```

Format with `cd frontend && npm run format`.

Behavioral tests cover scores, date boundaries, expiration/decay, fullscreen bounds/cooldown/restoration, collisions, full-board placement, persistence validation, and custom layout migration. Browser verification covers portrait/landscape bounds, touch placement, size feedback, Save/Cancel, reload persistence, board creation/switching, keyboard moves, interrupted drags, and storage failures. The intended LCD still needs a physical legibility and long-running check.

For a local production preview, run the backend without `DASHBOARD_DEV`, build the frontend, and run `npm run preview` in `frontend/` at http://127.0.0.1:4173. Raspberry Pi deployment and live external sources remain future work.
