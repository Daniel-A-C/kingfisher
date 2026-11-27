<script lang="ts">
  import type { Task } from '../lib/types';

  interface Props {
    task: Task;
    isActiveColumn: boolean;
    onSelect: (task: Task) => void;
    onDragStart: (task: Task, e: DragEvent) => void;
  }

  let { task, isActiveColumn, onSelect, onDragStart }: Props = $props();

  const priorityGlyph: Record<string, string> = {
    critical: '◆',
    high: '◈',
    medium: '✦',
    low: '❀',
  };
</script>

<article
  class="kf-card prio-{task.priority}"
  class:in-flight={isActiveColumn}
  draggable="true"
  ondragstart={(e) => onDragStart(task, e)}
>
  <!-- Delicate Art Nouveau specimen-plate corner flourishes -->
  <svg class="card-filigree tl" viewBox="0 0 22 22" aria-hidden="true">
    <path d="M2 18 C2 7, 7 2, 18 2" fill="none" stroke="rgba(246, 211, 136, 0.55)" stroke-width="1.1" />
    <circle cx="5.5" cy="5.5" r="1.2" fill="#f6d388" />
  </svg>
  <svg class="card-filigree br" viewBox="0 0 22 22" aria-hidden="true">
    <path d="M2 18 C2 7, 7 2, 18 2" fill="none" stroke="rgba(246, 211, 136, 0.45)" stroke-width="1.1" />
    <circle cx="5.5" cy="5.5" r="1.2" fill="#dca248" />
  </svg>

  <button
    type="button"
    class="card-body"
    onclick={() => onSelect(task)}
  >
    <div class="card-top">
      <span class="ticket-badge">✦ {task.id}</span>
      <span class="prio-gem" title="Priority: {task.priority}">
        <span class="gem-glyph">{priorityGlyph[task.priority] ?? '✦'}</span>
        <span class="gem-text">{task.priority}</span>
      </span>
    </div>

    <div class="card-rule" aria-hidden="true"></div>

    <h3 class="card-title">{task.title}</h3>

    {#if task.description}
      <p class="card-desc">{task.description}</p>
    {/if}

    {#if (task.tags && task.tags.length > 0) || task.agent_notes}
      <div class="card-footer">
        <div class="card-tags">
          {#each task.tags ?? [] as tag}
            <span class="tag-pill">{tag}</span>
          {/each}
        </div>
        {#if task.agent_notes}
          <span class="agent-badge" title="Agent notes attached">🪶</span>
        {/if}
      </div>
    {/if}
  </button>
</article>

<style>
  .kf-card {
    position: relative;
    border-radius: 14px;
    border: 1px solid rgba(220, 162, 72, 0.44);
    border-left: 3.5px solid var(--prio-color, var(--kf-gold));
    background:
      radial-gradient(ellipse at 100% 0%, rgba(246, 211, 136, 0.08), transparent 60%),
      linear-gradient(
        155deg,
        rgba(var(--kf-card-rgb), calc(var(--kf-opacity) * 0.94 + 0.16)) 0%,
        rgba(6, 18, 19, calc(var(--kf-opacity) * 0.88 + 0.14)) 100%
      );
    box-shadow:
      0 5px 16px rgba(0, 0, 0, 0.32),
      inset 0 0 0 1px rgba(246, 211, 136, 0.14),
      inset 0 1px 0 rgba(246, 211, 136, 0.24);
    transition: transform 150ms ease, border-color 150ms ease, box-shadow 150ms ease;
    cursor: grab;
    overflow: hidden;
  }

  .kf-card:active {
    cursor: grabbing;
  }

  .kf-card:hover {
    transform: translateY(-2px);
    border-color: rgba(246, 211, 136, 0.82);
    border-left-color: var(--prio-color, var(--kf-ochre));
    box-shadow:
      0 10px 26px rgba(0, 0, 0, 0.46),
      0 0 16px rgba(220, 162, 72, 0.18),
      inset 0 0 0 1px rgba(246, 211, 136, 0.28);
  }

  .kf-card.in-flight {
    border-color: rgba(70, 217, 236, 0.62);
    background:
      radial-gradient(ellipse at 100% 0%, rgba(70, 217, 236, 0.15), transparent 60%),
      linear-gradient(
        155deg,
        rgba(14, 42, 46, calc(var(--kf-opacity) * 0.94 + 0.18)) 0%,
        rgba(7, 22, 24, calc(var(--kf-opacity) * 0.9 + 0.14)) 100%
      );
    box-shadow:
      0 5px 20px rgba(0, 0, 0, 0.36),
      0 0 18px rgba(70, 217, 236, 0.18),
      inset 0 1px 0 rgba(126, 236, 250, 0.32);
  }

  .card-filigree {
    position: absolute;
    width: 18px;
    height: 18px;
    pointer-events: none;
    opacity: 0.72;
  }
  .card-filigree.tl { top: 2px; left: 4px; }
  .card-filigree.br { bottom: 2px; right: 2px; transform: scale(-1); }

  .card-body {
    display: flex;
    flex-direction: column;
    gap: 7px;
    width: 100%;
    padding: 12px 13px;
    border: none;
    background: transparent;
    color: inherit;
    text-align: left;
    cursor: inherit;
  }

  .card-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .ticket-badge {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    color: var(--kf-ochre);
  }

  .prio-gem {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 1px 7px;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--prio-color) 55%, transparent);
    background: color-mix(in srgb, var(--prio-color) 14%, rgba(6, 17, 18, 0.55));
    font-family: var(--font-mono);
    font-size: 9.5px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--prio-color);
  }

  .gem-glyph {
    font-size: 10px;
  }

  .card-rule {
    height: 1px;
    width: 100%;
    background: linear-gradient(
      90deg,
      rgba(220, 162, 72, 0.55) 0%,
      rgba(70, 217, 236, 0.25) 55%,
      transparent 100%
    );
  }

  .card-title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 14.5px;
    font-weight: 600;
    letter-spacing: 0.02em;
    line-height: 1.34;
    color: var(--kf-ivory);
    word-break: break-word;
  }

  .card-desc {
    margin: 0;
    font-size: 12px;
    line-height: 1.44;
    color: var(--kf-ivory-soft);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 6px;
    padding-top: 2px;
  }

  .card-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
  }

  .tag-pill {
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 1px 7px;
    border-radius: 999px;
    border: 1px solid rgba(220, 162, 72, 0.42);
    background: rgba(220, 162, 72, 0.12);
    color: var(--kf-ochre);
  }

  .in-flight .tag-pill {
    border-color: rgba(70, 217, 236, 0.45);
    background: rgba(30, 168, 184, 0.18);
    color: #c2f4fb;
  }

  .agent-badge {
    font-size: 13px;
    opacity: 0.95;
    flex-shrink: 0;
  }
</style>
