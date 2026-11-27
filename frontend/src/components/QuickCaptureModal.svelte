<script lang="ts">
  import { onMount } from 'svelte';
  import type { Column, Priority } from '../lib/types';
  import { attachLightDismissFallback } from '../lib/dialog';

  interface Props {
    open: boolean;
    columns: Column[];
    defaultStatus: string;
    projectName: string;
    onClose: () => void;
    onSubmit: (payload: {
      title: string;
      status: string;
      priority: Priority;
      tags: string[];
      description: string;
    }) => void;
  }

  let { open, columns, defaultStatus, projectName, onClose, onSubmit }: Props = $props();

  let dialogEl: HTMLDialogElement | undefined = $state();
  let titleInputEl: HTMLInputElement | undefined = $state();

  let title = $state('');
  let status = $state('ready');
  let priority = $state<Priority>('medium');
  let tagsRaw = $state('');
  let description = $state('');

  const priorities: Priority[] = ['low', 'medium', 'high', 'critical'];

  onMount(() => {
    if (!dialogEl) return;
    return attachLightDismissFallback(dialogEl);
  });

  $effect(() => {
    if (!dialogEl) return;
    if (open && !dialogEl.open) {
      title = '';
      status = defaultStatus || columns[0]?.id || 'ready';
      priority = 'medium';
      tagsRaw = '';
      description = '';
      dialogEl.showModal();
      queueMicrotask(() => titleInputEl?.focus());
    } else if (!open && dialogEl.open) {
      dialogEl.close();
    }
  });

  function handleClose() {
    if (dialogEl) dialogEl.returnValue = '';
    if (open) onClose();
  }

  function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    const trimmed = title.trim();
    if (!trimmed) return;
    const tags = tagsRaw
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    onSubmit({
      title: trimmed,
      status,
      priority,
      tags,
      description: description.trim(),
    });
    dialogEl?.close();
  }
</script>

<dialog
  bind:this={dialogEl}
  class="kf-dialog"
  aria-labelledby="capture-title"
  {...({ closedby: 'any' } as Record<string, string>)}
  onclose={handleClose}
>
  <form class="capture-form" onsubmit={handleSubmit}>
    <header class="modal-header">
      <div>
        <span class="modal-eyebrow">✦ {projectName}</span>
        <h1 id="capture-title" class="modal-title">Quick Capture</h1>
      </div>
      <button type="button" class="kf-btn" onclick={() => dialogEl?.close()}>Esc</button>
    </header>

    <div class="modal-body">
      <label class="field">
        <span class="field-label">Task Title</span>
        <input
          bind:this={titleInputEl}
          bind:value={title}
          class="kf-input"
          type="text"
          placeholder="What needs to be built or fixed?"
          required
        />
      </label>

      <div class="field-row">
        <label class="field">
          <span class="field-label">Column</span>
          <select bind:value={status} class="kf-select">
            {#each columns as col}
              <option value={col.id}>{col.name}</option>
            {/each}
          </select>
        </label>

        <div class="field">
          <span class="field-label">Priority</span>
          <div class="prio-pills">
            {#each priorities as p}
              <button
                type="button"
                class="prio-pill prio-{p}"
                class:selected={priority === p}
                onclick={() => (priority = p)}
              >
                {p}
              </button>
            {/each}
          </div>
        </div>
      </div>

      <label class="field">
        <span class="field-label">Tags (comma separated)</span>
        <input
          bind:value={tagsRaw}
          class="kf-input"
          type="text"
          placeholder="cli, ux, parser"
        />
      </label>

      <label class="field">
        <span class="field-label">Description</span>
        <textarea
          bind:value={description}
          class="kf-textarea"
          rows="3"
          placeholder="Context for yourself or a coding agent..."
        ></textarea>
      </label>
    </div>

    <footer class="modal-footer">
      <button type="button" class="kf-btn" onclick={() => dialogEl?.close()}>Cancel</button>
      <button type="submit" class="kf-btn kf-btn-primary">✦ Add to Board</button>
    </footer>
  </form>
</dialog>

<style>
  .capture-form {
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px 12px;
    border-bottom: 1px solid rgba(217, 161, 74, 0.28);
  }

  .modal-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--kf-cerulean);
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .modal-title {
    margin: 2px 0 0;
    font-family: var(--font-display);
    font-size: 20px;
    font-weight: 600;
    color: var(--kf-ochre);
    letter-spacing: 0.05em;
  }

  .modal-body {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 16px 20px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 5px;
    flex: 1;
  }

  .field-row {
    display: flex;
    gap: 12px;
    align-items: flex-end;
  }

  .field-label {
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--kf-ivory-dim);
  }

  .prio-pills {
    display: flex;
    gap: 5px;
  }

  .prio-pill {
    flex: 1;
    padding: 8px 6px;
    border-radius: 8px;
    border: 1px solid rgba(217, 161, 74, 0.28);
    background: rgba(6, 17, 18, 0.65);
    color: var(--kf-ivory-soft);
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: capitalize;
    cursor: pointer;
    transition: border-color 120ms ease, background 120ms ease, color 120ms ease;
  }

  .prio-pill.selected {
    border-color: var(--prio-color);
    background: color-mix(in srgb, var(--prio-color) 22%, rgba(6, 17, 18, 0.8));
    color: #ffffff;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding: 12px 20px 16px;
    border-top: 1px solid rgba(217, 161, 74, 0.22);
  }
</style>
