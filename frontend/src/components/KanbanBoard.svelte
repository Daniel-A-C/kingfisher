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

  function columnTheme(colID: string): { stroke: string; jewel: string; cls: string } {
    switch (colID.toLowerCase()) {
      case 'backlog':
        return { stroke: 'rgba(176, 138, 216, 0.72)', jewel: '#c7a4ee', cls: 'col-backlog' };
      case 'ready':
        return { stroke: 'rgba(232, 166, 74, 0.82)', jewel: '#f6d388', cls: 'col-ready' };
      case 'active':
        return { stroke: 'rgba(70, 217, 236, 0.88)', jewel: '#46d9ec', cls: 'col-active' };
      case 'done':
        return { stroke: 'rgba(78, 189, 146, 0.78)', jewel: '#6be0b0', cls: 'col-done' };
      default:
        return { stroke: 'rgba(220, 162, 72, 0.72)', jewel: '#f6d388', cls: '' };
    }
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
      {@const theme = columnTheme(col.id)}
      <div
        class="kf-column {theme.cls}"
        class:active-col={isActiveCol}
        class:drag-over={dragOverCol === col.id}
        ondragover={(e) => handleDragOver(col.id, e)}
        ondragleave={() => handleDragLeave(col.id)}
        ondrop={(e) => handleDrop(col.id, e)}
        role="region"
        aria-label={col.name}
      >
        <!-- Ref 4 Inspired Top Botanical Frieze Band -->
        <div class="col-frieze-bar" aria-hidden="true">
          <svg viewBox="0 0 300 16" preserveAspectRatio="none">
            <rect x="2" y="2" width="296" height="12" rx="6" fill="rgba(150, 38, 48, 0.22)" stroke="rgba(246, 211, 136, 0.45)" stroke-width="0.9" />
            <path
              d="M12 8 Q 36 2, 60 8 T 108 8 T 156 8 T 204 8 T 252 8 T 288 8"
              fill="none"
              stroke="#f6d388"
              stroke-width="1"
              opacity="0.72"
            />
            {#each [36, 84, 132, 168, 216, 264] as cx}
              <circle {cx} cy="8" r="2.2" fill={theme.jewel} stroke="#f6d388" stroke-width="0.6" />
            {/each}
          </svg>
        </div>

        <!-- Alphonse Mucha Sunburst & Lavirotte Lancet Arch Crown -->
        <div class="col-arch-header">
          <svg class="mucha-arch" viewBox="0 0 300 72" preserveAspectRatio="none" aria-hidden="true">
            <!-- Stained-glass scalloped halo fan -->
            <path
              d="M16 66 C16 20, 66 8, 150 8 C234 8, 284 20, 284 66 Z"
              fill="rgba(150, 38, 48, 0.14)"
            />
            <!-- Radiating stained-glass halo rays -->
            <g stroke={theme.stroke} stroke-width="0.75" opacity="0.46">
              <line x1="150" y1="24" x2="150" y2="3" />
              <line x1="150" y1="24" x2="122" y2="6" />
              <line x1="150" y1="24" x2="178" y2="6" />
              <line x1="150" y1="24" x2="94" y2="12" />
              <line x1="150" y1="24" x2="206" y2="12" />
              <line x1="150" y1="24" x2="64" y2="24" />
              <line x1="150" y1="24" x2="236" y2="24" />
            </g>
            <!-- Outer lancet arch -->
            <path
              d="M6 68 C6 18, 60 5, 150 5 C240 5, 294 18, 294 68"
              fill="none"
              stroke={theme.stroke}
              stroke-width="2"
            />
            <!-- Inner gilded halo ring -->
            <path
              d="M18 68 C18 26, 68 14, 150 14 C232 14, 282 26, 282 68"
              fill="none"
              stroke="rgba(246, 211, 136, 0.54)"
              stroke-width="1.2"
              stroke-dasharray="4 3"
            />
            <!-- Lavirotte wrought-iron botanical omega scrolls -->
            <path
              d="M28 52 C40 28, 72 20, 96 27 C80 36, 62 36, 52 48"
              fill="none"
              stroke="rgba(246, 211, 136, 0.68)"
              stroke-width="1.15"
            />
            <path
              d="M272 52 C260 28, 228 20, 204 27 C220 36, 238 36, 248 48"
              fill="none"
              stroke="rgba(246, 211, 136, 0.68)"
              stroke-width="1.15"
            />
            <!-- Crown jewel medallion -->
            <circle cx="150" cy="15" r="8.5" fill="rgba(8, 22, 26, 0.94)" stroke="#f6d388" stroke-width="1.4" />
            <circle cx="150" cy="15" r="4.2" fill={theme.jewel} />
            <!-- Corner finials -->
            <circle cx="22" cy="42" r="2.4" fill="#f6d388" opacity="0.82" />
            <circle cx="278" cy="42" r="2.4" fill="#f6d388" opacity="0.82" />
          </svg>

          <div class="col-title-row">
            <span class="col-glyph-seal" style="color: {theme.jewel}" aria-hidden="true">
              {columnGlyph(col.id)}
            </span>
            <h2 class="col-name">{col.name}</h2>
            <span class="col-count">{colTasks.length}</span>
            <button
              type="button"
              class="col-add-btn"
              title="Capture into {col.name}"
              aria-label="Capture into {col.name}"
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

        <!-- Bottom ornamental lithograph foot-rail -->
        <svg class="col-foot-rail" viewBox="0 0 300 12" preserveAspectRatio="none" aria-hidden="true">
          <path d="M14 6 Q 150 12, 286 6" fill="none" stroke="rgba(246, 211, 136, 0.45)" stroke-width="1" />
          <circle cx="150" cy="8.5" r="2.2" fill={theme.jewel} />
        </svg>
      </div>
    {/each}
  </div>

  <footer class="board-footer">
    <span class="footer-path" title={projectPath}>
      ✦ {projectPath}/.kingfisher.yaml
    </span>
    <span class="footer-hints">
      <span class="kf-kbd">⌘1–9</span> Tabs
      <span class="kf-kbd">⌘K</span> Aviary
      <span class="kf-kbd">⌘N</span> Capture
      <span class="kf-kbd">⌘,</span> Settings
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
    padding: 4px 16px 12px;
    gap: 8px;
  }

  .yaml-error-banner {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 14px;
    border-radius: 12px;
    border: 1px solid rgba(230, 93, 56, 0.78);
    background: rgba(68, 22, 14, 0.82);
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
    grid-template-columns: repeat(auto-fit, minmax(235px, 1fr));
    gap: 14px;
    flex: 1;
    min-height: 0;
    overflow-x: auto;
    padding: 2px;
  }

  .kf-column {
    display: flex;
    flex-direction: column;
    min-height: 0;
    border-radius: 26px 26px 14px 14px;
    border: 1.5px solid rgba(220, 162, 72, 0.45);
    background: linear-gradient(
      180deg,
      rgba(var(--kf-patina-rgb), calc(var(--kf-opacity) * 0.88 + 0.08)) 0%,
      rgba(var(--kf-moss-rgb), calc(var(--kf-opacity) * 0.78 + 0.05)) 100%
    );
    box-shadow:
      inset 0 0 0 1px rgba(246, 211, 136, 0.16),
      inset 0 24px 36px rgba(220, 162, 72, 0.06),
      0 8px 24px rgba(0, 0, 0, 0.28);
    transition: border-color 150ms ease, box-shadow 150ms ease;
    overflow: hidden;
  }

  .kf-column.col-backlog {
    border-color: rgba(176, 138, 216, 0.45);
    box-shadow:
      inset 0 24px 36px rgba(159, 121, 200, 0.08),
      inset 0 0 0 1px rgba(246, 211, 136, 0.14),
      0 8px 24px rgba(0, 0, 0, 0.28);
  }

  .kf-column.col-ready {
    border-color: rgba(220, 162, 72, 0.56);
    box-shadow:
      inset 0 24px 36px rgba(220, 162, 72, 0.1),
      inset 0 0 0 1px rgba(246, 211, 136, 0.2),
      0 8px 24px rgba(0, 0, 0, 0.28);
  }

  .kf-column.active-col {
    border-color: rgba(70, 217, 236, 0.62);
    background: linear-gradient(
      180deg,
      rgba(14, 54, 58, calc(var(--kf-opacity) * 0.86 + 0.1)) 0%,
      rgba(var(--kf-moss-rgb), calc(var(--kf-opacity) * 0.78 + 0.05)) 100%
    );
    box-shadow:
      inset 0 28px 40px rgba(70, 217, 236, 0.12),
      inset 0 0 0 1px rgba(70, 217, 236, 0.24),
      0 8px 28px rgba(0, 0, 0, 0.32);
  }

  .kf-column.col-done {
    border-color: rgba(78, 189, 146, 0.48);
    box-shadow:
      inset 0 24px 36px rgba(47, 158, 122, 0.09),
      inset 0 0 0 1px rgba(246, 211, 136, 0.16),
      0 8px 24px rgba(0, 0, 0, 0.28);
  }

  .kf-column.drag-over {
    border-color: var(--kf-cerulean);
    box-shadow:
      0 0 26px rgba(70, 217, 236, 0.42),
      inset 0 0 24px rgba(70, 217, 236, 0.16);
  }

  .col-frieze-bar {
    height: 15px;
    padding: 4px 10px 0;
    flex-shrink: 0;
  }

  .col-frieze-bar svg {
    width: 100%;
    height: 100%;
    display: block;
  }

  .col-arch-header {
    position: relative;
    padding: 24px 12px 11px;
    border-bottom: 1px solid rgba(246, 211, 136, 0.36);
    flex-shrink: 0;
  }

  .col-foot-rail {
    width: 100%;
    height: 12px;
    display: block;
    flex-shrink: 0;
    opacity: 0.78;
  }

  .mucha-arch {
    position: absolute;
    top: 2px;
    left: 0;
    width: 100%;
    height: 62px;
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

  .col-glyph-seal {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    border: 1px solid rgba(246, 211, 136, 0.48);
    background: rgba(7, 21, 21, 0.72);
    font-size: 13px;
    flex-shrink: 0;
  }

  .col-name {
    margin: 0;
    flex: 1;
    font-family: var(--font-display);
    font-size: 16px;
    font-weight: 600;
    letter-spacing: 0.07em;
    color: var(--kf-ivory);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .col-count {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 1px 8px;
    border-radius: 999px;
    border: 1px solid rgba(246, 211, 136, 0.5);
    background: rgba(8, 22, 22, 0.68);
    color: var(--kf-ochre);
  }

  .col-add-btn {
    width: 25px;
    height: 25px;
    border-radius: 50%;
    border: 1px solid rgba(220, 162, 72, 0.48);
    background: rgba(220, 162, 72, 0.14);
    color: var(--kf-ochre);
    font-size: 15px;
    line-height: 1;
    cursor: pointer;
    transition: background 120ms ease, border-color 120ms ease;
  }

  .col-add-btn:hover {
    background: rgba(70, 217, 236, 0.28);
    border-color: var(--kf-cerulean);
    color: #ffffff;
  }

  .col-cards {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 11px;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  .col-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 5px;
    padding: 34px 12px;
    color: var(--kf-ivory-dim);
    font-family: var(--font-display);
    font-style: italic;
    font-size: 13.5px;
    opacity: 0.72;
  }

  .empty-glyph {
    font-size: 20px;
    font-style: normal;
    color: var(--kf-gold);
    opacity: 0.62;
  }

  .board-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 2px 8px 0;
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
