const apiBase = '/api';
const statusBadge = document.getElementById('api-status');
const userGreeting = document.getElementById('user-greeting');
const authActionButton = document.getElementById('auth-action');
const devTools = document.getElementById('dev-tools');
const seedDbBtn = document.getElementById('seed-db');
const clearDbBtn = document.getElementById('clear-db');
const workspaceSelect = document.getElementById('workspace-select');
const refreshWorkspacesBtn = document.getElementById('refresh-workspaces');
const refreshDevicesBtn = document.getElementById('refresh-devices');
const recomputePlanBtn = document.getElementById('recompute-plan');
const openWorkspaceModalBtn = document.getElementById('open-workspace-modal');
const openTaskModalBtn = document.getElementById('open-task-modal');
const openOperatorModalBtn = document.getElementById('open-operator-modal');
const openDeviceModalBtn = document.getElementById('open-device-modal');
const openScheduleModalBtn = document.getElementById('open-schedule-modal');
const tasksDateInput = document.getElementById('tasks-date');
const tasksSortSelect = document.getElementById('tasks-sort');
const homeStats = document.getElementById('home-stats');
const upcomingTasks = document.getElementById('upcoming-tasks');
const entitySummary = document.getElementById('entity-summary');
const homeGantt = document.getElementById('home-gantt');
const tasksGantt = document.getElementById('tasks-gantt');
const homeGanttLegend = document.getElementById('home-gantt-legend');
const tasksGanttLegend = document.getElementById('tasks-gantt-legend');
const operatorsGanttLegend = document.getElementById('operators-gantt-legend');
const equipmentCards = document.getElementById('equipment-cards');
const equipmentList = document.getElementById('equipment-list');
const operatorsList = document.getElementById('operators-list');
const operatorsGantt = document.getElementById('operators-gantt');
const homeDateLabel = document.getElementById('home-date');
const profileSummary = document.getElementById('profile-summary');
const loginForm = document.getElementById('login-form');
const registerForm = document.getElementById('register-form');
const logoutButton = document.getElementById('logout-button');
const usersList = document.getElementById('users-list');
const adminPanel = document.getElementById('admin-panel');
const equipmentCharacteristicForm = document.getElementById('equipment-characteristic-form');
const deviceTypeForm = document.getElementById('device-type-form');
const taskTypeForm = document.getElementById('task-type-form');
const equipmentCharacteristicsList = document.getElementById('equipment-characteristics-list');
const deviceTypesList = document.getElementById('device-types-list');
const taskTypesList = document.getElementById('task-types-list');
const equipmentCharacteristicSelect = document.getElementById('equipment-characteristic-select');
const toastContainer = document.getElementById('toast-container');

const workspaceModal = document.getElementById('workspace-modal');
const taskModal = document.getElementById('task-modal');
const operatorModal = document.getElementById('operator-modal');
const deviceModal = document.getElementById('device-modal');
const scheduleModal = document.getElementById('schedule-modal');
const taskModalTitle = document.getElementById('task-modal-title');
const operatorModalTitle = document.getElementById('operator-modal-title');
const deviceModalTitle = document.getElementById('device-modal-title');
const scheduleModalTitle = document.getElementById('schedule-modal-title');

const workspaceForm = document.getElementById('workspace-form');
const taskForm = document.getElementById('task-form');
const operatorForm = document.getElementById('operator-form');
const deviceForm = document.getElementById('device-form');
const scheduleForm = document.getElementById('schedule-form');

const taskOperatorSelect = document.getElementById('task-operator');
const taskTypeSelect = document.getElementById('task-type');
const taskPrioritySelect = document.getElementById('task-priority');
const taskDeviceSelect = document.getElementById('task-device');
const deviceTypeSelect = document.getElementById('device-type');
const deviceStateSelect = document.getElementById('device-state');
const scheduleOperatorSelect = document.getElementById('schedule-operator');
const scheduleTypeSelect = document.getElementById('schedule-type');

const navLinks = document.querySelectorAll('.nav__link');
const pages = document.querySelectorAll('.page');

const state = {
  workspaces: [],
  devices: [],
  operators: [],
  tasks: [],
  deviceTypes: [],
  equipmentCharacteristics: [],
  deviceStates: [],
  priorities: [],
  taskTypes: [],
  operatorDevices: [],
  operatorCompetencies: [],
  userTasks: [],
  currentUser: null
};

const dayStartHour = 9;
const dayEndHour = 22;
const ganttHours = Array.from({ length: dayEndHour - dayStartHour }, (_, i) => dayStartHour + i);
const ganttTotalMinutes = (dayEndHour - dayStartHour) * 60;
const storedToken = localStorage.getItem('authToken');
let authToken = storedToken || '';
let autoRefreshTimer = null;
let ganttDragState = null;
let suppressGanttClick = false;
let pendingOverlapCheck = false;

async function fetchJSON(url, options) {
  const headers = new Headers(options?.headers || {});
  if (authToken) {
    headers.set('Authorization', `Bearer ${authToken}`);
  }
  const response = await fetch(url, { ...options, headers });
  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || 'Ошибка запроса');
  }
  if (response.status === 204) {
    return null;
  }
  return response.json();
}

function formatTime(value) {
  if (!value) return '—';
  const date = typeof value === 'string' ? new Date(value) : value;
  return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
}

function formatDate(value) {
  if (!value) return '—';
  const date = typeof value === 'string' ? new Date(value) : value;
  return date.toLocaleDateString('ru-RU');
}

function toLocalInputValue(date) {
  const pad = (num) => String(num).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

function toLocalDateTimeValue(date) {
  if (!date) return '';
  const pad = (num) => String(num).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(
    date.getHours()
  )}:${pad(date.getMinutes())}`;
}

function parseDateInput(value) {
  if (!value) return null;
  const [year, month, day] = value.split('-').map(Number);
  return new Date(year, month - 1, day);
}

function parseDateTimeInput(value) {
  if (!value) return null;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

function formatPhoneInput(value) {
  const digits = value.replace(/\D/g, '').replace(/^8/, '7');
  const cleaned = digits.startsWith('7') ? digits.slice(1) : digits;
  const parts = [
    cleaned.slice(0, 3),
    cleaned.slice(3, 6),
    cleaned.slice(6, 8),
    cleaned.slice(8, 10)
  ].filter(Boolean);
  let formatted = '+7';
  if (parts[0]) formatted += ` (${parts[0]}`;
  if (parts[0] && parts[0].length === 3) formatted += ')';
  if (parts[1]) formatted += ` ${parts[1]}`;
  if (parts[2]) formatted += `-${parts[2]}`;
  if (parts[3]) formatted += `-${parts[3]}`;
  return formatted;
}

function formatDocNumber(value, pad = false) {
  const digits = value.replace(/\D/g, '').slice(0, 6);
  if (!digits) return '';
  const padded = pad ? digits.padStart(3, '0') : digits;
  return `DOC-${padded}`;
}

function applyInputMasks() {
  const phoneInput = operatorForm?.querySelector('input[name="phone_number"]');
  if (phoneInput) {
    phoneInput.addEventListener('input', () => {
      phoneInput.value = formatPhoneInput(phoneInput.value);
    });
  }

  const docInput = taskForm?.querySelector('input[name="doc_num"]');
  if (docInput) {
    docInput.addEventListener('input', () => {
      docInput.value = formatDocNumber(docInput.value);
    });
    docInput.addEventListener('blur', () => {
      docInput.value = formatDocNumber(docInput.value, true);
    });
  }
}

function getWorkspaceId() {
  return Number(workspaceSelect.value);
}

function getStoredWorkspaceId() {
  return localStorage.getItem('workspaceId');
}

function setStoredWorkspaceId(id) {
  if (id) {
    localStorage.setItem('workspaceId', String(id));
  } else {
    localStorage.removeItem('workspaceId');
  }
}

function mapById(items) {
  return items.reduce((acc, item) => {
    acc[item.id] = item;
    return acc;
  }, {});
}

function showToast(message, variant = 'info') {
  if (!toastContainer) {
    alert(message);
    return;
  }
  const toast = document.createElement('div');
  toast.className = `toast toast--${variant}`;
  toast.textContent = message;
  toastContainer.appendChild(toast);
  requestAnimationFrame(() => {
    toast.classList.add('is-visible');
  });
  setTimeout(() => {
    toast.classList.remove('is-visible');
    toast.addEventListener(
      'transitionend',
      () => {
        toast.remove();
      },
      { once: true }
    );
  }, 5000);
}

function showSaveToast(label = 'Изменения') {
  showToast(`${label} сохранены.`, 'info');
}

function resetWorkspaceState() {
  state.devices = [];
  state.operators = [];
  state.tasks = [];
  state.deviceTypes = [];
  state.equipmentCharacteristics = [];
  state.taskTypes = [];
  state.operatorDevices = [];
  state.operatorCompetencies = [];
  state.userTasks = [];
  renderAll();
}

function openModal(modal) {
  if (!modal.open) {
    modal.showModal();
  }
}

function closeModal(modal) {
  if (modal.open) {
    modal.close();
  }
}

function setStatus(ok) {
  if (ok) {
    statusBadge.textContent = 'API: доступно';
    statusBadge.style.background = '#e9f7ec';
    statusBadge.style.color = '#1d7b3d';
  } else {
    statusBadge.textContent = 'API: недоступно';
    statusBadge.style.background = '#fdeaea';
    statusBadge.style.color = '#b91c1c';
  }
}

async function checkHealth() {
  try {
    await fetchJSON('/health');
    setStatus(true);
  } catch {
    setStatus(false);
  }
}

function setActivePage(page) {
  navLinks.forEach((link) => {
    link.classList.toggle('is-active', link.dataset.page === page);
  });
  pages.forEach((section) => {
    section.classList.toggle('is-active', section.dataset.page === page);
  });
}

function scheduleAutoRefresh() {
  if (autoRefreshTimer) clearTimeout(autoRefreshTimer);
  autoRefreshTimer = setTimeout(async () => {
    autoRefreshTimer = null;
    await refreshWorkspaceView();
  }, 200);
}

async function refreshWorkspaceView() {
  if (!getWorkspaceId()) {
    renderAll();
    return;
  }
  await loadWorkspaceData();
  await loadUsers();
}

async function loadWorkspaces() {
  const query = state.currentUser ? `?user_login=${encodeURIComponent(state.currentUser.login)}` : '';
  state.workspaces = await fetchJSON(`${apiBase}/workspaces/${query}`);
  workspaceSelect.innerHTML = '';
  if (!state.workspaces.length) {
    const opt = document.createElement('option');
    opt.textContent = 'Нет данных';
    opt.value = '';
    workspaceSelect.appendChild(opt);
    setStoredWorkspaceId('');
    resetWorkspaceState();
    return;
  }
  state.workspaces.forEach((workspace) => {
    const opt = document.createElement('option');
    opt.value = workspace.id;
    opt.textContent = `${workspace.name} (#${workspace.id})`;
    workspaceSelect.appendChild(opt);
  });
  const storedWorkspaceId = getStoredWorkspaceId();
  const storedWorkspace = state.workspaces.find((workspace) => String(workspace.id) === storedWorkspaceId);
  if (storedWorkspace) {
    workspaceSelect.value = String(storedWorkspace.id);
  } else if (!workspaceSelect.value) {
    workspaceSelect.value = String(state.workspaces[0].id);
  }
  setStoredWorkspaceId(workspaceSelect.value);
}

