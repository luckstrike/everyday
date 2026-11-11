import type { SourceStatus } from './types.ts';

export interface SourceNotice {
  title: string;
  detail: string;
}

// A widget whose data is typed in rather than connected can supply its own
// wording, since "not connected" and "waiting for the latest data" would mislead
// someone who has a file to fix.
export interface NoticeWording {
  unconfigured?: SourceNotice;
  unavailableDetail?: string;
}

// What a widget says in place of its content when its data source isn't
// delivering. Live and mock data need no notice: an empty list from a working
// source is a real answer, such as "no upcoming events".
export function sourceNotice(
  status: SourceStatus,
  name: string,
  wording: NoticeWording = {},
): SourceNotice | null {
  switch (status) {
    case 'unconfigured':
      return (
        wording.unconfigured ?? {
          title: 'Not connected',
          detail: 'Nothing is hooked up yet.',
        }
      );
    case 'unavailable':
      return {
        title: `${name} unavailable`,
        detail: wording.unavailableDetail ?? 'Waiting for the latest data.',
      };
    default:
      return null;
  }
}
