// The colour themes, in the order Appearance lists them. Each has a block of colours
// in app.css under [data-theme='<id>'], and the settings swatch reads its colours from
// that same block, so a theme is defined in one place.
export const themes = [
  { id: 'amber', label: 'Amber' },
  { id: 'evergreen', label: 'Evergreen' },
  { id: 'graphite', label: 'Graphite' },
  { id: 'indigo', label: 'Indigo' },
  { id: 'ocean', label: 'Ocean' },
  { id: 'plum', label: 'Plum' },
  { id: 'stone', label: 'Stone' },
  { id: 'turquoise', label: 'Turquoise' },
] as const;

export type Theme = (typeof themes)[number]['id'];
export const defaultTheme: Theme = 'evergreen';
export const THEME_KEY = 'everyday.theme.v1';

const isTheme = (value: unknown): value is Theme =>
  themes.some((theme) => theme.id === value);

// The saved theme, or the default when none is saved, the saved one is no longer a
// theme, or the browser will not say.
export function loadTheme(): Theme {
  try {
    const saved = localStorage.getItem(THEME_KEY);
    return isTheme(saved) ? saved : defaultTheme;
  } catch {
    return defaultTheme;
  }
}
