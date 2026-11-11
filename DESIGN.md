# Ambient dashboard

Reference: Paper Ambient Landscape supplied by the owner. A wall display in a softly lit home should feel quiet, readable, and stable.

Use system sans, normal weights, and a light clock face. Evergreen background oklch(24.30% .0194 167.17), raised surface oklch(29.53% .0268 163.93), sage accent oklch(82.38% .0619 130.22), primary text oklch(92.82% .0140 124.56), muted text oklch(74.91% .0230 149.61), outline oklch(39.45% .0284 156.49). All widgets have a 1px outline and 18px radius. No decorative shadows except floating controls.

Clock digits use tabular numbers with light tracking (-0.025em), keeping minutes distinct while retaining a compact clock face.

The clock shows the time and the weekday and date only.

The month calendar is its own widget, with Sunday-first weeks and today highlighted. When the widget is at least 280px tall it shows the whole month with room for a six-week month. Shorter widgets show only the current week. The month is read-only and updates with the local date; it is distinct from the Up next widget, which shows the next event.

The display is landscape only and uses one fixed layout on a 24 by 24 grid: Clock and Weather across the top, then Month calendar, Up next, and Bills below. Landscape uses 48px outer padding and 24px gaps. All widgets have a 1px outline and a background that matches the page. Widgets never change position, emphasis, or size at runtime, and nothing takes over the screen.

Appearance settings sit beside the location in the header. Evergreen remains the default; Turquoise uses deep teal surfaces and pale turquoise accents with the same spacing and contrast hierarchy. The browser stores the selected theme. A shared toast component dismisses after 4.5 seconds, pauses on hover/focus, and handles repeated notices.

Weather centers current conditions and forecast as one group with equal top/bottom breathing room and a shared left inset. The sun is optically aligned to forecast text. The forecast switch sits at the top right with a 44px touch target; compact weather reduces detail rather than clipping. When the data includes an alert, compact Weather replaces the forecast with alert text, and its shortest presentation shows the warning alone. Weather spacing and responsive rules live in its component.

Smaller widgets show deliberately reduced content rather than scaled content. Compact tiles keep core information: event and time, or bill and amount with due date. Secondary detail gives way before core content clips.
