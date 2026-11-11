# Ambient dashboard

Reference: Paper Ambient Landscape supplied by the owner. A wall display in a softly lit home should feel quiet, readable, and stable.

Use system sans, normal weights, and a light clock face. Evergreen background oklch(24.30% .0194 167.17), raised surface oklch(29.53% .0268 163.93), sage accent oklch(82.38% .0619 130.22), primary text oklch(92.82% .0140 124.56), muted text oklch(74.91% .0230 149.61), outline oklch(39.45% .0284 156.49). The raised surface is for controls and floating elements (hovered and selected buttons, the settings popover, the toast) and for the alternating fill of the event cards inside Up next; a widget's own background never uses it. All widgets have a 1px outline and 18px radius. No decorative shadows except floating controls.

Clock digits use tabular numbers with light tracking (-0.025em), keeping minutes distinct while retaining a compact clock face.

Every widget uses `WidgetFrame`, which owns outer padding, responsive padding rules, the header-to-content gap, and the remaining content area. Individual widgets must not duplicate or override that frame spacing. The frame renders the shared `WidgetHeader`: a muted, normal-weight heading at the top left, with an optional `actions` snippet aligned to the right. Use normal reading and keyboard order for controls; insert a new control before existing ones to grow the row from right to left while keeping the rightmost control anchored. Titles truncate before controls are squeezed. The actions row is 20px tall like the heading; buttons retain their 44px minimum touch target, extending into the surrounding spacing without pushing the heading down. Headers stay at the top in loading, empty, and not-connected states; the remaining content area centers those states independently.

To add a widget, supply a title, content through the default children snippet, and optionally header controls through the named actions snippet:

```svelte
<WidgetFrame title="Example">
  {#snippet actions()}
    <button onclick={refresh}>Refresh</button>
    <button onclick={openMenu}>Options</button>
  {/snippet}

  <p>Widget content</p>
</WidgetFrame>
```

Omit the actions snippet when there are no controls. Content is centered by default; a widget can give its own content wrapper `flex: 1` to fill the available space, as the event list does. Widget-specific spacing such as card padding and forecast gaps stays inside that content. The frame also exposes `--widget-content-height` for content sizing, so the clock does not repeat header or outer-padding calculations. `WidgetHeader` remains usable independently where a full widget frame is not needed.

The clock shows a "Clock" heading above the time and the weekday and date.

The month calendar is its own widget, with Sunday-first weeks and today highlighted. When the widget is at least 280px tall it shows the whole month with room for a six-week month. Shorter widgets show only the current week. The month is read-only and updates with the local date; it is distinct from the Up next widget, which shows the next event.

Up next lists up to three events under an "Up next" header, each as a card: the widget's 1px outline with a 12px radius, and every other card (the first, third) filled with the raised surface colour, so both themes follow their own palette. All-day events lead their day. The first card's time uses the accent colour. The list shows only whole cards, as many as the widget's height allows, so smaller displays show two or one. The empty and not-connected states keep their text vertically centered instead.

The display is landscape only and uses one fixed layout on a 24 by 24 grid: Clock and Weather across the top, then Month calendar, Up next, and Bills below. Landscape uses 48px outer padding and 24px gaps. All widgets have a 1px outline and a background that matches the page. Widgets never change position, emphasis, or size at runtime, and nothing takes over the screen.

Appearance settings sit beside the location in the header. Evergreen remains the default; Turquoise uses deep teal surfaces and pale turquoise accents with the same spacing and contrast hierarchy. The browser stores the selected theme. A shared toast component dismisses after 4.5 seconds, pauses on hover/focus, and handles repeated notices.

Weather centers current conditions and forecast in the content area below its header, with a shared left inset. The sun is optically aligned to forecast text. The forecast switch occupies the header's right-hand actions slot with a 44px touch target, rather than floating over the weather content. Widgets at most 230px tall show current conditions without the forecast or its switch. When the data includes an alert, compact Weather replaces the forecast with alert text, and its shortest presentation shows the warning alone. Weather spacing and responsive rules live in its component.

Smaller widgets show deliberately reduced content rather than scaled content. Compact tiles keep core information: event and time, or bill and amount with due date. Secondary detail gives way before core content clips.
