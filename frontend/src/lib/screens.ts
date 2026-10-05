import type { ScreenDefinition, ScreenId } from './types.ts';
import type { Candidate } from './priority.ts';
export interface ScreenSession {
  current: ScreenId;
  since: number;
  lastShown: Partial<Record<ScreenId, number>>;
  bypassed: string[];
  activeEvent?: string;
}
export function initialSession(now: number): ScreenSession {
  return {
    current: 'home',
    since: now,
    lastShown: { home: now },
    bypassed: [],
  };
}
export function selectScreen(
  session: ScreenSession,
  candidates: Candidate[],
  definitions: ScreenDefinition[],
  now: number,
): ScreenSession {
  const definition = definitions.find((s) => s.id === session.current)!;
  const elapsed = now - session.since;
  if (elapsed < definition.minDurationMs) return session;
  const current = candidates.find((c) => c.id === session.current)!;
  const eligible = candidates
    .filter((c) => c.id !== session.current)
    .filter((c) => {
      const spec = definitions.find((s) => s.id === c.id)!;
      const last = session.lastShown[c.id];
      return (
        last === undefined ||
        now - last >= spec.cooldownMs ||
        (c.urgent && !!c.eventId && !session.bypassed.includes(c.eventId))
      );
    });
  let next: Candidate | undefined;
  // A temporary takeover ends back at Home as soon as its minimum reading time permits.
  if (
    session.activeEvent &&
    current.eventId !== session.activeEvent &&
    session.current !== 'home'
  )
    next = candidates.find((c) => c.id === 'home');
  if (!next && elapsed < definition.maxDurationMs)
    next = eligible.find((c) => c.modifier >= 15 && c.score > current.score);
  if (!next && elapsed >= definition.maxDurationMs) {
    next =
      session.current !== 'home'
        ? eligible.find((c) => c.id === 'home')
        : (eligible.find((c) => c.modifier > 0) ??
          [...eligible].sort(
            (a, b) =>
              (session.lastShown[a.id] ?? -Infinity) -
                (session.lastShown[b.id] ?? -Infinity) ||
              b.score - a.score ||
              a.id.localeCompare(b.id),
          )[0]);
  }
  if (!next) return session;
  return {
    current: next.id,
    since: now,
    // Cooldown starts at exit, so a long display does not consume its own cooldown.
    lastShown: { ...session.lastShown, [session.current]: now, [next.id]: now },
    bypassed: [
      ...session.bypassed.filter((id) =>
        candidates.some((c) => c.eventId === id),
      ),
      ...(next.urgent && next.eventId ? [next.eventId] : []),
    ],
    activeEvent: next.eventId,
  };
}
