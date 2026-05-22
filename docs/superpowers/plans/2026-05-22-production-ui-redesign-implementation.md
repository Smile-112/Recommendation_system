# Production UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the existing single-page UI into a production control cockpit that exposes planning, schedules, operators, equipment, warnings, and recommendation explanations as one coherent workflow.

**Architecture:** Keep the current vanilla HTML/CSS/JavaScript app and improve it in place. Split rendering logic inside `app.js` into clearer helper functions without introducing a frontend framework, because the project already ships static assets from the Go binary. Backend changes should be limited to small recommendation/read endpoints only if the UI cannot access data already returned by existing APIs.

**Tech Stack:** Go 1.24 backend, PostgreSQL, Docker Compose, vanilla HTML/CSS/JavaScript frontend, browser verification at `http://localhost:8080`.

---

## File Structure

- Modify `recsys-backend/web/index.html`: replace the top-heavy header with an app shell, add production cockpit sections, add planner result/recommendation panels, refine task/operator/schedule modals.
- Modify `recsys-backend/web/styles.css`: replace the rounded card-heavy style with a compact operational design system, sidebar shell, timeline, status badges, dense tables, responsive rules.
- Modify `recsys-backend/web/app.js`: add derived planning state, render production cockpit, render recommendation panel, improve timeline rows, operator workload cards, schedule conflict hints, recompute result persistence.
- Modify `recsys-backend/internal/httpapi/handlers.go` only if recommendation history must be exposed separately.
- Modify `recsys-backend/internal/storage/repos.go` only if recommendation history needs a list method.
- Add no new framework or build step.

## Task 1: App Shell And Navigation

**Files:**
- Modify: `recsys-backend/web/index.html`
- Modify: `recsys-backend/web/styles.css`
- Modify: `recsys-backend/web/app.js`

- [ ] **Step 1: Rename the main navigation target from `home` to `plan` in HTML**

In `recsys-backend/web/index.html`, change the first nav button and first page section:

```html
<button class="nav__link is-active" data-page="plan">План производства</button>
```

```html
<section class="page is-active" data-page="plan">
```

- [ ] **Step 2: Replace the header layout with an app shell**

Wrap the existing header and main in a shell:

```html
<div class="app-shell">
  <aside class="app-sidebar">
    <div class="brand brand--sidebar">...</div>
    <div class="workspace-panel">...</div>
    <nav class="nav nav--sidebar">...</nav>
    <div class="sidebar-status">...</div>
  </aside>
  <div class="app-workspace">
    <main class="app-main">...</main>
  </div>
</div>
```

Keep existing element IDs (`workspace-select`, `api-status`, `auth-action`, `dev-tools`) so current JavaScript continues to work.

- [ ] **Step 3: Update navigation JavaScript aliases**

In `recsys-backend/web/app.js`, update any logic that expects `home` as the default page:

```js
const defaultPage = 'plan';
```

If there is no explicit variable, update calls that fallback to `'home'` so they fallback to `'plan'`.

- [ ] **Step 4: Add shell CSS**

In `recsys-backend/web/styles.css`, add:

```css
.app-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  background: #eef2f7;
}

.app-sidebar {
  position: sticky;
  top: 0;
  height: 100vh;
  display: grid;
  grid-template-rows: auto auto 1fr auto;
  gap: 18px;
  padding: 18px;
  background: #101827;
  color: #f8fafc;
}

.app-workspace {
  min-width: 0;
}

.app-main {
  padding: 20px;
}

.nav--sidebar {
  display: grid;
  gap: 6px;
}

.nav--sidebar .nav__link {
  width: 100%;
  padding: 10px 12px;
  border: 0;
  border-radius: 8px;
  color: #cbd5e1;
  text-align: left;
  background: transparent;
}

.nav--sidebar .nav__link.is-active {
  color: #ffffff;
  background: #243149;
  font-weight: 700;
}
```

- [ ] **Step 5: Verify shell**

Run:

```powershell
docker compose up -d --build
```

Expected: API starts and `http://localhost:8080` loads with left sidebar navigation.

- [ ] **Step 6: Commit**

```powershell
git add recsys-backend/web/index.html recsys-backend/web/styles.css recsys-backend/web/app.js
git commit -m "Redesign app shell navigation"
```

## Task 2: Production Plan Cockpit

**Files:**
- Modify: `recsys-backend/web/index.html`
- Modify: `recsys-backend/web/styles.css`
- Modify: `recsys-backend/web/app.js`

- [ ] **Step 1: Replace the old home cards with cockpit regions**

Inside `section[data-page="plan"]`, use this structure:

