<script lang="ts">
  import { onMount } from 'svelte';
  import { api, subscribeEvents } from './lib/api';
  import type { AppState, Priority, Task, TaskPatch } from './lib/types';
  import TabBar from './components/TabBar.svelte';
  import KanbanBoard from './components/KanbanBoard.svelte';
  import QuickCaptureModal from './components/QuickCaptureModal.svelte';
  import TaskDetailDrawer from './components/TaskDetailDrawer.svelte';
  import ProjectSwitcherModal from './components/ProjectSwitcherModal.svelte';

  let appState = $state<AppState>({
    projects: [],
    open_tabs: [],
    active_tab: '',
    boards: {},
    errors: {},
    preferences: { opacity: 0.55, scan_dirs: ['~/Documents/programs'] },
  });

  let switcherOpen = $state(false);
  let captureOpen = $state(false);
  let captureColumn = $state('ready');
  let selectedTaskID = $state<string | null>(null);
  let toastMessage = $state<string | null>(null);
  let toastTimer: number | undefined;

  let activeBoard = $derived(
    appState.active_tab ? appState.boards[appState.active_tab] ?? null : null
  );
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
    }, 3200);
  }

  function applyOpacityVar(val: number) {
    document.documentElement.style.setProperty('--kf-opacity', String(val));
  }

  onMount(() => {
    api
      .getState()
      .then((s) => {
        appState = s;
        applyOpacityVar(s.preferences.opacity);
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

      // Cmd + N -> Quick Capture
      if (meta && !e.shiftKey && e.key.toLowerCase() === 'n') {
        e.preventDefault();
        if (activeBoard) {
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
      if (!meta && !isTyping && !switcherOpen && !captureOpen && !selectedTaskID) {
        if (e.key === 'n' && activeBoard) {
          e.preventDefault();
          captureColumn = 'ready';
          captureOpen = true;
          return;
        }
        if (e.key === 'Escape') {
          e.preventDefault();
          api.hideWindow();
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
    appState = await api.setActiveTab(path);
  }

  async function handleOpenProject(path: string) {
    try {
      selectedTaskID = null;
      appState = await api.openProjectTab(path);
    } catch (err) {
      showToast(`Could not open project: ${err}`);
    }
  }

  async function handleCloseTab(path: string) {
    selectedTaskID = null;
    appState = await api.closeProjectTab(path);
  }

  async function handleInitFolder() {
    try {
      const dir = await api.pickFolder();
      if (!dir) return;
      appState = await api.initProject(dir, '', '', false);
      showToast('Project ready (.kingfisher.yaml ignored via .git/info/exclude)');
    } catch (err) {
      showToast(`Init failed: ${err}`);
    }
  }

  async function handleRescan() {
    appState = await api.scanProjects();
    showToast(`Found ${appState.projects.length} Kingfisher project(s)`);
  }

  async function handleOpacityChange(value: number) {
    applyOpacityVar(value);
    appState = {
      ...appState,
      preferences: { ...appState.preferences, opacity: value },
    };
    await api.setOpacity(value);
  }

  async function handleCreateTask(payload: {
    title: string;
    status: string;
    priority: Priority;
    tags: string[];
    description: string;
  }) {
    if (!appState.active_tab) return;
    try {
      const updated = await api.createTask(
        appState.active_tab,
        payload.title,
        payload.status,
        payload.priority,
        payload.description,
        payload.tags
      );
      appState = {
        ...appState,
        boards: { ...appState.boards, [appState.active_tab]: updated },
      };
    } catch (err) {
      showToast(`Failed to add task: ${err}`);
    }
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

<main class="kf-shell">
  <!-- Art Nouveau Botanical Corner Filigree -->
  {#each ['tl', 'tr', 'bl', 'br'] as pos}
    <svg class="kf-corner {pos}" viewBox="0 0 60 60" aria-hidden="true">
      <path
        d="M3 46 C3 16, 16 3, 46 3"
        fill="none"
        stroke="rgba(242, 204, 126, 0.72)"
        stroke-width="1.4"
      />
      <path
        d="M8 34 C10 18, 18 10, 34 8"
        fill="none"
        stroke="rgba(65, 211, 230, 0.55)"
        stroke-width="1"
      />
      <circle cx="11" cy="11" r="2.2" fill="#f2cc7e" />
    </svg>
  {/each}

  <TabBar
    openTabs={appState.open_tabs}
    activeTab={appState.active_tab}
    boards={appState.boards}
    opacity={appState.preferences.opacity}
    onSelectTab={handleSelectTab}
    onCloseTab={handleCloseTab}
    onOpenSwitcher={() => (switcherOpen = true)}
    onQuickCapture={() => {
      captureColumn = 'ready';
      captureOpen = true;
    }}
    onOpacityChange={handleOpacityChange}
    onHide={() => api.hideWindow()}
    onMinimize={() => api.minimizeWindow()}
    onMaximize={() => api.toggleMaximizeWindow()}
  />

  {#if activeBoard}
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
    columns={activeBoard?.columns ?? []}
    defaultStatus={captureColumn}
    projectName={activeBoard?.project.name ?? ''}
    onClose={() => (captureOpen = false)}
    onSubmit={handleCreateTask}
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
    padding: 40px 36px;
    border-radius: 160px 160px 24px 24px;
    border: 1.5px solid var(--kf-lead-gold);
    background: linear-gradient(
      180deg,
      rgba(16, 44, 46, calc(var(--kf-opacity) * 0.85 + 0.16)) 0%,
      rgba(8, 22, 22, calc(var(--kf-opacity) * 0.82 + 0.12)) 100%
    );
    box-shadow:
      0 18px 48px rgba(0, 0, 0, 0.35),
      inset 0 1px 0 rgba(242, 204, 126, 0.32);
  }

  .aviary-glyph {
    font-size: 28px;
    color: var(--kf-cerulean);
  }

  .aviary-title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 32px;
    letter-spacing: 0.14em;
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
    padding: 8px 16px;
    border-radius: 999px;
    border: 1px solid var(--kf-cerulean);
    background: rgba(10, 30, 32, 0.92);
    color: var(--kf-ivory);
    font-size: 12.5px;
    box-shadow: 0 10px 28px rgba(0, 0, 0, 0.5);
    z-index: 100;
  }
</style>
