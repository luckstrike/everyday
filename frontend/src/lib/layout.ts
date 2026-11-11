import type { Placement, WidgetId } from './types.ts';

// The dashboard uses one fixed landscape layout. Widgets are placed on a
// 24 by 24 grid, so a future admin panel can edit this same data.
export const GRID = 24;

export const layout: Placement[] = [
  { id: 'clock', x: 0, y: 0, w: 14, h: 12 },
  { id: 'weather', x: 14, y: 0, w: 10, h: 12 },
  { id: 'calendar', x: 0, y: 12, w: 7, h: 8 },
  { id: 'finance', x: 7, y: 12, w: 7, h: 8 },
  { id: 'discover', x: 14, y: 12, w: 10, h: 8 },
  { id: 'sports', x: 0, y: 20, w: 24, h: 4 },
];

export const widgetTitles: Record<WidgetId, string> = {
  clock: 'Clock',
  weather: 'Weather',
  calendar: 'Calendar',
  finance: 'Bills',
  discover: 'Around Dallas',
  sports: 'Sports',
};

export function overlaps(a: Placement, b: Placement): boolean {
  return (
    a.x < b.x + b.w && a.x + a.w > b.x && a.y < b.y + b.h && a.y + a.h > b.y
  );
}
