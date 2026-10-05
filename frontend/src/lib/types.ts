export type WidgetId = 'clock' | 'weather' | 'sports' | 'finance' | 'calendar';
export type ScreenId = 'home' | 'weather' | 'sports' | 'finance';
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
export interface ScreenDefinition {
  id: ScreenId;
  title: string;
  basePriority: number;
  cooldownMs: number;
  minDurationMs: number;
  maxDurationMs: number;
}
export interface Config {
  screens: ScreenDefinition[];
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
export type Layouts = Record<ScreenId, Placement[]>;
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
  events: AttentionEvent[];
}
