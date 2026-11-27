<script lang="ts">
  import type { Board } from '../lib/types';

  interface Props {
    openTabs: string[];
    activeTab: string;
    boards: Record<string, Board>;
    opacity: number;
    onSelectTab: (path: string) => void;
    onCloseTab: (path: string) => void;
    onOpenSwitcher: () => void;
    onQuickCapture: () => void;
    onOpacityChange: (value: number) => void;
    onHide: () => void;
    onMinimize: () => void;
    onMaximize: () => void;
  }

  let {
    openTabs,
    activeTab,
    boards,
    opacity,
    onSelectTab,
    onCloseTab,
    onOpenSwitcher,
    onQuickCapture,
    onOpacityChange,
    onHide,
    onMinimize,
    onMaximize,
  }: Props = $props();

  function tabLabel(path: string): { name: string; prefix: string; activeCount: number } {
    const b = boards[path];
    if (b) {
      const activeCount = b.tasks.filter((t) => t.status === 'active').length;
      return { name: b.project.name, prefix: b.project.prefix, activeCount };
    }
    const parts = path.split('/').filter(Boolean);
    return { name: parts[parts.length - 1] ?? path, prefix: 'KF', activeCount: 0 };
  }

  function shortcutHint(idx: number, total: number): string {
    if (idx < 8) return `⌘${idx + 1}`;
    if (idx === total - 1) return '⌘9';
    return '';
  }
</script>

