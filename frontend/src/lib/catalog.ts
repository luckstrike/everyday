import type { WidgetId } from './types.ts';
export const catalog: { id: WidgetId; title: string; description: string }[] = [
  { id: 'clock', title: 'Clock', description: 'Time and date' },
  {
    id: 'weather',
    title: 'Weather',
    description: 'Current conditions and forecast',
  },
  { id: 'calendar', title: 'Calendar', description: 'Your next event' },
  { id: 'finance', title: 'Bills', description: 'Upcoming payments' },
  { id: 'discover', title: 'Around Dallas', description: 'Local discoveries' },
  {
    id: 'sports',
    title: 'Sports',
    description: 'Dallas scores and game times',
  },
];
