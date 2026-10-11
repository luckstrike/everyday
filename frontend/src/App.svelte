<script lang="ts">
  import { onMount, tick } from 'svelte';
  import DashboardGrid from './components/DashboardGrid.svelte';
  import WidgetPicker from './components/WidgetPicker.svelte';
  import Icon from './components/Icon.svelte';
  import Toast from './components/Toast.svelte';
  import DisplaySettings from './components/DisplaySettings.svelte';
  import DiscardLayoutDialog from './components/DiscardLayoutDialog.svelte';
  import { loadTheme, THEME_KEY, type Theme } from './lib/theme.ts';
  import { catalog } from './lib/catalog.ts';
  import { getConfig, getState, getWidgets, simulate } from './lib/api.ts';
  import {
    BOARDS_KEY,
    initialBoards,
    migrateLayouts,
    parseBoards,
    validPositions,
    footprint,
    layoutsEqual,
    type BoardStore,
    type Orientation,
  } from './lib/boards.ts';
  import { billEvents } from './lib/bills.ts';
  import { rankWidgets } from './lib/priority.ts';
  import {
    initialPresentation,
    updatePresentation,
    attentionFor,
    displayLayout,
  } from './lib/presentation.ts';
  import type {
    Config,
    DashboardState,
    Placement,
    Size,
    WidgetDefinition,
    WidgetId,
  } from './lib/types.ts';
  let store = $state<BoardStore>(initialBoards());
  let orientation = $state<Orientation>('landscape');
  let dashboard = $state<DashboardState | null>(null),
    config = $state<Config | null>(null),
    definitions = $state<WidgetDefinition[]>([]);
  let now = $state(Date.now()),
    lastSuccess = $state(0);
  let editing = $state(false),
    picker = $state(false),
    boardMenu = $state(false),
    creating = $state(false),
    devPanel = $state(false),
    busy = $state(false);
  let draft = $state<Placement[]>([]),
    editOrientation = $state<Orientation>('landscape');
  let editEntry = $state<'add' | 'layout'>('layout');
  let discardOpen = $state(false);
  let pending = $state<{
    id: WidgetId;
    size: Size;
    w: number;
    h: number;
    existing: boolean;
  } | null>(null);
  let name = $state(''),
    error = $state(''),
    notice = $state('');
  let noticeToken = $state(0);
  let theme = $state<Theme>(loadTheme());
  $effect(() => {
    document.documentElement.dataset.theme = theme;
  });
  function notify(message: string) {
    notice = message;
    noticeToken++;
  }
  function chooseTheme(value: Theme) {
    theme = value;
    try {
      localStorage.setItem(THEME_KEY, value);
    } catch {
      notify('Theme applied for this session. Browser storage is unavailable.');
    }
  }
  const board = $derived(store.boards.find((b) => b.id === store.activeId)!);
  const saved = $derived(board.layouts[orientation]);
  const dirty = $derived(
    editing && !layoutsEqual(draft, board.layouts[editOrientation]),
  );
  const ranks = $derived(
    rankWidgets(
      definitions,
      [
        ...(dashboard?.events ?? []),
        ...billEvents(dashboard?.finance.bills ?? [], now),
      ],
      now,
    ),
  );
  let presentation = $state(initialPresentation());
  $effect(() => {
    const next = updatePresentation(presentation, ranks, now);
    if (next !== presentation) presentation = next;
  });
  const attention = $derived(attentionFor(presentation, ranks, saved));
  const visible = $derived(editing ? draft : displayLayout(saved, attention));
  const available = $derived(
    pending
      ? validPositions(
          pending.id,
          pending.size,
          editOrientation,
          draft.filter((p) => p.id !== pending?.id),
          definitions,
          { w: pending.w, h: pending.h },
        )
      : [],
  );
  const fits = $derived.by(() => {
    const result = { small: false, medium: false, large: false };
    if (pending)
      for (const size of ['small', 'medium', 'large'] as Size[])
        result[size] =
          validPositions(
            pending.id,
            size,
            editOrientation,
            draft.filter((p) => p.id !== pending?.id),
            definitions,
            size === pending.size
              ? { w: pending.w, h: pending.h }
              : footprint(pending.id, size, editOrientation),
          ).length > 0;
    return result;
  });
  const stale = $derived(lastSuccess > 0 && now - lastSuccess > 15000);
  const scenarios = [
    ['normal', 'Normal day'],
    ['rain', 'Rain soon'],
    ['severe', 'Severe weather'],
    ['starting', 'Game starting soon'],
    ['live', 'Live basketball'],
    ['final', 'Final score'],
    ['bill', 'Bill due tomorrow'],
    ['financial', 'Bill alert'],
    ['calendar', 'Event in 20 minutes'],
  ];
  async function refresh() {
    try {
      if (!config) {
        const [settings, widgets] = await Promise.all([
          getConfig(),
          getWidgets(),
        ]);
        definitions = widgets;
        config = settings;
        try {
          const raw = localStorage.getItem(BOARDS_KEY),
            legacy = localStorage.getItem('apartment-display.layouts.v1');
          if (raw) store = parseBoards(raw, widgets);
          else if (legacy) store = migrateLayouts(legacy, widgets);
        } catch {
          notify(
            'Saved boards could not be loaded. Using Everyday; your saved data has been retained.',
          );
        }
      }
      dashboard = await getState();
      lastSuccess = Date.now();
      error = '';
    } catch {
      error = 'Connection interrupted. Retrying automatically.';
    }
  }
  onMount(() => {
    const resize = () => {
      orientation =
        window.innerHeight > window.innerWidth ? 'portrait' : 'landscape';
    };
    resize();
    window.addEventListener('resize', resize);
    let stopped = false,
      timer: ReturnType<typeof setTimeout>;
    const poll = async () => {
      await refresh();
      if (!stopped) timer = setTimeout(poll, config?.pollIntervalMs ?? 5000);
    };
    void poll();
    const clock = setInterval(() => (now = Date.now()), 1000);
    const leave = (event: BeforeUnloadEvent) => {
      if (dirty) {
        event.preventDefault();
        event.returnValue = '';
      }
    };
    window.addEventListener('beforeunload', leave);
    return () => {
      stopped = true;
      clearTimeout(timer);
      clearInterval(clock);
      window.removeEventListener('resize', resize);
      window.removeEventListener('beforeunload', leave);
    };
  });
  function persist(next: BoardStore): boolean {
    try {
      localStorage.setItem(BOARDS_KEY, JSON.stringify(next));
      store = next;
      notice = '';
      return true;
    } catch {
      notify(
        'Could not save. Browser storage is unavailable; your edits are still here.',
      );
      return false;
    }
  }
  function beginEdit(showPicker = true) {
    if (!editing) {
      draft = saved.map((p) => ({ ...p }));
      editOrientation = orientation;
      editEntry = showPicker ? 'add' : 'layout';
      editing = true;
    }
    boardMenu = false;
    devPanel = false;
    picker = showPicker;
    pending = null;
    notice = '';
    if (showPicker) void focusPicker();
  }
  async function focusPicker() {
    await tick();
    (document.querySelector('.widget-picker') as HTMLElement | null)?.focus();
  }
  function closePicker() {
    if (editing && editEntry === 'add' && !dirty) {
      finishEdit();
      return;
    }
    picker = false;
    pending = null;
    void tick().then(() =>
      (document.querySelector('.add-widget') as HTMLButtonElement)?.focus(),
    );
  }
  function cancel() {
    if (dirty) discardOpen = true;
    else finishEdit();
  }
  function finishEdit() {
    const returnSelector = editEntry === 'add' ? '.add-widget' : '.edit-layout';
    discardOpen = false;
    editing = false;
    picker = false;
    pending = null;
    draft = [];
    notice = '';
    void tick().then(() =>
      (
        document.querySelector(returnSelector) as HTMLButtonElement | null
      )?.focus(),
    );
  }
  function save() {
    const next: BoardStore = {
      ...store,
      boards: store.boards.map((b) =>
        b.id === board.id
          ? {
              ...b,
              layouts: {
                ...b.layouts,
                [editOrientation]: draft.map((p) => ({ ...p })),
              },
            }
          : b,
      ),
    };
    if (persist(next)) {
      finishEdit();
      notify('Layout saved.');
    }
  }
  function chooseWidget(id: WidgetId) {
    pending = {
      id,
      size: 'small',
      ...footprint(id, 'small', editOrientation),
      existing: false,
    };
    void focusPicker();
  }
  function editWidget(item: Placement) {
    pending = {
      ...item,
      size:
        item.w >= 14 && item.h >= 12
          ? 'large'
          : item.w >= 10 && item.h >= 8
            ? 'medium'
            : 'small',
      existing: true,
    };
    picker = true;
    void focusPicker();
  }
  function sizeWidget(size: Size) {
    if (pending)
      pending = {
        ...pending,
        size,
        ...footprint(pending.id, size, editOrientation),
      };
  }
  function place(item: Placement) {
    const wasExisting = pending?.existing;
    draft = [...draft.filter((p) => p.id !== item.id), { ...item }];
    pending = null;
    picker = false;
    notify(
      `${catalog.find((w) => w.id === item.id)?.title ?? 'Widget'} ${wasExisting ? 'updated' : 'added to layout'}. Save to keep your changes.`,
    );
    void tick().then(() =>
      (
        document.querySelector(
          `[data-widget-id="${item.id}"]`,
        ) as HTMLButtonElement | null
      )?.focus(),
    );
  }
  function chooseBoard(id: string) {
    if (!editing && persist({ ...store, activeId: id })) {
      boardMenu = false;
      creating = false;
    }
  }
  function createBoard(event: SubmitEvent) {
    event.preventDefault();
    const value = name.trim();
    if (!value || store.boards.length >= 50) return;
    const next = {
      id: crypto.randomUUID(),
      name: value,
      layouts: { landscape: [], portrait: [] },
    };
    if (
      persist({ ...store, activeId: next.id, boards: [...store.boards, next] })
    ) {
      name = '';
      creating = false;
      beginEdit();
    }
  }
  async function scenario(value: string) {
    busy = true;
    try {
      dashboard = await simulate(value);
      lastSuccess = Date.now();
      error = '';
    } catch {
      notify('Simulation unavailable. Start the backend with DASHBOARD_DEV=1.');
    } finally {
      busy = false;
    }
  }
  async function fullscreen() {
    try {
      if (document.fullscreenElement) await document.exitFullscreen();
      else await document.documentElement.requestFullscreen();
      boardMenu = false;
    } catch {
      notify('Use your browser’s fullscreen command to fill the display.');
    }
  }
  function escape(event: KeyboardEvent) {
    if (event.key !== 'Escape') return;
    if (discardOpen) return;
    if (pending) pending = null;
    else if (picker) closePicker();
    else {
      boardMenu = false;
      devPanel = false;
    }
  }
