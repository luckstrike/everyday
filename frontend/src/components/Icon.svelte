<script lang="ts">
  import { iconPaths } from '../lib/icon-paths.ts';
  let { name, size = 24 }: { name: string; size?: number } = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width={1.5}
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"><path d={iconPaths[name] ?? iconPaths.clock} /></svg
>
