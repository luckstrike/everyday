// Which screen is showing. The dashboard is the default; settings opens over it and
// lives at an address (#/settings/bills) so it can be opened directly later, from
// another device. A plain hash needs no router.
export type Section = 'bills' | 'appearance';
export type View =
  { name: 'dashboard' } | { name: 'settings'; section: Section };

// The rail lists sections alphabetically, so a new one has an obvious place.
export const sections: { id: Section; label: string }[] = [
  { id: 'appearance', label: 'Appearance' },
  { id: 'bills', label: 'Bills' },
];

// Bills is what the gear opens, and what a bare or unknown address falls back to,
// whatever order the rail is in.
export const defaultSection: Section = 'bills';

export function parseHash(hash: string): View {
  const [screen, section] = hash.replace(/^#\/?/, '').split('/');
  if (screen !== 'settings') return { name: 'dashboard' };
  const known = sections.find((s) => s.id === section);
  return { name: 'settings', section: known ? known.id : defaultSection };
}

export function hashFor(view: View): string {
  return view.name === 'dashboard' ? '' : `#/settings/${view.section}`;
}