</script>

<svelte:head><title>{board.name} · Apartment</title></svelte:head>
<svelte:window
  onkeydown={escape}
  onclick={(e) => {
    if (
      !e
        .composedPath()
        .some(
          (node) =>
            node instanceof Element &&
            node.classList.contains('board-selector'),
        )
    ) {
      boardMenu = false;
      creating = false;
    }
  }}
/>
<main
  class:portrait={editing
    ? editOrientation === 'portrait'
    : orientation === 'portrait'}
  class:editing
>
  <header class="display-header">
    {#if editing}<div class="editor-title">
        <h1>Edit {board.name}</h1>
        <span class="quiet"
          >{pending
            ? 'Tap a highlighted position'
            : picker
              ? 'Choose a widget to preview its position'
              : 'Tap a handle to move or resize a widget'}<small
            >{editOrientation} layout</small
          ></span
        >
      </div>
      <div class="editor-actions">
        <button onclick={cancel}>Cancel</button><button
          class="primary"
          onclick={save}>Save layout</button
        >
      </div>
    {:else}<div class="board-selector">
        <button
          class="board-trigger"
          aria-expanded={boardMenu}
          aria-controls="board-menu"
          onclick={() => {
            boardMenu = !boardMenu;
            devPanel = false;
          }}>{board.name}<Icon name="down" size={16} /></button
        >
        {#if boardMenu}<div id="board-menu" class="board-menu popover">
            <p class="eyebrow">Your boards</p>
            <div class="board-list">
              {#each store.boards as b}<button
                  class:current={board.id === b.id}
                  onclick={() => chooseBoard(b.id)}
                  ><span
                    >{b.name}<small
                      >{b.layouts[orientation].length} widgets</small
                    ></span
                  >{#if board.id === b.id}<Icon
                      name="check"
                      size={18}
                    />{/if}</button
                >{/each}
            </div>
            {#if creating}<form onsubmit={createBoard}>
                <label for="board-name">Board name</label><input
                  id="board-name"
                  bind:value={name}
                  maxlength="40"
                  required
                  placeholder="e.g. A quiet morning"
                /><button class="primary" type="submit" disabled={!name.trim()}
                  >Create board</button
                >
              </form>{:else}<button
                disabled={store.boards.length >= 50}
                onclick={() => {
                  creating = true;
                  void tick().then(() =>
                    document.getElementById('board-name')?.focus(),
                  );
                }}><Icon name="plus" size={18} />Create board</button
              >{/if}
            <div class="board-menu-tools">
              <button onclick={() => beginEdit(false)}>Edit layout</button
              ><button onclick={fullscreen}>Fullscreen</button
              >{#if import.meta.env.DEV && config?.development}<button
                  onclick={() => {
                    devPanel = true;
                    boardMenu = false;
                  }}>Simulate</button
                >{/if}
            </div>
            <p class="menu-footnote">
              Boards switch only when you choose.<br />Each board fits on one
              screen.
            </p>
          </div>{/if}
      </div>
      <div class="header-tools">
        <span class="header-location"
          ><Icon name="map" size={16} />{dashboard?.weather.location ??
            'Dallas, TX'}</span
        ><DisplaySettings {theme} onchange={chooseTheme} />
      </div>{/if}
  </header>
  <div class="board-slot">
    {#if dashboard}{#key `${board.id}:${editing ? editOrientation : orientation}`}<DashboardGrid
          items={visible}
          {dashboard}
          {now}
          {editing}
          {definitions}
          {attention}
          previews={available}
          pendingId={pending?.id ?? null}
          onchange={(items) => {
            draft = items;
            pending = null;
          }}
          onplace={place}
          onselect={editWidget}
        />{/key}
    {:else}<div class="loading">
        <h1>Getting your day ready.</h1>
        <p class="quiet">{error || 'Connecting to your local dashboard…'}</p>
      </div>{/if}
  </div>
  <footer>
    <span
      >{error
        ? `${error}${dashboard ? ' Showing last received data.' : ''}`
        : stale
          ? 'Data delayed · showing last received data'
          : dashboard?.mock
            ? 'Mock data'
            : 'Local display'}</span
    >
    <div class="layout-tools">
      {#if !editing}<button
          class="edit-layout"
          disabled={!dashboard}
          onclick={() => beginEdit(false)}
          ><Icon name="edit" size={16} />Edit layout</button
        >{/if}
      <button
        class="add-widget"
        class:picker-open={picker}
        disabled={!dashboard}
        onclick={() => (picker ? closePicker() : beginEdit())}
        ><Icon name={picker ? 'close' : 'plus'} size={16} />{picker
          ? 'Close picker'
          : 'Add widget'}</button
      >
    </div>
  </footer>
  <Toast message={notice} token={noticeToken} ondismiss={() => (notice = '')} />
  {#if discardOpen}<DiscardLayoutDialog
      onkeep={() => (discardOpen = false)}
      ondiscard={finishEdit}
    />{/if}
  {#if picker}<WidgetPicker
      items={draft}
      selected={pending?.id ?? null}
      size={pending?.size ?? 'small'}
      existing={pending?.existing ?? false}
      {fits}
      onchoose={chooseWidget}
      onsize={sizeWidget}
      onclose={closePicker}
      onback={() => (pending = null)}
      onremove={() => {
        if (pending) {
          notify(
            `${catalog.find((w) => w.id === pending?.id)?.title ?? 'Widget'} removed. Save to keep your changes.`,
          );
          draft = draft.filter((p) => p.id !== pending?.id);
        }
        pending = null;
        picker = false;
      }}
    />{/if}
  {#if devPanel}<aside
      class="dev-panel popover"
      aria-label="Development simulations"
    >
      <div class="popover-heading">
        <h2>Scenario lab</h2>
        <button
          class="icon-button"
          aria-label="Close simulations"
          onclick={() => (devPanel = false)}><Icon name="close" /></button
        >
      </div>
      <p class="quiet">Mock events expire after 2 minutes. Boards stay put.</p>
      <div class="scenario-buttons">
        {#each scenarios as [value, label]}<button
            disabled={busy}
            class:selected={dashboard?.scenario === value}
            onclick={() => scenario(value)}>{label}</button
          >{/each}
      </div>
    </aside>{/if}
</main>
