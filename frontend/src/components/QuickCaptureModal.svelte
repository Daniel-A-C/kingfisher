<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api';
  import type { AIConfig, AIStatus } from '../lib/types';
  import { attachLightDismissFallback } from '../lib/dialog';

  interface Props {
    open: boolean;
    defaultStatus: string;
    projectName: string;
    aiConfig: AIConfig;
    onClose: () => void;
    onCapture: (rawContext: string, defaultStatus: string) => Promise<void>;
    onSaveAIConfig: (cfg: AIConfig) => Promise<void>;
  }

  let {
    open,
    defaultStatus,
    projectName,
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
  let aiStatus = $state<AIStatus | null>(null);

  let editProvider = $state('ollama');
  let editEndpoint = $state('http://127.0.0.1:11434');
  let editModel = $state('qwen2.5:1.5b');
  let editEnabled = $state(true);

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
      editProvider = aiConfig.provider || 'ollama';
      editEndpoint = aiConfig.endpoint || 'http://127.0.0.1:11434';
      editModel = aiConfig.model || 'qwen2.5:1.5b';
      editEnabled = aiConfig.enabled ?? true;
      dialogEl.showModal();
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
      await onCapture(trimmed, defaultStatus);
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
  class="kf-dialog capture-dialog"
  aria-labelledby="capture-title"
  {...({ closedby: 'any' } as Record<string, string>)}
  onclose={handleClose}
>
  <form class="capture-form" onsubmit={handleFormSubmit}>
    <header class="modal-header">
      <div>
        <span class="modal-eyebrow">✦ {projectName}</span>
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
        <button type="button" class="kf-btn" onclick={() => dialogEl?.close()}>Esc</button>
      </div>
    </header>

    <div class="modal-body">
      <textarea
        bind:this={textareaEl}
        bind:value={rawContext}
        class="kf-textarea context-box"
        rows="5"
        disabled={processing}
        placeholder="Dump your raw idea, bug, or context here…&#10;&#10;The local model will automatically generate a clean title, column, priority, tags, and formatted description."
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
        <span class="kf-kbd">⌘Enter</span> to distill &amp; file
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
</dialog>

<style>
  .capture-dialog {
    max-width: min(580px, 92vw);
  }

  .capture-form {
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 15px 20px 12px;
    border-bottom: 1px solid rgba(217, 161, 74, 0.26);
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .modal-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--kf-cerulean);
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .modal-title {
    margin: 2px 0 0;
    font-family: var(--font-display);
    font-size: 20px;
    font-weight: 600;
    color: var(--kf-ochre);
    letter-spacing: 0.05em;
  }

  .ai-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    border-radius: 999px;
    border: 1px solid rgba(217, 161, 74, 0.36);
    background: rgba(8, 22, 22, 0.65);
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
    border-color: rgba(65, 211, 230, 0.55);
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

  .modal-body {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px 20px;
  }

  .context-box {
    min-height: 132px;
    font-size: 14px;
    line-height: 1.55;
    padding: 12px 14px;
  }

  .ai-settings {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px;
    border-radius: 10px;
    border: 1px solid rgba(217, 161, 74, 0.28);
    background: rgba(6, 17, 18, 0.65);
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
    border-top: 1px solid rgba(217, 161, 74, 0.22);
  }

  .capture-hint {
    font-size: 11.5px;
    color: var(--kf-ivory-dim);
  }

  .footer-actions {
    display: flex;
    gap: 10px;
  }
</style>
