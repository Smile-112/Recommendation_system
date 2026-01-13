package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/brianvoe/gofakeit/v6"

	"recsys-backend/internal/storage"
)

// SeedDevRequest описывает параметры dev-наполнения (легко удалить при необходимости).
type SeedDevRequest struct {
	WorkspaceID int64 `json:"workspace_id"`
}

type SeedDevResponse struct {
	WorkspaceID   int64 `json:"workspace_id"`
	DeviceStateID int64 `json:"device_state_id"`
	PriorityID    int64 `json:"priority_id"`
	DeviceTypeID  int64 `json:"device_type_id"`
	DeviceID      int64 `json:"device_id"`
	OperatorID    int64 `json:"operator_id"`
	DeviceTaskID  int64 `json:"device_task_id"`
	UserTaskID    int64 `json:"user_task_id"`
}

// SeedDevData создает по одному экземпляру данных для UI-объектов (dev-only).
func (h *Handlers) SeedDevData(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	var req SeedDevRequest
	if err := decodeOptionalJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}

	workspaceID, err := h.resolveWorkspaceID(r.Context(), req.WorkspaceID, user.Login)
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}

	faker := gofakeit.New(time.Now().UnixNano())
	now := time.Now()

	deviceStateID, err := h.repos.CreateDeviceState(r.Context(), faker.RandomString([]string{
		"Готов",
		"В ремонте",
		"На обслуживании",
	}))
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	priorityID, err := h.repos.CreatePriority(r.Context(), faker.RandomString([]string{
		"Низкий",
		"Средний",
		"Высокий",
		"Критичный",
	}))
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	characteristicID, err := h.repos.CreateEquipmentCharacteristic(r.Context(), storage.EquipmentCharacteristic{
		Name:        faker.RandomString([]string{"Пластик", "Металл", "Смола", "Композит"}),
		WorkspaceID: workspaceID,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	deviceTypeName := fmt.Sprintf("3D-принтер %s", faker.RandomString([]string{"FDM", "SLA", "SLS"}))
	deviceTypeID, err := h.repos.CreateDeviceType(r.Context(), storage.DeviceType{
		Name:                      deviceTypeName,
		EquipmentCharacteristicID: characteristicID,
		WorkspaceID:               workspaceID,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	addInRec := true
	deviceName := fmt.Sprintf("%s %s", deviceTypeName, faker.RandomString([]string{"MK4", "Pro", "X2"}))
	deviceID, err := h.repos.CreateDevice(r.Context(), storage.Device{
		Name:           deviceName,
		PhotoURL:       "https://placehold.co/400x300?text=3D+Printer",
		AddInRecSystem: &addInRec,
		DeviceTypeID:   deviceTypeID,
		DeviceStateID:  deviceStateID,
		WorkspaceID:    workspaceID,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	operatorID, err := h.repos.CreateOperator(r.Context(), storage.Operator{
		FullName:    faker.Name(),
		PhoneNumber: faker.Numerify("+7 (###) ###-##-##"),
		WorkspaceID: workspaceID,
		UserLogin:   user.Login,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	if _, err := h.repos.CreateOperatorCompetency(r.Context(), storage.OperatorCompetency{
		WorkspaceID:  workspaceID,
		DeviceTypeID: deviceTypeID,
		OperatorID:   operatorID,
	}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	if _, err := h.repos.CreateOperatorDevice(r.Context(), storage.OperatorDevice{
		OperatorID: operatorID,
		DeviceID:   deviceID,
	}); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	taskTypeID, err := h.repos.CreateDeviceTaskType(r.Context(), storage.DeviceTaskType{
		Name:        faker.RandomString([]string{"Прототипирование", "Печать корпуса", "Мелкосерийное производство"}),
		WorkspaceID: workspaceID,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	durationMin := faker.Number(45, 180)
	setupMin := faker.Number(10, 45)
	unloadMin := faker.Number(5, 20)
	planStart := now.Add(time.Duration(faker.Number(1, 6)) * time.Hour)
	planEnd := planStart.Add(time.Duration(durationMin) * time.Minute)
	deadline := planEnd.Add(time.Duration(faker.Number(4, 48)) * time.Hour)
	docNum := fmt.Sprintf("DOC-%04d", faker.Number(1, 9999))

	deviceTaskID, err := h.repos.CreateDeviceTask(r.Context(), storage.DeviceTask{
		Name:             fmt.Sprintf("Изделие: %s", faker.RandomString([]string{"корпус", "шестерня", "держатель", "кожух"})),
		Deadline:         &deadline,
		Duration:         time.Duration(durationMin) * time.Minute,
		SetupTime:        time.Duration(setupMin) * time.Minute,
		UnloadTime:       time.Duration(unloadMin) * time.Minute,
		NeedOperator:     true,
		PhotoURL:         "https://placehold.co/600x400?text=3D+Task",
		PlanStart:        &planStart,
		PlanEnd:          &planEnd,
		DocNum:           docNum,
		CompletionMark:   "false",
		AddInRecSystem:   &addInRec,
		DeviceTaskTypeID: taskTypeID,
		WorkspaceID:      workspaceID,
		OperatorID:       operatorID,
		DeviceID:         deviceID,
		PriorityID:       priorityID,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	userTaskStart := planStart.Add(time.Duration(faker.Number(15, 60)) * time.Minute)
	userTaskEnd := userTaskStart.Add(time.Duration(faker.Number(30, 90)) * time.Minute)
	priorityValue := faker.Number(1, 3)
	completionMark := false

	userTaskID, err := h.repos.CreateUserTask(r.Context(), storage.UserTask{
		Name:           "Сменное задание: контроль качества",
		StartTime:      &userTaskStart,
		EndTime:        &userTaskEnd,
		Priority:       &priorityValue,
		CompletionMark: &completionMark,
		WorkspaceID:    workspaceID,
		DeviceTaskID:   &deviceTaskID,
		OperatorID:     operatorID,
	})
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, 201, SeedDevResponse{
		WorkspaceID:   workspaceID,
		DeviceStateID: deviceStateID,
		PriorityID:    priorityID,
		DeviceTypeID:  deviceTypeID,
		DeviceID:      deviceID,
		OperatorID:    operatorID,
		DeviceTaskID:  deviceTaskID,
		UserTaskID:    userTaskID,
	})
}

// ClearDevData удаляет тестовые данные (dev-only).
func (h *Handlers) ClearDevData(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	if err := h.repos.ClearAllData(r.Context()); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func decodeOptionalJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return nil
	}
	err := json.NewDecoder(r.Body).Decode(v)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func (h *Handlers) resolveWorkspaceID(ctx context.Context, workspaceID int64, userLogin string) (int64, error) {
	if workspaceID > 0 {
		if _, err := h.repos.GetWorkspace(ctx, workspaceID); err != nil {
			return 0, fmt.Errorf("workspace not found")
		}
		return workspaceID, nil
	}

	workspaces, err := h.repos.ListWorkspaces(ctx, &userLogin)
	if err != nil {
		return 0, err
	}
	if len(workspaces) > 0 {
		return workspaces[0].ID, nil
	}

	faker := gofakeit.New(time.Now().UnixNano())
	createdID, err := h.repos.CreateWorkspace(ctx, storage.Workspace{
		Name:      fmt.Sprintf("Тестовый цех %s", faker.RandomString([]string{"Север", "Центр", "Восток"})),
		UserLogin: userLogin,
	})
	if err != nil {
		return 0, err
	}
	return createdID, nil
}
