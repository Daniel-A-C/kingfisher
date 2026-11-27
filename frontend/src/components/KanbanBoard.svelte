<script lang="ts">
  import type { Board, Task } from '../lib/types';
  import { columnGlyph } from '../lib/dialog';
  import TaskCard from './TaskCard.svelte';

  interface Props {
    board: Board;
    projectPath: string;
    parseError?: string;
    onSelectTask: (task: Task) => void;
    onMoveTask: (taskID: string, targetStatus: string, targetIndex: number) => void;
    onQuickAddInColumn: (colID: string) => void;
  }

  let {
    board,
    projectPath,
    parseError,
    onSelectTask,
    onMoveTask,
    onQuickAddInColumn,
  }: Props = $props();

  let draggedTaskID = $state<string | null>(null);
  let dragOverCol = $state<string | null>(null);

  function tasksForColumn(colID: string): Task[] {
    return board.tasks.filter((t) => t.status === colID);
  }

  function handleDragStart(task: Task, e: DragEvent) {
    draggedTaskID = task.id;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', task.id);
    }
  }

  function handleDragOver(colID: string, e: DragEvent) {
    e.preventDefault();
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }
    dragOverCol = colID;
  }

  function handleDragLeave(colID: string) {
    if (dragOverCol === colID) {
      dragOverCol = null;
    }
  }

  function handleDrop(colID: string, e: DragEvent) {
    e.preventDefault();
    const id = draggedTaskID || e.dataTransfer?.getData('text/plain');
    draggedTaskID = null;
    dragOverCol = null;
    if (!id) return;
    onMoveTask(id, colID, -1);
  }
</script>

