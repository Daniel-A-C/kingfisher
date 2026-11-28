<script lang="ts">
  import { onMount } from 'svelte';
  import { api, subscribeEvents } from './lib/api';
  import type {
    AIConfig,
    AppState,
    Preferences,
    ProjectOverview,
    Task,
    TaskPatch,
  } from './lib/types';
  import TabBar from './components/TabBar.svelte';
  import MacWindowControls from './components/MacWindowControls.svelte';
  import KanbanBoard from './components/KanbanBoard.svelte';
  import HomeView from './components/HomeView.svelte';
  import QuickCaptureModal from './components/QuickCaptureModal.svelte';
  import TaskDetailDrawer from './components/TaskDetailDrawer.svelte';
  import ProjectSwitcherModal from './components/ProjectSwitcherModal.svelte';
  import SettingsModal from './components/SettingsModal.svelte';

  let appState = $state<AppState>({
    projects: [],
    open_tabs: [],
    active_tab: '',
    boards: {},
    errors: {},
    preferences: {
      opacity: 0.55,
      high_contrast: false,
      reduce_motion: false,
      show_header_opacity: true,
      scan_dirs: ['~/Documents/programs'],
      ai: {
        enabled: true,
        provider: 'ollama',
        endpoint: 'http://127.0.0.1:11434',
        model: 'qwen2.5:1.5b',
      },
    },
  });

  let switcherOpen = $state(false);
  let settingsOpen = $state(false);
  let captureOpen = $state(false);
  let homeOpen = $state(false);
  let projectOverviews = $state<ProjectOverview[]>([]);
  let overviewsLoading = $state(false);
  let captureColumn = $state('ready');
  let selectedTaskID = $state<string | null>(null);
  let toastMessage = $state<string | null>(null);
  let toastTimer: number | undefined;
  let summonState = $state<'idle' | 'activate' | 'raycast'>('idle');
  let summonTimer: number | undefined;

  function triggerSummonBloom(source: string) {
    summonState = 'idle';
    if (summonTimer) window.clearTimeout(summonTimer);
    requestAnimationFrame(() => {
      summonState = source === 'raycast' ? 'raycast' : 'activate';
      summonTimer = window.setTimeout(() => {
        summonState = 'idle';
      }, 820);
    });
  }

  let activeBoard = $derived(
    appState.active_tab ? appState.boards[appState.active_tab] ?? null : null
  );
  let showingHome = $derived(homeOpen || (!activeBoard && appState.projects.length > 0));
  let activeError = $derived(
    appState.active_tab ? appState.errors[appState.active_tab] : undefined
  );
  let selectedTask = $derived<Task | null>(
    activeBoard && selectedTaskID
      ? activeBoard.tasks.find((t) => t.id === selectedTaskID) ?? null
      : null
  );

  function showToast(msg: string) {
    toastMessage = msg;
    if (toastTimer) window.clearTimeout(toastTimer);
    toastTimer = window.setTimeout(() => {
      toastMessage = null;
    }, 3400);
  }

  function applyOpacityVar(val: number) {
    document.documentElement.style.setProperty('--kf-opacity', String(val));
  }

  async function refreshProjectOverviews() {
    overviewsLoading = true;
    try {
      projectOverviews = await api.getProjectOverviews();
    } catch (err) {
      showToast(`Could not load project telemetry: ${err}`);
    } finally {
      overviewsLoading = false;
    }
  }

  function openHomeMenu() {
    selectedTaskID = null;
    homeOpen = true;
    refreshProjectOverviews();
  }

  function toggleHomeMenu() {
    if (showingHome && activeBoard) {
      homeOpen = false;
    } else {
      openHomeMenu();
    }
  }

  onMount(() => {
    api
      .getState()
      .then((s) => {
        appState = s;
        applyOpacityVar(s.preferences.opacity);
        if (!s.active_tab && s.projects.length > 0) {
          refreshProjectOverviews();
        }
      })
      .catch((err) => showToast(String(err)));

    const unsub = subscribeEvents({
      onBoardUpdated: ({ path, board }) => {
        const nextErrors = { ...appState.errors };
        delete nextErrors[path];
        appState = {
          ...appState,
          boards: { ...appState.boards, [path]: board },
          errors: nextErrors,
        };
      },
      onBoardError: ({ path, error }) => {
        appState = {
          ...appState,
          errors: { ...appState.errors, [path]: error },
        };
      },
      onStateChanged: (state) => {
        appState = state;
        applyOpacityVar(state.preferences.opacity);
      },
      onWindowSummoned: ({ source }) => {
        triggerSummonBloom(source);
      },
    });

    const onKeyDown = (e: KeyboardEvent) => {
      const meta = e.metaKey || e.ctrlKey;
      const target = e.target as HTMLElement | null;
      const isTyping =
        target &&
        (target.tagName === 'INPUT' ||
          target.tagName === 'TEXTAREA' ||
          target.tagName === 'SELECT' ||
          target.isContentEditable);

      // Cmd + , -> Settings & Accessibility
      if (meta && !e.shiftKey && e.key === ',') {
        e.preventDefault();
        settingsOpen = true;
        return;
      }

      // Cmd + 0 or Cmd + Shift + H -> Toggle Home Menu
      if (
        (meta && !e.shiftKey && !e.altKey && e.key === '0') ||
        (meta && e.shiftKey && !e.altKey && e.key.toLowerCase() === 'h')
      ) {
        e.preventDefault();
        toggleHomeMenu();
        return;
      }

      // Cmd + 1..9 browser-style tab jumping
      if (meta && !e.shiftKey && !e.altKey && e.key >= '1' && e.key <= '9') {
        e.preventDefault();
        const tabs = appState.open_tabs;
        if (tabs.length === 0) return;
        const num = parseInt(e.key, 10);
        const idx = num === 9 ? tabs.length - 1 : num - 1;
        if (idx >= 0 && idx < tabs.length) {
          handleSelectTab(tabs[idx]);
        }
        return;
      }

      // Cmd + K or Cmd + T -> Quick Project Switcher
      if (meta && !e.shiftKey && (e.key.toLowerCase() === 'k' || e.key.toLowerCase() === 't')) {
        e.preventDefault();
        switcherOpen = true;
        return;
      }

      // Cmd + Shift + T -> Cycle stained-glass opacity presets
      if (meta && e.shiftKey && e.key.toLowerCase() === 't') {
        e.preventDefault();
        const cur = appState.preferences.opacity;
        const next = cur < 0.2 ? 0.55 : cur < 0.7 ? 0.85 : 0.0;
        handleOpacityChange(next);
        return;
      }

      // Cmd + N -> Project-Agnostic Capture
      if (meta && !e.shiftKey && e.key.toLowerCase() === 'n') {
        e.preventDefault();
        if (appState.projects.length > 0) {
          captureColumn = 'ready';
          captureOpen = true;
        } else {
          switcherOpen = true;
        }
        return;
      }

      // Cmd + W -> Close current tab (or hide if none)
      if (meta && !e.shiftKey && e.key.toLowerCase() === 'w') {
        e.preventDefault();
        if (appState.active_tab) {
          handleCloseTab(appState.active_tab);
        } else {
          api.hideWindow();
        }
        return;
      }

      // Single-key shortcuts when not typing and no modal open
      if (
        !meta &&
        !isTyping &&
        !switcherOpen &&
        !settingsOpen &&
        !captureOpen &&
        !selectedTaskID
      ) {
        if (e.key === 'n' && appState.projects.length > 0) {
          e.preventDefault();
          captureColumn = 'ready';
          captureOpen = true;
          return;
        }
        if (e.key === 'h') {
          e.preventDefault();
          toggleHomeMenu();
          return;
        }
        if (e.key === 'Escape') {
          e.preventDefault();
          if (homeOpen && activeBoard) {
            homeOpen = false;
          } else {
            api.hideWindow();
          }
        }
      }
    };

    window.addEventListener('keydown', onKeyDown);
    return () => {
      unsub();
      window.removeEventListener('keydown', onKeyDown);
    };
  });

  async function handleSelectTab(path: string) {
    selectedTaskID = null;
    homeOpen = false;
    appState = await api.setActiveTab(path);
  }

  async function handleOpenProject(path: string) {
    try {
      selectedTaskID = null;
      homeOpen = false;
      appState = await api.openProjectTab(path);
    } catch (err) {
      showToast(`Could not open project: ${err}`);
    }
  }

  async function handleCloseTab(path: string) {
    selectedTaskID = null;
    appState = await api.closeProjectTab(path);
    if (!appState.active_tab && appState.projects.length > 0) {
      refreshProjectOverviews();
    }
  }

  async function handleInitFolder() {
    try {
      const dir = await api.pickFolder();
      if (!dir) return;
      homeOpen = false;
      appState = await api.initProject(dir, '', '', false);
      showToast('Project ready (.kingfisher.yaml ignored via .git/info/exclude)');
    } catch (err) {
      showToast(`Init failed: ${err}`);
    }
  }

  async function handleRescan() {
    appState = await api.scanProjects();
    showToast(`Found ${appState.projects.length} Kingfisher project(s)`);
    if (showingHome) {
      await refreshProjectOverviews();
    }
  }

  async function handleOpacityChange(value: number) {
    applyOpacityVar(value);
    appState = {
      ...appState,
      preferences: { ...appState.preferences, opacity: value },
    };
    await api.setOpacity(value);
  }

  function handleLivePreviewPrefs(partial: Partial<Preferences>) {
    if (partial.opacity !== undefined) {
      applyOpacityVar(partial.opacity);
    }
    appState = {
      ...appState,
      preferences: { ...appState.preferences, ...partial },
    };
  }

  async function handleSavePreferences(prefs: Preferences) {
    applyOpacityVar(prefs.opacity);
    appState = await api.updatePreferences(prefs);
    showToast('✦ Preferences saved');
  }

  async function handleCaptureTask(rawContext: string, defaultStatus: string) {
    try {
      const res = await api.captureContext(rawContext, appState.active_tab, defaultStatus);
      appState = res.state;
      if (showingHome) {
        refreshProjectOverviews();
      }
      const engine = res.used_ai ? res.model || 'local AI' : 'heuristic fallback';
      if (res.tickets.length === 1) {
        const t = res.tickets[0];
        showToast(`✦ ${t.task.id} → ${t.project_name} / ${t.task.status} (${engine})`);
      } else {
        const summary = res.tickets.map((t) => `${t.task.id} (${t.project_name})`).join(', ');
        showToast(`✦ Filed ${res.tickets.length} tickets: ${summary}`);
      }
    } catch (err) {
      showToast(`Capture failed: ${err}`);
    }
  }

  async function handleSaveAIConfig(cfg: AIConfig) {
    const prefs = await api.updateAIConfig(cfg);
    appState = { ...appState, preferences: prefs };
  }

  async function handleSaveTask(taskID: string, patch: TaskPatch) {
    if (!appState.active_tab) return;
    try {
      const updated = await api.updateTask(appState.active_tab, taskID, patch);
      appState = {
        ...appState,
        boards: { ...appState.boards, [appState.active_tab]: updated },
      };
    } catch (err) {
      showToast(`Failed to update task: ${err}`);
    }
  }

  async function handleMoveTask(taskID: string, targetStatus: string, targetIndex: number) {
    if (!appState.active_tab) return;
    try {
      const updated = await api.moveTask(
        appState.active_tab,
        taskID,
        targetStatus,
        targetIndex
      );
      appState = {
        ...appState,
        boards: { ...appState.boards, [appState.active_tab]: updated },
      };
    } catch (err) {
      showToast(`Failed to move task: ${err}`);
    }
  }

  async function handleDeleteTask(taskID: string) {
    if (!appState.active_tab) return;
    try {
      const updated = await api.deleteTask(appState.active_tab, taskID);
      selectedTaskID = null;
      appState = {
        ...appState,
        boards: { ...appState.boards, [appState.active_tab]: updated },
      };
    } catch (err) {
      showToast(`Failed to delete task: ${err}`);
    }
  }
