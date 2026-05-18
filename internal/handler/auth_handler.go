package handler

import (
	"log/slog"

	"clinmitra/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates an AuthHandler backed by the given AuthService.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login authenticates a user with username/password and returns session info.
// Exposed to the Wails frontend via binding.
func (h *AuthHandler) Login(username, password string) (*service.AuthResponse, error) {
	slog.Info("login attempt", "username", username)
	result, err := h.authService.Login(username, password)
	if err != nil {
		slog.Warn("login failed", "username", username, "error", err.Error())
		return nil, safeError(err)
	}
	slog.Info("login successful", "username", username, "userId", result.User.ID, "role", result.User.Role)
	return result, nil
}

// Logout destroys the current user session.
func (h *AuthHandler) Logout() error {
	slog.Info("logout requested")
	err := h.authService.Logout()
	if err != nil {
		slog.Warn("logout failed", "error", err.Error())
		return safeError(err)
	}
	slog.Info("logout successful")
	return nil
}

// GetCurrentUser returns the currently authenticated user info, or
// a response with LoggedIn=false if no session is active.
func (h *AuthHandler) GetCurrentUser() (*service.AuthResponse, error) {
	result, err := h.authService.GetCurrentUser()
	if err != nil {
		slog.Error("get current user failed", "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// ChangePassword updates the password for the currently logged-in user
// after verifying the old password.
func (h *AuthHandler) ChangePassword(oldPassword, newPassword string) error {
	slog.Info("password change requested")
	err := h.authService.ChangePassword(oldPassword, newPassword)
	if err != nil {
		slog.Warn("password change failed", "error", err.Error())
		return safeError(err)
	}
	slog.Info("password changed successfully")
	return nil
}
