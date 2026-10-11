<script lang="ts">
  import { onMount } from 'svelte';
  let {
    onkeep,
    ondiscard,
  }: {
    onkeep: () => void;
    ondiscard: () => void;
  } = $props();
  let dialog: HTMLDialogElement;
  let keepButton: HTMLButtonElement;
  let discardButton: HTMLButtonElement;
  onMount(() => {
    const previousFocus = document.activeElement;
    dialog.showModal();
    keepButton.focus();
    return () => {
      dialog.close();
      if (previousFocus instanceof HTMLElement && previousFocus.isConnected)
        previousFocus.focus();
    };
  });
</script>

<dialog
  bind:this={dialog}
  aria-labelledby="discard-title"
  aria-describedby="discard-description"
  onkeydown={(event) => {
    if (event.key !== 'Tab') return;
    if (event.shiftKey && document.activeElement === discardButton) {
      event.preventDefault();
      keepButton.focus();
    } else if (!event.shiftKey && document.activeElement === keepButton) {
      event.preventDefault();
      discardButton.focus();
    }
  }}
  oncancel={(event) => {
    event.preventDefault();
    onkeep();
  }}
>
  <h2 id="discard-title">Discard layout changes?</h2>
  <p id="discard-description">
    Your unsaved changes will be lost. Your saved layout will stay as it was.
  </p>
  <div class="dialog-actions">
    <button class="danger" bind:this={discardButton} onclick={ondiscard}
      >Discard changes</button
    >
    <button class="primary" bind:this={keepButton} onclick={onkeep}
      >Keep editing</button
    >
  </div>
</dialog>

<style>
  dialog {
    width: min(440px, calc(100vw - 32px));
    max-height: calc(100dvh - 32px);
    overflow-y: auto;
    padding: 28px;
    color: var(--text);
    background: var(--surface);
    border: 1px solid var(--muted);
    border-radius: 20px;
  }
  dialog::backdrop {
    background: oklch(12% 0.012 160 / 0.65);
  }
  p {
    margin-top: 12px;
    color: var(--muted);
    font-size: 16px;
    line-height: 1.5;
  }
  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 24px;
  }
  @media (max-width: 480px) {
    dialog {
      padding: 24px;
    }
    .dialog-actions {
      gap: 8px;
    }
    button {
      padding-inline: 14px;
    }
  }
</style>
