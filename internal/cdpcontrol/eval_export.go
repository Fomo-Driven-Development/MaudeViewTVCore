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

var exportModule = null;
try { exportModule = _wpReq(183702); } catch(_) {}
if (!exportModule || typeof exportModule.exportData !== "function")
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"exportData unavailable (module 183702)"});

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
