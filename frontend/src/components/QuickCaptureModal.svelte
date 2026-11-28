<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api';
  import type { AIConfig, AIStatus, Priority } from '../lib/types';
  import { attachLightDismissFallback } from '../lib/dialog';

  interface Props {
    open: boolean;
    defaultStatus: string;
    projectName: string;
    projectCount: number;
    aiConfig: AIConfig;
    onClose: () => void;
    onCapture: (rawContext: string, defaultStatus: string) => Promise<void>;
    onSaveAIConfig: (cfg: AIConfig) => Promise<void>;
  }

  let {
    open,
    defaultStatus,
    projectName,
    projectCount,
    aiConfig,
    onClose,
    onCapture,
    onSaveAIConfig,
  }: Props = $props();

  let dialogEl: HTMLDialogElement | undefined = $state();
  let textareaEl: HTMLTextAreaElement | undefined = $state();

  let rawContext = $state('');
  let processing = $state(false);
  let showSettings = $state(false);
  let ringPaused = $state(false);
  let aiStatus = $state<AIStatus | null>(null);
  let openCount = $state(0);

  // Partially functional orbital ring controls
  let selectedColumn = $state('ready');
  let selectedPriority = $state<Priority | ''>('');

  let editProvider = $state('ollama');
  let editEndpoint = $state('http://127.0.0.1:11434');
  let editModel = $state('qwen2.5:1.5b');
  let editEnabled = $state(true);

  const columnNodes = [
    { id: 'backlog', label: 'Backlog', glyph: '◌' },
    { id: 'ready', label: 'Ready', glyph: '✦' },
    { id: 'active', label: 'In Flight', glyph: '❧' },
  ];

  const priorityNodes: { id: Priority; label: string; glyph: string }[] = [
    { id: 'critical', label: 'Critical', glyph: '◆' },
    { id: 'high', label: 'High', glyph: '◈' },
    { id: 'medium', label: 'Medium', glyph: '✦' },
    { id: 'low', label: 'Low', glyph: '❀' },
  ];

  onMount(() => {
    if (!dialogEl) return;
    return attachLightDismissFallback(dialogEl);
  });

  $effect(() => {
    if (!dialogEl) return;
    if (open && !dialogEl.open) {
      rawContext = '';
      processing = false;
      showSettings = false;
      selectedColumn = defaultStatus || 'ready';
      selectedPriority = '';
      editProvider = aiConfig.provider || 'ollama';
      editEndpoint = aiConfig.endpoint || 'http://127.0.0.1:11434';
      editModel = aiConfig.model || 'qwen2.5:1.5b';
      editEnabled = aiConfig.enabled ?? true;
      openCount += 1;
      dialogEl.showModal();
      // Immediate focus so user can type at t=0ms while the outer ring builds itself
      queueMicrotask(() => textareaEl?.focus());
      api
        .getAIStatus()
        .then((st) => (aiStatus = st))
        .catch(() => (aiStatus = null));
    } else if (!open && dialogEl.open) {
      dialogEl.close();
    }
  });

  function handleClose() {
    if (dialogEl) dialogEl.returnValue = '';
    if (open) onClose();
  }

  async function submitCapture() {
    const trimmed = rawContext.trim();
    if (!trimmed || processing) return;
    processing = true;
    try {
      const payload = selectedPriority ? `[priority:${selectedPriority}] ${trimmed}` : trimmed;
      await onCapture(payload, selectedColumn || defaultStatus);
      dialogEl?.close();
    } finally {
      processing = false;
    }
  }

  function handleFormSubmit(e: SubmitEvent) {
    e.preventDefault();
    submitCapture();
  }

  function handleKeyDown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      submitCapture();
    }
  }

  async function saveSettings() {
    await onSaveAIConfig({
      enabled: editEnabled,
      provider: editProvider,
      endpoint: editEndpoint.trim(),
      model: editModel.trim(),
    });
    aiStatus = await api.getAIStatus().catch(() => null);
    showSettings = false;
    queueMicrotask(() => textareaEl?.focus());
  }
</script>

<dialog
  bind:this={dialogEl}
  class="capture-stage-dialog"
  aria-labelledby="capture-title"
  {...({ closedby: 'any' } as Record<string, string>)}
  onclose={handleClose}
