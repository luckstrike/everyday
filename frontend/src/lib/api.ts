import type { Config, DashboardState, WidgetDefinition } from './types.ts';
async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/${path}`, {
    ...init,
    signal: AbortSignal.timeout(8000),
  });
  if (!response.ok)
    throw new Error(`Dashboard API returned ${response.status}`);
  return response.json() as Promise<T>;
}
export const getState = () => request<DashboardState>('state');
export const getConfig = () => request<Config>('config');
export const getWidgets = () => request<WidgetDefinition[]>('widgets');
