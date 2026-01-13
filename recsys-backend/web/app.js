const apiBase = '/api';
const statusBadge = document.getElementById('api-status');
const userGreeting = document.getElementById('user-greeting');
const authActionButton = document.getElementById('auth-action');
const workspaceSelect = document.getElementById('workspace-select');
const refreshWorkspacesBtn = document.getElementById('refresh-workspaces');
const refreshDevicesBtn = document.getElementById('refresh-devices');
const recomputePlanBtn = document.getElementById('recompute-plan');
const openWorkspaceModalBtn = document.getElementById('open-workspace-modal');
const openTaskModalBtn = document.getElementById('open-task-modal');
const openOperatorModalBtn = document.getElementById('open-operator-modal');
const openDeviceModalBtn = document.getElementById('open-device-modal');
const tasksDateInput = document.getElementById('tasks-date');
const homeStats = document.getElementById('home-stats');
const upcomingTasks = document.getElementById('upcoming-tasks');
const homeGantt = document.getElementById('home-gantt');
const tasksGantt = document.getElementById('tasks-gantt');
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
const deviceTypeForm = document.getElementById('device-type-form');
const taskTypeForm = document.getElementById('task-type-form');
const deviceTypesList = document.getElementById('device-types-list');
const taskTypesList = document.getElementById('task-types-list');
const equipmentCharacteristicSelect = document.getElementById('equipment-characteristic-select');

const workspaceModal = document.getElementById('workspace-modal');
const taskModal = document.getElementById('task-modal');
const operatorModal = document.getElementById('operator-modal');
const deviceModal = document.getElementById('device-modal');

const workspaceForm = document.getElementById('workspace-form');
const taskForm = document.getElementById('task-form');
const operatorForm = document.getElementById('operator-form');
const deviceForm = document.getElementById('device-form');

const taskOperatorSelect = document.getElementById('task-operator');
const taskTypeSelect = document.getElementById('task-type');
const taskPrioritySelect = document.getElementById('task-priority');
const taskDeviceSelect = document.getElementById('task-device');
const deviceTypeSelect = document.getElementById('device-type');
const deviceStateSelect = document.getElementById('device-state');

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

const ganttHours = Array.from({ length: 14 }, (_, i) => 9 + i);
const storedToken = localStorage.getItem('authToken');
let authToken = storedToken || '';

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
  const digits = value.replace(/\D/g, '').slice(0, 4);
  if (!digits) return '';
  const padded = pad ? digits.padStart(4, '0') : digits;
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

