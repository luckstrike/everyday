import type { Config, DashboardState } from './types.ts';
import type { BillEntry, ListedBill } from './bills-editor.ts';

// A failed request. When the server explains itself (it answers {"error": "..."}),
// the message is its explanation, which is written to be shown to a person.
export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/${path}`, {
    ...init,
    signal: AbortSignal.timeout(8000),
  });
  if (!response.ok) {
    let message = `Dashboard API returned ${response.status}`;
    try {
      const body = await response.json();
      if (typeof body?.error === 'string') message = body.error;
    } catch {
      // The body was not JSON; the status message stands.
    }
    throw new ApiError(response.status, message);
  }
  return response.json() as Promise<T>;
}
export const getState = () => request<DashboardState>('state');
export const getConfig = () => request<Config>('config');

// The bills as written in the bills file. problem is set when that file exists but
// cannot be read, in which case saving replaces it.
export interface BillsResponse {
  bills: ListedBill[];
  problem?: string;
}
export const getBills = () => request<BillsResponse>('bills');
export const saveBills = (bills: BillEntry[]) =>
  request<BillsResponse>('bills', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bills }),
  });