```html
<section class="plan-cockpit">
  <div class="plan-cockpit__main">
    <div class="metric-grid" id="plan-metrics"></div>
    <section class="panel panel--timeline">
      <div class="panel__header">
        <div>
          <h2>План на сегодня</h2>
          <p class="panel__subtitle" id="home-date"></p>
        </div>
        <div class="toolbar">
          <button class="button button--ghost" id="refresh-devices">Обновить</button>
          <button class="button" id="recompute-plan">Пересчитать план</button>
        </div>
      </div>
      <div class="gantt gantt--production" id="home-gantt"></div>
    </section>
  </div>
  <aside class="plan-cockpit__side">
    <section class="panel">
      <div class="panel__header"><h2>Рекомендация</h2></div>
      <div id="planner-recommendation-panel"></div>
    </section>
    <section class="panel">
      <div class="panel__header"><h2>Предупреждения</h2></div>
      <div id="planner-warning-panel"></div>
    </section>
  </aside>
</section>
```

- [ ] **Step 2: Add DOM references**

In `app.js`, add:

```js
const planMetrics = document.getElementById('plan-metrics');
const plannerRecommendationPanel = document.getElementById('planner-recommendation-panel');
const plannerWarningPanel = document.getElementById('planner-warning-panel');
```

- [ ] **Step 3: Add planner result state**

Extend `state`:

```js
lastPlannerResult: null
```

- [ ] **Step 4: Add derived metrics helper**

Add:

```js
function getPlanMetrics() {
  const scheduledTasks = state.tasks.filter((task) => task.plan_start && task.plan_end);
  const unscheduledTasks = state.tasks.filter((task) => !task.plan_start || !task.plan_end);
  const now = new Date();
  const deadlineRisks = state.tasks.filter((task) => {
    if (!task.deadline || !task.plan_end) return false;
    return new Date(task.plan_end) > new Date(task.deadline) && new Date(task.deadline) >= now;
  });
  const recommendationCount = state.lastPlannerResult?.recommendations?.length || 0;
  return [
    { label: 'В плане', value: scheduledTasks.length },
    { label: 'Риск срока', value: deadlineRisks.length, tone: deadlineRisks.length ? 'danger' : 'ok' },
    { label: 'Не назначено', value: unscheduledTasks.length, tone: unscheduledTasks.length ? 'warning' : 'ok' },
    { label: 'Рекомендаций', value: recommendationCount }
  ];
}
```

- [ ] **Step 5: Render metrics**

Add:

```js
function renderPlanMetrics() {
  if (!planMetrics) return;
  planMetrics.innerHTML = getPlanMetrics()
    .map((metric) => `
      <article class="metric-card metric-card--${metric.tone || 'neutral'}">
        <span>${metric.label}</span>
        <strong>${metric.value}</strong>
      </article>
    `)
    .join('');
}
```

Call `renderPlanMetrics()` from `renderWorkspace()`.

- [ ] **Step 6: Persist recompute result in UI**

In `recomputePlan()`, after the API response:

```js
state.lastPlannerResult = result;
```

Then call:

```js
renderPlannerPanels();
renderPlanMetrics();
```

- [ ] **Step 7: Render recommendation and warning panels**

Add:

```js
function renderPlannerPanels() {
  if (plannerRecommendationPanel) {
    const recommendations = state.lastPlannerResult?.recommendations || [];
    plannerRecommendationPanel.innerHTML = recommendations.length
      ? recommendations.slice(0, 4).map(renderRecommendationItem).join('')
      : '<div class="empty-state">Пересчитайте план, чтобы увидеть рекомендации.</div>';
  }
  if (plannerWarningPanel) {
    const warnings = state.lastPlannerResult?.warnings || [];
    plannerWarningPanel.innerHTML = warnings.length
      ? warnings.map((warning) => `<div class="warning-item">${escapeHTML(warning.message || warning.code)}</div>`).join('')
      : '<div class="empty-state empty-state--ok">Критических предупреждений нет.</div>';
  }
}

function renderRecommendationItem(item) {
  const task = state.tasks.find((candidate) => candidate.id === item.task_id);
  const device = state.devices.find((candidate) => candidate.id === item.device_id);
  const operator = state.operators.find((candidate) => candidate.id === item.operator_id);
  return `
    <article class="recommendation-item">
      <strong>${escapeHTML(task?.name || `Задача #${item.task_id}`)}</strong>
      <span>${escapeHTML(device?.name || 'Оборудование не выбрано')}</span>
      <span>${escapeHTML(operator?.full_name || 'Оператор не выбран')}</span>
      <small>${formatTime(item.plan_start)}–${formatTime(item.plan_end)}</small>
      ${item.explanation ? `<p>${escapeHTML(item.explanation)}</p>` : ''}
    </article>
  `;
}
```

