export type Theme = 'evergreen' | 'turquoise';
export const THEME_KEY = 'everyday.theme.v1';
export function loadTheme(): Theme {
  try {
    return localStorage.getItem(THEME_KEY) === 'turquoise'
      ? 'turquoise'
      : 'evergreen';
  } catch {
    return 'evergreen';
  }
}
