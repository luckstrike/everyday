export type WidgetId = 'clock' | 'calendar' | 'weather' | 'upnext' | 'finance';
export type Size = 'small' | 'medium' | 'large';
export interface Config {
  pollIntervalMs: number;
}
export interface Placement {
  id: WidgetId;
  x: number;
  y: number;
  w: number;
  h: number;
}
// Where a section's data came from; the backend sets one per section.
export type SourceStatus = 'live' | 'mock' | 'unavailable' | 'unconfigured';
export interface DashboardState {
  generatedAt: string;
  sources: {
    weather: SourceStatus;
    finance: SourceStatus;
    calendar: SourceStatus;
  };
  weather: {
    location: string;
    temperature: number;
    condition: string;
    high: number;
    low: number;
    precipitation: number;
    tomorrow: string;
    hourly: { label: string; temperature: number; condition: string }[];
    daily: {
      date: string;
      high: number;
      low: number;
      condition: string;
      precipitation: number;
    }[];
    alert: string;
  };
  finance: {
    bills: { name: string; amount: number; dueAt: string; autopay: boolean }[];
  };
  calendar: {
    title: string;
    location: string;
    startsAt: string;
    endsAt: string;
  }[];
}