function mapById(items) {
  return items.reduce((acc, item) => {
    acc[item.id] = item;
    return acc;
  }, {});
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

async function loadWorkspaces() {
  const query = state.currentUser ? `?user_login=${encodeURIComponent(state.currentUser.login)}` : '';
  state.workspaces = await fetchJSON(`${apiBase}/workspaces/${query}`);
  workspaceSelect.innerHTML = '';
  if (!state.workspaces.length) {
    const opt = document.createElement('option');
    opt.textContent = 'Нет данных';
    opt.value = '';
    workspaceSelect.appendChild(opt);
    return;
  }
  state.workspaces.forEach((workspace) => {
    const opt = document.createElement('option');
    opt.value = workspace.id;
    opt.textContent = `${workspace.name} (#${workspace.id})`;
    workspaceSelect.appendChild(opt);
  });
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

function buildGantt(container, tasks, labelFormatter) {
  if (!tasks.length) {
    container.innerHTML = '<div class="gantt__empty">Нет данных для отображения</div>';
    return;
  }

  const header = document.createElement('div');
  header.className = 'gantt__header';
  header.innerHTML = `<div>Название</div>${ganttHours
    .map((hour) => `<div>${hour.toString().padStart(2, '0')}:00</div>`)
    .join('')}`;

  container.innerHTML = '';
  container.appendChild(header);

  const dayStart = 9 * 60;
  const dayEnd = 22 * 60;
  const totalMinutes = dayEnd - dayStart;

  tasks.forEach((task) => {
    const row = document.createElement('div');
    row.className = 'gantt__row';

    const label = document.createElement('div');
    label.className = 'gantt__label';
    label.innerHTML = labelFormatter(task);

    const track = document.createElement('div');
    track.className = 'gantt__track';

    const startValue = task.plan_start ? new Date(task.plan_start) : null;
    const endValue = task.plan_end
      ? new Date(task.plan_end)
      : startValue && task.duration_min
      ? new Date(startValue.getTime() + task.duration_min * 60000)
      : null;

    if (startValue && endValue) {
      const startMinutes = Math.max(
        0,
        startValue.getHours() * 60 + startValue.getMinutes() - dayStart
      );
      const endMinutes = Math.min(
        totalMinutes,
        endValue.getHours() * 60 + endValue.getMinutes() - dayStart
      );
      const width = Math.max(4, ((endMinutes - startMinutes) / totalMinutes) * 100);
      const left = (startMinutes / totalMinutes) * 100;
      const bar = document.createElement('div');
      bar.className = `gantt__bar ${getTaskStatus(task)}`;
      bar.style.left = `${left}%`;
      bar.style.width = `${width}%`;
      bar.textContent = `${formatTime(startValue)} – ${formatTime(endValue)}`;
      track.appendChild(bar);
    }

    row.appendChild(label);
    row.appendChild(track);
    container.appendChild(row);
  });
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
      <div class="stat-card">
        <h4>${stat.label}</h4>
        <span>${stat.value}</span>
      </div>
    `
    )
    .join('');

  homeDateLabel.textContent = formatDate(today);
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
      return `
        <div class="table__row">
          <div>
            <strong>${task.name}</strong><br />
            <span class="muted">${operator ? operator.full_name : 'Оператор не назначен'} ·
            ${device ? device.name : 'Оборудование не выбрано'}</span>
            <div class="muted">Тип: ${taskType ? taskType.name : '—'} · Приоритет: ${
        priority ? priority.name : '—'
      }</div>
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
      ${task.name}
      <small>${operator ? operator.full_name : 'Оператор не назначен'} · ${
      taskType ? taskType.name : 'Тип не указан'
    }</small>
    `;
  });
}

function renderEquipment() {
  const deviceTypesById = mapById(state.deviceTypes);
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
    return `
      <div class="device-card">
        <img src="${device.photo_url || 'https://placehold.co/400x240?text=3D+Printer'}" alt="${device.name}" />
        <div>
          <strong>${device.name}</strong>
          <div class="muted">${deviceTypesById[device.device_type_id]?.name || 'Тип не указан'}</div>
        </div>
        <div class="device-card__status">
          <span>${deviceStatesById[device.device_state_id]?.name || 'Состояние неизвестно'}</span>
          <span>${activeTask ? 'В работе' : 'Простой'}</span>
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

      return `
        <div class="operator-card">
          <div class="operator-card__header">
            <strong>${operator.full_name}</strong>
            <span class="muted">${operator.phone_number}</span>
          </div>
          <div class="operator-card__meta">
            <div><strong>Компетенции:</strong> ${competencies.join(', ') || 'Не указаны'}</div>
            <div><strong>Отвечает за:</strong> ${responsibilities.join(', ') || 'Не указано'}</div>
            <div><strong>Ближайшая задача:</strong> ${nextTask ? nextTask.name : 'Нет'}</div>
            <div><strong>Плановый старт:</strong> ${nextTask ? formatTime(nextTask.plan_start) : '—'}</div>
          </div>
        </div>
      `;
    })
    .join('');
}

function renderOperatorsGantt() {
  const tasksByOperator = state.operators.map((operator) => {
    const tasks = state.tasks.filter((task) => task.operator_id === operator.id);
    return {
      operator,
      tasks
    };
  });

  const flattenedTasks = tasksByOperator.flatMap((row) =>
    row.tasks.map((task) => ({
      ...task,
      _label: row.operator.full_name
    }))
  );

  buildGantt(operatorsGantt, flattenedTasks, (task) => {
    return `
      ${task._label}
      <small>${task.name}</small>
    `;
  });
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
  populateSelect(
    equipmentCharacteristicSelect,
    state.equipmentCharacteristics,
    (c) => `${c.name} (#${c.id})`,
    'Выберите характеристику...'
  );
}

