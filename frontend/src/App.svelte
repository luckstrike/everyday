<script lang="ts">
  import { onMount } from 'svelte';
  import DashboardGrid from './components/DashboardGrid.svelte';
  import { getConfig, getState, getWidgets, simulate } from './lib/api.ts';
  import {
    addWidget,
    defaultLayouts,
    upgradeLayouts,
    parseLayouts,
    boardLayout,
    STORAGE_KEY,
    PINS_KEY,
    parsePins,
    pinnedLayout,
    portraitLayout,
  } from './lib/layout.ts';
  import { billEvents } from './lib/bills.ts';
  import { rankScreens } from './lib/priority.ts';
  import type {
    Config,
    DashboardState,
    Layouts,
    Placement,
    ScreenId,
    WidgetDefinition,
    WidgetId,
  } from './lib/types.ts';
  let pins = $state<WidgetId[]>([]);
  let portrait = $state(false),
    narrow = $state(false);
  let dashboard = $state<DashboardState | null>(null),
    config = $state<Config | null>(null),
    definitions = $state<WidgetDefinition[]>([]);
  let now = $state(Date.now());
  let currentScreen = $state<ScreenId>('home');
  let layouts = $state<Layouts>(defaultLayouts()),
    draft = $state<Layouts>(defaultLayouts());
  let editing = $state(false),
    controls = $state(false),
    devPanel = $state(false),
    busy = $state(false);
  let error = $state(''),
    notice = $state(''),
    lastSuccess = $state(0);
  const ranks = $derived(
    config
      ? rankScreens(
          config.screens,
          definitions,
          [
            ...(dashboard?.events ?? []),
            ...billEvents(dashboard?.finance.bills ?? [], now),
          ],
          now,
        )
      : [],
  );
  const attention = $derived(
    ranks.find((candidate) => candidate.modifier >= 15),
  );
  const projected = $derived(
    pinnedLayout(
      boardLayout(layouts[currentScreen], attention),
      pins,
      attention?.presentation === 'FULLSCREEN',
    ),
  );
  const visible = $derived(
    editing
      ? draft[currentScreen]
      : portrait
        ? portraitLayout(projected, narrow)
        : projected,
  );
  const currentTitle = $derived(
    config?.screens.find((s) => s.id === currentScreen)?.title ?? 'Home',
  );
  const stale = $derived(lastSuccess > 0 && now - lastSuccess > 15000);
  const scenarios = [
    ['normal', 'Normal day'],
    ['rain', 'Rain soon'],
    ['severe', 'Severe weather'],
    ['starting', 'Game starting soon'],
    ['live', 'Live basketball'],
    ['final', 'Recent final score'],
    ['bill', 'Bill due tomorrow'],
    ['financial', 'Financial alert'],
    ['calendar', 'Event in 20 minutes'],
  ];
  async function refresh() {
    try {
      if (!config) {
        const [settings, widgets] = await Promise.all([
          getConfig(),
          getWidgets(),
        ]);
        config = settings;
        definitions = widgets;
        try {
          const saved = localStorage.getItem(STORAGE_KEY);
          if (saved) layouts = upgradeLayouts(parseLayouts(saved, widgets));
        } catch {
          notice = 'Saved layout could not be loaded. Using defaults.';
        }
      }
      dashboard = await getState();
      lastSuccess = Date.now();
      error = '';
    } catch {
      error = 'Cannot reach the dashboard API. Retrying automatically.';
    }
  }
  onMount(() => {
    const resize = () => {
      portrait = window.innerHeight > window.innerWidth;
      narrow = window.innerWidth < 500;
    };
    resize();
    window.addEventListener('resize', resize);
    try {
      const saved = localStorage.getItem(PINS_KEY);
      if (saved) pins = parsePins(saved);
    } catch {
      notice = 'Saved pins could not be loaded.';
    }
    let stopped = false,
      timer: ReturnType<typeof setTimeout>;
    const poll = async () => {
      await refresh();
      if (!stopped) timer = setTimeout(poll, config?.pollIntervalMs ?? 5000);
    };
    void poll();
    const tick = setInterval(() => {
      now = Date.now();
    }, 1000);
    return () => {
      window.removeEventListener('resize', resize);
      stopped = true;
      clearTimeout(timer);
      clearInterval(tick);
    };
  });
  function choose(id: ScreenId) {
    currentScreen = id;
  }

  function togglePin(id: WidgetId) {
    const next = pins.includes(id)
      ? pins.filter((pin) => pin !== id)
      : [...pins, id];
    try {
      localStorage.setItem(PINS_KEY, JSON.stringify(next));
      pins = next;
      notice = '';
    } catch {
      notice = 'Could not save pins. Browser storage is unavailable.';
    }
  }
  function beginEdit() {
    draft = structuredClone($state.snapshot(layouts));
    editing = true;
    notice = '';
  }
  function save() {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(draft));
      layouts = structuredClone($state.snapshot(draft));
      editing = false;
      notice = 'Layout saved on this browser.';
    } catch {
      notice =
        'Could not save. Browser storage is unavailable; your edits are still here.';
    }
  }
  function cancel() {
    editing = false;
    notice = 'Changes discarded.';
  }
  function update(items: Placement[]) {
    draft = { ...draft, [currentScreen]: items };
  }
  function add(id: WidgetId) {
    const definition = definitions.find((d) => d.id === id)!;
    const next = addWidget(draft[currentScreen], definition);
    if (next) {
      update(next);
      notice = '';
    } else
      notice =
        'No room in the 12 × 12 grid. Remove or shrink another widget first.';
  }
  async function scenario(value: string) {
    busy = true;
    try {
      dashboard = await simulate(value);
      lastSuccess = Date.now();
      error = '';
    } catch {
      error =
        'Simulation could not be applied. Check that the backend uses DASHBOARD_DEV=1.';
    } finally {
      busy = false;
    }
  }
  async function fullscreen() {
    try {
      if (document.fullscreenElement) await document.exitFullscreen();
      else {
        await document.documentElement.requestFullscreen();
        controls = false;
        devPanel = false;
        window.scrollTo(0, 0);
      }
    } catch {
      notice = 'Use your browser’s fullscreen command to fill the display.';
    }
  }
