import type {
  AttentionEvent,
  Presentation,
  ScreenDefinition,
  ScreenId,
  WidgetDefinition,
  WidgetId,
} from './types.ts';
export const modifiers: Record<string, Record<string, number>> = {
  weather: { normal: 0, rain: 15, severe: 60 },
  sports: { today: 10, starting: 20, live: 40, final: 20 },
  finance: { week: 5, tomorrow: 25, alert: 40 },
  calendar: { today: 10, upcoming: 25, soon: 50 },
};
export const screenFor: Record<WidgetId, ScreenId> = {
  clock: 'home',
  calendar: 'home',
  weather: 'weather',
  sports: 'sports',
  finance: 'finance',
};
export function eventModifier(event: AttentionEvent, now: number): number {
  const start = Date.parse(event.startsAt),
    end = Date.parse(event.expiresAt);
  if (
    !Number.isFinite(start) ||
    !Number.isFinite(end) ||
    now < start ||
    now >= end ||
    end <= start
  )
    return 0;
  const value = modifiers[event.widget]?.[event.kind] ?? 0;
  return event.decay ? value * (1 - (now - start) / (end - start)) : value;
}
export interface Candidate {
  id: ScreenId;
  score: number;
  modifier: number;
  eventId?: string;
  widget?: WidgetId;
  presentation: Presentation;
  urgent: boolean;
}
export function rankScreens(
  screens: ScreenDefinition[],
  widgets: WidgetDefinition[],
  events: AttentionEvent[],
  now: number,
): Candidate[] {
  return screens
    .map((screen) => {
      const active = events
        .filter((e) => screenFor[e.widget] === screen.id)
        .map((event) => ({ event, value: eventModifier(event, now) }))
        .filter((e) => e.value > 0)
        .sort(
          (a, b) => b.value - a.value || a.event.id.localeCompare(b.event.id),
        );
      const top = active[0];
      const base = top
        ? (widgets.find((w) => w.id === top.event.widget)?.basePriority ??
          screen.basePriority)
        : screen.basePriority;
      const presentation: Presentation =
        top?.event.kind === 'severe'
          ? 'FULLSCREEN'
          : top && top.value >= 15
            ? 'EMPHASIZED'
            : 'NORMAL';
      return {
        id: screen.id,
        score: base + (top?.value ?? 0),
        modifier: top?.value ?? 0,
        eventId: top?.event.id,
        widget: top?.event.widget,
        presentation,
        urgent: (top?.value ?? 0) >= 50,
      };
    })
    .sort((a, b) => b.score - a.score || a.id.localeCompare(b.id));
}
