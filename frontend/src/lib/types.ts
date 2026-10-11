export type WidgetId =
  'clock' | 'weather' | 'sports' | 'finance' | 'calendar' | 'discover';
export type Size = 'small' | 'medium' | 'large';
export type Presentation = 'NORMAL' | 'EMPHASIZED' | 'FULLSCREEN';
export interface WidgetDefinition {
  id: WidgetId;
  title: string;
  basePriority: number;
  minWidth: number;
  minHeight: number;
  defaultWidth: number;
  defaultHeight: number;
}
export interface Config {
  pollIntervalMs: number;
  development: boolean;
}
export interface Placement {
  id: WidgetId;
  x: number;
  y: number;
  w: number;
  h: number;
}
export interface AttentionEvent {
  id: string;
  widget: WidgetId;
  kind: string;
  startsAt: string;
  expiresAt: string;
  decay: boolean;
}
export interface DashboardState {
  generatedAt: string;
  mock: boolean;
  scenario: string;
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
  sports: {
    league: string;
    home: string;
    homeCode: string;
    away: string;
    awayCode: string;
    homeScore: number;
    awayScore: number;
    status: string;
    detail: string;
    startsAt: string;
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
  localEvents: {
    id: string;
    title: string;
    venue: string;
    startsAt: string;
    category: string;
    price: string;
    reason: string;
  }[];
  events: AttentionEvent[];
}
