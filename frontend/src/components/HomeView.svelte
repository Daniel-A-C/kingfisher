<script lang="ts">
  import type { ProjectOverview } from '../lib/types';

  interface Props {
    overviews: ProjectOverview[];
    loading: boolean;
    hasActiveTab: boolean;
    onOpenProject: (path: string) => void;
    onInitFolder: () => void;
    onRefresh: () => void;
    onQuickCapture: () => void;
    onReturnToBoard: () => void;
  }

  let {
    overviews,
    loading,
    hasActiveTab,
    onOpenProject,
    onInitFolder,
    onRefresh,
    onQuickCapture,
    onReturnToBoard,
  }: Props = $props();

  type SortMode = 'recent' | 'name' | 'tasks' | 'size' | 'commits';
  let query = $state('');
  let sortMode = $state<SortMode>('recent');

  let totals = $derived.by(() => {
    let projects = overviews.length;
    let openTabs = 0;
    let totalTasks = 0;
    let readyTasks = 0;
    let activeTasks = 0;
    let doneTasks = 0;
    let totalBytes = 0;
    let totalFiles = 0;
    let totalCommits = 0;

    for (const p of overviews) {
      if (p.is_open) openTabs++;
      totalTasks += p.tasks.total;
      readyTasks += p.tasks.ready;
      activeTasks += p.tasks.active;
      doneTasks += p.tasks.done;
      totalBytes += p.git.total_bytes;
      totalFiles += p.git.tracked_files;
      totalCommits += p.git.commit_count;
    }

    return {
      projects,
      openTabs,
      totalTasks,
      readyTasks,
      activeTasks,
      doneTasks,
      totalBytes,
      totalFiles,
      totalCommits,
    };
  });

  let filteredAndSorted = $derived.by(() => {
    const q = query.trim().toLowerCase();
    const list = q
      ? overviews.filter(
          (p) =>
            p.name.toLowerCase().includes(q) ||
            p.prefix.toLowerCase().includes(q) ||
            p.path.toLowerCase().includes(q) ||
            p.git.branch.toLowerCase().includes(q) ||
            p.git.languages.some((l) => l.name.toLowerCase().includes(q))
        )
      : [...overviews];

    list.sort((a, b) => {
      switch (sortMode) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'tasks': {
          const openA = a.tasks.ready + a.tasks.active;
          const openB = b.tasks.ready + b.tasks.active;
          if (openA !== openB) return openB - openA;
          return b.tasks.total - a.tasks.total;
        }
        case 'size':
          return b.git.total_bytes - a.git.total_bytes;
        case 'commits':
          return b.git.commit_count - a.git.commit_count;
        case 'recent':
        default: {
          if (a.is_active !== b.is_active) return a.is_active ? -1 : 1;
          if (a.is_open !== b.is_open) return a.is_open ? -1 : 1;
          const tA = a.git.last_commit_at || a.last_opened || '';
          const tB = b.git.last_commit_at || b.last_opened || '';
          if (tA !== tB) return tB.localeCompare(tA);
          return a.name.localeCompare(b.name);
        }
      }
    });

    return list;
  });

  function formatBytes(bytes: number): string {
    if (!bytes || bytes <= 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    let idx = 0;
    let val = bytes;
    while (val >= 1024 && idx < units.length - 1) {
      val /= 1024;
      idx++;
    }
    return idx === 0 ? `${Math.round(val)} ${units[idx]}` : `${val.toFixed(val >= 100 ? 0 : 1)} ${units[idx]}`;
  }

  function formatRelativeTime(iso: string): string {
    if (!iso) return 'No commits yet';
    const ms = Date.parse(iso);
    if (Number.isNaN(ms)) return '';
    const diffSec = Math.max(0, Math.floor((Date.now() - ms) / 1000));
    if (diffSec < 60) return 'just now';
    const diffMin = Math.floor(diffSec / 60);
    if (diffMin < 60) return `${diffMin}m ago`;
    const diffHr = Math.floor(diffMin / 60);
    if (diffHr < 24) return `${diffHr}h ago`;
    const diffDay = Math.floor(diffHr / 24);
    if (diffDay < 30) return `${diffDay}d ago`;
    const diffMo = Math.floor(diffDay / 30);
    if (diffMo < 12) return `${diffMo}mo ago`;
    return `${Math.floor(diffMo / 12)}y ago`;
  }

  function compactPath(fullPath: string): string {
    return fullPath.replace(/^\/Users\/[^/]+/, '~');
  }
</script>

<section class="home-view kf-scroll">
  <!-- Ornate Art Nouveau Sunburst Header Banner -->
  <header class="home-banner">
    <div class="banner-title-block">
      <div class="banner-eyebrow">
        <span aria-hidden="true">❧</span>
        <span>PROJECT ATLAS &amp; GIT TELEMETRY</span>
        <span aria-hidden="true">☙</span>
      </div>
      <h1 class="home-title">The Grand Aviary</h1>
      <p class="home-subtitle">
        Overview of your registered repositories, live task queues, codebase size, and language makeup.
      </p>
    </div>

    <div class="banner-actions">
      {#if hasActiveTab}
        <button type="button" class="kf-btn" onclick={onReturnToBoard} title="Return to active board (⌘0)">
          ← Active Board
          <span class="kf-kbd">⌘0</span>
        </button>
      {/if}
      <button type="button" class="kf-btn" onclick={onRefresh} disabled={loading}>
        {loading ? '◌ Scanning…' : '↻ Refresh'}
      </button>
      <button type="button" class="kf-btn" onclick={onInitFolder}>
        + Init Folder
      </button>
      <button type="button" class="kf-btn kf-btn-primary" onclick={onQuickCapture}>
        ✦ Capture <span class="kf-kbd">⌘N</span>
      </button>
    </div>
  </header>

  <!-- Aggregate Stained-Glass Medallion Summary Strip -->
  <dl class="summary-medallions">
    <div class="medallion">
      <dt class="medallion-label">Projects</dt>
      <dd class="medallion-value">{totals.projects}</dd>
      <dd class="medallion-sub">{totals.openTabs} open in tabs</dd>
    </div>
    <div class="medallion">
      <dt class="medallion-label">Task Queue</dt>
      <dd class="medallion-value">{totals.readyTasks + totals.activeTasks}</dd>
      <dd class="medallion-sub">
        {totals.activeTasks} in flight · {totals.readyTasks} ready · {totals.doneTasks} landed
      </dd>
    </div>
    <div class="medallion">
      <dt class="medallion-label">Tracked Codebase</dt>
      <dd class="medallion-value">{formatBytes(totals.totalBytes)}</dd>
      <dd class="medallion-sub">{totals.totalFiles.toLocaleString()} tracked files</dd>
    </div>
    <div class="medallion">
      <dt class="medallion-label">Git History</dt>
      <dd class="medallion-value">{totals.totalCommits.toLocaleString()}</dd>
      <dd class="medallion-sub">commits across repositories</dd>
    </div>
  </dl>

  <!-- Filter & Sort Controls -->
  <div class="home-toolbar">
    <div class="search-field">
      <label for="home-search" class="visually-hidden">Filter projects by name, prefix, branch, or language</label>
      <input
        id="home-search"
        type="search"
        class="kf-input home-search-input"
        placeholder="Filter by project, branch, or language (e.g. Go, Svelte)…"
        bind:value={query}
      />
    </div>

    <div class="sort-field">
      <label for="home-sort" class="sort-label">Order</label>
      <select id="home-sort" class="kf-select home-sort-select" bind:value={sortMode}>
        <option value="recent">Recent Activity</option>
        <option value="tasks">Active &amp; Ready Tasks</option>
        <option value="size">Codebase Size</option>
        <option value="commits">Commit Count</option>
        <option value="name">Name (A–Z)</option>
      </select>
    </div>
  </div>

  <!-- Project Cards Grid -->
  {#if loading && overviews.length === 0}
    <div class="home-empty" role="status">
      <progress aria-label="Loading project overviews"></progress>
      <p>Inspecting repositories and Git telemetry…</p>
    </div>
  {:else if filteredAndSorted.length === 0}
    <div class="home-empty">
      <span class="empty-glyph" aria-hidden="true">✦</span>
      {#if overviews.length === 0}
        <h2>No Projects in the Aviary Yet</h2>
        <p>Initialize any local folder to create a gitignored <code>.kingfisher.yaml</code> board.</p>
        <button type="button" class="kf-btn kf-btn-primary" onclick={onInitFolder}>
          ✦ Open or Initialize a Project Folder
        </button>
      {:else}
        <h2>No Matching Projects</h2>
        <p>No projects matched "{query}".</p>
        <button type="button" class="kf-btn" onclick={() => (query = '')}>Clear Filter</button>
      {/if}
    </div>
  {:else}
    <div class="project-grid">
      {#each filteredAndSorted as proj (proj.path)}
        <article
          class="project-card"
          class:is-open={proj.is_open}
          class:is-active={proj.is_active}
        >
          <!-- Decorative Art Nouveau top arch accent -->
          <svg class="card-arch-ornament" viewBox="0 0 280 14" preserveAspectRatio="none" aria-hidden="true">
            <path d="M8 12 Q 140 -2, 272 12" fill="none" stroke="rgba(246, 211, 136, 0.52)" stroke-width="1.2" />
            <circle cx="140" cy="5" r="2.2" fill="#46d9ec" />
          </svg>

          <header class="card-head">
            <div class="card-title-row">
              <span class="card-prefix">{proj.prefix}</span>
              <h2 class="card-name">
                <button
                  type="button"
                  class="card-primary-btn"
                  onclick={() => onOpenProject(proj.path)}
                >
                  {proj.name}
                </button>
              </h2>
              {#if proj.is_active}
                <span class="status-pill active-pill">Active</span>
              {:else if proj.is_open}
                <span class="status-pill open-pill">Open</span>
              {/if}
            </div>

            <div class="git-branch-row">
              {#if proj.git.is_git_repo}
                <span class="branch-badge" title="Current Git branch">
                  ⎇ {proj.git.branch || 'HEAD'}
                </span>
                {#if proj.git.dirty_count > 0}
                  <span class="dirty-badge" title="{proj.git.dirty_count} uncommitted change(s)">
                    ● {proj.git.dirty_count} dirty
                  </span>
                {:else}
                  <span class="clean-badge">✓ clean</span>
                {/if}
              {:else}
                <span class="non-git-badge">Not a Git repo</span>
              {/if}
            </div>
          </header>

          <div class="card-body">
            <!-- Board Task Breakdown -->
            <div class="task-pills" aria-label="Task counts">
              <span class="t-pill t-active" title="In Flight">
                <strong>{proj.tasks.active}</strong> in flight
              </span>
              <span class="t-pill t-ready" title="Ready to Dive">
                <strong>{proj.tasks.ready}</strong> ready
              </span>
              <span class="t-pill t-done" title="Landed">
                <strong>{proj.tasks.done}</strong> done
              </span>
              {#if proj.tasks.high_priority_open > 0}
                <span class="t-pill t-high" title="High or critical open tasks">
                  ✦ {proj.tasks.high_priority_open} high
                </span>
              {/if}
            </div>

            {#if proj.tasks.next_task_id}
              <div class="next-up">
                <span class="next-label">Next:</span>
                <span class="next-id">{proj.tasks.next_task_id}</span>
                <span class="next-title">{proj.tasks.next_task_title}</span>
              </div>
            {/if}

            <!-- Git Size & Telemetry -->
            {#if proj.git.is_git_repo}
              <dl class="git-metrics">
                <div class="metric">
                  <dt>Size</dt>
                  <dd>{formatBytes(proj.git.total_bytes)}</dd>
                </div>
                <div class="metric">
                  <dt>Files</dt>
                  <dd>{proj.git.tracked_files.toLocaleString()}</dd>
                </div>
                <div class="metric">
                  <dt>Commits</dt>
                  <dd>{proj.git.commit_count.toLocaleString()}</dd>
                </div>
                <div class="metric">
                  <dt>Updated</dt>
                  <dd>{formatRelativeTime(proj.git.last_commit_at)}</dd>
                </div>
              </dl>

              <!-- Language Makeup Stained-Glass Bar -->
              {#if proj.git.languages.length > 0}
                <div class="makeup-section">
                  <div class="makeup-bar" aria-hidden="true">
                    {#each proj.git.languages as lang (lang.name)}
                      <span
                        class="makeup-segment"
                        style="width: {Math.max(lang.percent, 2)}%; background: {lang.color};"
                        title="{lang.name}: {lang.percent}% ({formatBytes(lang.bytes)}, {lang.files} files)"
                      ></span>
                    {/each}
                  </div>
                  <ul role="list" class="lang-legend" aria-label="Language makeup">
                    {#each proj.git.languages as lang (lang.name)}
                      <li class="lang-item">
                        <span class="lang-dot" style="background: {lang.color};" aria-hidden="true"></span>
                        <span class="lang-name">{lang.name}</span>
                        <span class="lang-pct">{lang.percent}%</span>
                      </li>
                    {/each}
                  </ul>
                </div>
              {/if}

              {#if proj.git.last_commit_msg}
                <p class="last-commit" title={proj.git.last_commit_msg}>
                  <span class="commit-glyph" aria-hidden="true">↳</span>
                  {proj.git.last_commit_msg}
                </p>
              {/if}
            {/if}
          </div>

          <footer class="card-footer">
            <span class="card-path" title={proj.path}>{compactPath(proj.path)}</span>
            <span class="open-cue" aria-hidden="true">
              {proj.is_active ? 'Viewing →' : 'Open Board →'}
            </span>
          </footer>
        </article>
      {/each}
    </div>
  {/if}
</section>

<style>
  .home-view {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 18px 26px 32px;
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .home-banner {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 16px;
    padding: 20px 24px;
    border-radius: 24px 24px 16px 16px;
    border: 1.5px solid var(--kf-lead-gold-bright);
    background:
      radial-gradient(ellipse at 18% 0%, rgba(70, 217, 236, 0.2), transparent 55%),
      radial-gradient(ellipse at 82% 100%, rgba(220, 162, 72, 0.16), transparent 55%),
      linear-gradient(
        180deg,
        rgba(14, 42, 44, calc(var(--kf-opacity) * 0.88)) 0%,
        rgba(8, 22, 23, calc(var(--kf-opacity) * 0.82)) 100%
      );
    box-shadow:
      0 14px 36px rgba(0, 0, 0, 0.34),
      inset 0 0 0 3px rgba(220, 162, 72, 0.12),
      inset 0 1px 0 rgba(246, 211, 136, 0.44);
  }

  .banner-title-block {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .banner-eyebrow {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.2em;
    color: var(--kf-cerulean);
  }

  .home-title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 28px;
    font-weight: 600;
    letter-spacing: 0.1em;
    color: var(--kf-ochre);
  }

  .home-subtitle {
    margin: 0;
    font-size: 13px;
    color: var(--kf-ivory-soft);
  }

  .banner-actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  /* Summary Medallions */
  .summary-medallions {
    margin: 0;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 12px;
  }

  .medallion {
    padding: 13px 16px;
    border-radius: 16px;
    border: 1px solid var(--kf-lead-gold);
    background: linear-gradient(
      180deg,
      rgba(13, 36, 38, calc(var(--kf-opacity) * 0.84)) 0%,
      rgba(7, 20, 21, calc(var(--kf-opacity) * 0.78)) 100%
    );
    box-shadow: inset 0 1px 0 rgba(246, 211, 136, 0.22);
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .medallion-label {
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--kf-ivory-dim);
  }

  .medallion-value {
    margin: 0;
    font-family: var(--font-display);
    font-size: 24px;
    font-weight: 600;
    color: var(--kf-ochre);
    line-height: 1.15;
  }

  .medallion-sub {
    margin: 0;
    font-size: 11.5px;
    color: var(--kf-ivory-soft);
  }

  /* Toolbar */
  .home-toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .search-field {
    flex: 1;
    min-width: 220px;
  }

  .home-search-input {
    border-radius: 999px;
    padding: 8px 16px;
  }

  .sort-field {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .sort-label {
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--kf-ivory-dim);
  }

  .home-sort-select {
    width: auto;
    min-width: 170px;
    border-radius: 999px;
    padding: 7px 14px;
    cursor: pointer;
  }

  /* Project Grid */
  .project-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(330px, 1fr));
    gap: 16px;
  }

  /* Container-query responsive Art Nouveau specimen card */
  .project-card {
    position: relative;
    container-type: inline-size;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 18px 18px 14px;
    border-radius: 22px 22px 14px 14px;
    border: 1.5px solid rgba(220, 162, 72, 0.5);
    background:
      radial-gradient(ellipse at 50% 0%, rgba(30, 168, 184, 0.14), transparent 60%),
      linear-gradient(
        180deg,
        rgba(14, 37, 39, calc(var(--kf-opacity) * 0.88)) 0%,
        rgba(8, 22, 23, calc(var(--kf-opacity) * 0.82)) 100%
      );
    box-shadow:
      0 10px 26px rgba(0, 0, 0, 0.32),
      inset 0 1px 0 rgba(246, 211, 136, 0.26);
    transition: border-color 150ms ease, transform 150ms ease, box-shadow 150ms ease;
  }

  .project-card.is-active {
    border-color: var(--kf-cerulean);
    box-shadow:
      0 12px 30px rgba(0, 0, 0, 0.4),
      0 0 18px rgba(70, 217, 236, 0.2),
      inset 0 1px 0 rgba(183, 245, 253, 0.4);
  }

  .project-card:has(.card-primary-btn:hover) {
    border-color: var(--kf-ochre);
    transform: translateY(-2px);
    box-shadow:
      0 16px 34px rgba(0, 0, 0, 0.42),
      0 0 18px rgba(220, 162, 72, 0.22),
      inset 0 1px 0 rgba(246, 211, 136, 0.48);
  }

  .project-card:has(.card-primary-btn:focus-visible) {
    outline: 2.5px solid var(--kf-cerulean);
    outline-offset: 3px;
  }

  .card-arch-ornament {
    position: absolute;
    top: 3px;
    left: 12%;
    width: 76%;
    height: 12px;
    pointer-events: none;
  }

  .card-head {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .card-title-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .card-prefix {
    font-family: var(--font-mono);
    font-size: 10.5px;
    font-weight: 600;
    color: var(--kf-ochre);
    padding: 2px 7px;
    border-radius: 999px;
    border: 1px solid rgba(220, 162, 72, 0.5);
    background: rgba(220, 162, 72, 0.14);
    flex-shrink: 0;
  }

  .card-name {
    margin: 0;
    flex: 1;
    min-width: 0;
    font-family: var(--font-display);
    font-size: 18px;
    font-weight: 600;
    letter-spacing: 0.04em;
    color: var(--kf-ivory);
  }

  /* Stretched-button primary action per modern-web-guidance card pattern */
  .card-primary-btn {
    background: transparent;
    border: none;
    padding: 0;
    margin: 0;
    color: inherit;
    font: inherit;
    letter-spacing: inherit;
    text-align: left;
    cursor: pointer;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
  }

  .card-primary-btn::after {
    content: '';
    position: absolute;
    inset: 0;
  }

  .card-primary-btn:focus-visible {
    outline: none; /* Focus ring drawn on .project-card */
  }

  .status-pill {
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 2px 8px;
    border-radius: 999px;
    flex-shrink: 0;
  }

  .active-pill {
    border: 1px solid var(--kf-cerulean);
    background: rgba(70, 217, 236, 0.2);
    color: #c7f7fe;
  }

  .open-pill {
    border: 1px solid rgba(220, 162, 72, 0.55);
    background: rgba(220, 162, 72, 0.14);
    color: var(--kf-ochre);
  }

  .git-branch-row {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    font-family: var(--font-mono);
    font-size: 11px;
  }

  .branch-badge {
    color: var(--kf-cerulean);
    background: rgba(70, 217, 236, 0.1);
    border: 1px solid rgba(70, 217, 236, 0.34);
    padding: 1px 8px;
    border-radius: 6px;
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .dirty-badge {
    color: #f6b368;
  }

  .clean-badge {
    color: var(--kf-sage);
  }

  .non-git-badge {
    color: var(--kf-ivory-dim);
    font-style: italic;
  }

  .card-body {
    display: flex;
    flex-direction: column;
    gap: 10px;
    flex: 1;
  }

  .task-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .t-pill {
    font-size: 11px;
    padding: 2px 8px;
    border-radius: 999px;
    border: 1px solid rgba(220, 162, 72, 0.28);
    background: rgba(8, 22, 23, 0.56);
    color: var(--kf-ivory-soft);
  }

  .t-pill strong {
    font-family: var(--font-mono);
    color: var(--kf-ivory);
  }

  .t-active {
    border-color: rgba(70, 217, 236, 0.45);
  }
  .t-active strong {
    color: var(--kf-cerulean);
  }

  .t-high {
    border-color: rgba(230, 93, 56, 0.55);
    color: #ffcabd;
  }

  .next-up {
    display: flex;
    align-items: baseline;
    gap: 6px;
    padding: 6px 10px;
    border-radius: 10px;
    border: 1px solid rgba(220, 162, 72, 0.26);
    background: rgba(6, 17, 18, 0.48);
    font-size: 11.5px;
    min-width: 0;
  }

  .next-label {
    color: var(--kf-ivory-dim);
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    flex-shrink: 0;
  }

  .next-id {
    font-family: var(--font-mono);
    color: var(--kf-ochre);
    font-size: 10.5px;
    flex-shrink: 0;
  }

  .next-title {
    color: var(--kf-ivory);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Git metrics grid adapts to card inline-size */
  .git-metrics {
    margin: 0;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 6px;
    padding: 8px 10px;
    border-radius: 10px;
    background: rgba(6, 17, 18, 0.46);
    border: 1px solid rgba(220, 162, 72, 0.22);
  }

  @container (max-width: 310px) {
    .git-metrics {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  .metric {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
  }

  .metric dt {
    font-family: var(--font-mono);
    font-size: 9.5px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--kf-ivory-dim);
  }

  .metric dd {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 12px;
    font-weight: 600;
    color: var(--kf-ivory);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Language Makeup Stained-Glass Bar */
  .makeup-section {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .makeup-bar {
    display: flex;
    height: 8px;
    border-radius: 999px;
    overflow: hidden;
    background: rgba(6, 17, 18, 0.65);
    border: 1px solid rgba(220, 162, 72, 0.34);
    gap: 1px;
  }

  .makeup-segment {
    height: 100%;
    display: block;
    min-width: 3px;
  }

  .lang-legend {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 5px 10px;
  }

  .lang-item {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 11px;
    color: var(--kf-ivory-soft);
  }

  .lang-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
    box-shadow: 0 0 6px rgba(0, 0, 0, 0.45);
  }

  .lang-pct {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--kf-ivory-dim);
  }

  .last-commit {
    margin: 0;
    font-size: 11.5px;
    color: var(--kf-ivory-dim);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .commit-glyph {
    color: var(--kf-ochre);
    margin-right: 4px;
  }

  .card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding-top: 8px;
    border-top: 1px solid rgba(220, 162, 72, 0.22);
    font-size: 11px;
  }

  .card-path {
    font-family: var(--font-mono);
    color: var(--kf-ivory-dim);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
  }

  .open-cue {
    font-family: var(--font-display);
    font-size: 12px;
    letter-spacing: 0.05em;
    color: var(--kf-ochre);
    flex-shrink: 0;
  }

  .home-empty {
    margin: auto;
    max-width: 440px;
    text-align: center;
    padding: 40px 28px;
    border-radius: 28px;
    border: 1.5px solid var(--kf-lead-gold);
    background: rgba(10, 28, 30, calc(var(--kf-opacity) * 0.85 + 0.14));
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
  }

  .empty-glyph {
    font-size: 28px;
    color: var(--kf-cerulean);
  }

  .home-empty h2 {
    margin: 0;
    font-family: var(--font-display);
    color: var(--kf-ochre);
    font-size: 20px;
  }

  .home-empty p {
    margin: 0;
    font-size: 13px;
    color: var(--kf-ivory-soft);
  }
</style>
