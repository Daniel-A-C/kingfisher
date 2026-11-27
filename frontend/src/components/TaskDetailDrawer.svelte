<script lang="ts">
  import { onMount } from 'svelte';
  import type { Column, Priority, Task, TaskPatch } from '../lib/types';
  import { attachLightDismissFallback } from '../lib/dialog';

  interface Props {
    task: Task | null;
    columns: Column[];
    onClose: () => void;
    onSave: (taskID: string, patch: TaskPatch) => void;
    onDelete: (taskID: string) => void;
  }

  let { task, columns, onClose, onSave, onDelete }: Props = $props();

  let dialogEl: HTMLDialogElement | undefined = $state();
  let title = $state('');
  let status = $state('');
  let priority = $state<Priority>('medium');
  let tagsRaw = $state('');
  let description = $state('');
  let rawInput = $state('');
  let agentNotes = $state('');
  let confirmDelete = $state(false);

  const priorities: Priority[] = ['low', 'medium', 'high', 'critical'];

  onMount(() => {
    if (!dialogEl) return;
    return attachLightDismissFallback(dialogEl);
  });

  $effect(() => {
    if (!dialogEl) return;
    if (task) {
      title = task.title;
      status = task.status;
      priority = task.priority;
      tagsRaw = (task.tags ?? []).join(', ');
      description = task.description ?? '';
      rawInput = task.raw_input ?? '';
      agentNotes = task.agent_notes ?? '';
      confirmDelete = false;
      if (!dialogEl.open) {
        dialogEl.showModal();
      }
    } else if (dialogEl.open) {
      dialogEl.close();
    }
  });

  function handleClose() {
    if (dialogEl) dialogEl.returnValue = '';
    if (task) onClose();
  }

  function handleSave(e: SubmitEvent) {
    e.preventDefault();
    if (!task) return;
    const trimmed = title.trim();
    if (!trimmed) return;
    const tags = tagsRaw
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    onSave(task.id, {
      title: trimmed,
      status,
      priority,
      tags,
      has_tags: true,
      description,
      raw_input: rawInput,
      agent_notes: agentNotes,
    });
    dialogEl?.close();
  }

  function handleDelete() {
    if (!task) return;
    if (!confirmDelete) {
      confirmDelete = true;
      return;
    }
    onDelete(task.id);
    dialogEl?.close();
  }
</script>

<dialog
  bind:this={dialogEl}
  class="kf-dialog kf-drawer"
  aria-labelledby="drawer-title"
  {...({ closedby: 'any' } as Record<string, string>)}
  onclose={handleClose}
>
  {#if task}
    <form class="drawer-form" onsubmit={handleSave}>
      <header class="drawer-header">
        <div>
          <span class="drawer-id">✦ {task.id}</span>
          <h1 id="drawer-title" class="drawer-heading">Specimen Ledger</h1>
        </div>
        <button type="button" class="kf-btn" onclick={() => dialogEl?.close()}>Close</button>
      </header>

      <div class="drawer-body kf-scroll">
        <label class="field">
          <span class="field-label">Title</span>
          <input bind:value={title} class="kf-input" type="text" required />
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

          <label class="field">
            <span class="field-label">Priority</span>
            <select bind:value={priority} class="kf-select">
              {#each priorities as p}
                <option value={p}>{p}</option>
              {/each}
            </select>
          </label>
        </div>

        <label class="field">
          <span class="field-label">Tags</span>
          <input bind:value={tagsRaw} class="kf-input" type="text" placeholder="cli, ux" />
        </label>

        <label class="field">
          <span class="field-label">Description</span>
          <textarea bind:value={description} class="kf-textarea" rows="5"></textarea>
        </label>

        <label class="field agent-field">
          <span class="field-label">🪶 Agent Notes (written by coding agents or kf move)</span>
          <textarea
            bind:value={agentNotes}
            class="kf-textarea agent-textarea"
            rows="4"
            placeholder="Agents append completion summaries, changed files, or blockers here..."
          ></textarea>
        </label>

        {#if rawInput || task.raw_input}
          <label class="field raw-input-field">
            <span class="field-label">📜 Original Human Input (verbatim capture)</span>
            <textarea
              bind:value={rawInput}
              class="kf-textarea raw-textarea"
              rows="4"
            ></textarea>
          </label>
        {/if}

        <div class="meta-timestamps">
          <span>Created: {task.created || '—'}</span>
          <span>Updated: {task.updated || '—'}</span>
        </div>
      </div>

      <footer class="drawer-footer">
        <button type="button" class="kf-btn kf-btn-danger" onclick={handleDelete}>
          {confirmDelete ? 'Confirm Delete?' : 'Delete'}
        </button>
        <div class="footer-right">
          <button type="button" class="kf-btn" onclick={() => dialogEl?.close()}>Cancel</button>
          <button type="submit" class="kf-btn kf-btn-primary">Save Changes</button>
        </div>
      </footer>
    </form>
  {/if}
</dialog>

<style>
  .drawer-form {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .drawer-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px 12px;
    border-bottom: 1px solid rgba(217, 161, 74, 0.28);
  }

  .drawer-id {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--kf-cerulean);
  }

  .drawer-heading {
    margin: 2px 0 0;
    font-family: var(--font-display);
    font-size: 20px;
    color: var(--kf-ochre);
    letter-spacing: 0.04em;
  }

  .drawer-body {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 16px 20px;
    flex: 1;
    overflow-y: auto;
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
  }

  .field-label {
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--kf-ivory-dim);
  }

  .agent-textarea {
    font-family: var(--font-mono);
    font-size: 12px;
    border-color: rgba(65, 211, 230, 0.4);
    background: rgba(8, 28, 30, 0.78);
  }

  .raw-textarea {
    font-size: 12.5px;
    line-height: 1.55;
    border-color: rgba(220, 162, 72, 0.48);
    border-left: 3px solid var(--kf-gold);
    background: rgba(24, 22, 14, 0.62);
    color: var(--kf-ivory-soft);
  }

  .meta-timestamps {
    display: flex;
    flex-direction: column;
    gap: 3px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    color: var(--kf-ivory-dim);
    padding-top: 4px;
  }

  .drawer-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 20px 16px;
    border-top: 1px solid rgba(217, 161, 74, 0.24);
  }

  .footer-right {
    display: flex;
    gap: 8px;
  }
</style>
