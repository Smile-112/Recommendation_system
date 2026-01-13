package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"recsys-backend/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  storage.User `json:"user"`
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if strings.TrimSpace(req.Login) == "" || strings.TrimSpace(req.Email) == "" || req.Password == "" {
		writeJSON(w, 400, map[string]any{"error": "login, email, password required"})
		return
	}

	userCount, err := h.repos.CountUsers(r.Context())
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	nextID, err := h.repos.NextUserID(r.Context())
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	user := storage.User{
		Login:   req.Login,
		ID:      nextID,
		Email:   req.Email,
		IsAdmin: userCount == 0,
	}
	if _, err := h.repos.CreateUser(r.Context(), user, passwordHash); err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	token, err := h.issueToken(user)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 201, AuthResponse{Token: token, User: user})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "bad json"})
		return
	}
	if strings.TrimSpace(req.Login) == "" || req.Password == "" {
		writeJSON(w, 400, map[string]any{"error": "login and password required"})
		return
	}
	user, passwordHash, err := h.repos.GetUserAuth(r.Context(), req.Login)
	if err != nil {
		writeJSON(w, 401, map[string]any{"error": "invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		writeJSON(w, 401, map[string]any{"error": "invalid credentials"})
		return
	}
	token, err := h.issueToken(user)
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, AuthResponse{Token: token, User: user})
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	token := parseToken(r)
	if token != "" {
		h.sessionsMu.Lock()
		delete(h.sessions, token)
		h.sessionsMu.Unlock()
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(r)
	if !ok {
		writeJSON(w, 401, map[string]any{"error": "unauthorized"})
		return
	}
	writeJSON(w, 200, user)
}

func (h *Handlers) requireAdmin(w http.ResponseWriter, r *http.Request) (storage.User, bool) {
	user, ok := h.currentUser(r)
	if !ok {
		writeJSON(w, 401, map[string]any{"error": "unauthorized"})
		return storage.User{}, false
	}
	if !user.IsAdmin {
		writeJSON(w, 403, map[string]any{"error": "admin required"})
		return storage.User{}, false
	}
	return user, true
}

func (h *Handlers) currentUser(r *http.Request) (storage.User, bool) {
	token := parseToken(r)
	if token == "" {
		return storage.User{}, false
	}
	h.sessionsMu.RLock()
	defer h.sessionsMu.RUnlock()
	user, ok := h.sessions[token]
	return user, ok
}

func parseToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (h *Handlers) issueToken(user storage.User) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	h.sessionsMu.Lock()
	h.sessions[token] = user
	h.sessionsMu.Unlock()
	return token, nil
}
