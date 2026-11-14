<script lang="ts">
  import { onMount } from 'svelte';
  import SettingsHeader from './SettingsHeader.svelte';
  import {
    ApiError,
    getLocation,
    saveLocation,
    searchLocations,
    type LocationResponse,
    type WeatherLocation,
  } from '../lib/api.ts';

  let { onchanged }: { onchanged: (location: WeatherLocation) => void } =
    $props();
  let settings = $state<LocationResponse | null>(null);
  let loading = $state(true);
  let loadError = $state('');
  let query = $state('');
  let searchedFor = $state('');
  let results = $state<WeatherLocation[]>([]);
  let selected = $state<number | undefined>();
  let searching = $state(false);
  let saving = $state(false);
  let searchError = $state('');
  let saveError = $state('');
  let saved = $state('');
  let searchVersion = 0;
  const candidate = $derived(
    selected === undefined ? undefined : results[selected],
  );

  function errorMessage(error: unknown, fallback: string) {
    return error instanceof ApiError ? error.message : fallback;
  }

  async function load() {
    loading = true;
    loadError = '';
    try {
      settings = await getLocation();
    } catch (error) {
      loadError = errorMessage(
        error,
        'Could not load your location. Check that the dashboard is running.',
      );
    } finally {
      loading = false;
    }
  }
  onMount(() => {
    void load();
    return () => {
      searchVersion++;
    };
  });

  function clearSearch() {
    searchVersion++;
    searching = false;
    results = [];
    selected = undefined;
    searchedFor = '';
    searchError = '';
    saveError = '';
    saved = '';
  }

  async function search(event: SubmitEvent) {
    event.preventDefault();
    const term = query.trim();
    if (term.length < 2 || saving) return;
    clearSearch();
    const version = ++searchVersion;
    searching = true;
    try {
      const response = await searchLocations(term);
      if (version !== searchVersion) return;
      results = response.locations;
      searchedFor = term;
    } catch (error) {
      if (version === searchVersion) {
        searchError = errorMessage(
          error,
          'City search is unavailable. Please try again.',
        );
      }
    } finally {
      if (version === searchVersion) searching = false;
    }
  }

  async function save() {
    if (!candidate || saving) return;
    saving = true;
    saveError = '';
    saved = '';
    try {
      settings = await saveLocation(candidate);
      onchanged(settings.location);
      saved = `Saved ${settings.location.name}. Weather is refreshing.`;
      results = [];
      selected = undefined;
      searchedFor = '';
      query = '';
    } catch (error) {
      saveError = errorMessage(
        error,
        'Could not confirm the save. Reload this section to check your location before trying again.',
      );
    } finally {
      saving = false;
    }
  }
</script>

<div class="location-settings">
  <SettingsHeader
    title="Location"
    description="Choose the city for your weather forecast."
  />
  {#if loading}
    <p class="quiet" role="status">Loading your location…</p>
  {:else if loadError}
    <p role="alert">{loadError}</p>
    <div><button onclick={load}>Try again</button></div>
  {:else if settings}
    <div class="current-location">
      <span class="quiet">Weather location</span>
      <p>{settings.location.name}</p>
      <p class="quiet">
        This changes weather only. The clock and calendar keep this display’s
        timezone.
      </p>
    </div>
    {#if !settings.editable}
      <p role="status">
        {settings.reason || 'Location editing is unavailable.'}
      </p>
    {:else}
      <form onsubmit={search}>
        <label for="city-search">City or postal code</label>
        <div class="search-row">
          <input
            id="city-search"
            type="search"
            bind:value={query}
            oninput={clearSearch}
            placeholder="e.g. Austin, TX"
            minlength="2"
            maxlength="100"
            required
            autocomplete="off"
            disabled={saving}
            aria-describedby="search-hint"
          />
          <button
            type="submit"
            disabled={searching || saving || query.trim().length < 2}
          >
            {searching ? 'Searching…' : 'Search'}
          </button>
        </div>
        <p id="search-hint" class="quiet">
          Include a state or country to narrow your search.
        </p>
      </form>
      <div aria-live="polite" role="status">
        {#if searching}<p class="quiet">Searching for “{query.trim()}”…</p>
        {:else if searchedFor && results.length === 0}<p>
            No locations found for “{searchedFor}”. Try a nearby city or add a
            state or country.
          </p>{/if}
      </div>
      {#if searchError}<p role="alert">{searchError}</p>{/if}
      {#if results.length > 0}
        <fieldset disabled={saving}>
          <legend>Choose a location</legend>
          <div class="results">
            {#each results as result, index}
              <label class:selected={selected === index}>
                <input
                  type="radio"
                  name="weather-location"
                  value={index}
                  bind:group={selected}
                  onchange={() => {
                    saveError = '';
                    saved = '';
                  }}
                />
                <span>{result.name}</span>
              </label>
            {/each}
          </div>
        </fieldset>
        <div class="save-row">
          <button
            class="primary"
            disabled={!candidate || saving}
            onclick={save}
          >
            {saving ? 'Saving…' : 'Save location'}
          </button>
          <p class="quiet">
            {candidate ? `Use ${candidate.name}` : 'Select a result to save.'}
          </p>
        </div>
      {/if}
      {#if saveError}<p role="alert">{saveError}</p>{/if}
      {#if saved}<p role="status">{saved}</p>{/if}
      <p class="attribution quiet">
        City search: <a
          href="https://open-meteo.com/en/docs/geocoding-api"
          target="_blank"
          rel="noreferrer">Open-Meteo</a
        >
        · Location data:
        <a href="https://www.geonames.org/" target="_blank" rel="noreferrer"
          >GeoNames</a
        >
      </p>
    {/if}
  {/if}
</div>

<style>
  .location-settings {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  .current-location {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .current-location > p:first-of-type {
    font-size: 24px;
    overflow-wrap: anywhere;
  }
  p {
    line-height: 1.5;
  }
  form {
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-width: 760px;
  }
  .search-row {
    display: flex;
    align-items: stretch;
    gap: 12px;
  }
  .search-row input {
    min-width: 0;
    flex: 1;
    min-height: 56px;
  }
  .search-row button {
    min-width: 112px;
  }
  fieldset {
    border: 0;
    padding: 0;
    min-width: 0;
  }
  legend {
    padding: 0 0 12px;
  }
  .results {
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-width: 760px;
  }
  .results label {
    display: flex;
    align-items: center;
    gap: 14px;
    min-height: 56px;
    padding: 14px 16px;
    border: 1px solid var(--line);
    border-radius: 12px;
    cursor: pointer;
    overflow-wrap: anywhere;
  }
  .results label:hover,
  .results label.selected {
    background: var(--surface);
  }
  .results label.selected {
    border-color: var(--accent);
  }
  input[type='radio'] {
    flex: none;
    width: 20px;
    height: 20px;
    min-height: 0;
    accent-color: var(--accent);
  }
  .save-row {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
  }
  .save-row button {
    min-height: 56px;
  }
  .save-row p {
    overflow-wrap: anywhere;
  }
  .attribution {
    font-size: 14px;
  }
  a {
    color: inherit;
    text-underline-offset: 3px;
  }
  .location-settings > div[role='status']:empty {
    display: none;
  }
  @media (max-width: 480px) {
    .search-row {
      flex-direction: column;
    }
  }
</style>
