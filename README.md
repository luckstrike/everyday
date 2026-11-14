# Everyday

A personal wall dashboard built with Svelte 5, TypeScript, Vite, and a small Go REST service. The ambient evergreen UI follows the approved Paper landscape prototype. Weather is **live** (Open-Meteo, no API key). Calendar reads from Google Calendar once you [connect it](#connect-google-calendar). Bills come from a small [bills file](#bills) you edit yourself. Set `EVERYDAY_MOCK=1` to fill the display with made-up data instead.

## Run on macOS

Requires Node.js 22.18+ (24 LTS recommended), npm, and Go 1.22+.

```sh
make install
make dev
```

Open **http://127.0.0.1:5173**. The API binds to **127.0.0.1:8080**. Both processes stay on loopback; Ctrl-C stops both. Ports are fixed so conflicts fail visibly. `GO=/absolute/path/to/go make dev` selects another Go installation.

Use the browser's fullscreen command to fill the display. For a wall session, prevent macOS display sleep through system settings or `caffeinate -d` in another terminal.

## Configure

Settings come from three places; later ones win: built-in defaults, then the config file, then environment variables. The config file holds no secrets. It is `everyday/config.json` in your user config directory (on macOS, `~/Library/Application Support/everyday/config.json`) unless `EVERYDAY_CONFIG` names another path. Every field is optional, and a misspelled field is an error:

```json
{
  "location": { "name": "Austin, TX", "latitude": 30.2672, "longitude": -97.7431 },
  "google": {
    "exclude": ["Birthdays"],
    "calendars": ["primary", "family@example.com"]
  }
}
```

| File field | Meaning |
| --- | --- |
| `location` | Weather location: a display name and coordinates, all three required. The default is Dallas, TX. |
| `google.exclude` | Calendars to leave out, by name or ID, ignoring case. Use it to drop noisy ones such as Birthdays. |
| `google.calendars` | Show only these calendar IDs. Normally leave this out; see [which calendars show](#which-calendars-show). |

| Environment variable | Meaning |
| --- | --- |
| `EVERYDAY_CONFIG` | Path of the config file. A missing file here is an error. |
| `EVERYDAY_LOCATION`, `EVERYDAY_LAT`, `EVERYDAY_LON` | Weather location, overriding the file. Set all three or none. |
| `EVERYDAY_GOOGLE_CALENDARS` | Comma-separated calendar IDs, overriding the file's `google.calendars`. |
| `EVERYDAY_GOOGLE_CREDENTIALS` | Path of the Google credentials file. Defaults to `everyday/google.json` in your user config directory. |
| `EVERYDAY_BILLS` | Path of the bills file. Defaults to `everyday/bills.json` in your user config directory. |
| `EVERYDAY_MOCK` | `1` or `true` serves mock data for every widget instead of live sources. Off by default. |

Invalid settings stop the server at startup with a message naming the problem.

## Connect Google Calendar

The backend reads your calendar through Google's API, and the browser never sees any credentials. Setup is a one-time job in Google's console plus one sign-in:

1. In the [Google Cloud console](https://console.cloud.google.com/), create a project and enable the **Google Calendar API**.
2. Configure the OAuth consent screen (Google Auth Platform): user type **External**, any app name, your own email. Add the scopes `https://www.googleapis.com/auth/calendar.events.readonly` (read events) and `https://www.googleapis.com/auth/calendar.calendarlist.readonly` (list which calendars you have). Neither can change anything.
3. **Publish the app** (status "In production"). In "Testing" status Google expires the sign-in after 7 days. An unverified personal app is fine; at sign-in Google warns that it hasn't verified the app, so choose Advanced and continue.
4. Create credentials: **OAuth client ID**, application type **Desktop app**. Google may show the client secret only once, so save the ID and secret in a JSON file outside this repository: `{"client_id": "...", "client_secret": "..."}`. Google's downloaded `client_secret_*.json` also works, if the client is a Desktop app (its JSON starts with `"installed"`; a Web application client starts with `"web"` and will not work).
5. Sign in. A browser opens; approve read-only access to your events:

   ```sh
   make google-auth ARGS="-client-file /path/to/client.json"
   ```

6. Restart `make dev`. The Up next widget now shows your events.

Sign-in saves the client and a refresh token to `everyday/google.json` in your user config directory, readable only by you. Move that file to the Raspberry Pi to deploy there. If Google ever revokes access, the server logs it and the widget says unavailable. Sign in again with a plain `make google-auth`, with no arguments, which reuses the saved client.

### Which calendars show

By default the dashboard mirrors Google Calendar: it shows every calendar ticked in Google Calendar's sidebar, including shared calendars and subscribed ones such as your country's holidays. To change what appears, tick or untick a calendar in Google Calendar (or the Google Calendar app); the dashboard follows within about five minutes. The server logs the calendars it is showing whenever that set changes.

- To drop a calendar that is ticked but unwanted, add its name or ID to `google.exclude` in the config file.
- To pin an exact list that ignores what is ticked, set `google.calendars` (or `EVERYDAY_GOOGLE_CALENDARS`) to calendar IDs, found under each calendar's settings as **Calendar ID**.
- If one calendar cannot be read, it is skipped with a warning in the log and the others still show. Listing calendars needs the second permission above; if you signed in before it existed, run `make google-auth` again.

## Bills

Bills are typed in, not fetched: there is no bank or budgeting-app connection, and the display shows only what is due, never spending. They live in `everyday/bills.json` in your user config directory (on macOS, `~/Library/Application Support/everyday/bills.json`), or wherever `EVERYDAY_BILLS` points:

```json
{
  "bills": [
    { "name": "Rent", "amount": 1850, "day": 1 },
    { "name": "Internet", "amount": 65, "day": 15 }
  ]
}
```

- Every bill repeats monthly, on `day` (1 to 31). Today's date counts as still due. A bill due on the 29th, 30th or 31st falls on the last day of a shorter month.
- `amount` is the usual amount, shown with cents. A bill whose amount changes each month shows its usual figure.
- A bill can be ticked as paid, from the circle beside it in [Settings](#settings). That saves `paidFor`, the due date it was paid for, and the bill counts as paid only while that is the due date it is showing, so the tick clears itself when the next month's bill comes round. Paying ahead works: tick it any time before the due date.
- On the dashboard, bills still to pay come first, soonest due, followed by the ones already paid, dimmed with a check and the word "Paid". When space is short, what is left to pay gets it.
- Files from before this change may carry an `autopay` field per bill. It is read and ignored, and the next save from Settings leaves it out.
- The file is re-read whenever it changes, so edits appear within a few seconds with no restart. Leaving the file out shows "No bills yet". A mistake in it (a misspelled field, a day outside 1 to 31, a negative amount) shows "Bills unavailable" on the display and logs the exact problem once.
- The widget shows as many whole cards as fit.
- You can edit the file by hand or from [Settings](#settings); both change the same file.

## Settings

The gear beside the location opens Settings as a full screen over the dashboard, which keeps updating underneath. It has two sections for now:

- **Bills**: add, edit and delete bills, listed by the day of the month they fall due, with bills that share a day grouped together. Each bill has a name, an amount, a due day (a stepper, so nothing needs typing), and a circle to tick when it is paid. A change is saved as one whole file, written atomically, and shows on the dashboard at once. If the bills file has a mistake in it, Settings shows what it is and lets you replace the file.
- **Appearance**: the theme.

Done or Escape closes it, and it closes itself after 90 seconds without a touch, so it never stays open on the wall. It has an address (`http://127.0.0.1:5173/#/settings/bills`), so a section can be opened directly.

There is no login yet, so Settings only works from the machine the dashboard runs on. The endpoints that change things (`PUT /api/bills`) reject a request unless it is addressed to `localhost`, `127.0.0.1` or `::1`, is JSON, and either names no origin or names this server. That keeps another web page in your browser from editing your bills, and it keeps the dashboard from being edited from elsewhere on your network by accident. If you put a reverse proxy in front of the server, it must pass the original `Host` header through. A real login is planned and will replace these checks.

## What it shows

The MVP is a landscape-only display with one fixed layout and five widgets:

| Widget | Shows |
| --- | --- |
| Clock | Time, weekday, and date |
| Month calendar | The current month, or only the current week when the widget is short |
| Weather | Current conditions with a five-day forecast (today plus four days), low → high temperature ranges, an hourly view via the **Days / Hours** button, and any alert |
| Up next | Your next calendar event |
| Bills | Your recurring bills as cards, what is left to pay first and what is already paid after, each with its amount and due date |

Each widget adapts to the space it receives. Smaller sizes show less detail instead of scaling content down. Tap the settings gear beside **Dallas, TX** to open [Settings](#settings), where you can edit your bills and choose **Evergreen** or **Turquoise**. The theme applies immediately and persists in this browser. Notices use a shared toast that dismisses after 4.5 seconds.

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
| `backend/internal/bills` | The bills file: reading it, saving it, and when each bill is next due |
| `backend/internal/fileutil` | Writing a file so a half-written one is never seen |
| `backend/internal/gcal` | Google OAuth sign-in, the Calendar client, and the event cache |
| `backend/cmd/googleauth` | The one-time `make google-auth` sign-in command |
| `backend/internal/mock` | Independent mock sources, used only with `EVERYDAY_MOCK` |
| `backend/internal/config` | Environment settings |
| `backend/internal/api` | REST handlers |
| `frontend/src/App.svelte` | Polling lifecycle, theme, header and footer |
| `frontend/src/lib/layout.ts` | The fixed layout and grid helpers |
| `frontend/src/lib/month.ts` | Month grid and current-week calculation |
| `frontend/src/components` | The grid, widgets, the settings panel, toast |

| Method | Path | Result |
| --- | --- | --- |
| GET | `/api/state` | Normalized state, with a `sources` status (`live`, `mock`, `unavailable`, `unconfigured`) for each section |
| GET | `/api/config` | Polling interval |
| GET | `/api/bills` | The bills as written in the file, in file order, plus a `problem` if the file cannot be read |
| PUT | `/api/bills` | Replaces the bills, with the new list returned. Invalid bills get a 400 that says why. |

Polling uses a five-second interval with an eight-second timeout and never overlaps requests. Timers clean up on unmount. The last successful state remains visible during outages; the footer reports delay or connection failure. The backend contacts Open-Meteo every 15 minutes (retrying after a minute on failure), not on each poll. A failed refresh keeps the last forecast, which is dropped after an hour so stale weather is never shown as current. Google Calendar is read every 5 minutes (retrying after a minute on failure), and its events are dropped after 30 minutes without a successful refresh. In mock mode, dates are anchored to backend startup rather than advancing on each poll.

Svelte is the only frontend runtime dependency. Go uses its standard library. No layout libraries, network fonts, databases, or cloud services are used.

## Verification

```sh
make check   # Go race tests + vet, Svelte/TS checks, Node behavioral tests
make build   # Go binary in /tmp, frontend build in frontend/dist
```

Format with `cd frontend && npm run format`.

Tests cover that the fixed layout stays inside the grid without overlapping, month dates and the current-week row, that mock dates stay stable while polling, Open-Meteo parsing, the bills file and due-date arithmetic (month ends, leap years, year rollover), that the weather and calendar caches keep the last good data through failures, the Google sign-in flow (PKCE, state checks, token refresh), and calendar event handling (all-day events lead their day, cancelled and declined events, merged calendars, mirroring and excluding calendars, one broken calendar being skipped), and the config file. The intended LCD still needs a physical legibility and long-running check.

For a local production preview, run the backend, build the frontend, and run `npm run preview` in `frontend/` at http://127.0.0.1:4173.

## After the MVP

Removed from the prototype and preserved in the git tag `prototype-v0`:

- Widget and layout editing (moving and resizing widgets)
- Multiple named boards and portrait layouts
- Sports and Around Dallas widgets (V1+ or V2)
- The priority and attention system

Editing calendars and location in Settings, a login so Settings can be used from other devices, iCloud calendars (Google only for now), and Raspberry Pi deployment also remain future work.
