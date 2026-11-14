<script lang="ts">
  import { onMount, tick } from 'svelte';
  import Icon from './Icon.svelte';
  import SettingsHeader from './SettingsHeader.svelte';
  import { ApiError, getBills, saveBills } from '../lib/api.ts';
  import {
    draftFrom,
    dueNote,
    emptyDraft,
    entryFrom,
    ordinal,
    validateDraft,
    type BillDraft,
    type BillEntry,
  } from '../lib/bills-editor.ts';
  import { money } from '../lib/format.ts';

  // onchanged runs after bills are saved, so the dashboard can show them at once.
  let { onchanged }: { onchanged: () => void } = $props();

  let loading = $state(true);
  // Why the bills cannot be edited at all, such as sample data being shown.
  let unavailable = $state('');
  let bills = $state<BillEntry[]>([]);
  // Set when the bills file exists but cannot be read; saving here replaces it.
  let problem = $state('');

  let editing = $state(false);
  let editIndex = $state<number | null>(null); // null while adding a new bill
  let draft = $state<BillDraft>(emptyDraft());
  let formError = $state('');
  let saving = $state(false);
  let confirmDelete = $state(false);
  let nameInput = $state<HTMLInputElement>();

  function apply(response: { bills: BillEntry[]; problem?: string }) {
    bills = response.bills;
    problem = response.problem ?? '';
  }

  onMount(async () => {
    try {
      apply(await getBills());
    } catch (error) {
      unavailable =
        error instanceof ApiError && error.status === 404
          ? 'Bills cannot be edited while sample data is showing.'
          : 'Could not load your bills. Check that the dashboard is running.';
    } finally {
      loading = false;
    }
  });

  async function startAdd() {
    editIndex = null;
    draft = emptyDraft();
    formError = '';
    confirmDelete = false;
    editing = true;
    await tick();
    nameInput?.focus();
  }

  function startEdit(index: number) {
    editIndex = index;
    draft = draftFrom(bills[index]);
    formError = '';
    confirmDelete = false;
    editing = true;
  }

  function cancel() {
    editing = false;
    formError = '';
  }

  // Every change replaces the whole list, so the file is always one consistent
  // version and there is nothing half-saved to recover from.
  async function persist(next: BillEntry[]) {
    saving = true;
    formError = '';
    try {
      apply(await saveBills(next));
      editing = false;
      onchanged();
    } catch (error) {
      formError =
        error instanceof ApiError
          ? error.message
          : 'Could not reach the dashboard. Try again.';
    } finally {
      saving = false;
      confirmDelete = false;
    }
  }

  function save() {
    const message = validateDraft(draft);
    if (message) {
      formError = message;
      return;
    }
    const entry = entryFrom(draft);
    void persist(
      editIndex === null
        ? [...bills, entry]
        : bills.map((bill, i) => (i === editIndex ? entry : bill)),
    );
  }

  function remove() {
    // A second tap confirms, so one stray touch cannot delete a bill.
    if (!confirmDelete) {
      confirmDelete = true;
      return;
    }
    void persist(bills.filter((_, i) => i !== editIndex));
  }

  // The Add button lives in the header once there are bills; before that the empty
  // state offers it, so the screen never shows two.
  const showAdd = $derived(
    !loading && !unavailable && !editing && bills.length > 0,
  );

  const stepDay = (by: number) => {
    draft.day = Math.min(31, Math.max(1, draft.day + by));
  };
</script>