If `escapeHTML` does not exist yet, add it:

```js
function escapeHTML(value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}
```

- [ ] **Step 8: Add cockpit CSS**

Add panel, metric, and recommendation styles:

```css
.plan-cockpit {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 16px;
  align-items: start;
}

.plan-cockpit__main,
.plan-cockpit__side {
  display: grid;
  gap: 16px;
}

.panel {
  background: #ffffff;
  border: 1px solid #dbe3ef;
  border-radius: 8px;
  padding: 16px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(120px, 1fr));
  gap: 12px;
}

.metric-card {
  min-height: 82px;
  display: grid;
  gap: 8px;
  padding: 14px;
  background: #ffffff;
  border: 1px solid #dbe3ef;
  border-radius: 8px;
}

.metric-card span {
  color: #64748b;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}

.metric-card strong {
  font-size: 28px;
}

.metric-card--danger strong { color: #dc2626; }
.metric-card--warning strong { color: #d97706; }
.metric-card--ok strong { color: #15803d; }

.recommendation-item,
.warning-item,
.empty-state {
  padding: 10px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
  display: grid;
  gap: 4px;
  font-size: 13px;
}
```

- [ ] **Step 9: Verify cockpit**

Open `http://localhost:8080`, run recompute, confirm metrics and side panels update.

- [ ] **Step 10: Commit**

```powershell
git add recsys-backend/web/index.html recsys-backend/web/styles.css recsys-backend/web/app.js
git commit -m "Add production planning cockpit"
```

## Task 3: Task Form And Task List UX

**Files:**
- Modify: `recsys-backend/web/index.html`
- Modify: `recsys-backend/web/styles.css`
- Modify: `recsys-backend/web/app.js`

- [ ] **Step 1: Reorder task modal fields**

In `#task-form`, group fields into:

```html
<fieldset class="form-section">
  <legend>Производственный запрос</legend>
  ...
</fieldset>
<fieldset class="form-section form-section--secondary">
  <legend>Назначение планировщика</legend>
  ...
</fieldset>
```

Put `operator_id` and `device_id` in the secondary fieldset with helper text:

```html
<p class="field-hint">Оставьте пустым, чтобы планировщик выбрал ресурс автоматически.</p>
```

- [ ] **Step 2: Make device/operator selects optional in browser validation**

Ensure task device/operator selects have no `required` attribute:

```html
<select name="device_id" id="task-device"></select>
<select name="operator_id" id="task-operator"></select>
```

- [ ] **Step 3: Add task status helpers**

In `app.js`, add:

```js
function getTaskDeadlineTone(task) {
  if (!task.deadline) return 'neutral';
  if (task.plan_end && new Date(task.plan_end) > new Date(task.deadline)) return 'danger';
  const hoursLeft = (new Date(task.deadline) - new Date()) / 36e5;
  if (hoursLeft <= 4) return 'warning';
  return 'ok';
}

function renderTaskBadge(task) {
  const tone = getTaskDeadlineTone(task);
  const label = tone === 'danger' ? 'Просрочка' : tone === 'warning' ? 'Скоро дедлайн' : 'В срок';
  return `<span class="status-badge status-badge--${tone}">${label}</span>`;
}
```

- [ ] **Step 4: Update task row/card rendering**

Where tasks are rendered in tables or upcoming lists, include:

```js
${renderTaskBadge(task)}
<span>${formatTime(task.plan_start)}–${formatTime(task.plan_end)}</span>
<span>${escapeHTML(device?.name || 'Планировщик выберет')}</span>
<span>${escapeHTML(operator?.full_name || 'Планировщик выберет')}</span>
```

- [ ] **Step 5: Add form and badge CSS**

```css
.form-section {
  border: 1px solid #dbe3ef;
  border-radius: 8px;
  padding: 14px;
  display: grid;
  gap: 12px;
}

.form-section legend {
  padding: 0 6px;
  font-weight: 800;
}

.field-hint {
  color: #64748b;
  font-size: 12px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  min-height: 24px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 800;
}

.status-badge--ok { color: #166534; background: #dcfce7; }
.status-badge--warning { color: #92400e; background: #fef3c7; }
.status-badge--danger { color: #991b1b; background: #fee2e2; }
.status-badge--neutral { color: #475569; background: #e2e8f0; }
```

- [ ] **Step 6: Verify task creation**

Create a task with characteristic and without device/operator. Run recompute. Expected: task receives recommended device/operator in UI.

- [ ] **Step 7: Commit**