<header class="kf-header" style="--wails-draggable: drag">
  <div class="kf-window-controls" style="--wails-draggable: no-drag">
    <button
      type="button"
      class="win-dot dot-close"
      title="Hide Kingfisher (Esc)"
      aria-label="Hide window"
      onclick={onHide}
    ></button>
    <button
      type="button"
      class="win-dot dot-min"
      title="Minimize"
      aria-label="Minimize window"
      onclick={onMinimize}
    ></button>
    <button
      type="button"
      class="win-dot dot-max"
      title="Toggle Maximize"
      aria-label="Toggle maximize"
      onclick={onMaximize}
    ></button>
  </div>

  <div class="kf-brand" title="Kingfisher">
    <svg class="brand-crest" viewBox="0 0 28 28" aria-hidden="true">
      <path
        d="M4 17 C9 12, 14 7, 24 5 C20 11, 18 16, 12 20 C9 22, 6 21, 4 17 Z"
        fill="url(#kfWing)"
        stroke="#f2cc7e"
        stroke-width="1.2"
      />
      <path
        d="M11 16 C15 14, 19 14, 24 18 C19 20, 14 22, 10 21"
        fill="none"
        stroke="#d9a14a"
        stroke-width="1.2"
        stroke-linecap="round"
      />
      <circle cx="17.5" cy="9.5" r="1.2" fill="#f6efe0" />
      <defs>
        <linearGradient id="kfWing" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#41d3e6" />
          <stop offset="100%" stop-color="#0f5f6b" />
        </linearGradient>
      </defs>
    </svg>
    <span class="brand-name">KINGFISHER</span>
  </div>

  <nav class="kf-tabs kf-scroll" aria-label="Project tabs" style="--wails-draggable: no-drag">
    {#each openTabs as path, i (path)}
      {@const info = tabLabel(path)}
      {@const isActive = path === activeTab}
      {@const hint = shortcutHint(i, openTabs.length)}
      <div class="kf-tab" class:active={isActive}>
        <button
          type="button"
          class="tab-select"
          onclick={() => onSelectTab(path)}
          title={path}
        >
          <span class="tab-prefix">{info.prefix}</span>
          <span class="tab-title">{info.name}</span>
          {#if info.activeCount > 0}
            <span class="tab-flight" title="{info.activeCount} in flight">❧ {info.activeCount}</span>
          {/if}
          {#if hint}
            <span class="tab-hint">{hint}</span>
          {/if}
        </button>
        <button
          type="button"
          class="tab-close"
          aria-label="Close {info.name} tab"
          onclick={() => onCloseTab(path)}
        >
          ×
        </button>
      </div>
    {/each}

    <button
      type="button"
      class="tab-add"
      title="Open or switch project (⌘T / ⌘K)"
      aria-label="Open project switcher"
      onclick={onOpenSwitcher}
    >
      +
    </button>
  </nav>

  <div class="kf-actions" style="--wails-draggable: no-drag">
    <label class="opacity-control" title="Stained-glass opacity (⌘⇧T to cycle)">
      <span class="opacity-glyph" aria-hidden="true">◐</span>
      <input
        type="range"
        min="0"
        max="0.85"
        step="0.05"
        value={opacity}
        aria-label="Background opacity"
        oninput={(e) => onOpacityChange(parseFloat((e.currentTarget as HTMLInputElement).value))}
      />
      <span class="opacity-pct">{Math.round(opacity * 100)}%</span>
    </label>

    {#if activeTab}
      <button
        type="button"
        class="kf-btn kf-btn-primary capture-btn"
        onclick={onQuickCapture}
        title="Quick Capture (⌘N)"
      >
        <span>✦ Dive In</span>
        <span class="kf-kbd">⌘N</span>
      </button>
    {/if}
  </div>
</header>

<div class="vine-divider" aria-hidden="true">
  <svg viewBox="0 0 1200 14" preserveAspectRatio="none">
    <path
      d="M0 7 Q 150 1, 300 7 T 600 7 T 900 7 T 1200 7"
      fill="none"
      stroke="rgba(217, 161, 74, 0.48)"
      stroke-width="1.2"
    />
    <path
      d="M0 7 Q 150 13, 300 7 T 600 7 T 900 7 T 1200 7"
      fill="none"
      stroke="rgba(65, 211, 230, 0.25)"
      stroke-width="0.9"
    />
    <circle cx="600" cy="7" r="3" fill="#f2cc7e" />
  </svg>
</div>

<style>
  .kf-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 16px 6px;
    min-height: 52px;
    background: linear-gradient(
      180deg,
      rgba(12, 32, 33, calc(var(--kf-opacity) * 0.85 + 0.12)) 0%,
      rgba(8, 22, 22, calc(var(--kf-opacity) * 0.4)) 100%
    );
  }

  .kf-window-controls {
    display: flex;
    align-items: center;
    gap: 7px;
    padding-right: 4px;
  }

  .win-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: 1px solid rgba(242, 204, 126, 0.45);
    cursor: pointer;
    padding: 0;
    transition: transform 120ms ease, filter 120ms ease;
  }
  .win-dot:hover {
    transform: scale(1.12);
    filter: brightness(1.2);
  }
  .dot-close { background: #d95d39; }
  .dot-min   { background: #d9a14a; }
  .dot-max   { background: #1ea8b8; }

  .kf-brand {
    display: flex;
    align-items: center;
    gap: 7px;
    padding-right: 6px;
    flex-shrink: 0;
  }

  .brand-crest {
    width: 24px;
    height: 24px;
  }

  .brand-name {
    font-family: var(--font-display);
    font-size: 13.5px;
    letter-spacing: 0.2em;
    color: var(--kf-ochre);
    font-weight: 600;
  }

  .kf-tabs {
    display: flex;
    align-items: flex-end;
    gap: 6px;
    flex: 1;
    min-width: 0;
    overflow-x: auto;
    padding-bottom: 2px;
  }

  .kf-tab {
    display: inline-flex;
    align-items: center;
    border-radius: 14px 14px 4px 4px;
    border: 1px solid rgba(217, 161, 74, 0.28);
    background: rgba(10, 26, 27, calc(var(--kf-opacity) * 0.72 + 0.08));
    transition: border-color 150ms ease, background 150ms ease, box-shadow 150ms ease;
    flex-shrink: 0;
  }

  .kf-tab:hover {
    border-color: rgba(242, 204, 126, 0.56);
    background: rgba(16, 42, 43, calc(var(--kf-opacity) * 0.8 + 0.14));
  }

  .kf-tab.active {
    border-color: var(--kf-cerulean);
    background: linear-gradient(
      180deg,
      rgba(30, 168, 184, calc(var(--kf-opacity) * 0.42 + 0.16)) 0%,
      rgba(12, 34, 36, calc(var(--kf-opacity) * 0.82 + 0.12)) 100%
    );
    box-shadow:
      0 0 14px rgba(65, 211, 230, 0.2),
      inset 0 1px 0 rgba(242, 204, 126, 0.42);
  }

  .tab-select {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 6px 8px 6px 11px;
    background: transparent;
    border: none;
    color: var(--kf-ivory);
    cursor: pointer;
    font-size: 12.5px;
  }

  .tab-prefix {
    font-family: var(--font-mono);
    font-size: 10.5px;
    color: var(--kf-ochre);
    padding: 1px 5px;
    border-radius: 6px;
    border: 1px solid rgba(217, 161, 74, 0.38);
    background: rgba(217, 161, 74, 0.12);
  }

  .kf-tab.active .tab-prefix {
    border-color: rgba(65, 211, 230, 0.55);
    color: #9bf1fc;
    background: rgba(65, 211, 230, 0.14);
  }

  .tab-title {
    font-family: var(--font-display);
    font-size: 13.5px;
    letter-spacing: 0.03em;
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tab-flight {
    font-size: 11px;
    color: var(--kf-cerulean);
    font-family: var(--font-mono);
  }

  .tab-hint {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--kf-ivory-dim);
    opacity: 0.8;
  }

  .tab-close {
    background: transparent;
    border: none;
    color: var(--kf-ivory-dim);
    font-size: 15px;
    line-height: 1;
    padding: 4px 8px 4px 2px;
    cursor: pointer;
    opacity: 0.7;
    transition: opacity 120ms ease, color 120ms ease;
  }

  .tab-close:hover {
    opacity: 1;
    color: var(--kf-ochre);
  }

  .tab-add {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    border: 1px dashed rgba(217, 161, 74, 0.5);
    background: rgba(217, 161, 74, 0.08);
    color: var(--kf-ochre);
    font-size: 16px;
    line-height: 1;
    cursor: pointer;
    flex-shrink: 0;
    transition: background 140ms ease, border-color 140ms ease;
  }

  .tab-add:hover {
    background: rgba(217, 161, 74, 0.22);
    border-style: solid;
    border-color: var(--kf-ochre);
  }

  .kf-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-shrink: 0;
  }

  .opacity-control {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 9px;
    border-radius: 999px;
    border: 1px solid rgba(217, 161, 74, 0.3);
    background: rgba(8, 22, 22, 0.42);
    font-size: 11px;
    color: var(--kf-ivory-soft);
  }

  .opacity-glyph {
    color: var(--kf-ochre);
    font-size: 12px;
  }

  .opacity-control input[type='range'] {
    width: 64px;
    accent-color: var(--kf-cerulean);
    cursor: pointer;
  }

  .opacity-pct {
    font-family: var(--font-mono);
    font-size: 10.5px;
    min-width: 28px;
    text-align: right;
  }

  .capture-btn {
    padding: 5px 11px;
  }

  .vine-divider {
    height: 10px;
    width: 100%;
    flex-shrink: 0;
    overflow: hidden;
  }

  .vine-divider svg {
    width: 100%;
    height: 100%;
    display: block;
  }
</style>