async function loadReferenceData() {
  state.deviceStates = await fetchJSON(`${apiBase}/device-states/`);
  state.priorities = await fetchJSON(`${apiBase}/priorities/`);
}

async function loadWorkspaceData() {
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return;
  const [
    devices,
    operators,
    tasks,
    deviceTypes,
    equipmentCharacteristics,
    taskTypes,
    operatorDevices,
    operatorCompetencies,
    userTasks
  ] = await Promise.all([
    fetchJSON(`${apiBase}/workspaces/${workspaceId}/devices`),
    fetchJSON(`${apiBase}/workspaces/${workspaceId}/operators`),
    fetchJSON(`${apiBase}/workspaces/${workspaceId}/device-tasks`),
    fetchJSON(`${apiBase}/workspaces/${workspaceId}/device-types`),
    fetchJSON(`${apiBase}/workspaces/${workspaceId}/equipment-characteristics`),
    fetchJSON(`${apiBase}/workspaces/${workspaceId}/device-task-types`),
    fetchJSON(`${apiBase}/workspaces/${workspaceId}/operator-devices`),
    fetchJSON(`${apiBase}/workspaces/${workspaceId}/operator-competencies`),
    fetchJSON(`${apiBase}/workspaces/${workspaceId}/user-tasks`)
  ]);
  state.devices = devices;
  state.operators = operators;
  state.tasks = tasks;
  state.deviceTypes = deviceTypes;
  state.equipmentCharacteristics = equipmentCharacteristics;
  state.taskTypes = taskTypes;
  state.operatorDevices = operatorDevices;
  state.operatorCompetencies = operatorCompetencies;
  state.userTasks = userTasks;
  renderAll();
  if (pendingOverlapCheck) {
    notifyAllTaskBreakOverlaps();
    pendingOverlapCheck = false;
  }
}

function getTasksForDate(date) {
  if (!date) return [];
  return state.tasks.filter((task) => {
    if (!task.plan_start && !task.plan_end && !task.deadline) return false;
    const dateValue = task.plan_start || task.plan_end || task.deadline;
    const taskDate = new Date(dateValue);
    return taskDate.toDateString() === date.toDateString();
  });
}

function getTaskStatus(task) {
  const now = new Date();
  if (task.plan_end && new Date(task.plan_end) < now) return 'done';
  if (task.plan_start && task.plan_end) {
    const start = new Date(task.plan_start);
    const end = new Date(task.plan_end);
    if (start <= now && now <= end) return 'progress';
  }
  return 'pending';
}

function getStatusLabel(status) {
  const labels = {
    done: 'Завершено',
    progress: 'В работе',
    pending: 'Ожидает',
    lunch: 'Обед',
    off: 'Не работает'
  };
  return labels[status] || status || 'Без статуса';
}

function getTaskTimeRange(task) {
  const startValue = task._start
    ? new Date(task._start)
    : task.plan_start
    ? new Date(task.plan_start)
    : null;
  const endValue = task._end
    ? new Date(task._end)
    : task.plan_end
    ? new Date(task.plan_end)
    : startValue && task.duration_min
    ? new Date(startValue.getTime() + task.duration_min * 60000)
    : null;
  return { startValue, endValue };
}

function getMinutesFromDayStart(date) {
  if (!date) return null;
  return date.getHours() * 60 + date.getMinutes() - dayStartHour * 60;
}

function buildDateFromMinutes(date, minutesFromStart) {
  const base = new Date(date);
  base.setHours(dayStartHour, 0, 0, 0);
  base.setMinutes(base.getMinutes() + minutesFromStart);
  return base;
}

function buildGantt(container, tasks, labelFormatter) {
  if (!tasks.length) {
    container.innerHTML = '<div class="gantt__empty">Нет данных для отображения</div>';
    return;
  }

  container.style.setProperty('--gantt-hours', ganttHours.length);
  const header = document.createElement('div');
  header.className = 'gantt__header';
  header.innerHTML = `<div>Название</div>${ganttHours
    .map((hour) => `<div>${hour.toString().padStart(2, '0')}:00</div>`)
    .join('')}`;

  container.innerHTML = '';
  container.appendChild(header);

  const dayStart = dayStartHour * 60;
  const dayEnd = dayEndHour * 60;
  const totalMinutes = dayEnd - dayStart;

  tasks.forEach((task) => {
    const row = document.createElement('div');
    row.className = 'gantt__row';

    const label = document.createElement('div');
    label.className = 'gantt__label';
    label.innerHTML = labelFormatter(task);

    const track = document.createElement('div');
    track.className = 'gantt__track';

    const { startValue, endValue } = getTaskTimeRange(task);

    if (startValue && endValue) {
      const startMinutes = Math.max(0, getMinutesFromDayStart(startValue));
      const endMinutes = Math.min(totalMinutes, getMinutesFromDayStart(endValue));
      const width = Math.max(4, ((endMinutes - startMinutes) / totalMinutes) * 100);
      const left = (startMinutes / totalMinutes) * 100;
      const bar = document.createElement('div');
      const status = task._status || getTaskStatus(task);
      bar.className = `gantt__bar ${status}`;
      bar.style.left = `${left}%`;
      bar.style.width = `${width}%`;
      bar.textContent = `${formatTime(startValue)} – ${formatTime(endValue)}`;
      bar.dataset.status = status;
      bar.title = `${formatTime(startValue)} – ${formatTime(endValue)} · ${getStatusLabel(
        status
      )}`;
      if (task.id) {
        bar.dataset.taskId = task.id;
        bar.classList.add('is-draggable');
      }
      if (task._userTaskId) {
        bar.dataset.userTaskId = task._userTaskId;
        bar.classList.add('is-draggable');
      }
      track.appendChild(bar);
    }

    row.appendChild(label);
    row.appendChild(track);
    container.appendChild(row);
  });
}

function buildTaskPayload(task, overrides) {
  return {
    name: task.name,
    doc_num: task.doc_num,
    photo_url: task.photo_url,
    deadline: task.deadline ? new Date(task.deadline) : null,
    operator_id: Number(task.operator_id || 0),
    device_id: Number(task.device_id || 0),
    priority_id: Number(task.priority_id || 0),
    device_task_type_id: Number(task.device_task_type_id || 0),
    duration_min: Number(task.duration_min || 0),
    setup_time_min: Number(task.setup_time_min || 0),
    unload_time_min: Number(task.unload_time_min || 0),
    need_operator: Boolean(task.need_operator),
    add_in_rec_system: Boolean(task.add_in_rec_system),
    plan_start: task.plan_start ? new Date(task.plan_start) : null,
    plan_end: task.plan_end ? new Date(task.plan_end) : null,
    ...overrides
  };
}

