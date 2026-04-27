package httpapi

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	r.Use(panicRecoveryMiddleware)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Get("/health", h.Health)

	r.Route("/api", func(api chi.Router) {
		api.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/logout", h.Logout)
			r.Get("/me", h.Me)
		})

		// Dev-only helpers (легко удалить при необходимости).
		api.Route("/dev", func(r chi.Router) {
			r.Post("/seed", h.SeedDevData)
			r.Post("/clear", h.ClearDevData)
		})

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

		// Catch-all for unknown API endpoints → JSON 404.
		api.NotFound(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "endpoint not found"})
		})
		api.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		})
	})

	r.Handle("/*", staticFileHandler("web"))

	return r
}

// panicRecoveryMiddleware catches panics, logs them, and returns an appropriate
// error response: HTML 500 page for browser requests, JSON for API calls.
func panicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v [%s %s]", rec, r.Method, r.URL.Path)
				if isAPIPath(r) {
					writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal server error"})
					return
				}
				serveErrorPage(w, http.StatusInternalServerError, "web/500.html")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// staticFileHandler serves files from webDir. Unknown paths get a custom 404 page
// instead of Go's plain-text "404 page not found".
func staticFileHandler(webDir string) http.Handler {
	fs := http.FileServer(http.Dir(webDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uPath := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		fsPath := filepath.Join(webDir, filepath.FromSlash(uPath))

		info, err := os.Stat(fsPath)
		if err != nil {
			serveErrorPage(w, http.StatusNotFound, filepath.Join(webDir, "404.html"))
			return
		}
		if info.IsDir() {
			if _, err := os.Stat(filepath.Join(fsPath, "index.html")); err != nil {
				serveErrorPage(w, http.StatusNotFound, filepath.Join(webDir, "404.html"))
				return
			}
		}
		fs.ServeHTTP(w, r)
	})
}

// serveErrorPage writes an HTML error file with the given HTTP status code.
func serveErrorPage(w http.ResponseWriter, code int, htmlFile string) {
	content, err := os.ReadFile(htmlFile)
	if err != nil {
		http.Error(w, http.StatusText(code), code)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(code)
	_, _ = w.Write(content)
}

func isAPIPath(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/")
}
