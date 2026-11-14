<script lang="ts">
  import { onMount } from 'svelte';
  import DashboardGrid from './components/DashboardGrid.svelte';
  import Icon from './components/Icon.svelte';
  import Toast from './components/Toast.svelte';
  import SettingsPanel from './components/SettingsPanel.svelte';
  import { hashFor, parseHash, type Section, type View } from './lib/route.ts';
  import { loadTheme, THEME_KEY, type Theme } from './lib/theme.ts';
  import { getConfig, getState } from './lib/api.ts';
  import type { Config, DashboardState } from './lib/types.ts';
  let dashboard = $state<DashboardState | null>(null),
    config = $state<Config | null>(null);
  let now = $state(Date.now()),
    lastSuccess = $state(0);
  let error = $state(''),
    notice = $state('');
  let noticeToken = $state(0);
  let theme = $state<Theme>(loadTheme());
  // The dashboard is home. Settings opens over it, and keeps polling underneath.
  let view = $state<View>(parseHash(location.hash));
  let settingsButton = $state<HTMLButtonElement>();
  function navigate(next: View) {
    if (next.name === 'dashboard') {
      history.pushState(null, '', location.pathname + location.search);
    } else {
      location.hash = hashFor(next);
    }
    view = next;
  }
  function closeSettings() {
    navigate({ name: 'dashboard' });
    settingsButton?.focus();
  }
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
  const stale = $derived(lastSuccess > 0 && now - lastSuccess > 15000);
  async function refresh() {
    try {
      if (!config) config = await getConfig();
      dashboard = await getState();
      lastSuccess = Date.now();
      error = '';
    } catch {
      error = 'Connection interrupted. Retrying automatically.';
    }
  }
  onMount(() => {
    let stopped = false,
      timer: ReturnType<typeof setTimeout>;
    const poll = async () => {
      await refresh();
      if (!stopped) timer = setTimeout(poll, config?.pollIntervalMs ?? 5000);
    };
    void poll();
    const clock = setInterval(() => (now = Date.now()), 1000);
    const followAddress = () => (view = parseHash(location.hash));
    window.addEventListener('hashchange', followAddress);
    return () => {
      window.removeEventListener('hashchange', followAddress);
      stopped = true;
      clearTimeout(timer);
      clearInterval(clock);
    };
  });
</script>

<svelte:head><title>Everyday</title></svelte:head>
<main>
  <header class="display-header">
    <div class="header-tools">
      {#if dashboard}<span class="header-location"
          ><Icon name="pin" size={16} />{dashboard.weather.location}</span
        >{/if}<button
        bind:this={settingsButton}
        class="icon-button settings-trigger"
        aria-label="Settings"
        onclick={() => navigate({ name: 'settings', section: 'bills' })}
        ><Icon name="settings" size={20} /></button
      >
    </div>
  </header>
  <div class="board-slot">
    {#if dashboard}<DashboardGrid {dashboard} {now} />
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
          : dashboard && Object.values(dashboard.sources).includes('mock')
            ? 'Mock data'
            : 'Local display'}</span
    >
  </footer>
  {#if view.name === 'settings'}<SettingsPanel
      section={view.section}
      {theme}
      onsection={(section: Section) => navigate({ name: 'settings', section })}
      onclose={closeSettings}
      onchoosetheme={chooseTheme}
      onbillschanged={refresh}
    />{/if}
  <Toast message={notice} token={noticeToken} ondismiss={() => (notice = '')} />
</main>
