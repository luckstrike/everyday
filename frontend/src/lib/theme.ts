export type Theme = 'evergreen' | 'turquoise';
export const THEME_KEY = 'apartment-display.theme.v1';
export function loadTheme(): Theme {
  try {
    return localStorage.getItem(THEME_KEY) === 'turquoise'
      ? 'turquoise'
      : 'evergreen';
  } catch {
    return 'evergreen';
  }
}