function buildUserTaskPayload(userTask, overrides) {
  return {
    name: userTask.name,
    start_time: userTask.start_time ? new Date(userTask.start_time) : null,
    end_time: userTask.end_time ? new Date(userTask.end_time) : null,
    priority: userTask.priority ?? null,
    completion_mark: userTask.completion_mark ?? null,
    device_task_id: userTask.device_task_id ?? null,
    operator_id: Number(userTask.operator_id || 0),
    ...overrides
  };
}

function hasDeviceOverlap(taskId, deviceId, start, end) {
  return state.tasks.some((task) => {
    if (task.id === taskId) return false;
    if (task.device_id !== deviceId) return false;
    if (!task.plan_start || !task.plan_end) return false;
    const otherStart = new Date(task.plan_start);
    const otherEnd = new Date(task.plan_end);
    return start < otherEnd && end > otherStart;
  });
}

async function updateTaskSchedule(task, newStart, newEnd) {
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return false;
  const payload = buildTaskPayload(task, { plan_start: newStart, plan_end: newEnd });
  await fetchJSON(`${apiBase}/device-tasks/${task.id}?workspace_id=${workspaceId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  await loadWorkspaceData();
  notifyTaskBreakOverlap(task, newStart, newEnd);
  return true;
}

async function updateUserTaskSchedule(userTask, newStart, newEnd) {
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return false;
  const payload = buildUserTaskPayload(userTask, { start_time: newStart, end_time: newEnd });
  await fetchJSON(`${apiBase}/user-tasks/${userTask.id}?workspace_id=${workspaceId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  await loadWorkspaceData();
  return true;
}

function startGanttDrag(event) {
  const bar = event.target.closest('.gantt__bar');
  if (!bar) return;
  const taskId = bar.dataset.taskId;
  const userTaskId = bar.dataset.userTaskId;
  if (!taskId && !userTaskId) return;
  let task = null;
  let startValue = null;
  let endValue = null;
  let isUserTask = false;
  if (userTaskId) {
    task = state.userTasks.find((item) => item.id === Number(userTaskId));
    if (!task?.start_time || !task?.end_time) return;
    startValue = new Date(task.start_time);
    endValue = new Date(task.end_time);
    isUserTask = true;
  } else {
    task = state.tasks.find((item) => item.id === Number(taskId));
    if (!task?.plan_start || !task?.plan_end) return;
    startValue = new Date(task.plan_start);
    endValue = new Date(task.plan_end);
  }
  const startMinutes = getMinutesFromDayStart(startValue);
  const endMinutes = getMinutesFromDayStart(endValue);
  if (startMinutes === null || endMinutes === null) return;
  const track = bar.parentElement;
  const trackRect = track.getBoundingClientRect();
  if (!trackRect.width) return;
  ganttDragState = {
    bar,
    task,
    startMinutes,
    durationMinutes: endMinutes - startMinutes,
    trackWidth: trackRect.width,
    startX: event.clientX,
    newStartMinutes: startMinutes,
    originalLeft: bar.style.left,
    pointerId: event.pointerId,
    moved: false,
    isUserTask
  };
  bar.classList.add('is-dragging');
  bar.setPointerCapture(event.pointerId);
  event.preventDefault();
}

function moveGanttDrag(event) {
  if (!ganttDragState) return;
  const delta = event.clientX - ganttDragState.startX;
  if (Math.abs(delta) > 3) {
    ganttDragState.moved = true;
  }
  const minutesDelta = (delta / ganttDragState.trackWidth) * ganttTotalMinutes;
  let newStartMinutes = ganttDragState.startMinutes + minutesDelta;
  newStartMinutes = Math.max(
    0,
    Math.min(ganttTotalMinutes - ganttDragState.durationMinutes, newStartMinutes)
  );
  ganttDragState.newStartMinutes = newStartMinutes;
  ganttDragState.bar.style.left = `${(newStartMinutes / ganttTotalMinutes) * 100}%`;
}

async function endGanttDrag() {
  if (!ganttDragState) return;
  const {
    bar,
    task,
    durationMinutes,
    newStartMinutes,
    originalLeft,
    moved,
    pointerId,
    isUserTask
  } = ganttDragState;
  ganttDragState = null;
  bar.classList.remove('is-dragging');
  if (pointerId !== undefined) {
    bar.releasePointerCapture?.(pointerId);
  }
  if (!moved) return;
  suppressGanttClick = true;
  setTimeout(() => {
    suppressGanttClick = false;
  }, 0);
  const roundedStartMinutes = Math.round(newStartMinutes);
  const baseDate = new Date(isUserTask ? task.start_time : task.plan_start);
  const newStart = buildDateFromMinutes(baseDate, roundedStartMinutes);
  const newEnd = new Date(newStart.getTime() + durationMinutes * 60000);
  try {
    if (!isUserTask) {
      if (hasDeviceOverlap(task.id, task.device_id, newStart, newEnd)) {
        alert('Нельзя пересекать задания на одном принтере.');
        bar.style.left = originalLeft;
        return;
      }
      await updateTaskSchedule(task, newStart, newEnd);
    } else {
      await updateUserTaskSchedule(task, newStart, newEnd);
    }
  } catch (error) {
    console.error(error);
    alert('Не удалось обновить время задачи.');
    bar.style.left = originalLeft;
  }
}

function renderHomeStats() {
  const today = new Date();
  const todayTasks = getTasksForDate(today);
  const now = new Date();
  const inProgress = todayTasks.filter((task) => {
    if (!task.plan_start || !task.plan_end) return false;
    const start = new Date(task.plan_start);
    const end = new Date(task.plan_end);
    return start <= now && now <= end;
  });
  const pending = todayTasks.filter((task) => task.plan_start && new Date(task.plan_start) > now);
  const completed = todayTasks.filter((task) => task.plan_end && new Date(task.plan_end) < now);

  const closestEnd = inProgress
    .map((task) => task.plan_end)
    .filter(Boolean)
    .sort()[0];
  const nextTask = pending
    .map((task) => task.plan_start)
    .filter(Boolean)
    .sort()[0];

  const loadPercent = state.devices.length
    ? Math.round((inProgress.length / state.devices.length) * 100)
    : 0;

  homeStats.innerHTML = [
    { label: 'Задания в работе', value: inProgress.length },
    { label: 'Ожидают выполнения', value: pending.length },
    { label: 'Выполнено сегодня', value: completed.length },
    { label: 'Ближайшее завершение', value: closestEnd ? formatTime(closestEnd) : '—' },
    { label: 'Следующее задание', value: nextTask ? formatTime(nextTask) : '—' },
    { label: 'Загрузка оборудования', value: `${loadPercent}%` }
  ]
    .map(
      (stat) => `
      <div class="stat-card clickable" data-stat="${stat.label}">
        <h4>${stat.label}</h4>
        <span>${stat.value}</span>
      </div>
    `
    )
    .join('');

  homeDateLabel.textContent = formatDate(today);
}

function renderEntitySummary() {
  if (!entitySummary) return;
  if (!getWorkspaceId()) {
    entitySummary.innerHTML = '<div class="gantt__empty">Выберите рабочее пространство</div>';
    return;
  }
  const items = [
    { label: 'Оборудование', value: state.devices.length },
    { label: 'Типы оборудования', value: state.deviceTypes.length },
    { label: 'Характеристики', value: state.equipmentCharacteristics.length },
    { label: 'Операторы', value: state.operators.length },
    { label: 'Задания', value: state.tasks.length },
    { label: 'Типы заданий', value: state.taskTypes.length },
    { label: 'Компетенции', value: state.operatorCompetencies.length },
    { label: 'Поручения', value: state.userTasks.length }
  ];
  entitySummary.innerHTML = items
    .map(
      (item) => `
        <div class="entity-card clickable" data-entity="${item.label}">
          <strong>${item.value}</strong>
          <span>${item.label}</span>
        </div>
      `
    )
    .join('');
}

function renderUpcomingTasks() {
  const operatorsById = mapById(state.operators);
  const devicesById = mapById(state.devices);
  const taskTypesById = mapById(state.taskTypes);
  const prioritiesById = mapById(state.priorities);
  const upcoming = state.tasks
    .filter((task) => task.plan_start)
    .sort((a, b) => new Date(a.plan_start) - new Date(b.plan_start))
    .slice(0, 4);

  if (!upcoming.length) {
    upcomingTasks.innerHTML = '<div class="gantt__empty">Нет ближайших задач</div>';
    return;
  }

  const header = `
    <div class="table__row">
      <strong>Название</strong>
      <strong>Начало</strong>
      <strong>Завершение</strong>
    </div>
  `;

  const rows = upcoming
    .map((task) => {
      const operator = operatorsById[task.operator_id];
      const device = devicesById[task.device_id];
      const taskType = taskTypesById[task.device_task_type_id];
      const priority = prioritiesById[task.priority_id];
      const priorityBadge = priority
        ? `<span class="badge badge--info">${priority.name}</span>`
        : '<span class="badge">Без приоритета</span>';
      return `
        <div class="table__row clickable" data-task-id="${task.id}">
          <div>
            <strong>${task.name}</strong><br />
            <span class="muted">${operator ? operator.full_name : 'Оператор не назначен'} ·
            ${device ? device.name : 'Оборудование не выбрано'}</span>
            <div class="muted">Тип: ${taskType ? taskType.name : '—'} · Приоритет: ${
        priority ? priority.name : '—'
      }</div>
            <div class="device-card__badges">${priorityBadge}</div>
          </div>
          <div>${formatTime(task.plan_start)}</div>
          <div>${formatTime(task.plan_end)}</div>
        </div>
      `;
    })
    .join('');

  upcomingTasks.innerHTML = header + rows;
}

function renderHomeGantt() {
  const todayTasks = getTasksForDate(new Date()).sort((a, b) => {
    const aDate = a.plan_start || a.deadline || 0;
    const bDate = b.plan_start || b.deadline || 0;
    return new Date(aDate) - new Date(bDate);
  });
  buildGantt(homeGantt, todayTasks, (task) => {
    const operator = state.operators.find((item) => item.id === task.operator_id);
    const taskType = state.taskTypes.find((item) => item.id === task.device_task_type_id);
    return `
      <span class="gantt__label-title">${task.name}</span>
      <small>${operator ? operator.full_name : 'Оператор не назначен'} · ${
      taskType ? taskType.name : 'Тип не указан'
    }</small>
    `;
  });
}

function renderEquipment() {
  const deviceTypesById = mapById(state.deviceTypes);
  const characteristicsById = mapById(state.equipmentCharacteristics);
  const deviceStatesById = mapById(state.deviceStates);
  const tasksByDevice = state.tasks.reduce((acc, task) => {
    if (!acc[task.device_id]) acc[task.device_id] = [];
    acc[task.device_id].push(task);
    return acc;
  }, {});

  const makeCard = (device) => {
    const deviceTasks = tasksByDevice[device.id] || [];
    const activeTask = deviceTasks.find((task) => getTaskStatus(task) === 'progress');
    const progressPercent = activeTask ? 70 : deviceTasks.length ? 40 : 0;
    const deviceType = deviceTypesById[device.device_type_id];
    const characteristicName =
      characteristicsById[deviceType?.equipment_characteristic_id]?.name || '—';
    const stateLabel = deviceStatesById[device.device_state_id]?.name || 'Состояние неизвестно';
    const stateBadgeClass = stateLabel.toLowerCase().includes('авар')
      ? 'badge--danger'
      : stateLabel.toLowerCase().includes('ремонт')
      ? 'badge--warning'
      : 'badge--success';
    const recBadge = device.add_in_rec_system
      ? '<span class="badge badge--info">В рекомендациях</span>'
      : '<span class="badge">Без рекомендаций</span>';
    return `
      <div class="device-card clickable" data-device-id="${device.id}">
        <img src="${device.photo_url || 'https://placehold.co/400x240?text=3D+Printer'}" alt="${device.name}" />
        <div>
          <strong>${device.name}</strong>
          <div class="muted">${deviceType?.name || 'Тип не указан'}</div>
          <div class="muted">Характеристика: ${characteristicName}</div>
        </div>
        <div class="device-card__badges">
          <span class="badge ${stateBadgeClass}">${stateLabel}</span>
          ${recBadge}
        </div>
        <div class="device-card__status">
          <span>${activeTask ? 'В работе' : 'Простой'}</span>
          <span>${activeTask ? `До ${formatTime(activeTask.plan_end)}` : 'Нет активных задач'}</span>
        </div>
        <div class="device-card__progress">
          <span style="width: ${progressPercent}%"></span>
        </div>
        <div class="muted">${activeTask ? `Задача: ${activeTask.name}` : 'Нет активных задач'}</div>
      </div>
    `;
  };

  const cardsHtml = state.devices.map(makeCard).join('');
  const empty = '<div class="gantt__empty">Нет добавленного оборудования</div>';
  equipmentCards.innerHTML = cardsHtml || empty;
  equipmentList.innerHTML = cardsHtml || empty;
}

function renderOperators() {
  const deviceTypesById = mapById(state.deviceTypes);
  const devicesById = mapById(state.devices);
  const tasksByOperator = state.tasks.reduce((acc, task) => {
    if (!acc[task.operator_id]) acc[task.operator_id] = [];
    acc[task.operator_id].push(task);
    return acc;
  }, {});

  if (!state.operators.length) {
    operatorsList.innerHTML = '<div class="gantt__empty">Нет операторов</div>';
    return;
  }

  operatorsList.innerHTML = state.operators
    .map((operator) => {
      const competencies = state.operatorCompetencies
        .filter((item) => item.operator_id === operator.id)
        .map((item) => deviceTypesById[item.device_type_id]?.name || `#${item.device_type_id}`);
      const responsibilities = state.operatorDevices
        .filter((item) => item.operator_id === operator.id)
        .map((item) => devicesById[item.device_id]?.name || `#${item.device_id}`);
      const tasks = tasksByOperator[operator.id] || [];
      const nextTask = tasks
        .filter((task) => task.plan_start)
        .sort((a, b) => new Date(a.plan_start) - new Date(b.plan_start))[0];
      const competenciesHtml = competencies.length
        ? competencies.map((item) => `<span class="chip">${item}</span>`).join('')
        : '<span class="chip">Не указаны</span>';
      const responsibilitiesHtml = responsibilities.length
        ? responsibilities.map((item) => `<span class="chip">${item}</span>`).join('')
        : '<span class="chip">Не указано</span>';

      return `
        <div class="operator-card clickable" data-operator-id="${operator.id}">
          <div class="operator-card__header">
            <strong>${operator.full_name}</strong>
            <span class="muted">${operator.phone_number}</span>
          </div>
          <div class="operator-card__meta">
            <div>
              <strong>Компетенции:</strong>
              <div class="chip-group">${competenciesHtml}</div>
            </div>
            <div>
              <strong>Отвечает за:</strong>
              <div class="chip-group">${responsibilitiesHtml}</div>
            </div>
            <div><strong>Ближайшая задача:</strong> ${nextTask ? nextTask.name : 'Нет'}</div>
            <div><strong>Плановый старт:</strong> ${nextTask ? formatTime(nextTask.plan_start) : '—'}</div>
          </div>
        </div>
      `;
    })
    .join('');
}

function buildOperatorsGantt(container, tasksByOperator) {
  if (!container) return;
  if (!tasksByOperator.length) {
    container.innerHTML = '<div class="gantt__empty">Нет данных для отображения</div>';
    return;
  }

  container.style.setProperty('--gantt-hours', ganttHours.length);
  const header = document.createElement('div');
  header.className = 'gantt__header';
  header.innerHTML = `<div>Оператор</div>${ganttHours
    .map((hour) => `<div>${hour.toString().padStart(2, '0')}:00</div>`)
    .join('')}`;

  container.innerHTML = '';
  container.appendChild(header);

  const dayStart = dayStartHour * 60;
  const dayEnd = dayEndHour * 60;
  const totalMinutes = dayEnd - dayStart;

  tasksByOperator.forEach((row) => {
    const rowEl = document.createElement('div');
    rowEl.className = 'gantt__row';

    const label = document.createElement('div');
    label.className = 'gantt__label';
    label.innerHTML = `
      <span class="gantt__label-title">${row.operator.full_name}</span>
      <small>${row.tasks.length + row.userTasks.length} задач(и)</small>
    `;

    const track = document.createElement('div');
    track.className = 'gantt__track gantt__track--stacked';

    const items = [
      ...row.tasks.map((task) => ({
        ...task,
        _status: getTaskStatus(task)
      })),
      ...row.userTasks.map((task) => ({
        ...task,
        _start: task.start_time,
        _end: task.end_time,
        _status: getUserTaskStatus(task),
        _userTaskId: task.id
      }))
    ];

    items.forEach((task, index) => {
      const { startValue, endValue } = getTaskTimeRange(task);
      if (!startValue || !endValue) return;
      const startMinutes = Math.max(0, getMinutesFromDayStart(startValue));
      const endMinutes = Math.min(totalMinutes, getMinutesFromDayStart(endValue));
      const width = Math.max(4, ((endMinutes - startMinutes) / totalMinutes) * 100);
      const left = (startMinutes / totalMinutes) * 100;
      const bar = document.createElement('div');
      const status = task._status || getTaskStatus(task);
      bar.className = `gantt__bar gantt__bar--stacked ${status}`;
      bar.style.left = `${left}%`;
      bar.style.width = `${width}%`;
      bar.style.top = `${12 + index * 32}px`;
      bar.textContent = `${formatTime(startValue)} – ${formatTime(endValue)}`;
      bar.dataset.status = status;
      bar.title = `${formatTime(startValue)} – ${formatTime(endValue)} · ${getStatusLabel(
        status
      )}`;
      if (task.id) {
        bar.dataset.taskId = task.id;
        bar.classList.add('is-draggable');
      }
      if (task._userTaskId) {
        bar.dataset.userTaskId = task._userTaskId;
        bar.classList.add('is-draggable');
      }
      track.appendChild(bar);
    });

    const minHeight = Math.max(50, items.length * 34 + 12);
    track.style.minHeight = `${minHeight}px`;

    rowEl.appendChild(label);
    rowEl.appendChild(track);
    container.appendChild(rowEl);
  });
}

function renderOperatorsGantt() {
  if (!operatorsGantt) return;
  if (!state.operators.length) {
    operatorsGantt.innerHTML = '<div class="gantt__empty">Добавьте операторов для отображения</div>';
    return;
  }

  const tasksByOperator = state.operators.map((operator) => {
    const tasks = state.tasks.filter((task) => task.operator_id === operator.id);
    const userTasks = state.userTasks.filter((task) => task.operator_id === operator.id);
    return {
      operator,
      tasks,
      userTasks
    };
  });

  buildOperatorsGantt(operatorsGantt, tasksByOperator);
}

function populateSelect(select, items, formatter, placeholder = 'Выберите...') {
  select.innerHTML = '';
  const emptyOption = document.createElement('option');
  emptyOption.value = '';
  emptyOption.textContent = placeholder;
  select.appendChild(emptyOption);
  items.forEach((item) => {
    const option = document.createElement('option');
    option.value = item.id;
    option.textContent = formatter(item);
    select.appendChild(option);
  });
}

function renderSelects() {
  populateSelect(taskOperatorSelect, state.operators, (o) => `${o.full_name} (#${o.id})`);
  populateSelect(taskTypeSelect, state.taskTypes, (t) => `${t.name} (#${t.id})`);
  populateSelect(taskPrioritySelect, state.priorities, (p) => `${p.name} (#${p.id})`);
  populateSelect(taskDeviceSelect, state.devices, (d) => `${d.name} (#${d.id})`);
  populateSelect(deviceTypeSelect, state.deviceTypes, (t) => `${t.name} (#${t.id})`);
  populateSelect(deviceStateSelect, state.deviceStates, (s) => `${s.name} (#${s.id})`);
  populateSelect(scheduleOperatorSelect, state.operators, (o) => `${o.full_name} (#${o.id})`);
  populateSelect(
    equipmentCharacteristicSelect,
    state.equipmentCharacteristics,
    (c) => `${c.name} (#${c.id})`,
    'Выберите характеристику...'
  );
}

function renderTasksPage() {
  const selectedDate = parseDateInput(tasksDateInput.value) || new Date();
  const sortMode = tasksSortSelect?.value || 'task';
  const devicesById = mapById(state.devices);
  const tasksForDate = getTasksForDate(selectedDate).sort((a, b) => {
    if (sortMode === 'device') {
      const aDevice = devicesById[a.device_id]?.name || '';
      const bDevice = devicesById[b.device_id]?.name || '';
      if (aDevice !== bDevice) {
        return aDevice.localeCompare(bDevice, 'ru');
      }
    } else {
      const aName = a.name || '';
      const bName = b.name || '';
      if (aName !== bName) {
        return aName.localeCompare(bName, 'ru');
      }
    }
    const aDate = a.plan_start || a.deadline || 0;
    const bDate = b.plan_start || b.deadline || 0;
    return new Date(aDate) - new Date(bDate);
  });
  buildGantt(tasksGantt, tasksForDate, (task) => {
    const operator = state.operators.find((item) => item.id === task.operator_id);
    const device = state.devices.find((item) => item.id === task.device_id);
    const taskType = state.taskTypes.find((item) => item.id === task.device_task_type_id);
    const mainLabel =
      sortMode === 'device'
        ? `<span class="gantt__label-text">${task.name}</span>`
        : `<span class="gantt__label-title">${task.name}</span>`;
    const deviceLabel =
      sortMode === 'device'
        ? `<span class="gantt__label-title">${device ? device.name : 'Оборудование не выбрано'}</span>`
        : `${device ? device.name : 'Оборудование не выбрано'}`;
    return `
      ${mainLabel}
      <small>${operator ? operator.full_name : 'Оператор не назначен'} · ${
      deviceLabel
    } · ${taskType ? taskType.name : 'Тип не указан'}</small>
    `;
  });
}

function renderReferenceTables() {
  const characteristicsById = mapById(state.equipmentCharacteristics);

  if (!state.equipmentCharacteristics.length) {
    equipmentCharacteristicsList.innerHTML =
      '<div class="gantt__empty">Характеристики оборудования не добавлены</div>';
  } else {
    const header = `
      <div class="table__row">
        <strong>Название</strong>
        <strong class="table__cell--center">Workspace</strong>
        <strong>Действия</strong>
      </div>
    `;
    const rows = state.equipmentCharacteristics
      .map(
        (item) => `
        <div class="table__row">
          <div><strong>${item.name}</strong><br /><span class="muted">#${item.id}</span></div>
          <div class="table__cell--center">${item.workspace_id || '—'}</div>
          <div class="table__actions">
            <button class="button button--ghost" data-delete-equipment-characteristic="${
              item.id
            }" type="button">Удалить</button>
          </div>
        </div>
      `
      )
      .join('');
    equipmentCharacteristicsList.innerHTML = header + rows;
  }

  if (!state.deviceTypes.length) {
    deviceTypesList.innerHTML = '<div class="gantt__empty">Оборудование не добавлено</div>';
  } else {
    const header = `
      <div class="table__row">
        <strong>Название</strong>
        <strong>Характеристика</strong>
        <strong>Действия</strong>
      </div>
    `;
    const rows = state.deviceTypes
      .map(
        (item) => `
        <div class="table__row">
          <div><strong>${item.name}</strong><br /><span class="muted">#${item.id}</span></div>
          <div>${characteristicsById[item.equipment_characteristic_id]?.name || '—'}</div>
          <div class="table__actions">
            <button class="button button--ghost" data-delete-device-type="${item.id}" type="button">Удалить</button>
          </div>
        </div>
      `
      )
      .join('');
    deviceTypesList.innerHTML = header + rows;
  }

  if (!state.taskTypes.length) {
    taskTypesList.innerHTML = '<div class="gantt__empty">Типы заданий не добавлены</div>';
  } else {
    const header = `
      <div class="table__row">
        <strong>Название</strong>
        <strong class="table__cell--center">Workspace</strong>
        <strong>Действия</strong>
      </div>
    `;
    const rows = state.taskTypes
      .map(
        (item) => `
        <div class="table__row">
          <div><strong>${item.name}</strong><br /><span class="muted">#${item.id}</span></div>
          <div class="table__cell--center">${item.workspace_id || '—'}</div>
          <div class="table__actions">
            <button class="button button--ghost" data-delete-task-type="${item.id}" type="button">Удалить</button>
          </div>
        </div>
      `
      )
      .join('');
    taskTypesList.innerHTML = header + rows;
  }
}

function renderAll() {
  renderSelects();
  renderHomeStats();
  renderEntitySummary();
  renderUpcomingTasks();
  renderHomeGantt();
  renderEquipment();
  renderOperators();
  renderOperatorsGantt();
  renderTasksPage();
  renderReferenceTables();
}

function getUserTaskStatus(task) {
  const name = (task.name || '').toLowerCase();
  if (name.includes('обед')) return 'lunch';
  if (name.includes('не работает')) return 'off';
  return 'off';
}

function isBreakUserTask(task) {
  const name = (task.name || '').toLowerCase();
  return name.includes('обед') || name.includes('не работает');
}

function findBreakOverlap(task, startTime, endTime) {
  if (!task?.operator_id || !startTime || !endTime) return null;
  const overlaps = state.userTasks.filter((userTask) => {
    if (userTask.operator_id !== task.operator_id) return false;
    if (!isBreakUserTask(userTask)) return false;
    if (!userTask.start_time || !userTask.end_time) return false;
    const breakStart = new Date(userTask.start_time);
    const breakEnd = new Date(userTask.end_time);
    return startTime < breakEnd && endTime > breakStart;
  });
  return overlaps.length ? overlaps : null;
}

function notifyTaskBreakOverlap(task, startTime, endTime) {
  const overlaps = findBreakOverlap(task, startTime, endTime);
  if (!overlaps) return;
  const operator = state.operators.find((item) => item.id === task.operator_id);
  const operatorName = operator?.full_name || `Оператор #${task.operator_id}`;
  showToast(
    `Задача «${task.name}» оператора ${operatorName} пересекается с его перерывом.`,
    'warning'
  );
}

function notifyAllTaskBreakOverlaps() {
  state.tasks.forEach((task) => {
    if (!task.plan_start || !task.plan_end) return;
    notifyTaskBreakOverlap(task, new Date(task.plan_start), new Date(task.plan_end));
  });
}

function resetTaskForm() {
  taskForm.reset();
  taskForm.dataset.mode = 'create';
  delete taskForm.dataset.taskId;
  if (taskModalTitle) taskModalTitle.textContent = 'Новое задание';
}

function resetOperatorForm() {
  operatorForm.reset();
  operatorForm.dataset.mode = 'create';
  delete operatorForm.dataset.operatorId;
  if (operatorModalTitle) operatorModalTitle.textContent = 'Новый оператор';
}

function resetDeviceForm() {
  deviceForm.reset();
  deviceForm.dataset.mode = 'create';
  delete deviceForm.dataset.deviceId;
  if (deviceModalTitle) deviceModalTitle.textContent = 'Новое оборудование';
}

function resetScheduleForm() {
  scheduleForm.reset();
  scheduleForm.dataset.mode = 'create';
  delete scheduleForm.dataset.userTaskId;
  if (scheduleModalTitle) scheduleModalTitle.textContent = 'Новый перерыв';
}

function fillTaskForm(task) {
  taskForm.dataset.mode = 'edit';
  taskForm.dataset.taskId = task.id;
  if (taskModalTitle) taskModalTitle.textContent = 'Редактировать задание';
  taskForm.elements.name.value = task.name || '';
  taskForm.elements.doc_num.value = task.doc_num || '';
  taskForm.elements.photo_url.value = task.photo_url || '';
  taskForm.elements.deadline.value = task.deadline ? toLocalDateTimeValue(new Date(task.deadline)) : '';
  taskForm.elements.operator_id.value = task.operator_id || '';
  taskForm.elements.device_task_type_id.value = task.device_task_type_id || '';
  taskForm.elements.priority_id.value = task.priority_id || '';
  taskForm.elements.device_id.value = task.device_id || '';
  taskForm.elements.duration_min.value = task.duration_min ?? '';
  taskForm.elements.setup_time_min.value = task.setup_time_min ?? '';
  taskForm.elements.unload_time_min.value = task.unload_time_min ?? '';
  taskForm.elements.plan_start.value = task.plan_start
    ? toLocalDateTimeValue(new Date(task.plan_start))
    : '';
  taskForm.elements.plan_end.value = task.plan_end ? toLocalDateTimeValue(new Date(task.plan_end)) : '';
  taskForm.elements.need_operator.checked = Boolean(task.need_operator);
  taskForm.elements.add_in_rec_system.checked = task.add_in_rec_system !== false;
}

function fillOperatorForm(operator) {
  operatorForm.dataset.mode = 'edit';
  operatorForm.dataset.operatorId = operator.id;
  if (operatorModalTitle) operatorModalTitle.textContent = 'Редактировать оператора';
  operatorForm.elements.full_name.value = operator.full_name || '';
  operatorForm.elements.phone_number.value = operator.phone_number || '';
  operatorForm.elements.user_login.value = operator.user_login || '';
}

function fillDeviceForm(device) {
  deviceForm.dataset.mode = 'edit';
  deviceForm.dataset.deviceId = device.id;
  if (deviceModalTitle) deviceModalTitle.textContent = 'Редактировать оборудование';
  deviceForm.elements.name.value = device.name || '';
  deviceForm.elements.photo_url.value = device.photo_url || '';
  deviceForm.elements.device_type_id.value = device.device_type_id || '';
  deviceForm.elements.device_state_id.value = device.device_state_id || '';
  deviceForm.elements.add_in_rec_system.checked = device.add_in_rec_system !== false;
}

function fillScheduleForm(userTask) {
  scheduleForm.dataset.mode = 'edit';
  scheduleForm.dataset.userTaskId = userTask.id;
  if (scheduleModalTitle) scheduleModalTitle.textContent = 'Редактировать перерыв';
  scheduleForm.elements.operator_id.value = userTask.operator_id || '';
  scheduleForm.elements.start_time.value = userTask.start_time
    ? toLocalDateTimeValue(new Date(userTask.start_time))
    : '';
  scheduleForm.elements.end_time.value = userTask.end_time
    ? toLocalDateTimeValue(new Date(userTask.end_time))
    : '';
  scheduleForm.elements.schedule_type.value = getUserTaskStatus(userTask);
}

function openTaskEditor(taskId) {
  const task = state.tasks.find((item) => item.id === Number(taskId));
  if (!task) return;
  fillTaskForm(task);
  openModal(taskModal);
}

function openDeviceEditor(deviceId) {
  const device = state.devices.find((item) => item.id === Number(deviceId));
  if (!device) return;
  fillDeviceForm(device);
  openModal(deviceModal);
}

function openOperatorEditor(operatorId) {
  const operator = state.operators.find((item) => item.id === Number(operatorId));
  if (!operator) return;
  fillOperatorForm(operator);
  openModal(operatorModal);
}

function openScheduleEditor(userTaskId) {
  const userTask = state.userTasks.find((item) => item.id === Number(userTaskId));
  if (!userTask) return;
  fillScheduleForm(userTask);
  openModal(scheduleModal);
}

function handleLegendHover(event, legend) {
  if (!legend) return;
  const bar = event.target.closest('.gantt__bar');
  if (!bar) return;
  const status = bar.dataset.status;
  legend.querySelectorAll('.legend__item').forEach((item) => {
    item.classList.toggle('is-active', item.dataset.status === status);
  });
}

function clearLegendHover(legend) {
  if (!legend) return;
  legend.querySelectorAll('.legend__item').forEach((item) => item.classList.remove('is-active'));
}

function saveAuth(token, user) {
  authToken = token || '';
  if (authToken) {
    localStorage.setItem('authToken', authToken);
  } else {
    localStorage.removeItem('authToken');
  }
  state.currentUser = user;
  updateAuthUI();
}

function updateAuthUI() {
  const user = state.currentUser;
  userGreeting.textContent = user ? `${user.login}` : 'Гость';
  authActionButton.textContent = user ? 'Профиль' : 'Войти';
  logoutButton.style.display = user ? 'inline-flex' : 'none';
  loginForm.parentElement.style.display = user ? 'none' : 'block';
  registerForm.parentElement.style.display = user ? 'none' : 'block';
  adminPanel.style.display = user?.is_admin ? 'block' : 'none';
  if (devTools) {
    devTools.style.display = user?.is_admin ? 'flex' : 'none';
  }
  if (user) {
    profileSummary.innerHTML = `
      <div>
        <strong>${user.login}</strong>
        <div class="muted">${user.email}</div>
      </div>
      <div class="badge ${user.is_admin ? 'badge--admin' : ''}">
        ${user.is_admin ? 'Администратор' : 'Пользователь'}
      </div>
    `;
  } else {
    profileSummary.innerHTML = '<p class="muted">Войдите, чтобы увидеть данные профиля.</p>';
  }
  prefillUserLoginInputs();
}

function prefillUserLoginInputs() {
  if (!state.currentUser) return;
  document.querySelectorAll('input[name="user_login"]').forEach((input) => {
    if (!input.value) {
      input.value = state.currentUser.login;
    }
  });
}

async function loadCurrentUser() {
  if (!authToken) {
    updateAuthUI();
    return;
  }
  try {
    const user = await fetchJSON(`${apiBase}/auth/me`);
    state.currentUser = user;
    updateAuthUI();
  } catch (error) {
    console.warn(error);
    saveAuth('', null);
  }
}

async function handleLogin(event) {
  event.preventDefault();
  if (!loginForm.reportValidity()) return;
  const formData = new FormData(loginForm);
  const payload = Object.fromEntries(formData.entries());
  const res = await fetchJSON(`${apiBase}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  saveAuth(res.token, res.user);
  loginForm.reset();
  await loadWorkspaces();
  await loadWorkspaceData();
  await loadUsers();
}

async function handleRegister(event) {
  event.preventDefault();
  if (!registerForm.reportValidity()) return;
  const formData = new FormData(registerForm);
  const payload = Object.fromEntries(formData.entries());
  const res = await fetchJSON(`${apiBase}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  saveAuth(res.token, res.user);
  registerForm.reset();
  await loadWorkspaces();
  await loadWorkspaceData();
  await loadUsers();
}

async function handleLogout() {
  try {
    await fetchJSON(`${apiBase}/auth/logout`, { method: 'POST' });
  } catch (error) {
    console.warn(error);
  }
  saveAuth('', null);
  await loadWorkspaces();
  await loadWorkspaceData();
}

async function loadUsers() {
  if (!state.currentUser?.is_admin) return;
  let users = [];
  try {
    users = await fetchJSON(`${apiBase}/users/`);
  } catch (error) {
    console.warn(error);
    usersList.innerHTML = '<div class="gantt__empty">Нет доступа к списку пользователей</div>';
    return;
  }
  if (!users.length) {
    usersList.innerHTML = '<div class="gantt__empty">Пользователи не найдены</div>';
    return;
  }
  const header = `
    <div class="table__row">
      <strong>Пользователь</strong>
      <strong>Роль</strong>
      <strong>Действия</strong>
    </div>
  `;
  const rows = users
    .map((user) => {
      const roleLabel = user.is_admin ? 'Администратор' : 'Пользователь';
      return `
        <div class="table__row">
          <div>
            <strong>${user.login}</strong><br />
            <span class="muted">${user.email}</span>
          </div>
          <div>${roleLabel}</div>
          <div class="table__actions">
            <button class="button button--ghost" type="button" data-toggle-admin="${user.login}">
              ${user.is_admin ? 'Снять права' : 'Сделать админом'}
            </button>
            <button class="button button--ghost" type="button" data-delete-user="${user.login}">
              Удалить
            </button>
          </div>
        </div>
      `;
    })
    .join('');
  usersList.innerHTML = header + rows;
  usersList.dataset.users = JSON.stringify(users);
}

async function createWorkspace(event) {
  event.preventDefault();
  if (!workspaceForm.reportValidity()) return;
  const formData = new FormData(workspaceForm);
  const payload = Object.fromEntries(formData.entries());
  await fetchJSON(`${apiBase}/workspaces/`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  closeModal(workspaceModal);
  workspaceForm.reset();
  await loadWorkspaces();
}

async function createTask(event) {
  event.preventDefault();
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return;
  if (!taskForm.reportValidity()) return;
  const formData = new FormData(taskForm);
  const payload = Object.fromEntries(formData.entries());
  payload.duration_min = Number(payload.duration_min || 0);
  payload.setup_time_min = Number(payload.setup_time_min || 0);
  payload.unload_time_min = Number(payload.unload_time_min || 0);
  payload.operator_id = Number(payload.operator_id || 0);
  payload.device_id = Number(payload.device_id || 0);
  payload.priority_id = Number(payload.priority_id || 0);
  payload.device_task_type_id = Number(payload.device_task_type_id || 0);
  payload.need_operator = formData.get('need_operator') === 'on';
  payload.add_in_rec_system = formData.get('add_in_rec_system') === 'on';
  payload.deadline = parseDateTimeInput(payload.deadline);
  payload.plan_start = parseDateTimeInput(payload.plan_start);
  payload.plan_end = parseDateTimeInput(payload.plan_end);
  if (payload.plan_start && payload.plan_end && payload.plan_end < payload.plan_start) {
    alert('Плановое завершение не может быть раньше начала.');
    return;
  }
  const isEdit = taskForm.dataset.mode === 'edit' && taskForm.dataset.taskId;
  const url = isEdit
    ? `${apiBase}/device-tasks/${taskForm.dataset.taskId}?workspace_id=${workspaceId}`
    : `${apiBase}/workspaces/${workspaceId}/device-tasks`;
  const method = isEdit ? 'PUT' : 'POST';
  const result = await fetchJSON(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  const taskId = isEdit ? Number(taskForm.dataset.taskId) : result?.id;
  closeModal(taskModal);
  resetTaskForm();
  await loadWorkspaceData();
  const updatedTask = state.tasks.find((item) => item.id === taskId);
  if (updatedTask?.plan_start && updatedTask?.plan_end) {
    notifyTaskBreakOverlap(updatedTask, new Date(updatedTask.plan_start), new Date(updatedTask.plan_end));
  }
  if (isEdit) {
    showSaveToast('Задание');
  }
}

async function createOperator(event) {
  event.preventDefault();
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return;
  if (!operatorForm.reportValidity()) return;
  const formData = new FormData(operatorForm);
  const payload = Object.fromEntries(formData.entries());
  const isEdit = operatorForm.dataset.mode === 'edit' && operatorForm.dataset.operatorId;
  const url = isEdit
    ? `${apiBase}/operators/${operatorForm.dataset.operatorId}?workspace_id=${workspaceId}`
    : `${apiBase}/workspaces/${workspaceId}/operators`;
  const method = isEdit ? 'PUT' : 'POST';
  await fetchJSON(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  closeModal(operatorModal);
  resetOperatorForm();
  await loadWorkspaceData();
  if (isEdit) {
    showSaveToast('Оператор');
  }
}

async function createDevice(event) {
  event.preventDefault();
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return;
  if (!deviceForm.reportValidity()) return;
  const formData = new FormData(deviceForm);
  const payload = Object.fromEntries(formData.entries());
  payload.device_type_id = Number(payload.device_type_id || 0);
  payload.device_state_id = Number(payload.device_state_id || 0);
  payload.add_in_rec_system = formData.get('add_in_rec_system') === 'on';
  const isEdit = deviceForm.dataset.mode === 'edit' && deviceForm.dataset.deviceId;
  const url = isEdit
    ? `${apiBase}/devices/${deviceForm.dataset.deviceId}?workspace_id=${workspaceId}`
    : `${apiBase}/workspaces/${workspaceId}/devices`;
  const method = isEdit ? 'PUT' : 'POST';
  await fetchJSON(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  closeModal(deviceModal);
  resetDeviceForm();
  await loadWorkspaceData();
  if (isEdit) {
    showSaveToast('Оборудование');
  }
}

async function createScheduleEntry(event) {
  event.preventDefault();
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return;
  if (!scheduleForm.reportValidity()) return;
  const formData = new FormData(scheduleForm);
  const payload = Object.fromEntries(formData.entries());
  payload.operator_id = Number(payload.operator_id || 0);
  if (scheduleForm.dataset.mode === 'edit') {
    const existingTask = state.userTasks.find(
      (item) => item.id === Number(scheduleForm.dataset.userTaskId)
    );
    payload.device_task_id = existingTask?.device_task_id ?? null;
  } else {
    payload.device_task_id = null;
  }
  payload.start_time = parseDateTimeInput(payload.start_time);
  payload.end_time = parseDateTimeInput(payload.end_time);
  payload.name = payload.schedule_type === 'lunch' ? 'Обед' : 'Не работает';
  delete payload.schedule_type;
  if (payload.start_time && payload.end_time && payload.end_time < payload.start_time) {
    alert('Завершение не может быть раньше начала.');
    return;
  }
  const isEdit = scheduleForm.dataset.mode === 'edit' && scheduleForm.dataset.userTaskId;
  const url = isEdit
    ? `${apiBase}/user-tasks/${scheduleForm.dataset.userTaskId}?workspace_id=${workspaceId}`
    : `${apiBase}/workspaces/${workspaceId}/user-tasks`;
  const method = isEdit ? 'PUT' : 'POST';
  await fetchJSON(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  closeModal(scheduleModal);
  resetScheduleForm();
  await loadWorkspaceData();
  if (isEdit) {
    showSaveToast('Перерыв');
  }
}

async function createDeviceType(event) {
  event.preventDefault();
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return;
  if (!deviceTypeForm.reportValidity()) return;
  const formData = new FormData(deviceTypeForm);
  const payload = Object.fromEntries(formData.entries());
  payload.equipment_characteristic_id = Number(payload.equipment_characteristic_id || 0);
  if (!payload.equipment_characteristic_id) {
    alert('Выберите характеристику оборудования.');
    return;
  }
  await fetchJSON(`${apiBase}/workspaces/${workspaceId}/device-types`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  deviceTypeForm.reset();
  await loadWorkspaceData();
}

async function createEquipmentCharacteristic(event) {
  event.preventDefault();
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return;
  if (!equipmentCharacteristicForm.reportValidity()) return;
  const formData = new FormData(equipmentCharacteristicForm);
  const payload = Object.fromEntries(formData.entries());
  await fetchJSON(`${apiBase}/workspaces/${workspaceId}/equipment-characteristics`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  equipmentCharacteristicForm.reset();
  await loadWorkspaceData();
}

async function createTaskType(event) {
  event.preventDefault();
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return;
  if (!taskTypeForm.reportValidity()) return;
  const formData = new FormData(taskTypeForm);
  const payload = Object.fromEntries(formData.entries());
  await fetchJSON(`${apiBase}/workspaces/${workspaceId}/device-task-types`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  taskTypeForm.reset();
  await loadWorkspaceData();
}

async function handleReferenceAction(event) {
  const deleteEquipmentCharacteristicId = event.target.dataset.deleteEquipmentCharacteristic;
  const deleteDeviceTypeId = event.target.dataset.deleteDeviceType;
  const deleteTaskTypeId = event.target.dataset.deleteTaskType;
  if (deleteEquipmentCharacteristicId) {
    await fetchJSON(`${apiBase}/equipment-characteristics/${deleteEquipmentCharacteristicId}`, {
      method: 'DELETE'
    });
    await loadWorkspaceData();
    return;
  }
  if (deleteDeviceTypeId) {
    await fetchJSON(`${apiBase}/device-types/${deleteDeviceTypeId}`, { method: 'DELETE' });
    await loadWorkspaceData();
    return;
  }
  if (deleteTaskTypeId) {
    await fetchJSON(`${apiBase}/device-task-types/${deleteTaskTypeId}`, { method: 'DELETE' });
    await loadWorkspaceData();
  }
}

async function handleAdminAction(event) {
  const toggleLogin = event.target.dataset.toggleAdmin;
  const deleteLogin = event.target.dataset.deleteUser;
  if (!toggleLogin && !deleteLogin) return;
  const users = JSON.parse(usersList.dataset.users || '[]');
  const user = users.find((item) => item.login === toggleLogin || item.login === deleteLogin);
  if (!user) return;
  if (toggleLogin) {
    const payload = {
      login: user.login,
      id: user.id,
      email: user.email,
      is_admin: !user.is_admin
    };
    await fetchJSON(`${apiBase}/users/${user.login}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    await loadUsers();
    return;
  }
  if (deleteLogin) {
    await fetchJSON(`${apiBase}/users/${user.login}`, { method: 'DELETE' });
    await loadUsers();
  }
}

async function recomputePlan() {
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return;
  await fetchJSON(`${apiBase}/plans/recompute`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ workspace_id: workspaceId })
  });
  pendingOverlapCheck = true;
  await loadWorkspaceData();
}

async function seedDatabase() {
  if (!state.currentUser?.is_admin) return;
  const workspaceId = getWorkspaceId();
  const payload = workspaceId ? { workspace_id: workspaceId } : {};
  try {
    await fetchJSON(`${apiBase}/dev/seed`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
  } catch (error) {
    console.warn(error);
    alert('Не удалось добавить тестовые данные.');
    return;
  }
  await loadReferenceData();
  await loadWorkspaces();
  await loadWorkspaceData();
  await loadUsers();
}

async function clearDatabase() {
  if (!state.currentUser?.is_admin) return;
  const confirmed = confirm('Очистить базу данных? Это удалит все тестовые данные из системы.');
  if (!confirmed) return;
  try {
    await fetchJSON(`${apiBase}/dev/clear`, { method: 'POST' });
  } catch (error) {
    console.warn(error);
    alert('Не удалось очистить базу данных.');
    return;
  }
  await loadReferenceData();
  await loadWorkspaces();
  await loadWorkspaceData();
  await loadUsers();
}

navLinks.forEach((link) => {
  link.addEventListener('click', () => {
    setActivePage(link.dataset.page);
    scheduleAutoRefresh();
  });
});

refreshWorkspacesBtn.addEventListener('click', async () => {
  await loadWorkspaces();
  await loadWorkspaceData();
});
refreshDevicesBtn?.addEventListener('click', loadWorkspaceData);
recomputePlanBtn?.addEventListener('click', recomputePlan);
authActionButton.addEventListener('click', () => setActivePage('profile'));
logoutButton.addEventListener('click', handleLogout);
seedDbBtn?.addEventListener('click', seedDatabase);
clearDbBtn?.addEventListener('click', clearDatabase);

openWorkspaceModalBtn.addEventListener('click', () => openModal(workspaceModal));
openTaskModalBtn?.addEventListener('click', () => {
  resetTaskForm();
  openModal(taskModal);
});
openOperatorModalBtn?.addEventListener('click', () => {
  resetOperatorForm();
  openModal(operatorModal);
});
openDeviceModalBtn?.addEventListener('click', () => {
  resetDeviceForm();
  openModal(deviceModal);
});
openScheduleModalBtn?.addEventListener('click', () => {
  resetScheduleForm();
  openModal(scheduleModal);
});

document.querySelectorAll('[data-close]').forEach((button) => {
  button.addEventListener('click', (event) => {
    const modal = event.target.closest('dialog');
    if (modal) closeModal(modal);
  });
});

workspaceForm.addEventListener('submit', createWorkspace);
taskForm.addEventListener('submit', createTask);
operatorForm.addEventListener('submit', createOperator);
deviceForm.addEventListener('submit', createDevice);
scheduleForm?.addEventListener('submit', createScheduleEntry);
equipmentCharacteristicForm.addEventListener('submit', createEquipmentCharacteristic);
deviceTypeForm.addEventListener('submit', createDeviceType);
taskTypeForm.addEventListener('submit', createTaskType);
loginForm.addEventListener('submit', handleLogin);
registerForm.addEventListener('submit', handleRegister);
equipmentCharacteristicsList.addEventListener('click', handleReferenceAction);
deviceTypesList.addEventListener('click', handleReferenceAction);
taskTypesList.addEventListener('click', handleReferenceAction);
usersList.addEventListener('click', handleAdminAction);

workspaceSelect.addEventListener('change', () => {
  setStoredWorkspaceId(workspaceSelect.value);
  loadWorkspaceData();
});
tasksDateInput.addEventListener('change', renderTasksPage);
tasksSortSelect?.addEventListener('change', renderTasksPage);

homeStats?.addEventListener('click', (event) => {
  const card = event.target.closest('[data-stat]');
  if (!card) return;
  setActivePage('tasks');
  tasksDateInput.value = toLocalInputValue(new Date());
  renderTasksPage();
});

entitySummary?.addEventListener('click', (event) => {
  const card = event.target.closest('[data-entity]');
  if (!card) return;
  const label = card.dataset.entity;
  const pageMap = {
    Оборудование: 'equipment',
    'Типы оборудования': 'references',
    Характеристики: 'references',
    Операторы: 'operators',
    Задания: 'tasks',
    'Типы заданий': 'references',
    Компетенции: 'operators',
    Поручения: 'tasks'
  };
  const targetPage = pageMap[label];
  if (targetPage) {
    setActivePage(targetPage);
    scheduleAutoRefresh();
  }
});

upcomingTasks?.addEventListener('click', (event) => {
  const row = event.target.closest('[data-task-id]');
  if (!row) return;
  openTaskEditor(row.dataset.taskId);
});

tasksGantt?.addEventListener('click', (event) => {
  if (suppressGanttClick) return;
  const bar = event.target.closest('.gantt__bar');
  if (!bar?.dataset.taskId) return;
  openTaskEditor(bar.dataset.taskId);
});

homeGantt?.addEventListener('click', (event) => {
  if (suppressGanttClick) return;
  const bar = event.target.closest('.gantt__bar');
  if (!bar?.dataset.taskId) return;
  setActivePage('tasks');
  tasksDateInput.value = toLocalInputValue(new Date());
  renderTasksPage();
  openTaskEditor(bar.dataset.taskId);
});

operatorsGantt?.addEventListener('click', (event) => {
  if (suppressGanttClick) return;
  const bar = event.target.closest('.gantt__bar');
  if (!bar?.dataset.userTaskId) return;
  openScheduleEditor(bar.dataset.userTaskId);
});

equipmentCards?.addEventListener('click', (event) => {
  const card = event.target.closest('[data-device-id]');
  if (!card) return;
  openDeviceEditor(card.dataset.deviceId);
});

equipmentList?.addEventListener('click', (event) => {
  const card = event.target.closest('[data-device-id]');
  if (!card) return;
  openDeviceEditor(card.dataset.deviceId);
});

operatorsList?.addEventListener('click', (event) => {
  const card = event.target.closest('[data-operator-id]');
  if (!card) return;
  openOperatorEditor(card.dataset.operatorId);
});

homeGantt?.addEventListener('mouseover', (event) => handleLegendHover(event, homeGanttLegend));
homeGantt?.addEventListener('mouseout', () => clearLegendHover(homeGanttLegend));
tasksGantt?.addEventListener('mouseover', (event) => handleLegendHover(event, tasksGanttLegend));
tasksGantt?.addEventListener('mouseout', () => clearLegendHover(tasksGanttLegend));
operatorsGantt?.addEventListener('mouseover', (event) =>
  handleLegendHover(event, operatorsGanttLegend)
);
operatorsGantt?.addEventListener('mouseout', () => clearLegendHover(operatorsGanttLegend));
homeGantt?.addEventListener('pointerdown', startGanttDrag);
tasksGantt?.addEventListener('pointerdown', startGanttDrag);
operatorsGantt?.addEventListener('pointerdown', startGanttDrag);
document.addEventListener('pointermove', moveGanttDrag);
document.addEventListener('pointerup', endGanttDrag);

document.addEventListener('click', (event) => {
  if (!event.target.closest('button')) return;
  scheduleAutoRefresh();
});

tasksDateInput.value = toLocalInputValue(new Date());

checkHealth();
applyInputMasks();
loadCurrentUser()
  .then(loadReferenceData)
  .then(loadWorkspaces)
  .then(loadWorkspaceData)
  .then(loadUsers)
  .catch((error) => {
    console.error(error);
  });