</script>

<main
  class="kf-shell"
  class:high-contrast={appState.preferences.high_contrast}
  class:reduce-motion={appState.preferences.reduce_motion}
  class:summoned={summonState !== 'idle'}
  class:summoned-raycast={summonState === 'raycast'}
>
  <!-- Horta Whiplash & Mucha Celestial Stained-Glass Backdrop Tapestry -->
  <svg class="kf-tapestry" viewBox="0 0 1400 900" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
    <defs>
      <radialGradient id="kfTapestryHalo" cx="50%" cy="46%" r="48%">
        <stop offset="0%" stop-color="rgba(150, 38, 48, 0.24)" />
        <stop offset="45%" stop-color="rgba(220, 162, 72, 0.1)" />
        <stop offset="78%" stop-color="rgba(30, 168, 184, 0.08)" />
        <stop offset="100%" stop-color="transparent" />
      </radialGradient>
    </defs>
    <!-- Central Mucha zodiacal watermark ring -->
    <circle cx="700" cy="430" r="340" fill="url(#kfTapestryHalo)" />
    <circle cx="700" cy="430" r="320" fill="none" stroke="rgba(246, 211, 136, 0.16)" stroke-width="1.4" />
    <circle cx="700" cy="430" r="292" fill="none" stroke="rgba(220, 162, 72, 0.12)" stroke-width="1" stroke-dasharray="6 8" />
    <circle cx="700" cy="430" r="248" fill="none" stroke="rgba(70, 217, 236, 0.1)" stroke-width="1" />
    <!-- Sweeping Horta whiplash mural tendrils -->
    <path
      d="M-20 760 C 180 620, 260 440, 420 490 C 560 530, 510 740, 700 680 C 890 620, 840 430, 1010 470 C 1170 510, 1240 660, 1420 590"
      fill="none"
      stroke="rgba(220, 162, 72, 0.14)"
      stroke-width="2.2"
    />
    <path
      d="M-20 260 C 200 180, 310 340, 480 230 C 620 140, 780 140, 920 230 C 1090 340, 1200 180, 1420 260"
      fill="none"
      stroke="rgba(70, 217, 236, 0.11)"
      stroke-width="1.6"
    />
  </svg>

  <!-- Raycast / Summon Entrance Sunburst Halo Bloom -->
  <svg class="kf-summon-bloom" viewBox="0 0 1200 800" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
    <g transform="translate(600 400)">
      <circle r="290" fill="none" stroke="#f6d388" stroke-width="2.5" stroke-dasharray="14 8" opacity="0.85" />
      <circle r="245" fill="none" stroke="#46d9ec" stroke-width="1.8" opacity="0.72" />
      <circle r="200" fill="none" stroke="#c83e47" stroke-width="1.6" stroke-dasharray="4 6" opacity="0.7" />
      {#each Array(12) as _, idx}
        <g transform="rotate({idx * 30})">
          <line x1="0" y1="-195" x2="0" y2="-330" stroke="#f6d388" stroke-width="1.4" opacity="0.6" />
          <circle cx="0" cy="-290" r="6" fill="rgba(11, 27, 32, 0.9)" stroke="#f6d388" stroke-width="1.4" />
          <circle cx="0" cy="-290" r="2.4" fill="#46d9ec" />
        </g>
      {/each}
    </g>
  </svg>

  <!-- Top Mucha Laurel & Acanthus Botanical Spandrels -->
  {#each ['tl', 'tr'] as pos}
    <svg class="kf-corner {pos}" viewBox="0 0 96 96" aria-hidden="true">
      <path d="M4 86 C4 26, 26 4, 86 4" fill="none" stroke="rgba(246, 211, 136, 0.88)" stroke-width="2" />
      <path d="M11 72 C11 32, 32 11, 72 11" fill="none" stroke="rgba(150, 38, 48, 0.72)" stroke-width="1.4" />
      <!-- Botanical laurel leaves -->
      <path d="M10 48 C18 36, 28 32, 34 36 C28 44, 18 48, 10 48 Z" fill="rgba(220, 162, 72, 0.28)" stroke="#f6d388" stroke-width="0.9" />
      <path d="M48 10 C36 18, 32 28, 36 34 C44 28, 48 18, 48 10 Z" fill="rgba(220, 162, 72, 0.28)" stroke="#f6d388" stroke-width="0.9" />
      <path d="M8 58 C24 52, 32 34, 24 24 C34 32, 52 24, 58 8" fill="none" stroke="rgba(70, 217, 236, 0.68)" stroke-width="1.2" />
      <path d="M10 10 C24 10, 30 19, 24 24 C19 30, 10 24, 10 10 Z" fill="rgba(150, 38, 48, 0.45)" stroke="#f6d388" stroke-width="1.1" />
      <circle cx="16" cy="16" r="3.4" fill="#46d9ec" stroke="#f6d388" stroke-width="1" />
      <circle cx="38" cy="11" r="2" fill="#f6d388" />
      <circle cx="11" cy="38" r="2" fill="#f6d388" />
    </svg>
  {/each}

  <!-- Bottom-Left Mucha Sunburst Rosette Roundel (inspired by Zodiac) -->
  <svg class="kf-corner bl" viewBox="0 0 96 96" aria-hidden="true">
    <path d="M4 10 C4 68, 28 92, 86 92" fill="none" stroke="rgba(246, 211, 136, 0.84)" stroke-width="1.8" />
    <circle cx="28" cy="68" r="17" fill="rgba(12, 26, 30, 0.85)" stroke="#f6d388" stroke-width="1.3" />
    <circle cx="28" cy="68" r="12" fill="rgba(150, 38, 48, 0.35)" stroke="#dca248" stroke-width="0.9" stroke-dasharray="2.5 2.5" />
    <circle cx="28" cy="68" r="5.5" fill="#f6d388" stroke="#962630" stroke-width="1" />
    <circle cx="52" cy="84" r="2" fill="#46d9ec" />
    <circle cx="12" cy="44" r="2" fill="#f6d388" />
  </svg>

  <!-- Bottom-Right Mucha Crescent Moon & Vine Roundel -->
  <svg class="kf-corner br" viewBox="0 0 96 96" aria-hidden="true">
    <path d="M92 10 C92 68, 68 92, 10 92" fill="none" stroke="rgba(246, 211, 136, 0.84)" stroke-width="1.8" />
    <circle cx="68" cy="68" r="17" fill="rgba(12, 26, 30, 0.85)" stroke="#f6d388" stroke-width="1.3" />
    <path d="M72 57 A11 11 0 1 0 72 79 A8.5 8.5 0 1 1 72 57 Z" fill="#f6d388" stroke="#dca248" stroke-width="0.6" />
    <circle cx="71" cy="68" r="2.2" fill="#46d9ec" />
    <circle cx="44" cy="84" r="2" fill="#46d9ec" />
    <circle cx="84" cy="44" r="2" fill="#f6d388" />
  </svg>

  <!-- Vertical side vine rails -->
  {#each ['left', 'right'] as side}
    <svg class="kf-side-rail {side}" viewBox="0 0 12 400" preserveAspectRatio="none" aria-hidden="true">
      <path
        d="M5 0 C10 50, 1 100, 6 150 C11 200, 1 250, 6 300 C10 350, 2 380, 5 400"
        fill="none"
        stroke="rgba(220, 162, 72, 0.55)"
        stroke-width="1.1"
      />
      <circle cx="6" cy="200" r="2.2" fill="#46d9ec" />
    </svg>
  {/each}

  <MacWindowControls
    onHide={() => api.hideWindow()}
    onMinimize={() => api.minimizeWindow()}
    onMaximize={() => api.toggleMaximizeWindow()}
  />

  <TabBar
    openTabs={appState.open_tabs}
    activeTab={appState.active_tab}
    homeActive={showingHome}
    boards={appState.boards}
    opacity={appState.preferences.opacity}
    showHeaderOpacity={appState.preferences.show_header_opacity ?? true}
    onToggleHome={toggleHomeMenu}
    onSelectTab={handleSelectTab}
    onCloseTab={handleCloseTab}
    onOpenSwitcher={() => (switcherOpen = true)}
    onOpenSettings={() => (settingsOpen = true)}
    onQuickCapture={() => {
      captureColumn = 'ready';
      captureOpen = true;
    }}
    onOpacityChange={handleOpacityChange}
  />

  {#if showingHome}
    <HomeView
      overviews={projectOverviews}
      loading={overviewsLoading}
      hasActiveTab={Boolean(activeBoard)}
      onOpenProject={handleOpenProject}
      onInitFolder={handleInitFolder}
      onRefresh={handleRescan}
      onQuickCapture={() => {
        captureColumn = 'ready';
        captureOpen = true;
      }}
      onReturnToBoard={() => (homeOpen = false)}
    />
  {:else if activeBoard}
    <KanbanBoard
      board={activeBoard}
      projectPath={appState.active_tab}
      parseError={activeError}
      onSelectTask={(t) => (selectedTaskID = t.id)}
      onMoveTask={handleMoveTask}
      onQuickAddInColumn={(colID) => {
        captureColumn = colID;
        captureOpen = true;
      }}
    />
  {:else}
    <section class="empty-aviary">
      <div class="aviary-arch">
        <span class="aviary-glyph">✦</span>
        <h1 class="aviary-title">Kingfisher</h1>
        <p class="aviary-subtitle">
          Single-file, gitignored project boards synced live with your coding agents.
        </p>
        <div class="aviary-buttons">
          <button type="button" class="kf-btn kf-btn-primary" onclick={handleInitFolder}>
            ✦ Open or Initialize a Project Folder
          </button>
          {#if appState.projects.length > 0}
            <button type="button" class="kf-btn" onclick={() => (switcherOpen = true)}>
              Browse {appState.projects.length} Discovered Project{appState.projects.length === 1 ? '' : 's'}
              <span class="kf-kbd">⌘K</span>
            </button>
          {/if}
        </div>
      </div>
    </section>
  {/if}

  {#if toastMessage}
    <div class="kf-toast" role="status">{toastMessage}</div>
  {/if}

  <QuickCaptureModal
    open={captureOpen}
    defaultStatus={captureColumn}
    projectName={activeBoard?.project.name ?? ''}
    projectCount={appState.projects.length}
    aiConfig={appState.preferences.ai}
    onClose={() => (captureOpen = false)}
    onCapture={handleCaptureTask}
    onSaveAIConfig={handleSaveAIConfig}
  />

  <TaskDetailDrawer
    task={selectedTask}
    columns={activeBoard?.columns ?? []}
    onClose={() => (selectedTaskID = null)}
    onSave={handleSaveTask}
    onDelete={handleDeleteTask}
  />

  <ProjectSwitcherModal
    open={switcherOpen}
    projects={appState.projects}
    openTabs={appState.open_tabs}
    onClose={() => (switcherOpen = false)}
    onSelectProject={handleOpenProject}
    onInitFolder={handleInitFolder}
    onRescan={handleRescan}
  />

  <SettingsModal
    open={settingsOpen}
    preferences={appState.preferences}
    onClose={() => (settingsOpen = false)}
    onLivePreview={handleLivePreviewPrefs}
    onSave={handleSavePreferences}
  />
</main>

<style>
  .empty-aviary {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .aviary-arch {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 12px;
    max-width: 460px;
    padding: 44px 38px;
    border-radius: 170px 170px 26px 26px;
    border: 2px solid var(--kf-lead-gold-bright);
    background: linear-gradient(
      180deg,
      rgba(16, 46, 48, calc(var(--kf-opacity) * 0.88 + 0.18)) 0%,
      rgba(7, 21, 21, calc(var(--kf-opacity) * 0.85 + 0.14)) 100%
    );
    box-shadow:
      0 22px 54px rgba(0, 0, 0, 0.42),
      inset 0 0 0 4px rgba(220, 162, 72, 0.14),
      inset 0 1px 0 rgba(246, 211, 136, 0.42);
  }

  .aviary-glyph {
    font-size: 30px;
    color: var(--kf-cerulean);
  }

  .aviary-title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 34px;
    letter-spacing: 0.16em;
    color: var(--kf-ochre);
    font-weight: 600;
  }

  .aviary-subtitle {
    margin: 0;
    font-size: 13.5px;
    line-height: 1.55;
    color: var(--kf-ivory-soft);
  }

  .aviary-buttons {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-top: 8px;
    width: 100%;
  }

  .kf-toast {
    position: fixed;
    bottom: 22px;
    left: 50%;
    transform: translateX(-50%);
    padding: 8px 18px;
    border-radius: 999px;
    border: 1.5px solid var(--kf-cerulean);
    background: rgba(9, 28, 30, 0.94);
    color: var(--kf-ivory);
    font-size: 12.5px;
    box-shadow:
      0 10px 30px rgba(0, 0, 0, 0.55),
      0 0 14px rgba(70, 217, 236, 0.22);
    z-index: 100;
  }
</style>
