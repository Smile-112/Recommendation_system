package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"recsys-backend/internal/service"
	"recsys-backend/internal/storage"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	repos   *storage.Repos
	planner *service.Planner
}

func NewHandlers(repos *storage.Repos, planner *service.Planner) *Handlers {
	return &Handlers{repos: repos, planner: planner}
}

// DeviceTaskDTO — DTO для Swagger (без time.Duration)
type DeviceTaskDTO struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Deadline      *time.Time `json:"deadline"`
	DurationMin   int        `json:"duration_min"`
	SetupTimeMin  int        `json:"setup_time_min"`
	UnloadTimeMin int        `json:"unload_time_min"`
	NeedOperator  bool       `json:"need_operator"`
	PlanStart     *time.Time `json:"plan_start"`
	PlanEnd       *time.Time `json:"plan_end"`
	DocNum        string     `json:"doc_num"`
	PriorityID    int64      `json:"priority_id"`
	OperatorID    int64      `json:"operator_id"`
	DeviceID      int64      `json:"device_id"`
	TaskTypeID    int64      `json:"device_task_type_id"`
	WorkspaceID   int64      `json:"workspace_id"`
}

// Health godoc
// @Summary      Проверка работоспособности
// @Description  Проверяет доступность API и соединение с БД
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]any
// @Failure      500  {object}  map[string]any
// @Router       /health [get]
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.repos.Ping(r.Context()); err != nil {
		writeJSON(w, 500, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ListDeviceTasks godoc
// @Summary      Список заданий по workspace
// @Tags         device_task
// @Produce      json
// @Param        workspaceId  path      int  true  "Workspace ID"
// @Success      200  {array}   DeviceTaskDTO
// @Failure      400  {object}  map[string]any
// @Failure      500  {object}  map[string]any
// @Router       /api/workspaces/{workspaceId}/device-tasks [get]
func (h *Handlers) ListDeviceTasks(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "workspaceId")
	workspaceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || workspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "invalid workspaceId"})
		return
	}

	tasks, err := h.repos.ListDeviceTasksForWorkspace(r.Context(), workspaceID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	dtos := make([]DeviceTaskDTO, 0, len(tasks))
	for _, t := range tasks {
		dtos = append(dtos, DeviceTaskDTO{
			ID:            t.ID,
			Name:          t.Name,
			Deadline:      t.Deadline,
			DurationMin:   int(t.Duration.Minutes()),
			SetupTimeMin:  int(t.SetupTime.Minutes()),
			UnloadTimeMin: int(t.UnloadTime.Minutes()),
			NeedOperator:  t.NeedOperator,
			PlanStart:     t.PlanStart,
			PlanEnd:       t.PlanEnd,
			DocNum:        t.DocNum,
			PriorityID:    t.PriorityID,
			OperatorID:    t.OperatorID,
			DeviceID:      t.DeviceID,
			TaskTypeID:    t.DeviceTaskTypeID,
			WorkspaceID:   t.WorkspaceID,
		})
	}

	writeJSON(w, 200, dtos)
}

// RecomputePlan godoc
// @Summary      Пересчитать рекомендации/план по workspace
// @Tags         planning
// @Accept       json
// @Produce      json
// @Param        body  body      service.RecomputeRequest  true  "workspace_id"
// @Success      200   {object}  service.RecomputeResult
// @Failure      400   {object}  map[string]any
// @Failure      500   {object}  map[string]any
// @Router       /api/plans/recompute [post]
func (h *Handlers) RecomputePlan(w http.ResponseWriter, r *http.Request) {
	var req service.RecomputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if req.WorkspaceID <= 0 {
		writeJSON(w, 400, map[string]any{"error": "workspace_id must be > 0"})
		return
	}

	res, err := h.planner.Recompute(r.Context(), req.WorkspaceID)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, res)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