<section class="board-wrap" aria-label="{board.project.name} Kanban board">
  {#if parseError}
    <div class="yaml-error-banner" role="status">
      <span class="banner-icon">⚠</span>
      <span>
        <strong>Invalid YAML detected in .kingfisher.yaml</strong> — showing last valid state, waiting for valid save.
        <code class="banner-detail">{parseError}</code>
      </span>
    </div>
  {/if}

  <div class="columns-grid kf-scroll">
    {#each board.columns as col (col.id)}
      {@const colTasks = tasksForColumn(col.id)}
      {@const isActiveCol = col.id === 'active'}
      <div
        class="kf-column"
        class:active-col={isActiveCol}
        class:drag-over={dragOverCol === col.id}
        ondragover={(e) => handleDragOver(col.id, e)}
        ondragleave={() => handleDragLeave(col.id)}
        ondrop={(e) => handleDrop(col.id, e)}
        role="region"
        aria-label={col.name}
      >
        <!-- Alphonse Mucha Halo Arch Header -->
        <div class="col-arch-header">
          <svg class="mucha-arch" viewBox="0 0 280 42" preserveAspectRatio="none" aria-hidden="true">
            <path
              d="M6 40 C6 12, 54 4, 140 4 C226 4, 274 12, 274 40"
              fill="none"
              stroke={isActiveCol ? 'rgba(65, 211, 230, 0.72)' : 'rgba(217, 161, 74, 0.62)'}
              stroke-width="1.5"
            />
            <path
              d="M18 40 C18 18, 62 11, 140 11 C218 11, 262 18, 262 40"
              fill="none"
              stroke="rgba(242, 204, 126, 0.28)"
              stroke-width="1"
            />
            <circle cx="140" cy="7" r="2.8" fill={isActiveCol ? '#41d3e6' : '#f2cc7e'} />
          </svg>

          <div class="col-title-row">
            <span class="col-glyph" aria-hidden="true">{columnGlyph(col.id)}</span>
            <h2 class="col-name">{col.name}</h2>
            <span class="col-count">{colTasks.length}</span>
            <button
              type="button"
              class="col-add-btn"
              title="Add task to {col.name}"
              aria-label="Add task to {col.name}"
              onclick={() => onQuickAddInColumn(col.id)}
            >
              +
            </button>
          </div>
        </div>

        <div class="col-cards kf-scroll">
          {#each colTasks as task (task.id)}
            <TaskCard
              {task}
              isActiveColumn={isActiveCol}
              onSelect={onSelectTask}
              onDragStart={handleDragStart}
            />
          {:else}
            <div class="col-empty">
              <span class="empty-glyph">{columnGlyph(col.id)}</span>
              <span>Still waters</span>
            </div>
          {/each}
        </div>
      </div>
    {/each}
  </div>

  <footer class="board-footer">
    <span class="footer-path" title={projectPath}>
      ✦ {projectPath}/.kingfisher.yaml
    </span>
    <span class="footer-hints">
      <span class="kf-kbd">⌘1–9</span> Tabs
      <span class="kf-kbd">⌘K</span> Switch
      <span class="kf-kbd">⌘N</span> Capture
      <span class="kf-kbd">⌘⇧T</span> Glass
      <span class="kf-kbd">Esc</span> Hide
    </span>
  </footer>
</section>

<style>
  .board-wrap {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    padding: 8px 14px 10px;
    gap: 8px;
  }

  .yaml-error-banner {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 14px;
    border-radius: 10px;
    border: 1px solid rgba(228, 93, 58, 0.72);
    background: rgba(68, 22, 14, 0.78);
    color: #ffd7c9;
    font-size: 12.5px;
  }

  .banner-icon {
    font-size: 15px;
    color: var(--kf-ochre);
  }

  .banner-detail {
    display: inline-block;
    margin-left: 6px;
    font-family: var(--font-mono);
    font-size: 11px;
    opacity: 0.85;
  }

  .columns-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
    gap: 12px;
    flex: 1;
    min-height: 0;
    overflow-x: auto;
    padding: 2px;
  }

  .kf-column {
    display: flex;
    flex-direction: column;
    min-height: 0;
    border-radius: 18px 18px 12px 12px;
    border: 1px solid rgba(217, 161, 74, 0.34);
    background: linear-gradient(
      180deg,
      rgba(var(--kf-patina-rgb), calc(var(--kf-opacity) * 0.86 + 0.06)) 0%,
      rgba(var(--kf-moss-rgb), calc(var(--kf-opacity) * 0.74 + 0.04)) 100%
    );
    box-shadow:
      inset 0 1px 0 rgba(242, 204, 126, 0.16),
      0 6px 20px rgba(0, 0, 0, 0.22);
    transition: border-color 150ms ease, box-shadow 150ms ease;
    overflow: hidden;
  }

  .kf-column.active-col {
    border-color: rgba(65, 211, 230, 0.44);
    background: linear-gradient(
      180deg,
      rgba(16, 52, 56, calc(var(--kf-opacity) * 0.82 + 0.08)) 0%,
      rgba(var(--kf-moss-rgb), calc(var(--kf-opacity) * 0.74 + 0.04)) 100%
    );
  }

  .kf-column.drag-over {
    border-color: var(--kf-cerulean);
    box-shadow:
      0 0 22px rgba(65, 211, 230, 0.34),
      inset 0 0 18px rgba(65, 211, 230, 0.12);
  }

  .col-arch-header {
    position: relative;
    padding: 14px 12px 10px;
    border-bottom: 1px solid rgba(217, 161, 74, 0.22);
    flex-shrink: 0;
  }

  .mucha-arch {
    position: absolute;
    top: 3px;
    left: 0;
    width: 100%;
    height: 36px;
    pointer-events: none;
  }

  .col-title-row {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    gap: 7px;
    padding-top: 6px;
  }

  .col-glyph {
    color: var(--kf-ochre);
    font-size: 14px;
  }

  .active-col .col-glyph {
    color: var(--kf-cerulean);
  }

  .col-name {
    margin: 0;
    flex: 1;
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 0.06em;
    color: var(--kf-ivory);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .col-count {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 1px 7px;
    border-radius: 999px;
    border: 1px solid rgba(217, 161, 74, 0.42);
    background: rgba(8, 22, 22, 0.55);
    color: var(--kf-ochre);
  }

  .col-add-btn {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    border: 1px solid rgba(217, 161, 74, 0.36);
    background: rgba(217, 161, 74, 0.1);
    color: var(--kf-ochre);
    font-size: 15px;
    line-height: 1;
    cursor: pointer;
    transition: background 120ms ease, border-color 120ms ease;
  }

  .col-add-btn:hover {
    background: rgba(65, 211, 230, 0.24);
    border-color: var(--kf-cerulean);
    color: #ffffff;
  }

  .col-cards {
    display: flex;
    flex-direction: column;
    gap: 9px;
    padding: 10px;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  .col-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: 28px 12px;
    color: var(--kf-ivory-dim);
    font-family: var(--font-display);
    font-style: italic;
    font-size: 13px;
    opacity: 0.65;
  }

  .empty-glyph {
    font-size: 18px;
    font-style: normal;
    color: var(--kf-gold);
    opacity: 0.55;
  }

  .board-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 2px 6px 0;
    font-size: 11px;
    color: var(--kf-ivory-dim);
    flex-shrink: 0;
  }

  .footer-path {
    font-family: var(--font-mono);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .footer-hints {
    display: flex;
    align-items: center;
    gap: 7px;
    flex-shrink: 0;
  }
</style>