>
  {#key openCount}
    <div
      class="capture-cosmos"
      class:is-processing={processing}
      class:ring-paused={ringPaused}
    >
      <!-- Self-Building Maximalist Art Nouveau Zodiacal / Mucha Halo Ring -->
      <svg
        class="ornate-ring-svg"
        viewBox="0 0 820 620"
        preserveAspectRatio="xMidYMid meet"
        aria-hidden="true"
      >
        <defs>
          <radialGradient id="capHaloCrimson" cx="50%" cy="50%" r="50%">
            <stop offset="56%" stop-color="rgba(150, 38, 48, 0.34)" />
            <stop offset="76%" stop-color="rgba(220, 162, 72, 0.16)" />
            <stop offset="92%" stop-color="rgba(30, 168, 184, 0.12)" />
            <stop offset="100%" stop-color="transparent" />
          </radialGradient>
          <linearGradient id="capGoldGrad" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#ffe7a6" />
            <stop offset="50%" stop-color="#dca248" />
            <stop offset="100%" stop-color="#9d712d" />
          </linearGradient>
        </defs>

        <!-- Glowing crimson & gold Mucha halo backdrop -->
        <ellipse class="ring-glow" cx="410" cy="310" rx="372" ry="282" fill="url(#capHaloCrimson)" />

        <!-- Stage 1: Inner gilded scalloped lace ring -->
        <g class="ring-layer layer-inner">
          <ellipse
            class="draw-stroke s1"
            cx="410"
            cy="310"
            rx="304"
            ry="220"
            fill="none"
            stroke="url(#capGoldGrad)"
            stroke-width="2"
            pathLength="1000"
          />
          <ellipse
            class="draw-stroke s1-b"
            cx="410"
            cy="310"
            rx="314"
            ry="230"
            fill="none"
            stroke="rgba(246, 211, 136, 0.55)"
            stroke-width="1"
            stroke-dasharray="4 6"
            pathLength="1000"
          />
        </g>

        <!-- Stage 2: 24 radiating sunburst rays & Lavirotte wrought-iron ribs -->
        <g class="ring-layer layer-rays" transform="translate(410 310)">
          {#each Array(24) as _, i}
            <line
              class="sun-ray"
              style="--ray-delay: {80 + i * 14}ms"
              x1={Math.cos((i * Math.PI) / 12) * 240}
              y1={Math.sin((i * Math.PI) / 12) * 185}
              x2={Math.cos((i * Math.PI) / 12) * 356}
              y2={Math.sin((i * Math.PI) / 12) * 268}
              stroke={i % 2 === 0 ? 'rgba(246, 211, 136, 0.52)' : 'rgba(70, 217, 236, 0.34)'}
              stroke-width={i % 2 === 0 ? '1.2' : '0.8'}
            />
          {/each}
        </g>

        <!-- Stage 3: Outer concentric Zodiacal double ring + botanical crest arcs -->
        <g class="ring-layer layer-outer">
          <ellipse
            class="draw-stroke s3"
            cx="410"
            cy="310"
            rx="354"
            ry="264"
            fill="none"
            stroke="url(#capGoldGrad)"
            stroke-width="2.4"
            pathLength="1000"
          />
          <ellipse
            class="draw-stroke s3-b"
            cx="410"
            cy="310"
            rx="368"
            ry="276"
            fill="none"
            stroke="rgba(70, 217, 236, 0.55)"
            stroke-width="1.2"
            pathLength="1000"
          />
          <!-- Top & bottom Lavirotte omega wrought-iron crowns -->
          <path
            class="draw-stroke s3-c"
            d="M310 56 C350 22, 382 14, 410 14 C438 14, 470 22, 510 56"
            fill="none"
            stroke="#f6d388"
            stroke-width="1.8"
            pathLength="240"
          />
          <path
            class="draw-stroke s3-c"
            d="M310 564 C350 598, 382 606, 410 606 C438 606, 470 598, 510 564"
            fill="none"
            stroke="#f6d388"
            stroke-width="1.8"
            pathLength="240"
          />
        </g>

        <!-- Stage 4: 12 Orbital Zodiacal Medallions blooming around the ring -->
        <g class="ring-layer layer-medallions" transform="translate(410 310)">
          {#each Array(12) as _, i}
            {@const angle = (i * Math.PI) / 6 - Math.PI / 2}
            {@const mx = Math.cos(angle) * 354}
            {@const my = Math.sin(angle) * 264}
            <g class="zodiac-bead" style="--bead-delay: {220 + i * 32}ms" transform="translate({mx} {my})">
              <circle r="11" fill="rgba(10, 25, 30, 0.94)" stroke="#f6d388" stroke-width="1.4" />
              <circle r="6.5" fill="rgba(150, 38, 48, 0.5)" stroke="#dca248" stroke-width="0.8" />
              <circle r="2.6" fill={i % 2 === 0 ? '#46d9ec' : '#f6d388'} />
            </g>
          {/each}
        </g>
      </svg>

      <!-- Functional Orbital Nodes mounted along the left & right arcs of the ring -->
      <div class="orbital-rail orbital-left" role="group" aria-label="Target column">
        <span class="orbital-heading">Column</span>
        {#each columnNodes as col, idx (col.id)}
          <button
            type="button"
            class="orbital-node"
            class:selected={selectedColumn === col.id}
            style="--node-delay: {260 + idx * 70}ms"
            aria-pressed={selectedColumn === col.id}
            onclick={() => {
              selectedColumn = col.id;
              textareaEl?.focus();
            }}
          >
            <span class="node-glyph" aria-hidden="true">{col.glyph}</span>
            <span class="node-label">{col.label}</span>
          </button>
        {/each}
      </div>

      <div class="orbital-rail orbital-right" role="group" aria-label="Priority hint">
        <span class="orbital-heading">Priority</span>
        {#each priorityNodes as pr, idx (pr.id)}
          <button
            type="button"
            class="orbital-node prio-{pr.id}"
            class:selected={selectedPriority === pr.id}
            style="--node-delay: {280 + idx * 65}ms"
            aria-pressed={selectedPriority === pr.id}
            onclick={() => {
              selectedPriority = selectedPriority === pr.id ? '' : pr.id;
              textareaEl?.focus();
            }}
          >
            <span class="node-glyph" aria-hidden="true">{pr.glyph}</span>
            <span class="node-label">{pr.label}</span>
          </button>
        {/each}
      </div>

      <!-- Center Illuminated Capture Cartouche (Snaps in immediately!) -->
      <form class="capture-cartouche" onsubmit={handleFormSubmit}>
        <!-- Top ornate keystone filigree -->
        <svg class="cartouche-crown" viewBox="0 0 420 20" preserveAspectRatio="none" aria-hidden="true">
          <path d="M12 18 Q 210 -4, 408 18" fill="none" stroke="#f6d388" stroke-width="1.5" />
          <path d="M90 18 Q 210 2, 330 18" fill="none" stroke="rgba(70, 217, 236, 0.55)" stroke-width="1" />
          <circle cx="210" cy="8" r="4.2" fill="#46d9ec" stroke="#f6d388" stroke-width="1.1" />
        </svg>

        <header class="modal-header">
          <div>
            <span class="modal-eyebrow">
              ✦ Auto-Routing · {projectCount} project{projectCount === 1 ? '' : 's'}{projectName ? ` (${projectName})` : ''}
            </span>
            <h1 id="capture-title" class="modal-title">Capture Context</h1>
          </div>
          <div class="header-right">
            <button
              type="button"
              class="ai-pill"
              class:online={aiStatus?.available}
              onclick={() => (showSettings = !showSettings)}
              title={aiStatus?.message ?? 'Configure local model'}
            >
              <span class="ai-dot"></span>
              <span>
                {#if aiStatus?.available}
                  {aiStatus.model}
                {:else}
                  Local AI · Fallback
                {/if}
              </span>
            </button>
            <button
              type="button"
              class="kf-btn ring-pause-btn"
              onclick={() => (ringPaused = !ringPaused)}
              title={ringPaused ? 'Resume ring motion' : 'Pause ring motion'}
              aria-label={ringPaused ? 'Play ring animation' : 'Pause ring animation'}
            >
              {ringPaused ? '▶' : '⏸'}
            </button>
            <button type="button" class="kf-btn" onclick={() => dialogEl?.close()}>Esc</button>
          </div>
        </header>

        <div class="modal-body">
          <label for="capture-context-input" class="visually-hidden">
            Capture context for one or more tickets
          </label>
          <textarea
            id="capture-context-input"
            bind:this={textareaEl}
            bind:value={rawContext}
            class="kf-textarea context-box"
            rows="5"
            disabled={processing}
            placeholder="Dump one or more ideas for any project…&#10;&#10;Mention project names to auto-route (e.g. 'For Rookery: fix wallpaper wake fade. Also in Kingfisher: add card keyboard nav.'). Distinct items split into separate tickets."
            onkeydown={handleKeyDown}
          ></textarea>

          {#if showSettings}
            <div class="ai-settings">
              <div class="settings-row">
                <label class="field">
                  <span class="field-label">Provider</span>
                  <select bind:value={editProvider} class="kf-select">
                    <option value="ollama">Ollama</option>
                    <option value="openai">OpenAI-compatible (llama-server / LM Studio)</option>
                  </select>
                </label>
                <label class="field">
                  <span class="field-label">Model</span>
                  <input bind:value={editModel} class="kf-input" type="text" placeholder="qwen2.5:1.5b" />
                </label>
              </div>
              <div class="settings-row">
                <label class="field">
                  <span class="field-label">Local Endpoint</span>
                  <input
                    bind:value={editEndpoint}
                    class="kf-input"
                    type="url"
                    placeholder="http://127.0.0.1:11434"
                  />
                </label>
                <button type="button" class="kf-btn kf-btn-primary save-ai-btn" onclick={saveSettings}>
                  Save Model
                </button>
              </div>
              {#if aiStatus?.message}
                <p class="ai-hint">{aiStatus.message}</p>
              {/if}
            </div>
          {/if}
        </div>

        <footer class="modal-footer">
          <span class="capture-hint">
            <span class="kf-kbd">⌘Enter</span> to distill →
            <strong>{selectedColumn}</strong>{selectedPriority ? ` · ${selectedPriority}` : ''}
          </span>
          <div class="footer-actions">
            <button type="button" class="kf-btn" onclick={() => dialogEl?.close()}>Cancel</button>
            <button
              type="submit"
              class="kf-btn kf-btn-primary"
              disabled={processing || !rawContext.trim()}
            >
              {#if processing}
                ✦ Distilling…
              {:else}
                ✦ Capture
              {/if}
            </button>
          </div>
        </footer>
      </form>
    </div>
  {/key}
</dialog>

<style>
  .capture-stage-dialog {
    padding: 0;
    border: none;
    background: transparent;
    color: var(--kf-ivory);
    max-width: min(860px, 96vw);
    width: 100%;
    max-height: 92vh;
    overflow: visible;
    box-shadow: none;
  }

  .capture-stage-dialog[open] {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .capture-stage-dialog::backdrop {
    background:
      radial-gradient(ellipse at 50% 50%, rgba(18, 12, 22, 0.56), rgba(4, 10, 14, 0.78));
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
  }

  .capture-cosmos {
    position: relative;
    width: 100%;
    max-width: 820px;
    min-height: 460px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 58px 124px;
  }

  /* Responsive collapse of orbital rails on narrower windows */
  @media (max-width: 780px) {
    .capture-cosmos {
      flex-direction: column;
      padding: 24px 16px;
      gap: 12px;
      min-height: 0;
    }
    .orbital-rail {
      position: static !important;
      transform: none !important;
      flex-direction: row !important;
      flex-wrap: wrap;
      justify-content: center;
    }
  }

  .ornate-ring-svg {
    position: absolute;
    inset: -24px -12px;
    width: calc(100% + 24px);
    height: calc(100% + 48px);
    pointer-events: none;
    z-index: 0;
  }

  /* Self-building ring animations (default static visible state; opt-in motion) */
  .draw-stroke {
    stroke-dasharray: 1000;
    stroke-dashoffset: 0;
  }
  .s3-c {
    stroke-dasharray: 240;
    stroke-dashoffset: 0;
  }

  @media (prefers-reduced-motion: no-preference) {
    .capture-cosmos:not(.ring-paused) .ring-glow {
      animation: capGlowIn 420ms ease-out both;
    }
    .capture-cosmos:not(.ring-paused) .s1 {
      animation: capDraw 340ms cubic-bezier(0.22, 1, 0.36, 1) both;
    }
    .capture-cosmos:not(.ring-paused) .s1-b {
      animation: capFadeScale 420ms ease-out 80ms both;
    }
    .capture-cosmos:not(.ring-paused) .sun-ray {
      animation: capRayBloom 360ms ease-out var(--ray-delay, 100ms) both;
    }
    .capture-cosmos:not(.ring-paused) .s3 {
      animation: capDraw 560ms cubic-bezier(0.22, 1, 0.36, 1) 140ms both;
    }
    .capture-cosmos:not(.ring-paused) .s3-b {
      animation: capDraw 620ms cubic-bezier(0.22, 1, 0.36, 1) 200ms both;
    }
    .capture-cosmos:not(.ring-paused) .s3-c {
      animation: capDrawShort 460ms ease-out 260ms both;
    }
    .capture-cosmos:not(.ring-paused) .zodiac-bead {
      animation: capBeadPop 320ms cubic-bezier(0.16, 1, 0.3, 1) var(--bead-delay, 240ms) both;
    }
    .capture-cosmos:not(.ring-paused) .orbital-node {
      animation: capNodeBloom 340ms cubic-bezier(0.16, 1, 0.3, 1) var(--node-delay, 260ms) both;
    }
    .capture-cosmos:not(.ring-paused) .capture-cartouche {
      animation: capCartoucheSnap 140ms cubic-bezier(0.16, 1, 0.3, 1) both;
    }
    .capture-cosmos.is-processing:not(.ring-paused) .layer-rays {
      animation: capAlchemicalSpin 18s linear infinite;
    }
  }

  @keyframes capGlowIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  @keyframes capDraw {
    from { stroke-dashoffset: 1000; opacity: 0.3; }
    to { stroke-dashoffset: 0; opacity: 1; }
  }
  @keyframes capDrawShort {
    from { stroke-dashoffset: 240; opacity: 0; }
    to { stroke-dashoffset: 0; opacity: 1; }
  }
  @keyframes capFadeScale {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  @keyframes capRayBloom {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  @keyframes capBeadPop {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  @keyframes capNodeBloom {
    from { opacity: 0; transform: scale(0.82); }
    to { opacity: 1; transform: scale(1); }
  }
  @keyframes capCartoucheSnap {
    from { opacity: 0; transform: scale(0.96); }
    to { opacity: 1; transform: scale(1); }
  }
  @keyframes capAlchemicalSpin {
    from { transform: translate(410px, 310px) rotate(0deg); }
    to { transform: translate(410px, 310px) rotate(360deg); }
  }

  /* Functional orbital node rails flanking the capture cartouche */
  .orbital-rail {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    display: flex;
    flex-direction: column;
    gap: 8px;
    z-index: 3;
  }

  .orbital-left {
    left: 6px;
    align-items: flex-end;
  }

  .orbital-right {
    right: 6px;
    align-items: flex-start;
  }

  .orbital-heading {
    font-family: var(--font-mono);
    font-size: 9.5px;
    text-transform: uppercase;
    letter-spacing: 0.16em;
    color: var(--kf-ochre);
    opacity: 0.88;
    padding: 0 6px;
  }

  .orbital-node {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 11px;
    border-radius: 999px;
    border: 1.5px solid rgba(246, 211, 136, 0.56);
    background: linear-gradient(160deg, rgba(16, 40, 46, 0.94), rgba(8, 20, 24, 0.96));
    color: var(--kf-ivory-soft);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    box-shadow:
      0 6px 18px rgba(0, 0, 0, 0.48),
      inset 0 1px 0 rgba(246, 211, 136, 0.28);
    transition: border-color 140ms ease, background 140ms ease, transform 120ms ease, box-shadow 140ms ease;
  }

  .orbital-node:hover {
    border-color: var(--kf-ochre);
    color: var(--kf-ivory);
    transform: scale(1.04);
  }

  .orbital-node:focus-visible {
    outline: 2.5px solid var(--kf-cerulean);
    outline-offset: 2px;
  }

  .orbital-node.selected {
    border-color: var(--kf-cerulean);
    background: linear-gradient(135deg, rgba(30, 168, 184, 0.45), rgba(150, 38, 48, 0.34));
    color: #ffffff;
    box-shadow:
      0 0 18px rgba(70, 217, 236, 0.36),
      inset 0 1px 0 rgba(246, 211, 136, 0.5);
  }

  .node-glyph {
    color: var(--prio-color, var(--kf-ochre));
    font-size: 11px;
  }

  /* Center Illuminated Mucha Cartouche */
  .capture-cartouche {
    position: relative;
    z-index: 2;
    width: 100%;
    max-width: 550px;
    display: flex;
    flex-direction: column;
    border-radius: 26px 26px 20px 20px;
    border: 2px solid var(--kf-lead-gold-bright);
    background:
      radial-gradient(ellipse at 20% 0%, rgba(70, 217, 236, 0.2), transparent 55%),
      radial-gradient(ellipse at 85% 100%, rgba(150, 38, 48, 0.22), transparent 55%),
      linear-gradient(165deg, rgba(14, 36, 42, 0.97) 0%, rgba(7, 19, 24, 0.99) 100%);
    box-shadow:
      0 26px 68px rgba(0, 0, 0, 0.78),
      0 0 28px rgba(220, 162, 72, 0.18),
      inset 0 0 0 4px rgba(220, 162, 72, 0.14),
      inset 0 0 0 6px rgba(150, 38, 48, 0.18),
      inset 0 1px 0 rgba(246, 211, 136, 0.48);
    overflow: hidden;
  }

  .cartouche-crown {
    width: 100%;
    height: 18px;
    display: block;
    flex-shrink: 0;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 4px 20px 12px;
    border-bottom: 1px solid rgba(246, 211, 136, 0.28);
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .modal-eyebrow {
    font-family: var(--font-mono);
    font-size: 10.5px;
    color: var(--kf-cerulean);
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .modal-title {
    margin: 2px 0 0;
    font-family: var(--font-display);
    font-size: 21px;
    font-weight: 600;
    color: var(--kf-ochre);
    letter-spacing: 0.06em;
  }

  .ai-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    border-radius: 999px;
    border: 1px solid rgba(217, 161, 74, 0.42);
    background: rgba(8, 22, 26, 0.72);
    color: var(--kf-ivory-soft);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
    transition: border-color 140ms ease, color 140ms ease;
  }

  .ai-pill:hover {
    border-color: var(--kf-ochre);
    color: var(--kf-ivory);
  }

  .ai-pill.online {
    border-color: rgba(65, 211, 230, 0.62);
    color: #b7eef5;
  }

  .ai-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--kf-gold);
  }

  .ai-pill.online .ai-dot {
    background: var(--kf-cerulean);
    box-shadow: 0 0 8px var(--kf-cerulean);
  }

  .ring-pause-btn {
    padding: 4px 8px;
    font-size: 11px;
    color: var(--kf-ochre);
  }

  .modal-body {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 14px 20px;
  }

  .context-box {
    min-height: 130px;
    font-size: 14px;
    line-height: 1.55;
    padding: 12px 14px;
    border-color: rgba(246, 211, 136, 0.44);
  }

  .ai-settings {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px;
    border-radius: 10px;
    border: 1px solid rgba(217, 161, 74, 0.28);
    background: rgba(6, 17, 20, 0.72);
  }

  .settings-row {
    display: flex;
    gap: 10px;
    align-items: flex-end;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
  }

  .field-label {
    font-size: 10.5px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--kf-ivory-dim);
  }

  .save-ai-btn {
    height: 36px;
    flex-shrink: 0;
  }

  .ai-hint {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--kf-ivory-dim);
  }

  .modal-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 20px 16px;
    border-top: 1px solid rgba(246, 211, 136, 0.24);
  }

  .capture-hint {
    font-size: 11.5px;
    color: var(--kf-ivory-dim);
  }

  .capture-hint strong {
    color: var(--kf-ochre);
    font-family: var(--font-mono);
    font-weight: 500;
  }

  .footer-actions {
    display: flex;
    gap: 10px;
  }
</style>
