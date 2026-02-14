//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestDataWindowProbe(t *testing.T) {
	resp := env.POST(t, env.chartPath("data-window/probe"), nil)
	requireStatus(t, resp, http.StatusOK)

	var probe struct {
		PanelVisible     bool           `json:"panel_visible"`
		DOMElements      []string       `json:"dom_elements"`
		CrosshairMethods []string       `json:"crosshair_methods"`
		LegendElements   []string       `json:"legend_elements"`
		ChartWidgetProps []string       `json:"chart_widget_props"`
		ModelProps       []string       `json:"model_props"`
		DataWindowState  map[string]any `json:"data_window_state"`
	}
	probe = decodeJSON[struct {
		PanelVisible     bool           `json:"panel_visible"`
		DOMElements      []string       `json:"dom_elements"`
		CrosshairMethods []string       `json:"crosshair_methods"`
		LegendElements   []string       `json:"legend_elements"`
		ChartWidgetProps []string       `json:"chart_widget_props"`
		ModelProps       []string       `json:"model_props"`
		DataWindowState  map[string]any `json:"data_window_state"`
	}](t, resp)

	t.Logf("data window probe:")
	t.Logf("  panel_visible: %v", probe.PanelVisible)
	t.Logf("  dom_elements: %v", probe.DOMElements)
	t.Logf("  crosshair_methods: %d items", len(probe.CrosshairMethods))
	for i, m := range probe.CrosshairMethods {
		if i < 10 {
			t.Logf("    %s", m)
		}
	}
	t.Logf("  legend_elements: %d items", len(probe.LegendElements))
	for i, el := range probe.LegendElements {
		if i < 5 {
			t.Logf("    %s", el)
		}
	}
	t.Logf("  chart_widget_props: %d items", len(probe.ChartWidgetProps))
	for i, p := range probe.ChartWidgetProps {
		if i < 10 {
			t.Logf("    %s", p)
		}
	}
	t.Logf("  model_props: %d items", len(probe.ModelProps))
	for i, p := range probe.ModelProps {
		if i < 10 {
			t.Logf("    %s", p)
		}
	}
}
