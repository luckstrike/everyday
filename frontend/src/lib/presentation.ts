import type { Candidate } from './priority.ts';
import type { Placement, WidgetId } from './types.ts';
export const presentationTiming = {
  minimumMs: 10000,
  maximumMs: 25000,
  cooldownMs: 65000,
};
export interface PresentationSession {
  active?: Candidate;
  since: number;
  lastEnded: Partial<Record<WidgetId, number>>;
  handled: string[];
}
export function initialPresentation(): PresentationSession {
  return { since: 0, lastEnded: {}, handled: [] };
}
// Ambient boards emphasize in place. Fullscreen timing remains available only
// through an explicit policy opt-in; priority alone never changes the layout.
export function updatePresentation(
  session: PresentationSession,
  ranks: Candidate[],
  now: number,
  allowFullscreen = false,
): PresentationSession {
  if (!allowFullscreen)
    return session.active ? { ...session, active: undefined } : session;
  if (session.active) {
    const elapsed = now - session.since;
    const ongoing = ranks.some(
      (c) =>
        c.eventId === session.active?.eventId &&
        c.presentation === 'FULLSCREEN',
    );
    if (
      elapsed < presentationTiming.minimumMs ||
      (ongoing && elapsed < presentationTiming.maximumMs)
    )
      return session;
    return {
      since: now,
      lastEnded: { ...session.lastEnded, [session.active.id]: now },
      handled: [...session.handled, session.active.eventId!].slice(-64),
    };
  }
  const next = ranks.find(
    (c) =>
      c.presentation === 'FULLSCREEN' &&
      c.eventId &&
      !session.handled.includes(c.eventId) &&
      now - (session.lastEnded[c.id] ?? -Infinity) >=
        presentationTiming.cooldownMs,
  );
  return next ? { ...session, active: next, since: now } : session;
}
export function attentionFor(
  session: PresentationSession,
  ranks: Candidate[],
  items: Placement[],
): Candidate | undefined {
  if (session.active) return session.active;
  const candidate = ranks.find(
    (c) => c.modifier >= 15 && items.some((p) => p.id === c.widget),
  );
  return candidate ? { ...candidate, presentation: 'EMPHASIZED' } : undefined;
}
// Always return a projection, never mutate the user's layout.
export function displayLayout(
  items: Placement[],
  attention?: Candidate,
): Placement[] {
  return attention?.presentation === 'FULLSCREEN' && attention.widget
    ? [{ id: attention.widget, x: 0, y: 0, w: 24, h: 24 }]
    : items.map((p) => ({ ...p }));
}
