# Documentation screenshots

These images show the application at a 1600 × 900 browser viewport in the default Evergreen theme.

For replacements, stop any existing server and run `EVERYDAY_MOCK=1 make run`. Confirm that the dashboard footer says **Mock data** before capturing:

- `dashboard.png`: `http://127.0.0.1:8080/`
- `appearance.png`: `http://127.0.0.1:8080/#/settings/appearance`

Use a fresh browser profile. Keep the screenshot framing consistent and include meaningful alt text in the README. Never capture a live calendar, real bills, credentials, or personal settings for the public documentation.

The sample events, bill names, and amounts come from `backend/internal/mock/services.go`; they aren't the owner's data. No holidays are included.
