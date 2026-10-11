import type {
  AttentionEvent,
  Presentation,
  WidgetDefinition,
  WidgetId,
} from './types.ts';
export const modifiers: Record<string, Record<string, number>> = {
  weather: { normal: 0, rain: 15, severe: 60 },
  sports: { today: 10, starting: 20, live: 40, final: 20 },
  finance: {
    week: 5,
    soon: 15,
    tomorrow: 25,
    today: 35,
    overdue: 40,
    alert: 40,
  },
  calendar: { today: 10, upcoming: 25, soon: 50 },
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
  id: WidgetId;
  score: number;
  modifier: number;
  eventId?: string;
  widget?: WidgetId;
  presentation: Presentation;
  urgent: boolean;
}
export function rankWidgets(
  widgets: WidgetDefinition[],
  events: AttentionEvent[],
  now: number,
): Candidate[] {
  return widgets
    .map((widget) => {
      const active = events
        .filter((e) => e.widget === widget.id)
        .map((event) => ({ event, value: eventModifier(event, now) }))
        .filter((e) => e.value > 0)
        .sort(
          (a, b) => b.value - a.value || a.event.id.localeCompare(b.event.id),
        );
      const top = active[0];
      const base = widget.basePriority;
      const presentation: Presentation =
        top?.event.kind === 'severe'
          ? 'FULLSCREEN'
          : top && top.value >= 15
            ? 'EMPHASIZED'
            : 'NORMAL';
      return {
        id: widget.id,
        score: base + (top?.value ?? 0),
        modifier: top?.value ?? 0,
        eventId: top?.event.id,
        widget: widget.id,
        presentation,
        urgent: (top?.value ?? 0) >= 50,
      };
    })
    .sort((a, b) => b.score - a.score || a.id.localeCompare(b.id));
}
