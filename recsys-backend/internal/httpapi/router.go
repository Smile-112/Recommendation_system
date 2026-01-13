package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Get("/health", h.Health)

	r.Route("/api", func(api chi.Router) {
		api.Route("/users", func(r chi.Router) {
			r.Get("/", h.ListUsers)
			r.Post("/", h.CreateUser)
			r.Get("/{login}", h.GetUser)
			r.Put("/{login}", h.UpdateUser)
			r.Delete("/{login}", h.DeleteUser)
		})

		api.Route("/workspaces", func(r chi.Router) {
			r.Get("/", h.ListWorkspaces)
			r.Post("/", h.CreateWorkspace)
			r.Route("/{workspaceId}", func(ws chi.Router) {
				ws.Get("/", h.GetWorkspace)
				ws.Put("/", h.UpdateWorkspace)
				ws.Delete("/", h.DeleteWorkspace)

				ws.Get("/device-tasks", h.ListDeviceTasks)
				ws.Post("/device-tasks", h.CreateDeviceTask)
				ws.Get("/device-task-types", h.ListDeviceTaskTypes)
				ws.Post("/device-task-types", h.CreateDeviceTaskType)
				ws.Get("/user-tasks", h.ListUserTasks)
				ws.Post("/user-tasks", h.CreateUserTask)

				ws.Get("/operators", h.ListOperators)
				ws.Post("/operators", h.CreateOperator)
				ws.Get("/operator-competencies", h.ListOperatorCompetencies)
				ws.Post("/operator-competencies", h.CreateOperatorCompetency)
				ws.Get("/operator-devices", h.ListOperatorDevices)
				ws.Post("/operator-devices", h.CreateOperatorDevice)

				ws.Get("/devices", h.ListDevices)
				ws.Post("/devices", h.CreateDevice)
				ws.Get("/device-types", h.ListDeviceTypes)
				ws.Post("/device-types", h.CreateDeviceType)
				ws.Get("/equipment-characteristics", h.ListEquipmentCharacteristics)
				ws.Post("/equipment-characteristics", h.CreateEquipmentCharacteristic)
			})
		})

		api.Route("/device-states", func(r chi.Router) {
			r.Get("/", h.ListDeviceStates)
			r.Post("/", h.CreateDeviceState)
			r.Put("/{stateId}", h.UpdateDeviceState)
			r.Delete("/{stateId}", h.DeleteDeviceState)
		})

		api.Route("/priorities", func(r chi.Router) {
			r.Get("/", h.ListPriorities)
			r.Post("/", h.CreatePriority)
			r.Put("/{priorityId}", h.UpdatePriority)
			r.Delete("/{priorityId}", h.DeletePriority)
		})

		api.Route("/device-task-types", func(r chi.Router) {
			r.Get("/{deviceTaskTypeId}", h.GetDeviceTaskType)
			r.Put("/{deviceTaskTypeId}", h.UpdateDeviceTaskType)
			r.Delete("/{deviceTaskTypeId}", h.DeleteDeviceTaskType)
		})

		api.Route("/device-tasks", func(r chi.Router) {
			r.Get("/{deviceTaskId}", h.GetDeviceTask)
			r.Put("/{deviceTaskId}", h.UpdateDeviceTask)
			r.Delete("/{deviceTaskId}", h.DeleteDeviceTask)
		})

		api.Route("/devices", func(r chi.Router) {
			r.Put("/{deviceId}", h.UpdateDevice)
			r.Delete("/{deviceId}", h.DeleteDevice)
		})

		api.Route("/device-types", func(r chi.Router) {
			r.Put("/{deviceTypeId}", h.UpdateDeviceType)
			r.Delete("/{deviceTypeId}", h.DeleteDeviceType)
		})

		api.Route("/equipment-characteristics", func(r chi.Router) {
			r.Put("/{characteristicId}", h.UpdateEquipmentCharacteristic)
			r.Delete("/{characteristicId}", h.DeleteEquipmentCharacteristic)
		})

		api.Route("/operators", func(r chi.Router) {
			r.Put("/{operatorId}", h.UpdateOperator)
			r.Delete("/{operatorId}", h.DeleteOperator)
		})

		api.Route("/operator-competencies", func(r chi.Router) {
			r.Put("/{competencyId}", h.UpdateOperatorCompetency)
			r.Delete("/{competencyId}", h.DeleteOperatorCompetency)
		})

		api.Route("/operator-devices", func(r chi.Router) {
			r.Put("/{operatorDeviceId}", h.UpdateOperatorDevice)
			r.Delete("/{operatorDeviceId}", h.DeleteOperatorDevice)
		})

		api.Route("/user-tasks", func(r chi.Router) {
			r.Put("/{userTaskId}", h.UpdateUserTask)
			r.Delete("/{userTaskId}", h.DeleteUserTask)
		})

		api.Post("/plans/recompute", h.RecomputePlan)
	})

	r.Handle("/*", http.FileServer(http.Dir("web")))

	return r
}
