import type { Placement, WidgetId } from './types.ts';

// The dashboard uses one fixed landscape layout. Widgets are placed on a
// 24 by 24 grid, so a future admin panel can edit this same data.
export const GRID = 24;

export const layout: Placement[] = [
  { id: 'clock', x: 0, y: 0, w: 10, h: 10 },
  { id: 'weather', x: 10, y: 0, w: 14, h: 10 },
  { id: 'calendar', x: 0, y: 10, w: 10, h: 14 },
  { id: 'upnext', x: 10, y: 10, w: 7, h: 14 },
  { id: 'finance', x: 17, y: 10, w: 7, h: 14 },
];

export const widgetTitles: Record<WidgetId, string> = {
  clock: 'Clock',
  calendar: 'Month calendar',
  weather: 'Weather',
  upnext: 'Up next',
  finance: 'Bills',
};

export function overlaps(a: Placement, b: Placement): boolean {
  return (
    a.x < b.x + b.w && a.x + a.w > b.x && a.y < b.y + b.h && a.y + a.h > b.y
  );
}
