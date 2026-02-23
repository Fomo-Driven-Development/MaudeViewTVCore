package cdpcontrol

func jsExportChartData() string {
	return wrapJSEvalAsync(jsPreamble + `
var cw = chart && chart._chartWidget ? chart._chartWidget : null;
if (!cw) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"_chartWidget unavailable"});

// Ensure webpack require is cached
var _wpReq = window.__tvAgentWpRequire || null;
if (!_wpReq) {
  var _ca = window.webpackChunktradingview;
  if (_ca && Array.isArray(_ca)) {
    try { _ca.push([["__tvExport_" + Date.now()], {}, function(r) { _wpReq = r; }]); } catch(_) {}
    if (_wpReq) window.__tvAgentWpRequire = _wpReq;
  }
}
if (!_wpReq) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"webpack require unavailable"});

// Step 1: Ensure chunk 43619 is loaded. This is a no-op if the chunk is in webpack's
// initial list (already bundled), or triggers a network fetch if it's truly lazy.
// Either way, it guarantees module 853336's factory is registered in _wpReq.m.
try { await _wpReq.e(43619); } catch(_) {}

// Step 2: Initialize module 853336 (body-scroll-lock — a dependency of module 183702).
// Requiring it runs its factory and caches it so module 183702 can find it.
try { _wpReq(853336); } catch(_) {}

// Step 3: Require module 183702, which exports exportData.
var exportModule = null;
try {
  var _m = _wpReq(183702);
  if (_m && typeof _m.exportData === "function") exportModule = _m;
} catch(_) {}

// Fallback: scan executed module cache. Handles cases where module IDs changed between
// TradingView builds but the user already opened the export dialog at least once.
if (!exportModule && _wpReq.c) {
  for (var _k in _wpReq.c) {
    try {
      var _x = _wpReq.c[_k] && _wpReq.c[_k].exports;
      if (_x && typeof _x.exportData === "function") { exportModule = _x; break; }
    } catch(_) {}
  }
}

if (!exportModule || typeof exportModule.exportData !== "function")
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"exportData unavailable"});

// Two levels of model() — confirmed by reading module 276449 source
var innerModel = cw.model().model();
var result = await exportModule.exportData(innerModel);

// Remap schema keys to snake_case
var schema = [];
var timeColIdx = -1;
for (var i = 0; i < result.schema.length; i++) {
  var col = result.schema[i];
  if (col.type === "time" || col.type === "userTime") timeColIdx = i;
  schema.push({
    type: col.type || "",
    source_type: col.sourceType || "",
    source_id: col.sourceId || "",
    source_title: col.sourceTitle || "",
    plot_title: col.plotTitle || "",
    plot_id: col.plotId || ""
  });
}

// Convert object-keyed rows to arrays
var cols = schema.length;
var bars = [];
for (var i = 0; i < result.data.length; i++) {
  var row = result.data[i];
  var arr = new Array(cols);
  for (var j = 0; j < cols; j++) {
    var v = row[j];
    arr[j] = (v === undefined) ? null : v;
  }
  bars.push(arr);
}

// Metadata
var symbol = "", resolution = "";
try {
  var ms = innerModel.mainSeries();
  symbol = ms.actualSymbol ? String(ms.actualSymbol()) : "";
  var props = ms.properties ? ms.properties() : null;
  if (props && props.childs && props.childs().interval)
    resolution = String(props.childs().interval.value());
} catch(_) {}

return JSON.stringify({ok:true,data:{
  symbol: symbol,
  resolution: resolution,
  bar_count: bars.length,
  time_col_idx: timeColIdx,
  schema: schema,
  bars: bars
}});
`)
}
