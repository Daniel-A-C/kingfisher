<script lang="ts">
  import { onMount } from 'svelte';
  import type { ProjectEntry } from '../lib/types';
  import { attachLightDismissFallback } from '../lib/dialog';

  interface Props {
    open: boolean;
    projects: ProjectEntry[];
    openTabs: string[];
    onClose: () => void;
    onSelectProject: (path: string) => void;
    onInitFolder: () => void;
    onRescan: () => void;
  }

  let {
    open,
    projects,
    openTabs,
    onClose,
    onSelectProject,
    onInitFolder,
    onRescan,
  }: Props = $props();

  let dialogEl: HTMLDialogElement | undefined = $state();
  let searchInputEl: HTMLInputElement | undefined = $state();
  let query = $state('');
  let selectedIdx = $state(0);

  let filtered = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) return projects;
    return projects.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.slug.toLowerCase().includes(q) ||
        p.prefix.toLowerCase().includes(q) ||
        p.path.toLowerCase().includes(q)
    );
  });

  onMount(() => {
    if (!dialogEl) return;
    return attachLightDismissFallback(dialogEl);
  });

  $effect(() => {
    if (!dialogEl) return;
    if (open && !dialogEl.open) {
      query = '';
      selectedIdx = 0;
      dialogEl.showModal();
      queueMicrotask(() => searchInputEl?.focus());
    } else if (!open && dialogEl.open) {
      dialogEl.close();
    }
  });

  $effect(() => {
    if (selectedIdx >= filtered.length) {
      selectedIdx = Math.max(0, filtered.length - 1);
    }
  });

  function handleClose() {
    if (dialogEl) dialogEl.returnValue = '';
    if (open) onClose();
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (filtered.length > 0) selectedIdx = (selectedIdx + 1) % filtered.length;
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (filtered.length > 0)
        selectedIdx = (selectedIdx - 1 + filtered.length) % filtered.length;
    } else if (e.key === 'Enter' && filtered[selectedIdx]) {
      e.preventDefault();
      pick(filtered[selectedIdx].path);
    }
  }

  function pick(path: string) {
    onSelectProject(path);
    dialogEl?.close();
  }
</script>

<dialog
  bind:this={dialogEl}
  class="kf-dialog"
  aria-labelledby="switcher-title"
  {...({ closedby: 'any' } as Record<string, string>)}
  onclose={handleClose}
>
  <div class="switcher-wrap">
    <header class="switcher-header">
      <h1 id="switcher-title" class="switcher-title">✦ Project Aviary</h1>
      <div class="switcher-actions">
        <button type="button" class="kf-btn" onclick={onRescan}>Rescan</button>
        <button
          type="button"
          class="kf-btn kf-btn-primary"
          onclick={() => {
            dialogEl?.close();
            onInitFolder();
          }}
        >
          + Open / Init Folder
        </button>
      </div>
    </header>

    <div class="search-box">
      <input
        bind:this={searchInputEl}
        bind:value={query}
        class="kf-input"
        type="search"
        placeholder="Fuzzy jump to a project..."
        onkeydown={handleKeyDown}
      />
    </div>

    <ul class="project-list kf-scroll">
      {#each filtered as proj, idx (proj.path)}
        {@const isOpen = openTabs.includes(proj.path)}
        <li>
          <button
            type="button"
            class="project-item"
            class:selected={idx === selectedIdx}
            onclick={() => pick(proj.path)}
            onmouseenter={() => (selectedIdx = idx)}
          >
            <span class="proj-prefix">{proj.prefix}</span>
            <span class="proj-info">
              <span class="proj-name">{proj.name}</span>
              <span class="proj-path">{proj.path}</span>
            </span>
            {#if isOpen}
              <span class="open-pill">Open</span>
            {/if}
          </button>
        </li>
      {:else}
        <li class="empty-state">
          No matching <code>.kingfisher.yaml</code> projects found. Click
          <strong>+ Open / Init Folder</strong> to initialize any repo.
        </li>
      {/each}
    </ul>
  </div>
</dialog>

<style>
  .switcher-wrap {
    display: flex;
    flex-direction: column;
    max-height: 76vh;
  }

  .switcher-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px 12px;
    border-bottom: 1px solid rgba(217, 161, 74, 0.25);
  }

  .switcher-title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 19px;
    color: var(--kf-ochre);
    letter-spacing: 0.06em;
  }

  .switcher-actions {
    display: flex;
    gap: 8px;
  }

  .search-box {
    padding: 12px 20px 8px;
  }

  .project-list {
    list-style: none;
    margin: 0;
    padding: 4px 12px 14px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .project-item {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    padding: 10px 12px;
    border-radius: 10px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--kf-ivory);
    text-align: left;
    cursor: pointer;
    transition: background 120ms ease, border-color 120ms ease;
  }

  .project-item.selected,
  .project-item:hover {
    background: rgba(30, 168, 184, 0.2);
    border-color: rgba(65, 211, 230, 0.5);
  }

  .proj-prefix {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 2px 7px;
    border-radius: 6px;
    border: 1px solid rgba(217, 161, 74, 0.45);
    color: var(--kf-ochre);
    background: rgba(217, 161, 74, 0.12);
    flex-shrink: 0;
  }

  .proj-info {
    display: flex;
    flex-direction: column;
    min-width: 0;
    flex: 1;
  }

  .proj-name {
    font-family: var(--font-display);
    font-size: 15px;
    letter-spacing: 0.03em;
  }

  .proj-path {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--kf-ivory-dim);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .open-pill {
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 2px 7px;
    border-radius: 999px;
    border: 1px solid rgba(65, 211, 230, 0.45);
    color: var(--kf-cerulean);
  }

  .empty-state {
    padding: 26px 16px;
    text-align: center;
    color: var(--kf-ivory-soft);
    font-size: 13px;
  }
</style>
