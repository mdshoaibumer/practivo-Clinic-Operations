package handler

import (
	"log/slog"

	"clinmitra/internal/service"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

// NewDashboardHandler creates a DashboardHandler backed by the given service.
func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetDashboardStats returns aggregated statistics for the dashboard
// (today's appointments, revenue, total patients, outstanding balance).
func (h *DashboardHandler) GetDashboardStats() (*service.DashboardStats, error) {
	result, err := h.dashboardService.GetDashboardStats()
	if err != nil {
		slog.Error("get dashboard stats failed", "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// GetDailyReport returns a daily collection report with payment details.
func (h *DashboardHandler) GetDailyReport(date string) (*service.DailyReport, error) {
	slog.Info("generating daily report", "date", date)
	result, err := h.dashboardService.GetDailyReport(date)
	if err != nil {
		slog.Error("daily report failed", "date", date, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}

// GetMonthlyReport returns a monthly revenue and outstanding summary.
func (h *DashboardHandler) GetMonthlyReport(year, month int) (*service.MonthlyReport, error) {
	slog.Info("generating monthly report", "year", year, "month", month)
	result, err := h.dashboardService.GetMonthlyReport(year, month)
	if err != nil {
		slog.Error("monthly report failed", "year", year, "month", month, "error", err.Error())
		return nil, safeError(err)
	}
	return result, nil
}
