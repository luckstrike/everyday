import type { SourceStatus } from './types.ts';

export interface SourceNotice {
  title: string;
  detail: string;
}

// What a widget says in place of its content when its data source isn't
// delivering. Live and mock data need no notice: an empty list from a working
// source is a real answer, such as "no upcoming events".
export function sourceNotice(
  status: SourceStatus,
  name: string,
): SourceNotice | null {
  switch (status) {
    case 'unconfigured':
      return { title: 'Not connected', detail: 'Nothing is hooked up yet.' };
    case 'unavailable':
      return {
        title: `${name} unavailable`,
        detail: 'Waiting for the latest data.',
      };
    default:
      return null;
  }
}
