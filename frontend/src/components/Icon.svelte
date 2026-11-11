<script lang="ts">
  let { name, size = 24 }: { name: string; size?: number } = $props();
  const paths: Record<string, string> = {
    settings:
      'M9 3h6l.5 2.5 2 1.2 2.4-.8 2 3.5-1.9 1.6v2l1.9 1.6-2 3.5-2.4-.8-2 1.2L15 21H9l-.5-2.5-2-1.2-2.4.8-2-3.5L4 13v-2L2.1 9.4l2-3.5 2.4.8 2-1.2L9 3 M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0',
    clock: 'M12 8v4l3 2 M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0',
    weather:
      'M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5 M17 12a5 5 0 1 1-10 0 5 5 0 0 1 10 0',
    pin: 'M19 10c0 5-7 11-7 11S5 15 5 10a7 7 0 1 1 14 0 M14 10a2 2 0 1 1-4 0 2 2 0 0 1 4 0',
    close: 'M5 5l14 14M19 5L5 19',
    down: 'M8 10l4 4 4-4',
    'arrow-right': 'M4 12h16m-6-6 6 6-6 6',
    check: 'M5 12l4 4L19 6',
  };
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
  aria-hidden="true"><path d={paths[name] ?? paths.clock} /></svg
>
