package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dgnsrekt/tv_agent/internal/cdpcontrol"
)

type stubService struct{}

func (s *stubService) ListCharts(ctx context.Context) ([]cdpcontrol.ChartInfo, error) {
	return []cdpcontrol.ChartInfo{}, nil
}
func (s *stubService) GetActiveChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error) {
	return cdpcontrol.ActiveChartInfo{}, nil
}
func (s *stubService) GetSymbolInfo(ctx context.Context, chartID string) (cdpcontrol.SymbolInfo, error) {
	return cdpcontrol.SymbolInfo{}, nil
}
func (s *stubService) GetSymbol(ctx context.Context, chartID string) (string, error) { return "", nil }
func (s *stubService) SetSymbol(ctx context.Context, chartID, symbol string) (string, error) {
	return symbol, nil
}
func (s *stubService) GetResolution(ctx context.Context, chartID string) (string, error) {
	return "", nil
}
func (s *stubService) SetResolution(ctx context.Context, chartID, resolution string) (string, error) {
	return resolution, nil
}
func (s *stubService) ExecuteAction(ctx context.Context, chartID, actionID string) error { return nil }
func (s *stubService) ListStudies(ctx context.Context, chartID string) ([]cdpcontrol.Study, error) {
	return []cdpcontrol.Study{}, nil
}
func (s *stubService) AddStudy(ctx context.Context, chartID, name string, inputs map[string]any, forceOverlay bool) (cdpcontrol.Study, error) {
	return cdpcontrol.Study{}, nil
}
func (s *stubService) RemoveStudy(ctx context.Context, chartID, studyID string) error { return nil }
func (s *stubService) GetStudyInputs(ctx context.Context, chartID, studyID string) (cdpcontrol.StudyDetail, error) {
	return cdpcontrol.StudyDetail{}, nil
}
func (s *stubService) ModifyStudyInputs(ctx context.Context, chartID, studyID string, inputs map[string]any) (cdpcontrol.StudyDetail, error) {
	return cdpcontrol.StudyDetail{}, nil
}
func (s *stubService) ListWatchlists(ctx context.Context) ([]cdpcontrol.WatchlistInfo, error) {
	return []cdpcontrol.WatchlistInfo{}, nil
}
func (s *stubService) GetActiveWatchlist(ctx context.Context) (cdpcontrol.WatchlistDetail, error) {
	return cdpcontrol.WatchlistDetail{}, nil
}
func (s *stubService) SetActiveWatchlist(ctx context.Context, id string) (cdpcontrol.WatchlistInfo, error) {
	return cdpcontrol.WatchlistInfo{}, nil
}
func (s *stubService) GetWatchlist(ctx context.Context, id string) (cdpcontrol.WatchlistDetail, error) {
	return cdpcontrol.WatchlistDetail{}, nil
}
func (s *stubService) CreateWatchlist(ctx context.Context, name string) (cdpcontrol.WatchlistInfo, error) {
	return cdpcontrol.WatchlistInfo{}, nil
}
func (s *stubService) RenameWatchlist(ctx context.Context, id, name string) (cdpcontrol.WatchlistInfo, error) {
	return cdpcontrol.WatchlistInfo{}, nil
}
func (s *stubService) DeleteWatchlist(ctx context.Context, id string) error { return nil }

func TestDocsDarkMode(t *testing.T) {
	h := NewServer(&stubService{})
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `data-theme="dark"`) {
		t.Fatalf("docs missing dark theme marker")
	}
}
