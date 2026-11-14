# Configuration and technical reference

Start with the [setup guide](../README.md) for your first run, or the [Raspberry Pi guide](raspberry-pi.md) for a local display. This page keeps the detailed configuration, bill behavior, and architecture in one place.

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

To change weather location without a restart, tap the location name on the dashboard or open **Settings → Location** (`#/settings/location`). Search for a city or postal code, select the result with the right region/country, then save. The backend uses Open-Meteo's geocoding API (location data from GeoNames), which needs no API key for this non-commercial dashboard. The selected name and coordinates are written atomically to the config file, preserving Google settings, and weather reloads immediately. Old-location forecasts are cleared, including late responses from in-flight requests. A lookup failure leaves the saved location unchanged; a failed forecast is retried automatically. This setting changes weather only, not the clock or calendar timezone.

Environment variables retain priority: when the location variables are set, Location explains that they must be removed and the server restarted before editing here. Location editing is disabled in mock mode. Editing the config file by hand still requires restarting the server to load the change.

## Connect Google Calendar

The backend reads your calendar through Google's API, and the browser never sees any credentials. Setup is a one-time job in Google's console plus one sign-in:

1. In the [Google Cloud console](https://console.cloud.google.com/), create a project and enable the **Google Calendar API**.
2. Configure the OAuth consent screen (Google Auth Platform): user type **External**, any app name, your own email. Add the scopes `https://www.googleapis.com/auth/calendar.events.readonly` (read events) and `https://www.googleapis.com/auth/calendar.calendarlist.readonly` (list which calendars you have). Neither can change anything.
3. **Publish the app** (status "In production"). For these Calendar scopes, Google documents a seven-day refresh-token lifetime in "Testing" status; see [Google’s token-expiration guidance](https://developers.google.com/identity/protocols/oauth2#expiration). An unverified personal app may show a warning: continue only if it is the OAuth client you created and you recognize the requested permissions. Publishing does not guarantee that a sign-in can never expire or be revoked.
4. Create credentials: **OAuth client ID**, application type **Desktop app**. Google may show the client secret only once, so save the ID and secret in a JSON file outside this repository: `{"client_id": "...", "client_secret": "..."}`. Google's downloaded `client_secret_*.json` also works, if the client is a Desktop app (its JSON starts with `"installed"`; a Web application client starts with `"web"` and will not work).
5. Sign in. A browser opens; approve read-only access to your events:

   ```sh
   make google-auth ARGS="-client-file /path/to/client.json"
   ```

6. Restart the dashboard server (`make run` for normal use, or `make dev` for development). The Up next widget now shows your events.

Sign-in saves the client and a refresh token to `everyday/google.json` in your user config directory, readable only by you. For a Pi, authorize from its desktop browser or follow the credential-transfer notes in the [Raspberry Pi guide](raspberry-pi.md#bring-over-existing-data-optional). Treat this file as a secret. If Google ever revokes access, the server logs it and the widget says unavailable. Sign in again with a plain `make google-auth`, with no arguments, which reuses the saved client.

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
- `amount` is the usual amount, shown with cents. Leave it out for a bill with no fixed amount, such as a credit card: its card says "Due" where the figure would be (and nothing once it is paid), and it is otherwise a bill like any other, including going overdue. An amount of 0 means the same as none.
- A bill can be ticked as paid, from the circle beside it in [Settings](#settings). That saves `paidFor`, the due date it was paid for, and the bill counts as paid only while that is the due date it is showing, so the tick clears itself when the next month's bill comes round. Paying ahead works: tick it any time before the due date.
- A bill whose due date passes unpaid becomes **overdue**. There is no queue behind this: it is worked out each time from the bill's day of the month, its `paidFor` date, and `addedOn`, the day it was first saved. The oldest missed due date at or after `addedOn` that is later than `paidFor` is the one shown, with how many days late it is. One card per bill, however many months were missed, and it stays until ticked (or the bill is deleted), when a single tick settles every missed month. A bill only starts being tracked from the day it was saved, so a new bill is never late for dates before it existed. Bills that were in the file before this existed have no `addedOn` and are not tracked until their next save from Settings.
- On the dashboard, overdue bills come first (the longest overdue on top), then bills still to pay, soonest due, then the ones already paid, dimmed with a check and the word "Paid". An overdue bill replaces its next-due card rather than adding a second one, and a quiet "N overdue" count sits in the widget header. When space is short, what needs doing gets it.
- Files from before this change may carry an `autopay` field per bill. It is read and ignored, and the next save from Settings leaves it out.
- The file is re-read whenever it changes, so edits appear within a few seconds with no restart. Leaving the file out shows "No bills yet". A mistake in it (a misspelled field, a day outside 1 to 31, a negative amount) shows "Bills unavailable" on the display and logs the exact problem once.
- The widget shows as many whole cards as fit.
- You can edit the file by hand or from [Settings](#settings); both change the same file.

## Settings

The gear beside the location opens Settings as a full screen over the dashboard, which keeps updating underneath. Tap the location name to open Location directly. It has three sections:

- **Bills**: add, edit and delete bills, listed by the day of the month they fall due, with bills that share a day grouped together. Each bill has a name, an amount (or an "Amount varies" switch, for a card with no fixed amount), a due day (a stepper, so nothing needs typing), and a circle to tick when it is paid. A change is saved as one whole file, written atomically, and shows on the dashboard at once. If the bills file has a mistake in it, Settings shows what it is and lets you replace the file.
- **Appearance**: the theme.
- **Location**: search for a city or postal code, choose a matching region/country, then save the weather location. Search happens on submission, not on every keystroke.

Done or Escape closes it, and it closes itself after 90 seconds without a touch, so it never stays open on the wall. It has an address (`http://127.0.0.1:8080/#/settings/bills`), so a section can be opened directly.

There is no login yet, so Settings only works from the machine the dashboard runs on. The endpoints that change things (`PUT /api/bills`, `PUT /api/location`) reject a request unless it is addressed to `localhost`, `127.0.0.1` or `::1`, is JSON, and either names no origin or names this server. That keeps another web page in your browser from editing your bills, and it keeps the dashboard from being edited from elsewhere on your network by accident. These checks are not a substitute for authentication. Do not expose the server through a public reverse proxy, tunnel, or port forward. Authenticated remote access requires separate work.

## What it shows

The MVP is a landscape-only display with one fixed layout and five widgets:

| Widget | Shows |
| --- | --- |
| Clock | Time, weekday, and date |
| Month calendar | The current month, or only the current week when the widget is short |
| Weather | Current conditions with a five-day forecast (today plus four days), low → high temperature ranges, an hourly view via the **Days / Hours** button, and any alert |
| Up next | Upcoming Google Calendar events, as many whole cards as fit |
| Bills | Your recurring bills as cards: overdue first, then what is left to pay, then what is already paid, each with its amount and due date |

Each widget adapts to the space it receives. Smaller sizes show less detail instead of scaling content down. Tap the settings gear beside **Dallas, TX** to open [Settings](#settings), where you can edit your bills, choose a weather location, and choose a colour theme: **Amber**, **Evergreen**, **Graphite**, **Indigo**, **Ocean**, **Plum**, **Stone** or **Turquoise**. The theme applies immediately and persists in this browser. Notices use a shared toast that dismisses after 4.5 seconds.

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
| `backend/internal/config` | Config-file and environment settings, including saved weather location |
| `backend/internal/web` | Built frontend assets served alongside the API |
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
| GET | `/api/location` | Saved weather location, whether it is editable, and why not if locked |
| GET | `/api/locations?q=…` | Up to eight matching city/postal-code locations with coordinates, through Open-Meteo |
| PUT | `/api/location` | Saves a name, latitude and longitude; returns saved settings and starts a weather refresh |

Polling uses a five-second interval with an eight-second timeout. Saves request an immediate refresh; any older in-flight response is ignored. Timers clean up on unmount. The last successful state remains visible during outages; the footer reports delay or connection failure. The backend contacts Open-Meteo every 15 minutes (retrying after a minute on failure), not on each poll. A failed refresh keeps the last forecast, which is dropped after an hour so stale weather is never shown as current. Google Calendar is read every 5 minutes (retrying after a minute on failure), and its events are dropped after 30 minutes without a successful refresh. In mock mode, dates are anchored to backend startup rather than advancing on each poll.

Svelte is the only frontend runtime dependency. Go uses its standard library. No layout libraries, network fonts, databases, or cloud hosting are used. Weather and city search use Open-Meteo through the backend.

## Verification

```sh
make check   # Go race tests + vet, Svelte/TS checks, Node behavioral tests
make build   # Go binary in build/everyday, frontend build in frontend/dist
```

Format with `cd frontend && npm run format`.

Tests cover that the fixed layout stays inside the grid without overlapping, month dates and the current-week row, that mock dates stay stable while polling, Open-Meteo parsing, the bills file and due-date arithmetic (month ends, leap years, year rollover), that the weather and calendar caches keep the last good data through failures, the Google sign-in flow (PKCE, state checks, token refresh), and calendar event handling (all-day events lead their day, cancelled and declined events, merged calendars, mirroring and excluding calendars, one broken calendar being skipped), and the config file. The intended LCD still needs a physical legibility and long-running check.

Use `make run` for a production-style local session with Go serving the built frontend and API together. Missing assets and unknown API routes return 404 rather than the app HTML; directory listings and hidden files are not served. HTML is not stored in the browser cache, and other assets revalidate when requested. The intended LCD still needs a physical legibility check and an overnight reliability run before appliance deployment.

## After the MVP

Removed from the prototype and preserved in the git tag `prototype-v0`:

- Widget and layout editing (moving and resizing widgets)
- Multiple named boards and portrait layouts
- Sports and Around Dallas widgets (V1+ or V2)
- The priority and attention system

Editing calendars in Settings, a login so Settings can be used from other devices, and iCloud calendars (Google only for now) remain future work. Raspberry Pi deployment has an [initial manual guide](raspberry-pi.md), but has not yet been verified on physical hardware.