<div class="bills-editor">
  <SettingsHeader
    title={editing ? (editIndex === null ? 'New bill' : 'Edit bill') : 'Bills'}
    description={editing
      ? 'What it is, how much it is, and which day of the month it is due.'
      : 'Each bill repeats every month on its due day.'}
  >
    {#snippet actions()}{#if showAdd}<button class="primary" onclick={startAdd}
          >Add bill</button
        >{/if}{/snippet}
  </SettingsHeader>

  {#if loading}
    <p class="quiet">Loading your bills…</p>
  {:else if unavailable}
    <p class="quiet" role="status">{unavailable}</p>
  {:else if editing}
    <form
      class="bill-form"
      onsubmit={(event) => {
        event.preventDefault();
        save();
      }}
    >
      <label class="field">
        <span>Name</span>
        <input
          bind:this={nameInput}
          bind:value={draft.name}
          autocomplete="off"
          maxlength="60"
          placeholder="Rent, Internet, Phone…"
        />
      </label>
      <div class="pair">
        <label class="field">
          <span>Amount</span>
          <input
            bind:value={draft.amount}
            inputmode="decimal"
            autocomplete="off"
            placeholder="0.00"
          />
        </label>
        <div class="field">
          <span id="day-label">Due day</span>
          <div class="stepper" role="group" aria-labelledby="day-label">
            <button
              type="button"
              aria-label="Earlier day"
              disabled={draft.day <= 1}
              onclick={() => stepDay(-1)}>−</button
            >
            <output aria-live="polite">{draft.day}</output>
            <button
              type="button"
              aria-label="Later day"
              disabled={draft.day >= 31}
              onclick={() => stepDay(1)}>+</button
            >
          </div>
        </div>
      </div>
      <p class="quiet note">{dueNote(draft.day)}</p>
      <button
        type="button"
        class="switch"
        role="switch"
        aria-checked={draft.autopay}
        onclick={() => (draft.autopay = !draft.autopay)}
      >
        <span class="track" aria-hidden="true"><span class="thumb"></span></span
        >
        <span>Autopay</span>
      </button>
      {#if formError}<p class="error" role="alert">{formError}</p>{/if}
      <div class="actions">
        <button type="submit" class="primary" disabled={saving}
          >{saving ? 'Saving…' : 'Save'}</button
        >
        <button type="button" onclick={cancel} disabled={saving}>Cancel</button>
        {#if editIndex !== null}<button
            type="button"
            class="danger"
            onclick={remove}
            disabled={saving}
            >{confirmDelete ? 'Tap again to delete' : 'Delete'}</button
          >{/if}
      </div>
    </form>
  {:else}
    {#if problem}<p class="error" role="alert">
        The saved bills file has a problem: {problem}. Saving here replaces it.
      </p>{/if}
    {#if formError}<p class="error" role="alert">{formError}</p>{/if}
    {#if bills.length === 0}
      <div class="empty-state">
        <p class="empty-title">No bills yet</p>
        <p class="quiet">Add one and it appears on the dashboard.</p>
        <button class="primary" onclick={startAdd}>Add your first bill</button>
      </div>
    {:else}
      <ul class="bill-list">
        {#each bills as bill, i}<li>
            <button class="bill-row" onclick={() => startEdit(i)}>
              <span class="bill-main">
                <span class="bill-name">{bill.name}</span>
                <span class="quiet"
                  >The {ordinal(bill.day)} · {bill.autopay
                    ? 'Autopay on'
                    : 'Manual payment'}</span
                >
              </span>
              <span class="bill-amount">{money(bill.amount)}</span>
              <Icon name="arrow-right" size={18} />
            </button>
          </li>{/each}
      </ul>
    {/if}
  {/if}
</div>

<style>
  /* The section fills the space beside the rail, up to the same right gutter as the
     Done button, so its header action lines up with it. */
  .bills-editor {
    display: flex;
    flex-direction: column;
    gap: 24px;
    min-height: 0;
  }
  /* Nothing here yet: say so, and offer the one thing to do about it. */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  .empty-title {
    font-size: 20px;
  }
  .empty-state button {
    margin-top: 12px;
  }
  .error {
    color: oklch(78% 0.1 40);
    line-height: 1.4;
  }
  .bill-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  /* Rows echo the Bills widget's cards: outlined, every other one filled. */
  .bill-row {
    width: 100%;
    min-height: 64px;
    padding: 10px 16px;
    justify-content: flex-start;
    gap: 16px;
    text-align: left;
    border-radius: 14px;
  }
  .bill-list li:nth-child(odd) .bill-row {
    background: var(--surface);
  }
  .bill-main {
    flex: 1 1 0;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .bill-name {
    font-size: 18px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .bill-main .quiet {
    font-size: 14px;
  }
  .bill-amount {
    font-size: 18px;
    font-variant-numeric: tabular-nums;
  }
  .bill-row :global(svg) {
    color: var(--muted);
  }

  .bill-form {
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 520px;
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }
  .field > span {
    font-size: 14px;
    color: var(--muted);
  }
  .field input {
    min-height: 52px;
    font-size: 18px;
    border-radius: 12px;
  }
  .pair {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }
  .stepper {
    display: flex;
    align-items: center;
    justify-content: space-between;
    min-height: 52px;
    border: 1px solid var(--line);
    border-radius: 12px;
    background: var(--bg);
  }
  .stepper button {
    min-width: 56px;
    min-height: 50px;
    padding: 0;
    border: 0;
    border-radius: 12px;
    font-size: 24px;
  }
  .stepper output {
    font-size: 20px;
    font-variant-numeric: tabular-nums;
  }
  .note {
    font-size: 14px;
    margin-top: -6px;
  }
  .switch {
    justify-content: flex-start;
    gap: 14px;
    align-self: flex-start;
    padding: 6px 16px 6px 8px;
  }
  .track {
    width: 52px;
    height: 30px;
    border-radius: 15px;
    border: 1px solid var(--line);
    background: var(--bg);
    position: relative;
    transition: background 0.15s;
  }
  .thumb {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: var(--muted);
    transition:
      transform 0.15s,
      background 0.15s;
  }
  .switch[aria-checked='true'] .track {
    background: var(--surface);
    border-color: var(--accent);
  }
  .switch[aria-checked='true'] .thumb {
    transform: translateX(22px);
    background: var(--accent);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 4px;
  }
  .actions button {
    min-width: 96px;
  }
  .danger {
    margin-left: auto;
    color: oklch(78% 0.1 40);
    border-color: oklch(50% 0.08 40);
  }
  @media (max-height: 560px) {
    .bills-editor {
      gap: 14px;
    }
    .bill-form {
      gap: 10px;
    }
    .field input,
    .stepper {
      min-height: 48px;
    }
    .stepper button {
      min-height: 46px;
    }
    .note {
      margin-top: -4px;
    }
    .bill-row {
      min-height: 56px;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .track,
    .thumb {
      transition: none;
    }
  }
</style>
