<script lang="ts">
  let { name, size = 24 }: { name: string; size?: number } = $props();
  const paths: Record<string, string> = {
    clock: 'M12 8v4l3 2 M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0',
    weather:
      'M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5 M17 12a5 5 0 1 1-10 0 5 5 0 0 1 10 0',
    calendar: 'M7 2v4m10-4v4M3 9h18M5 4h14a2 2 0 0 1 2 2v14H3V6a2 2 0 0 1 2-2',
    finance: 'M8 3h8a2 2 0 0 1 2 2v16H6V5a2 2 0 0 1 2-2 M9 8h6m-6 4h6m-6 4h3',
    discover:
      'M19 10c0 5-7 11-7 11S5 15 5 10a7 7 0 1 1 14 0 M14 10a2 2 0 1 1-4 0 2 2 0 0 1 4 0',
    sports:
      'M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0 M3 12h18M12 3v18M6 5c7 5 7 9 0 14M18 5c-7 5-7 9 0 14',
    plus: 'M12 5v14M5 12h14',
    close: 'M5 5l14 14M19 5L5 19',
    down: 'M8 10l4 4 4-4',
    arrow: 'M5 12h14m-5-5 5 5-5 5',
    check: 'M5 12l4 4L19 6',
    grip: 'M9 5h.01M15 5h.01M9 12h.01M15 12h.01M9 19h.01M15 19h.01',
  };
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width={name === 'grip' ? 3 : 1.5}
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"><path d={paths[name] ?? paths.clock} /></svg
>
