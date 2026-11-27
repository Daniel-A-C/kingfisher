<script lang="ts">
  import type { Task } from '../lib/types';

  interface Props {
    task: Task;
    isActiveColumn: boolean;
    onSelect: (task: Task) => void;
    onDragStart: (task: Task, e: DragEvent) => void;
  }

  let { task, isActiveColumn, onSelect, onDragStart }: Props = $props();

  const priorityLabel: Record<string, string> = {
    critical: 'Critical',
    high: 'High',
    medium: 'Medium',
    low: 'Low',
  };
</script>

<article
  class="kf-card prio-{task.priority}"
  class:in-flight={isActiveColumn}
  draggable="true"
  ondragstart={(e) => onDragStart(task, e)}
>
  <button
    type="button"
    class="card-body"
    onclick={() => onSelect(task)}
  >
    <div class="card-top">
      <span class="ticket-badge">✦ {task.id}</span>
      <span class="prio-gem" title="Priority: {priorityLabel[task.priority] ?? task.priority}">
        <span class="gem-dot"></span>
        <span class="gem-text">{task.priority}</span>
      </span>
    </div>

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
    border-radius: 12px;
    border: 1px solid rgba(217, 161, 74, 0.34);
    border-left: 3px solid var(--prio-color, var(--kf-gold));
    background: linear-gradient(
      155deg,
      rgba(var(--kf-card-rgb), calc(var(--kf-opacity) * 0.92 + 0.14)) 0%,
      rgba(7, 19, 20, calc(var(--kf-opacity) * 0.85 + 0.12)) 100%
    );
    box-shadow:
      0 4px 14px rgba(0, 0, 0, 0.28),
      inset 0 1px 0 rgba(242, 204, 126, 0.14);
    transition: transform 140ms ease, border-color 140ms ease, box-shadow 140ms ease;
    cursor: grab;
  }

  .kf-card:active {
    cursor: grabbing;
  }

  .kf-card:hover {
    transform: translateY(-2px);
    border-color: rgba(242, 204, 126, 0.72);
    border-left-color: var(--prio-color, var(--kf-ochre));
    box-shadow:
      0 8px 22px rgba(0, 0, 0, 0.42),
      0 0 14px rgba(217, 161, 74, 0.14);
  }

  .kf-card.in-flight {
    border-color: rgba(65, 211, 230, 0.48);
    box-shadow:
      0 4px 18px rgba(0, 0, 0, 0.32),
      0 0 16px rgba(65, 211, 230, 0.14),
      inset 0 1px 0 rgba(65, 211, 230, 0.25);
  }

  .card-body {
    display: flex;
    flex-direction: column;
    gap: 7px;
    width: 100%;
    padding: 11px 12px;
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
    letter-spacing: 0.04em;
    color: var(--kf-ochre);
  }

  .prio-gem {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--prio-color);
  }

  .gem-dot {
    width: 6px;
    height: 6px;
    transform: rotate(45deg);
    background: var(--prio-color);
    box-shadow: 0 0 6px var(--prio-color);
  }

  .card-title {
    margin: 0;
    font-size: 13.5px;
    font-weight: 500;
    line-height: 1.38;
    color: var(--kf-ivory);
    word-break: break-word;
  }

  .card-desc {
    margin: 0;
    font-size: 12px;
    line-height: 1.42;
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
    padding: 1px 6px;
    border-radius: 999px;
    border: 1px solid rgba(65, 211, 230, 0.34);
    background: rgba(30, 168, 184, 0.14);
    color: #b7eef5;
  }

  .agent-badge {
    font-size: 12px;
    opacity: 0.9;
    flex-shrink: 0;
  }
</style>
