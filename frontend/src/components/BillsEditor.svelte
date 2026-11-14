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
    groupByDay,
    ordinal,
    toEntry,
    validateDraft,
    withPaid,
    type BillDraft,
    type BillEntry,
    type ListedBill,
  } from '../lib/bills-editor.ts';
  import { shortDate } from '../lib/format.ts';
  import { amountText } from '../lib/bills.ts';

  // onchanged runs after bills are saved, so the dashboard can show them at once.
  let { onchanged }: { onchanged: () => void } = $props();

  let loading = $state(true);
  // Why the bills cannot be edited at all, such as sample data being shown.
  let unavailable = $state('');
  let bills = $state<ListedBill[]>([]);
  // Set when the bills file exists but cannot be read; saving here replaces it.
  let problem = $state('');

  let editing = $state(false);
  let editIndex = $state<number | null>(null); // null while adding a new bill
  let draft = $state<BillDraft>(emptyDraft());
  let formError = $state('');
  let saving = $state(false);
  let confirmDelete = $state(false);
  let nameInput = $state<HTMLInputElement>();

  function apply(response: { bills: ListedBill[]; problem?: string }) {
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
    // An edit keeps the bill's paid mark, so fixing a typo does not un-pay it.
    const entry = entryFrom(
      draft,
      editIndex === null ? undefined : bills[editIndex],
    );
    const saved = bills.map(toEntry);
    void persist(
      editIndex === null
        ? [...saved, entry]
        : saved.map((bill, i) => (i === editIndex ? entry : bill)),
    );
  }

  // Ticking a bill saves at once: it is a quick yes or no, not something to fill in.
  function setPaid(index: number, paid: boolean) {
    void persist(
      bills.map((bill, i) =>
        i === index ? withPaid(bill, paid) : toEntry(bill),
      ),
    );
  }

  function remove() {
    // A second tap confirms, so one stray touch cannot delete a bill.
    if (!confirmDelete) {
      confirmDelete = true;
      return;
    }
    void persist(bills.filter((_, i) => i !== editIndex).map(toEntry));
  }

  // The Add button lives in the header once there are bills; before that the empty
  // state offers it, so the screen never shows two.
  const showAdd = $derived(
    !loading && !unavailable && !editing && bills.length > 0,
  );

  // Shown by the day of the month each falls due. Edits still go by position in the
  // saved list, which groupByDay carries along with each bill.
  const groups = $derived(groupByDay(bills));

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
            placeholder={draft.varies ? 'Shows as Due' : '0.00'}
            disabled={draft.varies}
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
      <button
        type="button"
        class="varies-toggle"
        role="switch"
        aria-checked={draft.varies}
        onclick={() => (draft.varies = !draft.varies)}
      >
        <span class="tick"><Icon name="check" size={14} /></span>
        <span>Amount varies, like a credit card</span>
      </button>
      <p class="quiet note">{dueNote(draft.day)}</p>
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
      <div class="day-groups">
        {#each groups as group (group.day)}<section
            class="day-group"
            class:shared={group.bills.length > 1}
            aria-label={`Bills due on day ${group.day}`}
          >
            <h4 class="day-label">{ordinal(group.day)}</h4>
            <ul class="bill-list">
              {#each group.bills as { bill, index, position } (index)}<li
                  class="bill-card"
                  class:filled={position % 2 === 0}
                  class:paid={bill.paid}
                  class:overdue={bill.overdue}
                >
                  <button
                    class="paid-toggle"
                    role="switch"
                    aria-checked={bill.paid}
                    aria-label={`${bill.name}: ${bill.paid ? 'paid' : bill.overdue ? 'overdue' : 'not paid'}`}
                    disabled={saving}
                    onclick={() => setPaid(index, !bill.paid)}
                  >
                    <span class="tick"><Icon name="check" size={14} /></span>
                  </button>
                  <button class="bill-row" onclick={() => startEdit(index)}>
                    <span class="bill-main">
                      <span class="bill-name">{bill.name}</span>
                      <span class="bill-status"
                        >{#if bill.paid}<span class="paid-word">Paid</span> · {shortDate(
                            bill.due,
                          )}{:else if bill.overdue}<span class="overdue-word"
                            ><Icon name="alert" size={14} />Overdue</span
                          >
                          · was due {shortDate(bill.due)}{:else}Due {shortDate(
                            bill.due,
                          )}{/if}</span
                      >
                    </span>
                    <span class="bill-amount"
                      >{amountText(bill.amount ?? 0, bill.paid)}</span
                    >
                    <Icon name="arrow-right" size={18} />
                  </button>
                </li>{/each}
            </ul>
          </section>{/each}
      </div>
    {/if}
  {/if}
</div>

<style>
  /* The section fills the space beside the rail, up to the same right gutter as the
     Done button, so its header action lines up with it. */
  .bills-editor {
    /* One height for a bill row and for the date beside it, so the date is centred
       on its first bill however the text inside wraps; one gap between bills. */
    --bill-gap: 12px;
    --bill-row-height: 66px;
    /* The date column and the gap after it. */
    --date-col: 3rem;
    --date-gap: 16px;
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
    color: var(--warn);
    line-height: 1.4;
  }
  /* Bills sit under the day of the month they fall due: the date down the left, the
     bills due that day beside it. Every bill is the same distance from the next,
     inside a group or between groups, so the list reads as one rhythm and the dates
     only label it. */
  .day-groups {
    display: flex;
    flex-direction: column;
    gap: var(--bill-gap);
  }
  .day-group {
    position: relative;
    display: grid;
    grid-template-columns: var(--date-col) minmax(0, 1fr);
    gap: var(--date-gap);
    align-items: start;
  }
  /* Bills that share a date are tied to it by a dotted line that drops out of the
     bottom of the date, centred under it, and runs down to the last bill. The first
     bill is already beside the date, so only the bills after it get a dot, each level
     with the middle of its card. */
  .day-group.shared {
    --line-x: calc(var(--date-col) / 2);
    --line-color: color-mix(in oklab, var(--muted) 55%, transparent);
  }
  .day-group.shared::before {
    content: '';
    position: absolute;
    left: calc(var(--line-x) - 1px);
    /* Starts a little below the date's text, which is centred on the first row. */
    top: calc(var(--bill-row-height) / 2 + 18px);
    bottom: calc(var(--bill-row-height) / 2);
    border-left: 2px dotted var(--line-color);
  }
  .day-group.shared .bill-card:not(:first-child)::before {
    content: '';
    position: absolute;
    /* Measured from just inside the card's 1px border, which is a date column and a
       gap in from the group's edge. */
    left: calc(var(--line-x) - var(--date-col) - var(--date-gap) - 1px - 4px);
    top: calc(50% - 4px);
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--muted);
  }
  .day-label {
    margin: 0;
    height: var(--bill-row-height);
    display: flex;
    align-items: center;
    /* Centred, so the dotted line drops from the middle of the date's bottom. */
    justify-content: center;
    font-size: 18px;
    font-weight: 400;
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }
  .bill-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--bill-gap);
  }
  /* A bill is a card like the Bills widget's: outlined, every other one filled. It
     holds the paid toggle and, beside it, the row that opens the bill for editing. */
  .bill-card {
    position: relative;
    display: flex;
    align-items: center;
    gap: 6px;
    height: var(--bill-row-height);
    padding: 0 6px 0 10px;
    border: 1px solid var(--line);
    border-radius: 14px;
  }
  /* Every other bill in date order is filled, across the groups. */
  .bill-card.filled {
    background: var(--surface);
  }
  .bill-row {
    flex: 1 1 0;
    min-width: 0;
    height: 100%;
    min-height: 0;
    padding: 0 10px;
    justify-content: flex-start;
    gap: 16px;
    text-align: left;
    border: 0;
    border-radius: 10px;
  }
  /* The open area stays transparent on hover and focus. The card paints the highlight
     instead, so it fills the whole row, edge to edge, with the row's own corners: a
     highlight on the inner button would stop short of the card's edge and have
     different corners. */
  .bill-row:hover:not(:disabled) {
    background: transparent;
  }
  .bill-card {
    transition: background 0.12s;
  }
  /* Hover only where there is a pointer to hover with. On a touch screen a tap would
     otherwise leave the row stuck highlighted. */
  @media (hover: hover) {
    .bill-card:has(.bill-row:hover) {
      background: var(--surface);
    }
    .bill-card.filled:has(.bill-row:hover) {
      background: color-mix(in oklch, var(--surface), var(--text) 7%);
    }
  }
  /* A keyboard focus ring goes around the whole row, not the button inside it. It is
     drawn just inside the card: the list scrolls, and anything drawn outside the
     card's edge would be clipped there. */
  .bill-row:focus-visible {
    outline: none;
  }
  .bill-card:has(.bill-row:focus-visible) {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }
  @media (prefers-reduced-motion: reduce) {
    .bill-card {
      transition: none;
    }
  }
  /* The tick: an empty ring until the bill is paid, then filled with the theme's
     accent. The accent already means "good" throughout, and a tick shape means it
     without relying on colour alone. */
  /* The button keeps a full 44px to tap, as a touch target should; only the ring
     drawn inside it is small. */
  .paid-toggle {
    flex-shrink: 0;
    width: 44px;
    height: 44px;
    min-width: 44px;
    padding: 0;
    border: 0;
    border-radius: 50%;
    background: transparent;
  }
  .paid-toggle:hover:not(:disabled) {
    background: transparent;
  }
  .tick {
    width: 24px;
    height: 24px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1.5px solid var(--muted);
    border-radius: 50%;
    color: transparent;
    transition:
      background 0.15s,
      border-color 0.15s,
      color 0.15s;
  }
  .paid-toggle:hover:not(:disabled) .tick {
    color: var(--muted);
  }
  .paid-toggle[aria-checked='true'] .tick,
  .varies-toggle[aria-checked='true'] .tick {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--bg);
  }
  .paid-toggle:focus-visible {
    outline-offset: 0;
  }
  @media (prefers-reduced-motion: reduce) {
    .tick {
      transition: none;
    }
  }
  /* A paid bill recedes: its name and amount go quiet, since there is nothing left
     to do about it, and only the "Paid" word keeps the accent. */
  .bill-card.paid .bill-name,
  .bill-card.paid .bill-amount {
    color: var(--muted);
  }
  .bill-status {
    font-size: 14px;
    color: var(--muted);
  }
  .paid-word {
    color: var(--accent);
  }
  /* A missed bill says so in words and an icon as well as the warm colour, and its
     tick ring is drawn in that colour, since the tick is what clears it. */
  .overdue-word {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--warn);
  }
  .bill-card.overdue .tick {
    border-color: var(--warn);
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
  /* A switch in the form, drawn with the same ring as the paid tick. */
  .varies-toggle {
    justify-content: flex-start;
    gap: 10px;
    min-height: 44px;
    padding: 0 12px 0 10px;
    border-radius: 12px;
    align-self: flex-start;
    text-align: left;
  }
  .varies-toggle:hover .tick {
    color: var(--muted);
  }
  .varies-toggle[aria-checked='true']:hover .tick {
    color: var(--bg);
  }
  .note {
    font-size: 14px;
    margin-top: -6px;
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
    color: var(--warn);
    border-color: var(--warn-line);
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
    .bills-editor {
      --bill-gap: 8px;
      --bill-row-height: 58px;
    }
  }
  /* Too narrow for a date column: the date heads its bills instead. */
  @media (max-width: 640px) {
    /* With the date above its bills, a group needs air before the next date, or the
       date would look like it belongs to the bills above it. */
    .day-groups {
      gap: calc(var(--bill-gap) + 8px);
    }
    /* The date is above its bills here, so there is no gap to run a line down. */
    .day-group.shared::before,
    .day-group.shared .bill-card::before {
      display: none;
    }
    .day-group {
      grid-template-columns: minmax(0, 1fr);
      gap: 4px;
    }
    .day-label {
      height: auto;
      justify-content: flex-start;
      font-size: 16px;
    }
  }
</style>
