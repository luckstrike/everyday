export type WidgetId = 'clock' | 'calendar' | 'weather' | 'upnext' | 'finance';
export type Size = 'small' | 'medium' | 'large';
// A kind of weather to draw; the backend picks it from the weather code.
// The two night kinds are only used for the current conditions; a forecast day is a daytime picture.
export type WeatherIcon =
  | 'clear'
  | 'clear-night'
  | 'partly-cloudy'
  | 'partly-cloudy-night'
  | 'cloudy'
  | 'fog'
  | 'rain'
  | 'snow'
  | 'storm';
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
    icon: WeatherIcon;
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
      icon: WeatherIcon;
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
    allDay: boolean;
  }[];
}
