# Ambient dashboard

Reference: Paper Ambient Portrait, Ambient Landscape, Everyday Board Picker, and Touch Pick a Widget supplied by the owner. A wall display in a softly lit home should feel quiet, readable, and stable.

Use system sans, normal weights, and a light clock face. Evergreen background oklch(24.30% .0194 167.17), raised surface oklch(29.53% .0268 163.93), sage accent oklch(82.38% .0619 130.22), primary text oklch(92.82% .0140 124.56), muted text oklch(74.91% .0230 149.61), outline oklch(39.45% .0284 156.49). All widgets have a 1px outline and 18px radius. No decorative shadows except floating controls.

Clock digits use tabular numbers with light tracking (-0.025em), keeping minutes distinct while retaining a compact clock face.

Landscape uses 48px outer padding and 24px gaps. Portrait uses 32px padding and 20px gaps. Clock and weather anchor the upper area; calendar, bills and local discovery sit below; sports is a quiet full-width strip. All widget backgrounds match the board, including Around Dallas, which has no Sample pick label. Priority emphasis gently pulses only the border three times over nine seconds, then keeps a steady accent border until emphasis ends. Reduced motion uses the steady border immediately. Priority never brightens widget backgrounds or text. The Add widget action sits bottom right.

Named boards switch manually. Each board saves independent portrait and landscape arrangements on a bounded 24 by 24 grid. The grid supports placement and collision checks; it never grows, auto-fills gaps, or rearranges existing tiles. Selecting a catalog widget reveals its actual footprint, size choices and valid placement. Touch, keyboard and optional pointer dragging share the same validation. Editing is a draft until Save; Cancel restores the saved board. Controls have at least 44px effective touch targets.

Edit layout is a visible secondary action beside Add widget. It opens editing directly, with accent widget outlines, handles and explicit Save/Cancel controls. Widget content retains its normal padding and size in edit mode. The picker uses 24px outer padding (20px on narrow displays), aligned catalog icons and a divider below its instructions.

Footer layout actions are compact 44px touch targets with 13px labels, 16px icons and 10px horizontal padding. Edit layout uses the same subtle 1px outline color as the widgets; Add widget retains the accent fill.

Closing an untouched picker opened through Add widget returns to the dashboard. Closing a picker within Edit layout, or after a layout change, keeps the editor open. Cancel exits immediately when the draft matches the saved layout, including reverted edits. Otherwise a focused, native modal asks whether to keep editing or discard changes; Escape keeps editing. Only Save layout persists the draft. Browser leave protection applies only to actual unsaved layout changes.

Appearance settings sit beside the location in the header. Evergreen remains the default; Turquoise uses deep teal surfaces and pale turquoise accents with the same spacing and contrast hierarchy. The browser stores the selected theme independently from board geometry. Weather centers current conditions and forecast as one group with equal top/bottom breathing room and a shared left inset. The sun is optically aligned to forecast text. The forecast switch sits at the top right with a 44px touch target; compact weather reduces detail rather than clipping. Weather spacing and responsive rules live in its component. A shared toast component dismisses after 4.5 seconds, pauses on hover/focus, and handles repeated notices.