function renderTasksPage() {
  const selectedDate = parseDateInput(tasksDateInput.value) || new Date();
  const tasksForDate = getTasksForDate(selectedDate).sort((a, b) => {
    const aDate = a.plan_start || a.deadline || 0;
    const bDate = b.plan_start || b.deadline || 0;
    return new Date(aDate) - new Date(bDate);
  });
  buildGantt(tasksGantt, tasksForDate, (task) => {
    const operator = state.operators.find((item) => item.id === task.operator_id);
    const device = state.devices.find((item) => item.id === task.device_id);
    const taskType = state.taskTypes.find((item) => item.id === task.device_task_type_id);
    return `
      ${task.name}
      <small>${operator ? operator.full_name : 'Оператор не назначен'} · ${
      device ? device.name : 'Оборудование не выбрано'
    } · ${taskType ? taskType.name : 'Тип не указан'}</small>
    `;
  });
}

function renderReferenceTables() {
  const characteristicsById = mapById(state.equipmentCharacteristics);

  if (!state.deviceTypes.length) {
    deviceTypesList.innerHTML = '<div class="gantt__empty">Типы оборудования не добавлены</div>';
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
        <strong>Workspace</strong>
        <strong>Действия</strong>
      </div>
    `;
    const rows = state.taskTypes
      .map(
        (item) => `
        <div class="table__row">
          <div><strong>${item.name}</strong><br /><span class="muted">#${item.id}</span></div>
          <div>${item.workspace_id || '—'}</div>
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
  renderUpcomingTasks();
  renderHomeGantt();
  renderEquipment();
  renderOperators();
  renderOperatorsGantt();
  renderTasksPage();
  renderReferenceTables();
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
  await fetchJSON(`${apiBase}/workspaces/${workspaceId}/device-tasks`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  closeModal(taskModal);
  taskForm.reset();
  await loadWorkspaceData();
}

async function createOperator(event) {
  event.preventDefault();
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return;
  if (!operatorForm.reportValidity()) return;
  const formData = new FormData(operatorForm);
  const payload = Object.fromEntries(formData.entries());
  await fetchJSON(`${apiBase}/workspaces/${workspaceId}/operators`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  closeModal(operatorModal);
  operatorForm.reset();
  await loadWorkspaceData();
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
  await fetchJSON(`${apiBase}/workspaces/${workspaceId}/devices`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  closeModal(deviceModal);
  deviceForm.reset();
  await loadWorkspaceData();
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
  const deleteDeviceTypeId = event.target.dataset.deleteDeviceType;
  const deleteTaskTypeId = event.target.dataset.deleteTaskType;
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
  await loadWorkspaceData();
}

navLinks.forEach((link) => {
  link.addEventListener('click', () => setActivePage(link.dataset.page));
});

refreshWorkspacesBtn.addEventListener('click', loadWorkspaces);
refreshDevicesBtn?.addEventListener('click', loadWorkspaceData);
recomputePlanBtn?.addEventListener('click', recomputePlan);
authActionButton.addEventListener('click', () => setActivePage('profile'));
logoutButton.addEventListener('click', handleLogout);

openWorkspaceModalBtn.addEventListener('click', () => openModal(workspaceModal));
openTaskModalBtn?.addEventListener('click', () => openModal(taskModal));
openOperatorModalBtn?.addEventListener('click', () => openModal(operatorModal));
openDeviceModalBtn?.addEventListener('click', () => openModal(deviceModal));

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
deviceTypeForm.addEventListener('submit', createDeviceType);
taskTypeForm.addEventListener('submit', createTaskType);
loginForm.addEventListener('submit', handleLogin);
registerForm.addEventListener('submit', handleRegister);
deviceTypesList.addEventListener('click', handleReferenceAction);
taskTypesList.addEventListener('click', handleReferenceAction);
usersList.addEventListener('click', handleAdminAction);

workspaceSelect.addEventListener('change', loadWorkspaceData);
tasksDateInput.addEventListener('change', renderTasksPage);

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