</script>

<svelte:head><title>{currentTitle} · Apartment</title></svelte:head>
<main class:display-mode={!editing} class:portrait>
  <header class="display-header">
    <div class="wordmark">
      <span class="brand-symbol" aria-hidden="true">⌂</span> Apartment
      <span class="header-divider">/</span>
      <span class="screen-name">{currentTitle}</span>
    </div>
    <div class="header-status">
      <span class="status-dot" class:offline={stale || !!error}></span><span
        >{stale ? 'Data delayed' : 'Local display'}</span
      ><span class="mock-label">MOCK DATA</span><button
        class="settings-button"
        aria-label="Display controls"
        aria-expanded={controls}
        onclick={() => (controls = !controls)}>•••</button
      >
    </div>
  </header>
  {#if controls || editing}
    <nav class="control-bar" aria-label="Display controls">
      <div class="screen-buttons">
        {#each config?.screens ?? [] as screen}<button
            class:selected={currentScreen === screen.id}
            onclick={() => choose(screen.id)}>{screen.title}</button
          >{/each}
      </div>
      <div class="control-actions">
        {#if !editing}<button onclick={beginEdit} disabled={!dashboard}
            >Edit layout</button
          ><button onclick={fullscreen}>Fullscreen</button
          >{#if import.meta.env.DEV && config?.development}<button
              class:selected={devPanel}
              onclick={() => (devPanel = !devPanel)}>Simulate</button
            >{/if}{:else}<button
            onclick={() => {
              draft = {
                ...draft,
                [currentScreen]: defaultLayouts()[currentScreen],
              };
            }}>Restore this screen</button
          ><button onclick={cancel}>Cancel</button><button
            class="primary"
            onclick={save}>Save layout</button
          >{/if}
      </div>
      {#if !editing}
        <div class="pin-controls">
          <span class="eyebrow">Keep visible</span>
          {#each definitions as widget}<button
              aria-pressed={pins.includes(widget.id)}
              class:selected={pins.includes(widget.id)}
              onclick={() => togglePin(widget.id)}
              >{pins.includes(widget.id)
                ? 'Pinned: '
                : 'Pin '}{widget.title}</button
            >{/each}
          <p>
            Pinned widgets stay across screens. Fullscreen alerts temporarily
            take precedence.
          </p>
        </div>
      {/if}
    </nav>
  {/if}
  {#if editing}<div class="editor-toolbar">
      <span class="eyebrow">Edit layout · {currentTitle}</span><label
        >Add widget <select
          aria-label="Add widget"
          value=""
          onchange={(e) => {
            if (e.currentTarget.value) add(e.currentTarget.value as WidgetId);
            e.currentTarget.value = '';
          }}
          ><option value="">Choose a widget…</option
          >{#each definitions.filter((d) => !draft[currentScreen].some((p) => p.id === d.id)) as widget}<option
              value={widget.id}>{widget.title}</option
            >{/each}</select
        ></label
      >
    </div>{/if}
  {#if error}<p class="connection-message" role="status">
      {error}{dashboard
        ? ' Showing the last received data.'
        : ' Start the Go backend on port 8080.'}
    </p>{/if}
  {#if notice && (controls || editing)}<p class="notice" role="status">
      {notice}
    </p>{/if}
  {#if dashboard}
    {#if !editing && attention}<div
        class="attention-banner"
        class:severe={attention?.presentation === 'FULLSCREEN'}
      >
        <span
          >{attention?.presentation === 'FULLSCREEN'
            ? 'Weather alert'
            : attention?.widget === 'calendar'
              ? 'Coming up soon'
              : attention?.widget === 'finance'
                ? 'A financial reminder'
                : 'Worth a look'}</span
        ><span>Mock data · attention needed</span>
      </div>{/if}
    <div class="board-slot">
      {#key currentScreen}<DashboardGrid
          items={visible}
          {dashboard}
          {now}
          {editing}
          {definitions}
          {attention}
          onchange={update}
        />{/key}
    </div>
  {:else}<div class="loading">
      <p class="eyebrow">Apartment display</p>
      <h1>Getting your day ready.</h1>
      <p class="quiet">Connecting to your local dashboard…</p>
    </div>{/if}
  <footer>
    <div class="screen-indicators" aria-label="Current screen">
      <span class="rotation-label"
        >{editing ? 'Editing layout' : 'Your board'}{pins.length
          ? ` · ${pins.length} pinned`
          : ''}</span
      >
    </div>
    <span
      >{new Date(now).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
      })} <span class="footer-separator">·</span>
      {new Intl.DateTimeFormat('en-US', { timeZoneName: 'short' })
        .formatToParts(now)
        .find((p) => p.type === 'timeZoneName')?.value}</span
    >
  </footer>
  {#if import.meta.env.DEV && devPanel && config?.development}
    <aside class="dev-panel" aria-label="Development simulations">
      <div>
        <p class="eyebrow">Development · Scenario lab</p>
        <p>Events expire after 2 minutes. Highlights stay on this board.</p>
      </div>
      <div class="scenario-buttons">
        {#each scenarios as [value, label]}<button
            disabled={busy}
            class:selected={dashboard?.scenario === value}
            onclick={() => scenario(value)}>{label}</button
          >{/each}
      </div>
      <div class="priority-debug">
        {#each ranks as rank}<span
            >{rank.id} <strong>{Math.round(rank.score)}</strong>
            {rank.presentation.toLowerCase()}</span
          >{/each}
      </div>
    </aside>
  {/if}
</main>