```powershell
git add recsys-backend/web/index.html recsys-backend/web/styles.css recsys-backend/web/app.js
git commit -m "Improve production task workflow"
```

## Task 4: Operator And Schedule Workload

**Files:**
- Modify: `recsys-backend/web/app.js`
- Modify: `recsys-backend/web/styles.css`
- Modify: `recsys-backend/web/index.html`

- [ ] **Step 1: Add operator workload helper**

In `app.js`, add:

```js
function getOperatorWorkload(operatorId) {
  const tasks = state.tasks.filter((task) => task.operator_id === operatorId && task.plan_start && task.plan_end);
  const breaks = state.userTasks.filter((task) => task.operator_id === operatorId);
  const nextTask = tasks
    .filter((task) => new Date(task.plan_end) >= new Date())
    .sort((a, b) => new Date(a.plan_start) - new Date(b.plan_start))[0];
  return { tasks, breaks, nextTask };
}
```

- [ ] **Step 2: Enrich operator cards**

In `renderOperators()`, include:

```js
const workload = getOperatorWorkload(operator.id);
const nextTaskText = workload.nextTask ? workload.nextTask.name : 'Нет ближайшей задачи';
```

And render:

```html
<div class="operator-card__load">
  <span>${workload.tasks.length} задач в плане</span>
  <span>${workload.breaks.length} интервалов недоступности</span>
  <strong>${escapeHTML(nextTaskText)}</strong>
</div>
```

- [ ] **Step 3: Add schedule conflict helper**

Add:

```js
function taskOverlapsOperatorBreak(task) {
  if (!task.operator_id || !task.plan_start || !task.plan_end) return false;
  const start = new Date(task.plan_start);
  const end = new Date(task.plan_end);
  return state.userTasks.some((entry) => {
    if (entry.operator_id !== task.operator_id || !entry.start_time || !entry.end_time) return false;
    return start < new Date(entry.end_time) && end > new Date(entry.start_time);
  });
}
```

- [ ] **Step 4: Mark conflicting task bars**

Where Gantt bars are built, add a class when `taskOverlapsOperatorBreak(task)` returns true:

```js
const conflictClass = taskOverlapsOperatorBreak(task) ? ' gantt__bar--conflict' : '';
```

Then include it in the bar class string.

- [ ] **Step 5: Add CSS**

```css
.operator-card__load {
  display: grid;
  gap: 4px;
  padding-top: 10px;
  border-top: 1px solid #e2e8f0;
  color: #475569;
  font-size: 13px;
}

.gantt__bar--conflict {
  outline: 2px solid #dc2626;
  outline-offset: 2px;
}
```

- [ ] **Step 6: Verify schedule**

Add an unavailable interval for an operator with a planned task. Expected: schedule shows unavailable bar and conflicting task bar receives conflict outline.

- [ ] **Step 7: Commit**

```powershell
git add recsys-backend/web/index.html recsys-backend/web/styles.css recsys-backend/web/app.js
git commit -m "Show operator workload and schedule conflicts"
```

## Task 5: Equipment And References Planning Settings

**Files:**
- Modify: `recsys-backend/web/index.html`
- Modify: `recsys-backend/web/app.js`
- Modify: `recsys-backend/web/styles.css`

- [ ] **Step 1: Add planning settings section to references**

In references page, add:

```html
<div class="panel">
  <div class="panel__header">
    <h2>Веса критериев планировщика</h2>
  </div>
  <div class="table table--wide" id="planning-weights-list"></div>
</div>
<div class="panel">
  <div class="panel__header">
    <h2>Оценки оборудования по характеристикам</h2>
  </div>
  <div class="table table--wide" id="device-characteristic-scores-list"></div>
</div>
```

- [ ] **Step 2: Add DOM references and state fields**

```js
const planningWeightsList = document.getElementById('planning-weights-list');
const deviceCharacteristicScoresList = document.getElementById('device-characteristic-scores-list');
```

Extend state:

```js
planningWeights: [],
deviceCharacteristicScores: [],
planningCriteria: []
```

- [ ] **Step 3: Load planning settings with workspace data**

In `loadWorkspaceData()`, add requests:

```js
fetchJSON(`${apiBase}/planning-criteria`),
fetchJSON(`${apiBase}/workspaces/${workspaceId}/planning-weights`),
fetchJSON(`${apiBase}/workspaces/${workspaceId}/device-characteristic-scores`)
```

Assign results to state.

- [ ] **Step 4: Render planning settings tables**

Add:

```js
function renderPlanningSettings() {
  if (planningWeightsList) {
    planningWeightsList.innerHTML = renderTable(
      ['Критерий', 'Вес'],
      state.planningWeights.map((weight) => [
        escapeHTML(state.planningCriteria.find((item) => item.code === weight.criterion_code)?.name || weight.criterion_code),
        Number(weight.weight).toFixed(2)
      ])
    );
  }
  if (deviceCharacteristicScoresList) {
    const devicesById = mapById(state.devices);
    const characteristicsById = mapById(state.equipmentCharacteristics);
    deviceCharacteristicScoresList.innerHTML = renderTable(
      ['Оборудование', 'Характеристика', 'Оценка'],
      state.deviceCharacteristicScores.map((score) => [
        escapeHTML(devicesById[score.device_id]?.name || `#${score.device_id}`),
        escapeHTML(characteristicsById[score.equipment_characteristic_id]?.name || `#${score.equipment_characteristic_id}`),
        Number(score.score).toFixed(0)
      ])
    );
  }
}
```

If there is no reusable `renderTable`, add a local helper:

```js
function renderTable(headers, rows) {
  if (!rows.length) return '<div class="empty-state">Нет данных</div>';
  return `
    <table>
      <thead><tr>${headers.map((header) => `<th>${escapeHTML(header)}</th>`).join('')}</tr></thead>
      <tbody>${rows.map((row) => `<tr>${row.map((cell) => `<td>${cell}</td>`).join('')}</tr>`).join('')}</tbody>
    </table>
  `;
}
```

- [ ] **Step 5: Add equipment score to equipment card**

In equipment rendering, show the best score:

```js
const scores = state.deviceCharacteristicScores.filter((score) => score.device_id === device.id);
const bestScore = scores.sort((a, b) => Number(b.score) - Number(a.score))[0];
```

Render:

```html
<span>Оценка планировщика: ${bestScore ? Number(bestScore.score).toFixed(0) : 'не задана'}</span>
```

- [ ] **Step 6: Verify references**

Open references screen. Expected: weights and device-characteristic scores are visible in compact tables.

- [ ] **Step 7: Commit**

```powershell
git add recsys-backend/web/index.html recsys-backend/web/styles.css recsys-backend/web/app.js
git commit -m "Expose planning settings in UI"
```

## Task 6: Responsive Polish And Browser Verification

**Files:**
- Modify: `recsys-backend/web/styles.css`
- Modify: `recsys-backend/web/app.js` only for final UI defects

- [ ] **Step 1: Add responsive shell rules**

```css
@media (max-width: 1100px) {
  .app-shell {
    grid-template-columns: 1fr;
  }

  .app-sidebar {
    position: static;
    height: auto;
    grid-template-rows: auto;
  }

  .nav--sidebar {
    display: flex;
    overflow-x: auto;
  }

  .plan-cockpit {
    grid-template-columns: 1fr;
  }

  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .app-main {
    padding: 12px;
  }

  .metric-grid {
    grid-template-columns: 1fr;
  }

  .panel {
    padding: 12px;
  }
}
```

- [ ] **Step 2: Run backend tests**

```powershell
docker run --rm -v ${PWD}/recsys-backend:/src -w /src golang:1.24-alpine go test ./...
```

Expected: all packages pass.

- [ ] **Step 3: Rebuild app**

```powershell
docker compose up -d --build
```

Expected: `recsys-api` and `recsys-postgres` are running.

- [ ] **Step 4: Browser smoke test**

Open `http://localhost:8080` and verify:

- sidebar is visible on desktop;
- plan screen has metrics, timeline, recommendation panel;
- recompute updates the panel;
- tasks page opens;
- equipment page opens;
- operators page opens;
- schedule page opens;
- references page shows planning settings;
- no obvious text overlap at desktop width;
- narrow viewport stacks content vertically.

- [ ] **Step 5: Final commit**

```powershell
git add recsys-backend/web/styles.css recsys-backend/web/app.js
git commit -m "Polish production UI responsiveness"
```

## Self-Review

- Spec coverage:
  - App shell and navigation: Task 1.
  - Production plan cockpit: Task 2.
  - Task UX: Task 3.
  - Planner recompute explanations and warnings: Task 2.
  - Operator UX: Task 4.
  - Schedule UX: Task 4.
  - Equipment and planning references: Task 5.
  - Responsive visual system and verification: Task 6.
- Placeholder scan:
  - No TBD/TODO placeholders.
  - Each code-changing step includes concrete code or concrete edit instructions.
- Type consistency:
  - Frontend fields use existing JSON names: `plan_start`, `plan_end`, `device_id`, `operator_id`, `equipment_characteristic_id`, `criterion_code`, `weight`, `score`.
  - Existing element IDs are preserved where current JavaScript depends on them.
