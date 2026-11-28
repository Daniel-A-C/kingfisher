<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api';
  import type { AIStatus, Preferences } from '../lib/types';
  import { attachLightDismissFallback } from '../lib/dialog';

  interface Props {
    open: boolean;
    preferences: Preferences;
    onClose: () => void;
    onLivePreview: (partial: Partial<Preferences>) => void;
    onSave: (prefs: Preferences) => Promise<void>;
  }

  let { open, preferences, onClose, onLivePreview, onSave }: Props = $props();

  let dialogEl: HTMLDialogElement | undefined = $state();
  let aiStatus = $state<AIStatus | null>(null);

  let opacity = $state(0.35);
  let blur = $state(0.25);
  let highContrast = $state(false);
  let reduceMotion = $state(false);
  let scanDirsRaw = $state('~/Documents/programs');

  let aiEnabled = $state(true);
  let aiProvider = $state('ollama');
  let aiEndpoint = $state('http://127.0.0.1:11434');
  let aiModel = $state('qwen2.5:1.5b');
  let saving = $state(false);
  let initialOpacity = 0.35;
  let initialBlur = 0.25;
  let initialHighContrast = false;
  let initialReduceMotion = false;
  let didSave = false;

  onMount(() => {
    if (!dialogEl) return;
    return attachLightDismissFallback(dialogEl);
  });

  $effect(() => {
    if (!dialogEl) return;
    if (open && !dialogEl.open) {
      opacity = preferences.opacity;
      blur = preferences.blur ?? 0.25;
      highContrast = preferences.high_contrast;
      reduceMotion = preferences.reduce_motion;
      initialOpacity = opacity;
      initialBlur = blur;
      initialHighContrast = highContrast;
      initialReduceMotion = reduceMotion;
      didSave = false;
      scanDirsRaw = (preferences.scan_dirs ?? []).join('\n');
      aiEnabled = preferences.ai?.enabled ?? true;
      aiProvider = preferences.ai?.provider || 'ollama';
      aiEndpoint = preferences.ai?.endpoint || 'http://127.0.0.1:11434';
      aiModel = preferences.ai?.model || 'qwen2.5:1.5b';
      dialogEl.showModal();
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
    if (!didSave) {
      onLivePreview({
        opacity: initialOpacity,
        blur: initialBlur,
        high_contrast: initialHighContrast,
        reduce_motion: initialReduceMotion,
      });
    }
    if (open) onClose();
  }

  function previewOpacity(val: number) {
    opacity = val;
    onLivePreview({ opacity: val });
  }

  function previewBlur(val: number) {
    blur = val;
    onLivePreview({ blur: val });
  }

  function previewHighContrast(val: boolean) {
    highContrast = val;
    onLivePreview({ high_contrast: val });
  }

  function previewReduceMotion(val: boolean) {
    reduceMotion = val;
    onLivePreview({ reduce_motion: val });
  }

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    saving = true;
    try {
      const scan_dirs = scanDirsRaw
        .split('\n')
        .map((s) => s.trim())
        .filter(Boolean);
      await onSave({
        opacity,
        blur,
        high_contrast: highContrast,
        reduce_motion: reduceMotion,
        show_header_opacity: false,
        scan_dirs,
        ai: {
          enabled: aiEnabled,
          provider: aiProvider,
          endpoint: aiEndpoint.trim(),
          model: aiModel.trim(),
        },
      });
      didSave = true;
      dialogEl?.close();
    } finally {
      saving = false;
    }
  }
</script>

<dialog
  bind:this={dialogEl}
  class="kf-dialog settings-dialog"
  aria-labelledby="settings-title"
  {...({ closedby: 'any' } as Record<string, string>)}
  onclose={handleClose}
>
  <form class="settings-form" onsubmit={handleSubmit}>
    <header class="modal-header">
      <div>
        <span class="modal-eyebrow">✦ Atelier Preferences</span>
        <h1 id="settings-title" class="modal-title">Settings &amp; Accessibility</h1>
      </div>
      <button type="button" class="kf-btn" onclick={() => dialogEl?.close()}>Esc</button>
    </header>

    <div class="settings-body kf-scroll">
      <!-- Accessibility & Stained-Glass Transparency -->
      <section class="settings-section">
        <h2 class="section-heading">◈ Stained-Glass Transparency &amp; Accessibility</h2>

        <div class="opacity-block">
          <div class="opacity-label-row">
            <label for="settings-opacity-range" class="field-label">
              Window Glass Opacity
            </label>
            <span class="opacity-readout">
              {opacity <= 0.001 ? '0% · Fully Transparent' : `${Math.round(opacity * 100)}%`}
            </span>
          </div>
          <input
            id="settings-opacity-range"
            type="range"
            min="0"
            max="1"
            step="0.01"
            value={opacity}
            class="settings-range"
            oninput={(e) => previewOpacity(parseFloat((e.currentTarget as HTMLInputElement).value))}
          />
          <div class="preset-row">
            <button
              type="button"
              class="kf-btn preset-btn"
              class:active-preset={opacity <= 0.001}
              onclick={() => previewOpacity(0)}
            >
              Clear (0%)
            </button>
            <button
              type="button"
              class="kf-btn preset-btn"
              class:active-preset={opacity > 0.001 && opacity <= 0.15}
              onclick={() => previewOpacity(0.1)}
            >
              Whisper (10%)
            </button>
            <button
              type="button"
              class="kf-btn preset-btn"
              class:active-preset={opacity > 0.15 && opacity <= 0.35}
              onclick={() => previewOpacity(0.25)}
            >
              Light (25%)
            </button>
            <button
              type="button"
              class="kf-btn preset-btn"
              class:active-preset={opacity > 0.35 && opacity <= 0.65}
              onclick={() => previewOpacity(0.55)}
            >
              Stained (55%)
            </button>
            <button
              type="button"
              class="kf-btn preset-btn"
              class:active-preset={opacity >= 0.8}
              onclick={() => previewOpacity(0.85)}
            >
              Deep (85%)
            </button>
          </div>
          <p class="opacity-help">
            Scales smoothly in 1% steps from completely clear glass (0%) to solid stained glass (100%).
          </p>
        </div>

        <div class="opacity-block blur-block">
          <div class="opacity-label-row">
            <label for="settings-blur-range" class="field-label">
              Background Blur
            </label>
            <span class="opacity-readout">
              {blur <= 0.001 ? '0% · Sharp' : `${Math.round(blur * 100)}%`}
            </span>
          </div>
          <input
            id="settings-blur-range"
            type="range"
            min="0"
            max="1"
            step="0.01"
            value={blur}
            class="settings-range"
            oninput={(e) => previewBlur(parseFloat((e.currentTarget as HTMLInputElement).value))}
          />
          <div class="preset-row">
            <button
              type="button"
              class="kf-btn preset-btn"
              class:active-preset={blur <= 0.001}
              onclick={() => previewBlur(0)}
            >
              Sharp (0%)
            </button>
            <button
              type="button"
              class="kf-btn preset-btn"
              class:active-preset={blur > 0.001 && blur <= 0.35}
              onclick={() => previewBlur(0.25)}
            >
              Soft (25%)
            </button>
            <button
              type="button"
              class="kf-btn preset-btn"
              class:active-preset={blur > 0.35 && blur <= 0.7}
              onclick={() => previewBlur(0.55)}
            >
              Frosted (55%)
            </button>
            <button
              type="button"
              class="kf-btn preset-btn"
              class:active-preset={blur > 0.7}
              onclick={() => previewBlur(1)}
            >
              Deep Blur (100%)
            </button>
          </div>
          <p class="opacity-help">
            Controls behind-window macOS blur independently of opacity (just like iTerm2).
          </p>
        </div>

        <label class="toggle-row">
          <input
            type="checkbox"
            checked={highContrast}
            onchange={(e) => previewHighContrast((e.currentTarget as HTMLInputElement).checked)}
          />
          <span class="toggle-copy">
            <strong>High-Contrast Typography &amp; Leaded Borders</strong>
            <small>Strengthens text halos, ivory brightness, and gilded borders for bright wallpapers.</small>
          </span>
        </label>

        <label class="toggle-row">
          <input
            type="checkbox"
            checked={reduceMotion}
            onchange={(e) => previewReduceMotion((e.currentTarget as HTMLInputElement).checked)}
          />
          <span class="toggle-copy">
            <strong>Reduce Motion</strong>
            <small>Disables card hover lifts and transitions across the board.</small>
          </span>
        </label>
      </section>

      <!-- Local AI Capture -->
      <section class="settings-section">
        <div class="section-header-row">
          <h2 class="section-heading">❧ Local Open-Source AI Capture</h2>
          {#if aiStatus}
            <span class="status-badge" class:online={aiStatus.available}>
              {aiStatus.available ? `Online · ${aiStatus.model}` : 'Offline · Fallback'}
            </span>
          {/if}
        </div>

        <label class="toggle-row">
          <input type="checkbox" bind:checked={aiEnabled} />
          <span class="toggle-copy">
            <strong>Enable Local LLM Triage on Capture</strong>
            <small>Automatically extracts title, column, priority, tags, and formatting.</small>
          </span>
        </label>

        <div class="field-row">
          <label class="field">
            <span class="field-label">Provider</span>
            <select bind:value={aiProvider} class="kf-select">
              <option value="ollama">Ollama</option>
              <option value="openai">OpenAI-compatible (llama-server / LM Studio)</option>
            </select>
          </label>
          <label class="field">
            <span class="field-label">Model</span>
            <input bind:value={aiModel} class="kf-input" type="text" placeholder="qwen2.5:1.5b" />
          </label>
        </div>

        <label class="field">
          <span class="field-label">Local Endpoint</span>
          <input
            bind:value={aiEndpoint}
            class="kf-input"
            type="url"
            placeholder="http://127.0.0.1:11434"
          />
        </label>
      </section>

      <!-- Project Auto-Discovery -->
      <section class="settings-section">
        <h2 class="section-heading">❀ Project Discovery Directories</h2>
        <label class="field">
          <span class="field-label">Root Folders to Scan for .kingfisher.yaml (one per line)</span>
          <textarea
            bind:value={scanDirsRaw}
            class="kf-textarea"
            rows="2"
            placeholder="~/Documents/programs"
          ></textarea>
        </label>
      </section>
    </div>

    <footer class="modal-footer">
      <span class="shortcut-note"><span class="kf-kbd">⌘,</span> Settings · <span class="kf-kbd">⌘⇧T</span> Cycle Glass</span>
      <div class="footer-actions">
        <button type="button" class="kf-btn" onclick={() => dialogEl?.close()}>Close</button>
        <button type="submit" class="kf-btn kf-btn-primary" disabled={saving}>
          {saving ? 'Saving…' : '✦ Save Preferences'}
        </button>
      </div>
    </footer>
  </form>
</dialog>

<style>
  .settings-dialog {
    max-width: min(620px, 92vw);
  }

  .settings-form {
    display: flex;
    flex-direction: column;
    max-height: 84vh;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 22px 12px;
    border-bottom: 1px solid rgba(217, 161, 74, 0.35);
  }

  .modal-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--kf-cerulean);
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }

  .modal-title {
    margin: 2px 0 0;
    font-family: var(--font-display);
    font-size: 21px;
    color: var(--kf-ochre);
    letter-spacing: 0.05em;
  }

  .settings-body {
    display: flex;
    flex-direction: column;
    gap: 18px;
    padding: 16px 22px;
    overflow-y: auto;
  }

  .settings-section {
    display: flex;
    flex-direction: column;
    gap: 11px;
    padding: 14px 16px;
    border-radius: 14px;
    border: 1px solid rgba(217, 161, 74, 0.28);
    background: rgba(7, 19, 20, 0.56);
    box-shadow: inset 0 1px 0 rgba(242, 204, 126, 0.12);
  }

  .section-header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .section-heading {
    margin: 0;
    font-family: var(--font-display);
    font-size: 15px;
    letter-spacing: 0.05em;
    color: var(--kf-ochre);
  }

  .status-badge {
    font-family: var(--font-mono);
    font-size: 10.5px;
    padding: 2px 9px;
    border-radius: 999px;
    border: 1px solid rgba(217, 161, 74, 0.45);
    color: var(--kf-ivory-soft);
  }

  .status-badge.online {
    border-color: var(--kf-cerulean);
    color: #b7eef5;
    background: rgba(30, 168, 184, 0.18);
  }

  .opacity-block {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .opacity-label-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .opacity-readout {
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--kf-ochre);
  }

  .settings-range {
    width: 100%;
    accent-color: var(--kf-cerulean);
    cursor: pointer;
  }

  .preset-row {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .preset-btn {
    flex: 1;
    font-size: 11.5px;
    padding: 6px 10px;
  }

  .preset-btn.active-preset {
    border-color: var(--kf-cerulean);
    background: rgba(30, 168, 184, 0.28);
    color: #ffffff;
  }

  .opacity-help {
    margin: 0;
    font-size: 11.5px;
    color: var(--kf-ivory-dim);
    line-height: 1.4;
  }

  .toggle-row {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    cursor: pointer;
    padding: 4px 2px;
  }

  .toggle-row input[type='checkbox'] {
    margin-top: 3px;
    accent-color: var(--kf-cerulean);
    width: 15px;
    height: 15px;
    cursor: pointer;
  }

  .toggle-copy {
    display: flex;
    flex-direction: column;
    gap: 2px;
    font-size: 13px;
    color: var(--kf-ivory);
  }

  .toggle-copy small {
    font-size: 11.5px;
    color: var(--kf-ivory-dim);
  }

  .field-row {
    display: flex;
    gap: 10px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
  }

  .field-label {
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--kf-ivory-dim);
  }

  .modal-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 22px 16px;
    border-top: 1px solid rgba(217, 161, 74, 0.28);
  }

  .shortcut-note {
    font-size: 11.5px;
    color: var(--kf-ivory-dim);
  }

  .footer-actions {
    display: flex;
    gap: 10px;
  }
</style>
