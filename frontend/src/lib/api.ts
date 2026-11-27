import type {
  AIConfig,
  AIStatus,
  AppState,
  Board,
  BoardErrorEvent,
  BoardUpdateEvent,
  CaptureBatchResponse,
  CaptureResponse,
  Preferences,
  TaskPatch,
} from './types';

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          GetState(): Promise<AppState>;
          ScanProjects(): Promise<AppState>;
          OpenProjectTab(path: string): Promise<AppState>;
          CloseProjectTab(path: string): Promise<AppState>;
          SetActiveTab(path: string): Promise<AppState>;
          ReorderTabs(paths: string[]): Promise<AppState>;
          PickFolder(): Promise<string>;
          InitProject(
            dir: string,
            name: string,
            prefix: string,
            useGitignore: boolean
          ): Promise<AppState>;
          CaptureContext(
            rawContext: string,
            activeProjectPath: string,
            defaultStatus: string
          ): Promise<CaptureBatchResponse>;
          CaptureTask(
            projectPath: string,
            rawContext: string,
            defaultStatus: string
          ): Promise<CaptureResponse>;
          GetAIStatus(): Promise<AIStatus>;
          UpdateAIConfig(cfg: AIConfig): Promise<Preferences>;
          CreateTask(
            projectPath: string,
            title: string,
            status: string,
            priority: string,
            description: string,
            tags: string[]
          ): Promise<Board>;
          UpdateTask(
            projectPath: string,
            taskID: string,
            patch: TaskPatch
          ): Promise<Board>;
          MoveTask(
            projectPath: string,
            taskID: string,
            targetStatus: string,
            targetIndex: number
          ): Promise<Board>;
          DeleteTask(projectPath: string, taskID: string): Promise<Board>;
          SetOpacity(opacity: number): Promise<Preferences>;
          UpdatePreferences(prefs: Preferences): Promise<AppState>;
          UpdateScanDirs(dirs: string[]): Promise<AppState>;
          HideWindow(): Promise<void>;
          MinimizeWindow(): Promise<void>;
          ToggleMaximizeWindow(): Promise<void>;
        };
      };
    };
    runtime?: {
      EventsOn(eventName: string, callback: (...data: any[]) => void): () => void;
    };
  }
}

function backend() {
  const app = window.go?.main?.App;
  if (!app) {
    throw new Error('Kingfisher Wails backend is not attached');
  }
  return app;
}

export const api = {
  getState: () => backend().GetState(),
  scanProjects: () => backend().ScanProjects(),
  openProjectTab: (path: string) => backend().OpenProjectTab(path),
  closeProjectTab: (path: string) => backend().CloseProjectTab(path),
  setActiveTab: (path: string) => backend().SetActiveTab(path),
  reorderTabs: (paths: string[]) => backend().ReorderTabs(paths),
  pickFolder: () => backend().PickFolder(),
  initProject: (dir: string, name: string, prefix: string, useGitignore: boolean) =>
    backend().InitProject(dir, name, prefix, useGitignore),
  captureContext: (rawContext: string, activeProjectPath: string, defaultStatus: string) =>
    backend().CaptureContext(rawContext, activeProjectPath, defaultStatus),
  captureTask: (projectPath: string, rawContext: string, defaultStatus: string) =>
    backend().CaptureTask(projectPath, rawContext, defaultStatus),
  getAIStatus: () => backend().GetAIStatus(),
  updateAIConfig: (cfg: AIConfig) => backend().UpdateAIConfig(cfg),
  createTask: (
    projectPath: string,
    title: string,
    status: string,
    priority: string,
    description: string,
    tags: string[]
  ) => backend().CreateTask(projectPath, title, status, priority, description, tags),
  updateTask: (projectPath: string, taskID: string, patch: TaskPatch) =>
    backend().UpdateTask(projectPath, taskID, patch),
  moveTask: (
    projectPath: string,
    taskID: string,
    targetStatus: string,
    targetIndex: number
  ) => backend().MoveTask(projectPath, taskID, targetStatus, targetIndex),
  deleteTask: (projectPath: string, taskID: string) =>
    backend().DeleteTask(projectPath, taskID),
  setOpacity: (opacity: number) => backend().SetOpacity(opacity),
  updatePreferences: (prefs: Preferences) => backend().UpdatePreferences(prefs),
  updateScanDirs: (dirs: string[]) => backend().UpdateScanDirs(dirs),
  hideWindow: () => backend().HideWindow(),
  minimizeWindow: () => backend().MinimizeWindow(),
  toggleMaximizeWindow: () => backend().ToggleMaximizeWindow(),
};

export function subscribeEvents(handlers: {
  onBoardUpdated: (ev: BoardUpdateEvent) => void;
  onBoardError: (ev: BoardErrorEvent) => void;
  onStateChanged: (state: AppState) => void;
}): () => void {
  const rt = window.runtime;
  if (!rt) return () => {};

  const unsubs = [
    rt.EventsOn('board:updated', (ev: BoardUpdateEvent) => handlers.onBoardUpdated(ev)),
    rt.EventsOn('board:error', (ev: BoardErrorEvent) => handlers.onBoardError(ev)),
    rt.EventsOn('state:changed', (state: AppState) => handlers.onStateChanged(state)),
  ];
  return () => unsubs.forEach((u) => u && u());
}
