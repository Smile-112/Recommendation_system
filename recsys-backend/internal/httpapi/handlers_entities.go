package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"recsys-backend/internal/storage"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type NameRequest struct {
	Name string `json:"name"`
}

type UserRequest struct {
	Login    string `json:"login"`
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type WorkspaceRequest struct {
	Name      string `json:"name"`
	UserLogin string `json:"user_login"`
}

type EquipmentCharacteristicRequest struct {
	Name string `json:"name"`
}

type DeviceTypeRequest struct {
	Name                      string `json:"name"`
	EquipmentCharacteristicID int64  `json:"equipment_characteristic_id"`
}

type DeviceRequest struct {
	Name           string `json:"name"`
	PhotoURL       string `json:"photo_url"`
	AddInRecSystem *bool  `json:"add_in_rec_system"`
	DeviceTypeID   int64  `json:"device_type_id"`
	DeviceStateID  int64  `json:"device_state_id"`
}

type OperatorRequest struct {
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	UserLogin   string `json:"user_login"`
}

type OperatorCompetencyRequest struct {
	DeviceTypeID int64 `json:"device_type_id"`
	OperatorID   int64 `json:"operator_id"`
}

type OperatorDeviceRequest struct {
	OperatorID int64 `json:"operator_id"`
	DeviceID   int64 `json:"device_id"`
}

type DeviceTaskTypeRequest struct {
	Name string `json:"name"`
}

type DeviceTaskRequest struct {
	Name             string     `json:"name"`
	Deadline         *time.Time `json:"deadline"`
	DurationMin      int        `json:"duration_min"`
	SetupTimeMin     int        `json:"setup_time_min"`
	UnloadTimeMin    int        `json:"unload_time_min"`
	NeedOperator     bool       `json:"need_operator"`
	PhotoURL         string     `json:"photo_url"`
	PlanStart        *time.Time `json:"plan_start"`
	PlanEnd          *time.Time `json:"plan_end"`
	DocNum           string     `json:"doc_num"`
	CompletionMark   string     `json:"completion_mark"`
	AddInRecSystem   *bool      `json:"add_in_rec_system"`
	DeviceTaskTypeID int64      `json:"device_task_type_id"`
	OperatorID       int64      `json:"operator_id"`
	DeviceID         int64      `json:"device_id"`
	PriorityID       int64      `json:"priority_id"`
}

type UserTaskRequest struct {
	Name           string     `json:"name"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	Priority       *int       `json:"priority"`
	CompletionMark *bool      `json:"completion_mark"`
	DeviceTaskID   int64      `json:"device_task_id"`
	OperatorID     int64      `json:"operator_id"`
}

func parseIDParam(r *http.Request, key string) (int64, error) {
	idStr := chi.URLParam(r, key)
	return strconv.ParseInt(idStr, 10, 64)
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func hashPassword(raw string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func minutesToDuration(minutes int) time.Duration {
	if minutes < 0 {
		return 0
	}
	return time.Duration(minutes) * time.Minute
}

// ListUsers godoc
// @Summary     Список пользователей
// @Tags        users
// @Produce     json
// @Success     200  {array}   storage.User
// @Failure     500  {object}  map[string]any
// @Router      /api/users [get]
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repos.ListUsers(r.Context())
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, users)
}

// GetUser godoc
// @Summary     Получить пользователя по логину
// @Tags        users
// @Produce     json
// @Param       login  path      string  true  "User login"
// @Success     200    {object}  storage.User
// @Failure     400    {object}  map[string]any
// @Failure     500    {object}  map[string]any
// @Router      /api/users/{login} [get]
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	login := chi.URLParam(r, "login")
	if strings.TrimSpace(login) == "" {
		writeJSON(w, 400, map[string]any{"error": "login required"})
		return
	}
	user, err := h.repos.GetUser(r.Context(), login)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, user)
}

// CreateUser godoc
// @Summary     Создать пользователя
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       body  body      UserRequest  true  "User payload"
// @Success     201   {object}  map[string]any
// @Failure     400   {object}  map[string]any
// @Failure     500   {object}  map[string]any
// @Router      /api/users [post]
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if req.Login == "" || req.Password == "" || req.Email == "" {
		writeJSON(w, 400, map[string]any{"error": "login, password, email required"})
		return
	}
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	id, err := h.repos.CreateUser(r.Context(), storage.User{
		Login: req.Login,
		ID:    req.ID,
		Email: req.Email,
	}, passwordHash)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateUser godoc
// @Summary     Обновить пользователя
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       login  path      string       true  "User login"
// @Param       body   body      UserRequest  true  "User payload"
// @Success     200    {object}  map[string]any
// @Failure     400    {object}  map[string]any
// @Failure     500    {object}  map[string]any
// @Router      /api/users/{login} [put]
func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	login := chi.URLParam(r, "login")
	if login == "" {
		writeJSON(w, 400, map[string]any{"error": "login required"})
		return
	}
	var req UserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	var passwordHash *string
	if req.Password != "" {
		hash, err := hashPassword(req.Password)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		passwordHash = &hash
	}
	if err := h.repos.UpdateUser(r.Context(), storage.User{
		Login: login,
		ID:    req.ID,
		Email: req.Email,
	}, passwordHash); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteUser godoc
// @Summary     Удалить пользователя
// @Tags        users
// @Produce     json
// @Param       login  path      string  true  "User login"
// @Success     200    {object}  map[string]any
// @Failure     400    {object}  map[string]any
// @Failure     500    {object}  map[string]any
// @Router      /api/users/{login} [delete]
func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	login := chi.URLParam(r, "login")
	if login == "" {
		writeJSON(w, 400, map[string]any{"error": "login required"})
		return
	}
	if err := h.repos.DeleteUser(r.Context(), login); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListWorkspaces godoc
// @Summary     Список рабочих пространств
// @Tags        workspaces
// @Produce     json
// @Param       user_login  query     string  false  "User login"
// @Success     200         {array}   storage.Workspace
// @Failure     500         {object}  map[string]any
// @Router      /api/workspaces [get]
func (h *Handlers) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	var userLogin *string
	if q := r.URL.Query().Get("user_login"); q != "" {
		userLogin = &q
	}
	items, err := h.repos.ListWorkspaces(r.Context(), userLogin)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// GetWorkspace godoc
// @Summary     Получить рабочее пространство
// @Tags        workspaces
// @Produce     json
// @Param       workspaceId  path      int  true  "Workspace ID"
// @Success     200          {object}  storage.Workspace
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId} [get]
func (h *Handlers) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "workspaceId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	item, err := h.repos.GetWorkspace(r.Context(), id)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, item)
}

// CreateWorkspace godoc
// @Summary     Создать рабочее пространство
// @Tags        workspaces
// @Accept      json
// @Produce     json
// @Param       body  body      WorkspaceRequest  true  "Workspace payload"
// @Success     201   {object}  map[string]any
// @Failure     400   {object}  map[string]any
// @Failure     500   {object}  map[string]any
// @Router      /api/workspaces [post]
func (h *Handlers) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req WorkspaceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if req.Name == "" || req.UserLogin == "" {
		writeJSON(w, 400, map[string]any{"error": "name and user_login required"})
		return
	}
	id, err := h.repos.CreateWorkspace(r.Context(), storage.Workspace{Name: req.Name, UserLogin: req.UserLogin})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateWorkspace godoc
// @Summary     Обновить рабочее пространство
// @Tags        workspaces
// @Accept      json
// @Produce     json
// @Param       workspaceId  path      int               true  "Workspace ID"
// @Param       body         body      WorkspaceRequest  true  "Workspace payload"
// @Success     200          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId} [put]
func (h *Handlers) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "workspaceId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	var req WorkspaceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if err := h.repos.UpdateWorkspace(r.Context(), storage.Workspace{ID: id, Name: req.Name, UserLogin: req.UserLogin}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteWorkspace godoc
// @Summary     Удалить рабочее пространство
// @Tags        workspaces
// @Produce     json
// @Param       workspaceId  path      int  true  "Workspace ID"
// @Success     200          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId} [delete]
func (h *Handlers) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "workspaceId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	if err := h.repos.DeleteWorkspace(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListDeviceStates godoc
// @Summary     Список состояний оборудования
// @Tags        device_states
// @Produce     json
// @Success     200  {array}   storage.DeviceState
// @Failure     500  {object}  map[string]any
// @Router      /api/device-states [get]
func (h *Handlers) ListDeviceStates(w http.ResponseWriter, r *http.Request) {
	items, err := h.repos.ListDeviceStates(r.Context())
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// CreateDeviceState godoc
// @Summary     Создать состояние оборудования
// @Tags        device_states
// @Accept      json
// @Produce     json
// @Param       body  body      NameRequest  true  "Device state payload"
// @Success     201   {object}  map[string]any
// @Failure     400   {object}  map[string]any
// @Failure     500   {object}  map[string]any
// @Router      /api/device-states [post]
func (h *Handlers) CreateDeviceState(w http.ResponseWriter, r *http.Request) {
	var req NameRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if req.Name == "" {
		writeJSON(w, 400, map[string]any{"error": "name required"})
		return
	}
	id, err := h.repos.CreateDeviceState(r.Context(), req.Name)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateDeviceState godoc
// @Summary     Обновить состояние оборудования
// @Tags        device_states
// @Accept      json
// @Produce     json
// @Param       stateId  path      int          true  "State ID"
// @Param       body     body      NameRequest  true  "Device state payload"
// @Success     200      {object}  map[string]any
// @Failure     400      {object}  map[string]any
// @Failure     500      {object}  map[string]any
// @Router      /api/device-states/{stateId} [put]
func (h *Handlers) UpdateDeviceState(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "stateId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid stateId"})
		return
	}
	var req NameRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if err := h.repos.UpdateDeviceState(r.Context(), id, req.Name); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteDeviceState godoc
// @Summary     Удалить состояние оборудования
// @Tags        device_states
// @Produce     json
// @Param       stateId  path      int  true  "State ID"
// @Success     200      {object}  map[string]any
// @Failure     400      {object}  map[string]any
// @Failure     500      {object}  map[string]any
// @Router      /api/device-states/{stateId} [delete]
func (h *Handlers) DeleteDeviceState(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "stateId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid stateId"})
		return
	}
	if err := h.repos.DeleteDeviceState(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListPriorities godoc
// @Summary     Список приоритетов
// @Tags        priorities
// @Produce     json
// @Success     200  {array}   storage.Priority
// @Failure     500  {object}  map[string]any
// @Router      /api/priorities [get]
func (h *Handlers) ListPriorities(w http.ResponseWriter, r *http.Request) {
	items, err := h.repos.ListPriorities(r.Context())
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// CreatePriority godoc
// @Summary     Создать приоритет
// @Tags        priorities
// @Accept      json
// @Produce     json
// @Param       body  body      NameRequest  true  "Priority payload"
// @Success     201   {object}  map[string]any
// @Failure     400   {object}  map[string]any
// @Failure     500   {object}  map[string]any
// @Router      /api/priorities [post]
func (h *Handlers) CreatePriority(w http.ResponseWriter, r *http.Request) {
	var req NameRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if req.Name == "" {
		writeJSON(w, 400, map[string]any{"error": "name required"})
		return
	}
	id, err := h.repos.CreatePriority(r.Context(), req.Name)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdatePriority godoc
// @Summary     Обновить приоритет
// @Tags        priorities
// @Accept      json
// @Produce     json
// @Param       priorityId  path      int          true  "Priority ID"
// @Param       body        body      NameRequest  true  "Priority payload"
// @Success     200         {object}  map[string]any
// @Failure     400         {object}  map[string]any
// @Failure     500         {object}  map[string]any
// @Router      /api/priorities/{priorityId} [put]
func (h *Handlers) UpdatePriority(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "priorityId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid priorityId"})
		return
	}
	var req NameRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if err := h.repos.UpdatePriority(r.Context(), id, req.Name); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeletePriority godoc
// @Summary     Удалить приоритет
// @Tags        priorities
// @Produce     json
// @Param       priorityId  path      int  true  "Priority ID"
// @Success     200         {object}  map[string]any
// @Failure     400         {object}  map[string]any
// @Failure     500         {object}  map[string]any
// @Router      /api/priorities/{priorityId} [delete]
func (h *Handlers) DeletePriority(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "priorityId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid priorityId"})
		return
	}
	if err := h.repos.DeletePriority(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListEquipmentCharacteristics godoc
// @Summary     Список характеристик оборудования
// @Tags        equipment_characteristics
// @Produce     json
// @Param       workspaceId  path      int  true  "Workspace ID"
// @Success     200          {array}   storage.EquipmentCharacteristic
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/equipment-characteristics [get]
func (h *Handlers) ListEquipmentCharacteristics(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	items, err := h.repos.ListEquipmentCharacteristics(r.Context(), workspaceID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// CreateEquipmentCharacteristic godoc
// @Summary     Создать характеристику оборудования
// @Tags        equipment_characteristics
// @Accept      json
// @Produce     json
// @Param       workspaceId  path      int                           true  "Workspace ID"
// @Param       body         body      EquipmentCharacteristicRequest  true  "Characteristic payload"
// @Success     201          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/equipment-characteristics [post]
func (h *Handlers) CreateEquipmentCharacteristic(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	var req EquipmentCharacteristicRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	id, err := h.repos.CreateEquipmentCharacteristic(r.Context(), storage.EquipmentCharacteristic{Name: req.Name, WorkspaceID: workspaceID})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateEquipmentCharacteristic godoc
// @Summary     Обновить характеристику оборудования
// @Tags        equipment_characteristics
// @Accept      json
// @Produce     json
// @Param       characteristicId  path      int                           true  "Characteristic ID"
// @Param       workspace_id      query     int                           true  "Workspace ID"
// @Param       body              body      EquipmentCharacteristicRequest  true  "Characteristic payload"
// @Success     200               {object}  map[string]any
// @Failure     400               {object}  map[string]any
// @Failure     500               {object}  map[string]any
// @Router      /api/equipment-characteristics/{characteristicId} [put]
func (h *Handlers) UpdateEquipmentCharacteristic(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "characteristicId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid characteristicId"})
		return
	}
	var req EquipmentCharacteristicRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		writeJSON(w, 400, map[string]any{"error": "workspace_id required"})
		return
	}
	workspaceID, err := strconv.ParseInt(workspaceIDStr, 10, 64)
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspace_id"})
		return
	}
	if err := h.repos.UpdateEquipmentCharacteristic(r.Context(), storage.EquipmentCharacteristic{ID: id, Name: req.Name, WorkspaceID: workspaceID}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteEquipmentCharacteristic godoc
// @Summary     Удалить характеристику оборудования
// @Tags        equipment_characteristics
// @Produce     json
// @Param       characteristicId  path      int  true  "Characteristic ID"
// @Success     200               {object}  map[string]any
// @Failure     400               {object}  map[string]any
// @Failure     500               {object}  map[string]any
// @Router      /api/equipment-characteristics/{characteristicId} [delete]
func (h *Handlers) DeleteEquipmentCharacteristic(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "characteristicId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid characteristicId"})
		return
	}
	if err := h.repos.DeleteEquipmentCharacteristic(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListDeviceTypes godoc
// @Summary     Список типов оборудования
// @Tags        device_types
// @Produce     json
// @Param       workspaceId  path      int  true  "Workspace ID"
// @Success     200          {array}   storage.DeviceType
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/device-types [get]
func (h *Handlers) ListDeviceTypes(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	items, err := h.repos.ListDeviceTypes(r.Context(), workspaceID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// CreateDeviceType godoc
// @Summary     Создать тип оборудования
// @Tags        device_types
// @Accept      json
// @Produce     json
// @Param       workspaceId  path      int              true  "Workspace ID"
// @Param       body         body      DeviceTypeRequest  true  "Device type payload"
// @Success     201          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/device-types [post]
func (h *Handlers) CreateDeviceType(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	var req DeviceTypeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	id, err := h.repos.CreateDeviceType(r.Context(), storage.DeviceType{Name: req.Name, EquipmentCharacteristicID: req.EquipmentCharacteristicID, WorkspaceID: workspaceID})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateDeviceType godoc
// @Summary     Обновить тип оборудования
// @Tags        device_types
// @Accept      json
// @Produce     json
// @Param       deviceTypeId  path      int               true  "Device type ID"
// @Param       workspace_id  query     int               true  "Workspace ID"
// @Param       body          body      DeviceTypeRequest  true  "Device type payload"
// @Success     200           {object}  map[string]any
// @Failure     400           {object}  map[string]any
// @Failure     500           {object}  map[string]any
// @Router      /api/device-types/{deviceTypeId} [put]
func (h *Handlers) UpdateDeviceType(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "deviceTypeId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid deviceTypeId"})
		return
	}
	var req DeviceTypeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		writeJSON(w, 400, map[string]any{"error": "workspace_id required"})
		return
	}
	workspaceID, err := strconv.ParseInt(workspaceIDStr, 10, 64)
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspace_id"})
		return
	}
	if err := h.repos.UpdateDeviceType(r.Context(), storage.DeviceType{ID: id, Name: req.Name, EquipmentCharacteristicID: req.EquipmentCharacteristicID, WorkspaceID: workspaceID}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteDeviceType godoc
// @Summary     Удалить тип оборудования
// @Tags        device_types
// @Produce     json
// @Param       deviceTypeId  path      int  true  "Device type ID"
// @Success     200           {object}  map[string]any
// @Failure     400           {object}  map[string]any
// @Failure     500           {object}  map[string]any
// @Router      /api/device-types/{deviceTypeId} [delete]
func (h *Handlers) DeleteDeviceType(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "deviceTypeId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid deviceTypeId"})
		return
	}
	if err := h.repos.DeleteDeviceType(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListDevices godoc
// @Summary     Список оборудования
// @Tags        devices
// @Produce     json
// @Param       workspaceId  path      int  true  "Workspace ID"
// @Success     200          {array}   storage.Device
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/devices [get]
func (h *Handlers) ListDevices(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	items, err := h.repos.ListDevices(r.Context(), workspaceID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// CreateDevice godoc
// @Summary     Создать оборудование
// @Tags        devices
// @Accept      json
// @Produce     json
// @Param       workspaceId  path      int            true  "Workspace ID"
// @Param       body         body      DeviceRequest  true  "Device payload"
// @Success     201          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/devices [post]
func (h *Handlers) CreateDevice(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	var req DeviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if req.Name == "" || req.PhotoURL == "" {
		writeJSON(w, 400, map[string]any{"error": "name and photo_url required"})
		return
	}
	id, err := h.repos.CreateDevice(r.Context(), storage.Device{
		Name:           req.Name,
		PhotoURL:       req.PhotoURL,
		AddInRecSystem: req.AddInRecSystem,
		DeviceTypeID:   req.DeviceTypeID,
		DeviceStateID:  req.DeviceStateID,
		WorkspaceID:    workspaceID,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateDevice godoc
// @Summary     Обновить оборудование
// @Tags        devices
// @Accept      json
// @Produce     json
// @Param       deviceId      path      int            true  "Device ID"
// @Param       workspace_id  query     int            true  "Workspace ID"
// @Param       body          body      DeviceRequest  true  "Device payload"
// @Success     200           {object}  map[string]any
// @Failure     400           {object}  map[string]any
// @Failure     500           {object}  map[string]any
// @Router      /api/devices/{deviceId} [put]
func (h *Handlers) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "deviceId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid deviceId"})
		return
	}
	var req DeviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		writeJSON(w, 400, map[string]any{"error": "workspace_id required"})
		return
	}
	workspaceID, err := strconv.ParseInt(workspaceIDStr, 10, 64)
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspace_id"})
		return
	}
	if err := h.repos.UpdateDevice(r.Context(), storage.Device{
		ID:             id,
		Name:           req.Name,
		PhotoURL:       req.PhotoURL,
		AddInRecSystem: req.AddInRecSystem,
		DeviceTypeID:   req.DeviceTypeID,
		DeviceStateID:  req.DeviceStateID,
		WorkspaceID:    workspaceID,
	}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteDevice godoc
// @Summary     Удалить оборудование
// @Tags        devices
// @Produce     json
// @Param       deviceId  path      int  true  "Device ID"
// @Success     200       {object}  map[string]any
// @Failure     400       {object}  map[string]any
// @Failure     500       {object}  map[string]any
// @Router      /api/devices/{deviceId} [delete]
func (h *Handlers) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "deviceId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid deviceId"})
		return
	}
	if err := h.repos.DeleteDevice(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListOperators godoc
// @Summary     Список операторов
// @Tags        operators
// @Produce     json
// @Param       workspaceId  path      int  true  "Workspace ID"
// @Success     200          {array}   storage.Operator
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/operators [get]
func (h *Handlers) ListOperators(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	items, err := h.repos.ListOperators(r.Context(), workspaceID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// CreateOperator godoc
// @Summary     Создать оператора
// @Tags        operators
// @Accept      json
// @Produce     json
// @Param       workspaceId  path      int              true  "Workspace ID"
// @Param       body         body      OperatorRequest  true  "Operator payload"
// @Success     201          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/operators [post]
func (h *Handlers) CreateOperator(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	var req OperatorRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if req.FullName == "" || req.PhoneNumber == "" || req.UserLogin == "" {
		writeJSON(w, 400, map[string]any{"error": "full_name, phone_number, user_login required"})
		return
	}
	id, err := h.repos.CreateOperator(r.Context(), storage.Operator{
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		WorkspaceID: workspaceID,
		UserLogin:   req.UserLogin,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateOperator godoc
// @Summary     Обновить оператора
// @Tags        operators
// @Accept      json
// @Produce     json
// @Param       operatorId    path      int              true  "Operator ID"
// @Param       workspace_id  query     int              true  "Workspace ID"
// @Param       body          body      OperatorRequest  true  "Operator payload"
// @Success     200           {object}  map[string]any
// @Failure     400           {object}  map[string]any
// @Failure     500           {object}  map[string]any
// @Router      /api/operators/{operatorId} [put]
func (h *Handlers) UpdateOperator(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "operatorId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid operatorId"})
		return
	}
	var req OperatorRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		writeJSON(w, 400, map[string]any{"error": "workspace_id required"})
		return
	}
	workspaceID, err := strconv.ParseInt(workspaceIDStr, 10, 64)
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspace_id"})
		return
	}
	if err := h.repos.UpdateOperator(r.Context(), storage.Operator{
		ID:          id,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		WorkspaceID: workspaceID,
		UserLogin:   req.UserLogin,
	}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteOperator godoc
// @Summary     Удалить оператора
// @Tags        operators
// @Produce     json
// @Param       operatorId  path      int  true  "Operator ID"
// @Success     200         {object}  map[string]any
// @Failure     400         {object}  map[string]any
// @Failure     500         {object}  map[string]any
// @Router      /api/operators/{operatorId} [delete]
func (h *Handlers) DeleteOperator(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "operatorId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid operatorId"})
		return
	}
	if err := h.repos.DeleteOperator(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListOperatorCompetencies godoc
// @Summary     Список компетенций операторов
// @Tags        operator_competencies
// @Produce     json
// @Param       workspaceId  path      int  true  "Workspace ID"
// @Success     200          {array}   storage.OperatorCompetency
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/operator-competencies [get]
func (h *Handlers) ListOperatorCompetencies(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	items, err := h.repos.ListOperatorCompetencies(r.Context(), workspaceID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// CreateOperatorCompetency godoc
// @Summary     Создать компетенцию оператора
// @Tags        operator_competencies
// @Accept      json
// @Produce     json
// @Param       workspaceId  path      int                         true  "Workspace ID"
// @Param       body         body      OperatorCompetencyRequest  true  "Competency payload"
// @Success     201          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/operator-competencies [post]
func (h *Handlers) CreateOperatorCompetency(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	var req OperatorCompetencyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	id, err := h.repos.CreateOperatorCompetency(r.Context(), storage.OperatorCompetency{
		WorkspaceID:  workspaceID,
		DeviceTypeID: req.DeviceTypeID,
		OperatorID:   req.OperatorID,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateOperatorCompetency godoc
// @Summary     Обновить компетенцию оператора
// @Tags        operator_competencies
// @Accept      json
// @Produce     json
// @Param       competencyId  path      int                         true  "Competency ID"
// @Param       workspace_id  query     int                         true  "Workspace ID"
// @Param       body          body      OperatorCompetencyRequest  true  "Competency payload"
// @Success     200           {object}  map[string]any
// @Failure     400           {object}  map[string]any
// @Failure     500           {object}  map[string]any
// @Router      /api/operator-competencies/{competencyId} [put]
func (h *Handlers) UpdateOperatorCompetency(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "competencyId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid competencyId"})
		return
	}
	var req OperatorCompetencyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		writeJSON(w, 400, map[string]any{"error": "workspace_id required"})
		return
	}
	workspaceID, err := strconv.ParseInt(workspaceIDStr, 10, 64)
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspace_id"})
		return
	}
	if err := h.repos.UpdateOperatorCompetency(r.Context(), storage.OperatorCompetency{
		ID:           id,
		WorkspaceID:  workspaceID,
		DeviceTypeID: req.DeviceTypeID,
		OperatorID:   req.OperatorID,
	}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteOperatorCompetency godoc
// @Summary     Удалить компетенцию оператора
// @Tags        operator_competencies
// @Produce     json
// @Param       competencyId  path      int  true  "Competency ID"
// @Success     200           {object}  map[string]any
// @Failure     400           {object}  map[string]any
// @Failure     500           {object}  map[string]any
// @Router      /api/operator-competencies/{competencyId} [delete]
func (h *Handlers) DeleteOperatorCompetency(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "competencyId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid competencyId"})
		return
	}
	if err := h.repos.DeleteOperatorCompetency(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListOperatorDevices godoc
// @Summary     Список связок оператор-оборудование
// @Tags        operator_devices
// @Produce     json
// @Param       workspaceId  path      int  true  "Workspace ID"
// @Success     200          {array}   storage.OperatorDevice
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/operator-devices [get]
func (h *Handlers) ListOperatorDevices(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	items, err := h.repos.ListOperatorDevices(r.Context(), workspaceID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// CreateOperatorDevice godoc
// @Summary     Создать связку оператор-оборудование
// @Tags        operator_devices
// @Accept      json
// @Produce     json
// @Param       workspaceId  path      int                      true  "Workspace ID"
// @Param       body         body      OperatorDeviceRequest  true  "Operator device payload"
// @Success     201          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/operator-devices [post]
func (h *Handlers) CreateOperatorDevice(w http.ResponseWriter, r *http.Request) {
	var req OperatorDeviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	id, err := h.repos.CreateOperatorDevice(r.Context(), storage.OperatorDevice{OperatorID: req.OperatorID, DeviceID: req.DeviceID})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateOperatorDevice godoc
// @Summary     Обновить связку оператор-оборудование
// @Tags        operator_devices
// @Accept      json
// @Produce     json
// @Param       operatorDeviceId  path      int                      true  "Operator device ID"
// @Param       body              body      OperatorDeviceRequest  true  "Operator device payload"
// @Success     200               {object}  map[string]any
// @Failure     400               {object}  map[string]any
// @Failure     500               {object}  map[string]any
// @Router      /api/operator-devices/{operatorDeviceId} [put]
func (h *Handlers) UpdateOperatorDevice(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "operatorDeviceId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid operatorDeviceId"})
		return
	}
	var req OperatorDeviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if err := h.repos.UpdateOperatorDevice(r.Context(), storage.OperatorDevice{ID: id, OperatorID: req.OperatorID, DeviceID: req.DeviceID}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteOperatorDevice godoc
// @Summary     Удалить связку оператор-оборудование
// @Tags        operator_devices
// @Produce     json
// @Param       operatorDeviceId  path      int  true  "Operator device ID"
// @Success     200               {object}  map[string]any
// @Failure     400               {object}  map[string]any
// @Failure     500               {object}  map[string]any
// @Router      /api/operator-devices/{operatorDeviceId} [delete]
func (h *Handlers) DeleteOperatorDevice(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "operatorDeviceId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid operatorDeviceId"})
		return
	}
	if err := h.repos.DeleteOperatorDevice(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListDeviceTaskTypes godoc
// @Summary     Список типов задач оборудования
// @Tags        device_task_types
// @Produce     json
// @Param       workspaceId  path      int  true  "Workspace ID"
// @Success     200          {array}   storage.DeviceTaskType
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/device-task-types [get]
func (h *Handlers) ListDeviceTaskTypes(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	items, err := h.repos.ListDeviceTaskTypes(r.Context(), workspaceID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// GetDeviceTaskType godoc
// @Summary     Получить тип задачи оборудования
// @Tags        device_task_types
// @Produce     json
// @Param       deviceTaskTypeId  path      int  true  "Device task type ID"
// @Success     200               {object}  storage.DeviceTaskType
// @Failure     400               {object}  map[string]any
// @Failure     500               {object}  map[string]any
// @Router      /api/device-task-types/{deviceTaskTypeId} [get]
func (h *Handlers) GetDeviceTaskType(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "deviceTaskTypeId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid deviceTaskTypeId"})
		return
	}
	item, err := h.repos.GetDeviceTaskType(r.Context(), id)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, item)
}

// CreateDeviceTaskType godoc
// @Summary     Создать тип задачи оборудования
// @Tags        device_task_types
// @Accept      json
// @Produce     json
// @Param       workspaceId  path      int                   true  "Workspace ID"
// @Param       body         body      DeviceTaskTypeRequest  true  "Device task type payload"
// @Success     201          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/device-task-types [post]
func (h *Handlers) CreateDeviceTaskType(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	var req DeviceTaskTypeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	id, err := h.repos.CreateDeviceTaskType(r.Context(), storage.DeviceTaskType{Name: req.Name, WorkspaceID: workspaceID})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateDeviceTaskType godoc
// @Summary     Обновить тип задачи оборудования
// @Tags        device_task_types
// @Accept      json
// @Produce     json
// @Param       deviceTaskTypeId  path      int                   true  "Device task type ID"
// @Param       workspace_id      query     int                   true  "Workspace ID"
// @Param       body              body      DeviceTaskTypeRequest  true  "Device task type payload"
// @Success     200               {object}  map[string]any
// @Failure     400               {object}  map[string]any
// @Failure     500               {object}  map[string]any
// @Router      /api/device-task-types/{deviceTaskTypeId} [put]
func (h *Handlers) UpdateDeviceTaskType(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "deviceTaskTypeId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid deviceTaskTypeId"})
		return
	}
	var req DeviceTaskTypeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		writeJSON(w, 400, map[string]any{"error": "workspace_id required"})
		return
	}
	workspaceID, err := strconv.ParseInt(workspaceIDStr, 10, 64)
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspace_id"})
		return
	}
	if err := h.repos.UpdateDeviceTaskType(r.Context(), storage.DeviceTaskType{ID: id, Name: req.Name, WorkspaceID: workspaceID}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteDeviceTaskType godoc
// @Summary     Удалить тип задачи оборудования
// @Tags        device_task_types
// @Produce     json
// @Param       deviceTaskTypeId  path      int  true  "Device task type ID"
// @Success     200               {object}  map[string]any
// @Failure     400               {object}  map[string]any
// @Failure     500               {object}  map[string]any
// @Router      /api/device-task-types/{deviceTaskTypeId} [delete]
func (h *Handlers) DeleteDeviceTaskType(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "deviceTaskTypeId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid deviceTaskTypeId"})
		return
	}
	if err := h.repos.DeleteDeviceTaskType(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// CreateDeviceTask godoc
// @Summary     Создать задачу оборудования
// @Tags        device_tasks
// @Accept      json
// @Produce     json
// @Param       workspaceId  path      int               true  "Workspace ID"
// @Param       body         body      DeviceTaskRequest  true  "Device task payload"
// @Success     201          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/device-tasks [post]
func (h *Handlers) CreateDeviceTask(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	var req DeviceTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if req.Name == "" || req.PhotoURL == "" || req.DocNum == "" {
		writeJSON(w, 400, map[string]any{"error": "name, photo_url, doc_num required"})
		return
	}
	completion := req.CompletionMark
	if completion == "" {
		completion = "false"
	}
	id, err := h.repos.CreateDeviceTask(r.Context(), storage.DeviceTask{
		Name:             req.Name,
		Deadline:         req.Deadline,
		Duration:         minutesToDuration(req.DurationMin),
		SetupTime:        minutesToDuration(req.SetupTimeMin),
		UnloadTime:       minutesToDuration(req.UnloadTimeMin),
		NeedOperator:     req.NeedOperator,
		PhotoURL:         req.PhotoURL,
		PlanStart:        req.PlanStart,
		PlanEnd:          req.PlanEnd,
		DocNum:           req.DocNum,
		CompletionMark:   completion,
		AddInRecSystem:   req.AddInRecSystem,
		DeviceTaskTypeID: req.DeviceTaskTypeID,
		WorkspaceID:      workspaceID,
		OperatorID:       req.OperatorID,
		DeviceID:         req.DeviceID,
		PriorityID:       req.PriorityID,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// GetDeviceTask godoc
// @Summary     Получить задачу оборудования
// @Tags        device_tasks
// @Produce     json
// @Param       deviceTaskId  path      int  true  "Device task ID"
// @Success     200           {object}  DeviceTaskDTO
// @Failure     400           {object}  map[string]any
// @Failure     500           {object}  map[string]any
// @Router      /api/device-tasks/{deviceTaskId} [get]
func (h *Handlers) GetDeviceTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "deviceTaskId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid deviceTaskId"})
		return
	}
	item, err := h.repos.GetDeviceTask(r.Context(), id)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, DeviceTaskDTO{
		ID:            item.ID,
		Name:          item.Name,
		Deadline:      item.Deadline,
		DurationMin:   int(item.Duration.Minutes()),
		SetupTimeMin:  int(item.SetupTime.Minutes()),
		UnloadTimeMin: int(item.UnloadTime.Minutes()),
		NeedOperator:  item.NeedOperator,
		PlanStart:     item.PlanStart,
		PlanEnd:       item.PlanEnd,
		DocNum:        item.DocNum,
		PriorityID:    item.PriorityID,
		OperatorID:    item.OperatorID,
		DeviceID:      item.DeviceID,
		TaskTypeID:    item.DeviceTaskTypeID,
		WorkspaceID:   item.WorkspaceID,
	})
}

// UpdateDeviceTask godoc
// @Summary     Обновить задачу оборудования
// @Tags        device_tasks
// @Accept      json
// @Produce     json
// @Param       deviceTaskId  path      int               true  "Device task ID"
// @Param       workspace_id  query     int               true  "Workspace ID"
// @Param       body          body      DeviceTaskRequest  true  "Device task payload"
// @Success     200           {object}  map[string]any
// @Failure     400           {object}  map[string]any
// @Failure     500           {object}  map[string]any
// @Router      /api/device-tasks/{deviceTaskId} [put]
func (h *Handlers) UpdateDeviceTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "deviceTaskId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid deviceTaskId"})
		return
	}
	var req DeviceTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		writeJSON(w, 400, map[string]any{"error": "workspace_id required"})
		return
	}
	workspaceID, err := strconv.ParseInt(workspaceIDStr, 10, 64)
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspace_id"})
		return
	}
	completion := req.CompletionMark
	if completion == "" {
		completion = "false"
	}
	if err := h.repos.UpdateDeviceTask(r.Context(), storage.DeviceTask{
		ID:               id,
		Name:             req.Name,
		Deadline:         req.Deadline,
		Duration:         minutesToDuration(req.DurationMin),
		SetupTime:        minutesToDuration(req.SetupTimeMin),
		UnloadTime:       minutesToDuration(req.UnloadTimeMin),
		NeedOperator:     req.NeedOperator,
		PhotoURL:         req.PhotoURL,
		PlanStart:        req.PlanStart,
		PlanEnd:          req.PlanEnd,
		DocNum:           req.DocNum,
		CompletionMark:   completion,
		AddInRecSystem:   req.AddInRecSystem,
		DeviceTaskTypeID: req.DeviceTaskTypeID,
		WorkspaceID:      workspaceID,
		OperatorID:       req.OperatorID,
		DeviceID:         req.DeviceID,
		PriorityID:       req.PriorityID,
	}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteDeviceTask godoc
// @Summary     Удалить задачу оборудования
// @Tags        device_tasks
// @Produce     json
// @Param       deviceTaskId  path      int  true  "Device task ID"
// @Success     200           {object}  map[string]any
// @Failure     400           {object}  map[string]any
// @Failure     500           {object}  map[string]any
// @Router      /api/device-tasks/{deviceTaskId} [delete]
func (h *Handlers) DeleteDeviceTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "deviceTaskId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid deviceTaskId"})
		return
	}
	if err := h.repos.DeleteDeviceTask(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListUserTasks godoc
// @Summary     Список пользовательских задач
// @Tags        user_tasks
// @Produce     json
// @Param       workspaceId  path      int  true  "Workspace ID"
// @Success     200          {array}   storage.UserTask
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/user-tasks [get]
func (h *Handlers) ListUserTasks(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	items, err := h.repos.ListUserTasks(r.Context(), workspaceID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, items)
}

// CreateUserTask godoc
// @Summary     Создать пользовательскую задачу
// @Tags        user_tasks
// @Accept      json
// @Produce     json
// @Param       workspaceId  path      int              true  "Workspace ID"
// @Param       body         body      UserTaskRequest  true  "User task payload"
// @Success     201          {object}  map[string]any
// @Failure     400          {object}  map[string]any
// @Failure     500          {object}  map[string]any
// @Router      /api/workspaces/{workspaceId}/user-tasks [post]
func (h *Handlers) CreateUserTask(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := parseIDParam(r, "workspaceId")
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}
	var req UserTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	id, err := h.repos.CreateUserTask(r.Context(), storage.UserTask{
		Name:           req.Name,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Priority:       req.Priority,
		CompletionMark: req.CompletionMark,
		WorkspaceID:    workspaceID,
		DeviceTaskID:   req.DeviceTaskID,
		OperatorID:     req.OperatorID,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// UpdateUserTask godoc
// @Summary     Обновить пользовательскую задачу
// @Tags        user_tasks
// @Accept      json
// @Produce     json
// @Param       userTaskId    path      int              true  "User task ID"
// @Param       workspace_id  query     int              true  "Workspace ID"
// @Param       body          body      UserTaskRequest  true  "User task payload"
// @Success     200           {object}  map[string]any
// @Failure     400           {object}  map[string]any
// @Failure     500           {object}  map[string]any
// @Router      /api/user-tasks/{userTaskId} [put]
func (h *Handlers) UpdateUserTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "userTaskId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid userTaskId"})
		return
	}
	var req UserTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		writeJSON(w, 400, map[string]any{"error": "workspace_id required"})
		return
	}
	workspaceID, err := strconv.ParseInt(workspaceIDStr, 10, 64)
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspace_id"})
		return
	}
	if err := h.repos.UpdateUserTask(r.Context(), storage.UserTask{
		ID:             id,
		Name:           req.Name,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Priority:       req.Priority,
		CompletionMark: req.CompletionMark,
		WorkspaceID:    workspaceID,
		DeviceTaskID:   req.DeviceTaskID,
		OperatorID:     req.OperatorID,
	}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// DeleteUserTask godoc
// @Summary     Удалить пользовательскую задачу
// @Tags        user_tasks
// @Produce     json
// @Param       userTaskId  path      int  true  "User task ID"
// @Success     200         {object}  map[string]any
// @Failure     400         {object}  map[string]any
// @Failure     500         {object}  map[string]any
// @Router      /api/user-tasks/{userTaskId} [delete]
func (h *Handlers) DeleteUserTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "userTaskId")
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid userTaskId"})
		return
	}
	if err := h.repos.DeleteUserTask(r.Context(), id); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
