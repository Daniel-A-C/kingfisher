export interface ProjectMeta {
  name: string;
  prefix: string;
  next_id: number;
}

export interface Column {
  id: string;
  name: string;
}

export type Priority = 'low' | 'medium' | 'high' | 'critical';

export interface Task {
  id: string;
  title: string;
  status: string;
  priority: Priority;
  tags: string[];
  created: string;
  updated: string;
  description: string;
  raw_input?: string;
  agent_notes: string;
}

export interface Board {
  project: ProjectMeta;
  columns: Column[];
  tasks: Task[];
}

export interface TaskPatch {
  title?: string;
  status?: string;
  priority?: Priority;
  tags?: string[];
  has_tags?: boolean;
  description?: string;
  raw_input?: string;
  agent_notes?: string;
}

export interface ProjectEntry {
  slug: string;
  name: string;
  prefix: string;
  path: string;
  last_opened: string;
  pinned: boolean;
}

export interface AIConfig {
  enabled: boolean;
  provider: string;
  endpoint: string;
  model: string;
  timeout_seconds?: number;
}

export interface AIStatus {
  available: boolean;
  provider: string;
  model: string;
  endpoint: string;
  models: string[];
  message: string;
}

export interface CaptureResponse {
  board: Board;
  task: Task;
  used_ai: boolean;
  model?: string;
}

export interface CapturedTicket {
  project_path: string;
  project_name: string;
  project_slug: string;
  task: Task;
}

export interface CaptureBatchResponse {
  state: AppState;
  tickets: CapturedTicket[];
  used_ai: boolean;
  model?: string;
}

export interface Preferences {
  opacity: number;
  high_contrast: boolean;
  reduce_motion: boolean;
  show_header_opacity?: boolean;
  scan_dirs: string[];
  ai: AIConfig;
}

export interface AppState {
  projects: ProjectEntry[];
  open_tabs: string[];
  active_tab: string;
  boards: Record<string, Board>;
  errors: Record<string, string>;
  preferences: Preferences;
}

export interface BoardUpdateEvent {
  path: string;
  board: Board;
}

export interface BoardErrorEvent {
  path: string;
  error: string;
}

export interface LanguageStat {
  name: string;
  color: string;
  bytes: number;
  files: number;
  percent: number;
}

export interface GitStats {
  is_git_repo: boolean;
  branch: string;
  dirty_count: number;
  commit_count: number;
  last_commit_at: string;
  last_commit_msg: string;
  tracked_files: number;
  total_bytes: number;
  languages: LanguageStat[];
}

export interface TaskSummary {
  total: number;
  backlog: number;
  ready: number;
  active: number;
  done: number;
  high_priority_open: number;
  next_task_id?: string;
  next_task_title?: string;
}

export interface ProjectOverview {
  slug: string;
  name: string;
  prefix: string;
  path: string;
  last_opened: string;
  is_open: boolean;
  is_active: boolean;
  tasks: TaskSummary;
  git: GitStats;
}
