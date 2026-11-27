<script lang="ts">
  import type { Board } from '../lib/types';

  interface Props {
    openTabs: string[];
    activeTab: string;
    boards: Record<string, Board>;
    opacity: number;
    showHeaderOpacity: boolean;
    onSelectTab: (path: string) => void;
    onCloseTab: (path: string) => void;
    onOpenSwitcher: () => void;
    onOpenSettings: () => void;
    onQuickCapture: () => void;
    onOpacityChange: (value: number) => void;
  }

  let {
    openTabs,
    activeTab,
    boards,
    opacity,
    showHeaderOpacity,
    onSelectTab,
    onCloseTab,
    onOpenSwitcher,
    onOpenSettings,
    onQuickCapture,
    onOpacityChange,
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
  <div class="kf-brand" title="Kingfisher">
    <svg class="brand-crest" viewBox="0 0 38 38" aria-hidden="true">
      <circle cx="19" cy="19" r="17" fill="rgba(12, 42, 45, 0.72)" stroke="#dca248" stroke-width="1.3" />
      <circle cx="19" cy="19" r="14" fill="none" stroke="rgba(246, 211, 136, 0.35)" stroke-width="0.8" stroke-dasharray="2 2" />
      <path
        d="M6 23 C12 16, 18 9, 31 7 C26 14, 23 20, 15 26 C11 28, 8 27, 6 23 Z"
        fill="url(#kfWingJewel)"
        stroke="#f6d388"
        stroke-width="1.1"
      />
      <path
        d="M14 21 C19 19, 25 19, 31 24 C24 26, 18 28, 13 27"
        fill="url(#kfBreastCopper)"
        stroke="#f6d388"
        stroke-width="0.9"
      />
      <circle cx="22.5" cy="12.5" r="1.4" fill="#fffdf7" />
      <defs>
        <linearGradient id="kfWingJewel" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#6ae8f7" />
          <stop offset="55%" stop-color="#1ea8b8" />
          <stop offset="100%" stop-color="#0b4e59" />
        </linearGradient>
        <linearGradient id="kfBreastCopper" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#f6d388" />
          <stop offset="100%" stop-color="#c46d34" />
        </linearGradient>
      </defs>
    </svg>
    <div class="brand-wordmark">
      <span class="brand-name">KINGFISHER</span>
      <span class="brand-flourish" aria-hidden="true">❧ ✦ ☙</span>
    </div>
  </div>

  <nav class="kf-tabs kf-scroll" aria-label="Project tabs" style="--wails-draggable: no-drag">
    {#each openTabs as path, i (path)}
      {@const info = tabLabel(path)}
      {@const isActive = path === activeTab}
      {@const hint = shortcutHint(i, openTabs.length)}
      <div class="kf-tab" class:active={isActive}>
        {#if isActive}
          <svg class="tab-crown" viewBox="0 0 120 8" preserveAspectRatio="none" aria-hidden="true">
            <path d="M4 7 Q 60 -2, 116 7" fill="none" stroke="#f6d388" stroke-width="1.3" />
            <circle cx="60" cy="2.5" r="2" fill="#46d9ec" />
          </svg>
        {/if}
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
    {#if showHeaderOpacity}
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
    {/if}

    <button
      type="button"
      class="kf-btn settings-btn"
      onclick={onOpenSettings}
      title="Settings &amp; Accessibility (⌘,)"
      aria-label="Open settings"
    >
      ⚙
    </button>

    <button
      type="button"
      class="kf-btn kf-btn-primary capture-btn"
      onclick={onQuickCapture}
      title="Project-Agnostic Capture (⌘N)"
    >
      <span>✦ Capture</span>
      <span class="kf-kbd">⌘N</span>
    </button>
  </div>
</header>

<!-- Ornate Symmetrical Art Nouveau Floral & Whiplash Frieze -->
<div class="vine-frieze" aria-hidden="true">
  <svg viewBox="0 0 1200 22" preserveAspectRatio="none">
    <defs>
      <linearGradient id="friezeGold" x1="0" y1="0" x2="1" y2="0">
        <stop offset="0%" stop-color="rgba(220,162,72,0.18)" />
        <stop offset="20%" stop-color="rgba(220,162,72,0.72)" />
        <stop offset="50%" stop-color="#f6d388" />
        <stop offset="80%" stop-color="rgba(220,162,72,0.72)" />
        <stop offset="100%" stop-color="rgba(220,162,72,0.18)" />
      </linearGradient>
    </defs>
    <!-- Primary whiplash vine -->
    <path
      d="M12 11 C 90 2, 160 20, 240 11 C 320 2, 400 20, 480 11 C 535 5, 565 16, 600 11 C 635 16, 665 5, 720 11 C 800 20, 880 2, 960 11 C 1040 20, 1110 2, 1188 11"
      fill="none"
      stroke="url(#friezeGold)"
      stroke-width="1.4"
    />
    <!-- Counter-curving teal tendril -->
    <path
      d="M36 11 C 120 20, 200 2, 280 11 C 360 20, 440 2, 520 11 C 555 15, 578 7, 600 11 C 622 7, 645 15, 680 11 C 760 2, 840 20, 920 11 C 1000 2, 1080 20, 1164 11"
      fill="none"
      stroke="rgba(70, 217, 236, 0.38)"
      stroke-width="1"
    />
    <!-- Center lotus & kingfisher jewel medallion -->
    <path d="M582 11 Q 600 1, 618 11 Q 600 21, 582 11 Z" fill="rgba(14, 48, 52, 0.9)" stroke="#f6d388" stroke-width="1.2" />
    <circle cx="600" cy="11" r="3.4" fill="#46d9ec" stroke="#f6d388" stroke-width="0.9" />
    <circle cx="566" cy="11" r="1.8" fill="#f6d388" />
    <circle cx="634" cy="11" r="1.8" fill="#f6d388" />
    <circle cx="300" cy="11" r="1.6" fill="rgba(246, 211, 136, 0.7)" />
    <circle cx="900" cy="11" r="1.6" fill="rgba(246, 211, 136, 0.7)" />
  </svg>
</div>

<style>
  .kf-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 16px 6px 80px; /* Clears native macOS TitleBarHiddenInset traffic lights */
    min-height: 58px;
    background: linear-gradient(
      180deg,
      rgba(12, 34, 35, calc(var(--kf-opacity) * 0.88 + 0.14)) 0%,
      rgba(7, 21, 21, calc(var(--kf-opacity) * 0.45)) 100%
    );
  }

  .kf-brand {
    display: flex;
    align-items: center;
    gap: 8px;
    padding-right: 6px;
    flex-shrink: 0;
  }

  .brand-crest {
    width: 32px;
    height: 32px;
    filter: drop-shadow(0 2px 6px rgba(0, 0, 0, 0.55));
  }

  .brand-wordmark {
    display: flex;
    flex-direction: column;
    align-items: center;
    line-height: 1;
  }

  .brand-name {
    font-family: var(--font-display);
    font-size: 13.5px;
    letter-spacing: 0.22em;
    color: var(--kf-ochre);
    font-weight: 600;
  }

  .brand-flourish {
    font-size: 8.5px;
    letter-spacing: 0.25em;
    color: var(--kf-cerulean);
    margin-top: 2px;
    opacity: 0.85;
  }

  .kf-tabs {
    display: flex;
    align-items: flex-end;
    gap: 7px;
    flex: 1;
    min-width: 0;
    overflow-x: auto;
    padding: 4px 2px 2px;
  }

  .kf-tab {
    position: relative;
    display: inline-flex;
    align-items: center;
    border-radius: 18px 18px 5px 5px;
    border: 1px solid rgba(220, 162, 72, 0.34);
    background: linear-gradient(
      180deg,
      rgba(14, 36, 37, calc(var(--kf-opacity) * 0.78 + 0.1)) 0%,
      rgba(8, 22, 23, calc(var(--kf-opacity) * 0.7 + 0.08)) 100%
    );
    box-shadow: inset 0 1px 0 rgba(246, 211, 136, 0.16);
    transition: border-color 150ms ease, background 150ms ease, box-shadow 150ms ease;
    flex-shrink: 0;
  }

  .tab-crown {
    position: absolute;
    top: -6px;
    left: 10%;
    width: 80%;
    height: 8px;
    pointer-events: none;
  }

  .kf-tab:hover {
    border-color: rgba(246, 211, 136, 0.65);
    background: rgba(18, 46, 48, calc(var(--kf-opacity) * 0.84 + 0.15));
  }

  .kf-tab.active {
    border-color: var(--kf-cerulean);
    background: linear-gradient(
      180deg,
      rgba(30, 168, 184, calc(var(--kf-opacity) * 0.48 + 0.2)) 0%,
      rgba(11, 32, 35, calc(var(--kf-opacity) * 0.86 + 0.12)) 100%
    );
    box-shadow:
      0 0 16px rgba(70, 217, 236, 0.24),
      inset 0 1px 0 rgba(246, 211, 136, 0.52);
  }

  .tab-select {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 6px 8px 6px 12px;
    background: transparent;
    border: none;
    color: var(--kf-ivory);
    cursor: pointer;
    font-size: 12.5px;
  }

  .tab-prefix {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--kf-ochre);
    padding: 1px 6px;
    border-radius: 999px;
    border: 1px solid rgba(220, 162, 72, 0.45);
    background: rgba(220, 162, 72, 0.14);
  }

  .kf-tab.active .tab-prefix {
    border-color: rgba(70, 217, 236, 0.65);
    color: #b7f5fd;
    background: rgba(70, 217, 236, 0.16);
  }

  .tab-title {
    font-family: var(--font-display);
    font-size: 14px;
    letter-spacing: 0.04em;
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
    opacity: 0.85;
  }

  .tab-close {
    background: transparent;
    border: none;
    color: var(--kf-ivory-dim);
    font-size: 15px;
    line-height: 1;
    padding: 4px 8px 4px 2px;
    cursor: pointer;
    opacity: 0.72;
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
    border: 1px dashed rgba(220, 162, 72, 0.56);
    background: rgba(220, 162, 72, 0.1);
    color: var(--kf-ochre);
    font-size: 16px;
    line-height: 1;
    cursor: pointer;
    flex-shrink: 0;
    transition: background 140ms ease, border-color 140ms ease;
  }

  .tab-add:hover {
    background: rgba(220, 162, 72, 0.24);
    border-style: solid;
    border-color: var(--kf-ochre);
  }

  .kf-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .opacity-control {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 9px;
    border-radius: 999px;
    border: 1px solid rgba(220, 162, 72, 0.36);
    background: rgba(8, 22, 22, 0.48);
    font-size: 11px;
    color: var(--kf-ivory-soft);
  }

  .opacity-glyph {
    color: var(--kf-ochre);
    font-size: 12px;
  }

  .opacity-control input[type='range'] {
    width: 60px;
    accent-color: var(--kf-cerulean);
    cursor: pointer;
  }

  .opacity-pct {
    font-family: var(--font-mono);
    font-size: 10.5px;
    min-width: 28px;
    text-align: right;
  }

  .settings-btn {
    width: 31px;
    height: 31px;
    padding: 0;
    font-size: 15px;
    color: var(--kf-ochre);
  }

  .capture-btn {
    padding: 5px 12px;
  }

  .vine-frieze {
    height: 18px;
    width: 100%;
    flex-shrink: 0;
    overflow: hidden;
  }

  .vine-frieze svg {
    width: 100%;
    height: 100%;
    display: block;
  }
</style>
