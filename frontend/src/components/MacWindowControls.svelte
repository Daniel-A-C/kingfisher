<script lang="ts">
  interface Props {
    onHide: () => void;
    onMinimize: () => void;
    onMaximize: () => void;
  }

  let { onHide, onMinimize, onMaximize }: Props = $props();
</script>

<div
  class="mac-cutout-zone"
  style="--wails-draggable: no-drag"
  role="group"
  aria-label="Window controls"
>
  <div class="mac-cutout">
    <!-- Subtle Art Nouveau gilded corner filigree inside the rectangular cutout -->
    <svg class="cutout-filigree" viewBox="0 0 74 28" preserveAspectRatio="none" aria-hidden="true">
      <path
        d="M2 24 C2 26, 4 26, 6 26 L62 26 C68 26, 72 22, 72 16 L72 2"
        fill="none"
        stroke="rgba(246, 211, 136, 0.78)"
        stroke-width="1.2"
      />
      <path
        d="M54 24 C61 24, 66 19, 68 12"
        fill="none"
        stroke="rgba(70, 217, 236, 0.55)"
        stroke-width="0.9"
      />
      <circle cx="67" cy="21" r="1.4" fill="#f6d388" />
    </svg>

    <button
      type="button"
      class="mac-jewel jewel-close"
      title="Hide Kingfisher (Esc)"
      aria-label="Close or hide window"
      onclick={onHide}
    >
      <svg viewBox="0 0 12 12" aria-hidden="true">
        <path d="M3.5 3.5 L8.5 8.5 M8.5 3.5 L3.5 8.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
      </svg>
    </button>

    <button
      type="button"
      class="mac-jewel jewel-min"
      title="Minimize window"
      aria-label="Minimize window"
      onclick={onMinimize}
    >
      <svg viewBox="0 0 12 12" aria-hidden="true">
        <path d="M3 6 H9" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
      </svg>
    </button>

    <button
      type="button"
      class="mac-jewel jewel-max"
      title="Toggle maximize"
      aria-label="Toggle maximize window"
      onclick={onMaximize}
    >
      <svg viewBox="0 0 12 12" aria-hidden="true">
        <path d="M3.5 8.5 L8.5 3.5 M5.5 3.5 H8.5 V6.5 M6.5 8.5 H3.5 V5.5" fill="none" stroke="currentColor" stroke-width="1.25" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>
  </div>
</div>

<style>
  /* Top-left hover sensor region aligned inside the native rounded corner */
  .mac-cutout-zone {
    position: absolute;
    top: 0;
    left: 0;
    width: 82px;
    height: 36px;
    z-index: 45;
    display: flex;
    align-items: flex-start;
    justify-content: flex-start;
    padding: 0;
  }

  /* Tiny rectangular Art Nouveau stained-glass cutout revealed on hover or focus */
  .mac-cutout {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    width: 74px;
    height: 28px;
    padding: 0 9px 1px 8px;
    border-radius: 9px 0 11px 0;
    background:
      radial-gradient(ellipse at 20% 0%, rgba(70, 217, 236, 0.24), transparent 70%),
      linear-gradient(155deg, rgba(16, 46, 48, 0.94) 0%, rgba(7, 21, 22, 0.96) 100%);
    border-right: 1.5px solid var(--kf-lead-gold-bright);
    border-bottom: 1.5px solid var(--kf-lead-gold-bright);
    box-shadow:
      0 6px 18px rgba(0, 0, 0, 0.52),
      0 0 12px rgba(220, 162, 72, 0.2),
      inset 0 0 0 1px rgba(246, 211, 136, 0.22);
    opacity: 0;
    transform: translate(-4px, -4px) scale(0.94);
    transform-origin: top left;
    transition: opacity 160ms ease-out, transform 180ms ease-out;
  }

  .mac-cutout-zone:hover .mac-cutout,
  .mac-cutout-zone:focus-within .mac-cutout {
    opacity: 1;
    transform: translate(0, 0) scale(1);
  }

  @media (hover: none) {
    .mac-cutout {
      opacity: 1;
      transform: none;
    }
  }

  .cutout-filigree {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
  }

  .mac-jewel {
    position: relative;
    z-index: 1;
    width: 13px;
    height: 13px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid rgba(246, 211, 136, 0.72);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    color: rgba(12, 22, 22, 0.88);
    box-shadow:
      inset 0 1px 1px rgba(255, 255, 255, 0.42),
      0 1px 3px rgba(0, 0, 0, 0.55);
    transition: transform 120ms ease-out, box-shadow 140ms ease-out, filter 140ms ease-out;
  }

  .mac-jewel svg {
    width: 9px;
    height: 9px;
    opacity: 0;
    transition: opacity 120ms ease-out;
  }

  .mac-cutout:hover .mac-jewel svg,
  .mac-cutout:focus-within .mac-jewel svg {
    opacity: 0.92;
  }

  .mac-jewel:hover {
    filter: brightness(1.14);
    box-shadow:
      inset 0 1px 1px rgba(255, 255, 255, 0.55),
      0 0 8px rgba(246, 211, 136, 0.48);
  }

  .mac-jewel:active {
    transform: scale(0.88);
  }

  .mac-jewel:focus-visible {
    outline: 2px solid var(--kf-cerulean);
    outline-offset: 2px;
  }

  /* Kingfisher stained-glass jewel variants of macOS traffic lights */
  .jewel-close {
    background: radial-gradient(circle at 35% 30%, #ff8b6b, #e65d38 65%, #a8371c 100%);
  }

  .jewel-min {
    background: radial-gradient(circle at 35% 30%, #ffe39c, #dca248 65%, #9d712d 100%);
  }

  .jewel-max {
    background: radial-gradient(circle at 35% 30%, #7ef0db, #2f9e7a 65%, #166650 100%);
  }
</style>
