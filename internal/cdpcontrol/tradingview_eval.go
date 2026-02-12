package cdpcontrol

import "fmt"

// jsPreamble is the common setup that resolves the TradingView API and active chart.
const jsPreamble = `
var api = window.TradingViewApi;
var chart = api && typeof api.activeChart === "function" ? api.activeChart() : null;`

func jsGetSymbol() string {
	return wrapJSEval(jsPreamble + `
var symbol = "";
if (api && typeof api.getSymbol === "function") symbol = String(api.getSymbol() || "");
if (!symbol && chart && typeof chart.symbol === "function") symbol = String(chart.symbol() || "");
if (!symbol && chart && chart.symbol) symbol = String(chart.symbol || "");
if (!symbol) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"symbol getter unavailable"});
return JSON.stringify({ok:true,data:{symbol:symbol}});
`)
}

func jsSetSymbol(symbol string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var requested = %s;
var changed = false;
if (api && typeof api.setSymbol === "function") {
  api.setSymbol(requested);
  changed = true;
} else if (chart && typeof chart.setSymbol === "function") {
  chart.setSymbol(requested);
  changed = true;
}
if (!changed) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setSymbol unavailable"});
var current = requested;
if (api && typeof api.getSymbol === "function") current = String(api.getSymbol() || requested);
else if (chart && typeof chart.symbol === "function") current = String(chart.symbol() || requested);
return JSON.stringify({ok:true,data:{current_symbol:current}});
`, jsString(symbol)))
}

func jsGetResolution() string {
	return wrapJSEval(jsPreamble + `
var resolution = "";
if (api && typeof api.getResolution === "function") resolution = String(api.getResolution() || "");
if (!resolution && chart && typeof chart.resolution === "function") resolution = String(chart.resolution() || "");
if (!resolution && chart && chart.resolution) resolution = String(chart.resolution || "");
if (!resolution) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"resolution getter unavailable"});
return JSON.stringify({ok:true,data:{resolution:resolution}});
`)
}

func jsSetResolution(resolution string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var requested = %s;
var changed = false;
if (api && typeof api.setResolution === "function") {
  api.setResolution(requested);
  changed = true;
} else if (chart && typeof chart.setResolution === "function") {
  chart.setResolution(requested);
  changed = true;
}
if (!changed) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setResolution unavailable"});
return JSON.stringify({ok:true,data:{}});
`, jsString(resolution)))
}

func jsExecuteAction(actionID string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var action = %s;
if (api && typeof api.executeActionById === "function") {
  api.executeActionById(action);
  return JSON.stringify({ok:true,data:{status:"executed"}});
}
if (chart && typeof chart.executeActionById === "function") {
  chart.executeActionById(action);
  return JSON.stringify({ok:true,data:{status:"executed"}});
}
if (api && typeof api.executeAction === "function") {
  api.executeAction(action);
  return JSON.stringify({ok:true,data:{status:"executed"}});
}
return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"action execution unavailable"});
`, jsString(actionID)))
}

func jsListStudies() string {
	return wrapJSEval(jsPreamble + `
if (!chart || typeof chart.getAllStudies !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getAllStudies unavailable"});
}
var items = chart.getAllStudies() || [];
var studies = [];
for (var i = 0; i < items.length; i++) {
  var it = items[i] || {};
  studies.push({id:String(it.id || it.entityId || ""), name:String(it.name || it.title || "")});
}
return JSON.stringify({ok:true,data:{studies:studies}});
`)
}

func jsAddStudy(name string, inputs map[string]any, forceOverlay bool) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
var name = %s;
var inputs = %s;
var forceOverlay = %t;
if (!chart || typeof chart.createStudy !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"createStudy unavailable"});
}
var id = "";
try {
  id = await chart.createStudy(name, forceOverlay, false, inputs) || "";
} catch (_) {
  id = await chart.createStudy(name) || "";
}
if (!id) id = "study:" + name;
return JSON.stringify({ok:true,data:{study:{id:String(id),name:String(name)}}});
`, jsString(name), jsJSON(inputs), forceOverlay))
}

func jsGetSymbolInfo() string {
	return wrapJSEval(jsPreamble + `
if (!api && !chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"TradingView API unavailable"});
var info = null;
if (chart && typeof chart.symbolExt === "function") {
  try { info = chart.symbolExt(); } catch(_) {}
}
if (!info && api && typeof api.getSymbolInfo === "function") {
  try { info = api.getSymbolInfo(); } catch(_) {}
}
var sym = "";
if (info && info.symbol) sym = String(info.symbol);
if (!sym && chart && typeof chart.symbol === "function") sym = String(chart.symbol() || "");
if (!sym && api && typeof api.getSymbol === "function") sym = String(api.getSymbol() || "");
if (!sym) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"symbol info unavailable"});
var i = info || {};
var result = {symbol:sym};
result.name = String(i.name || i.full_name || "");
result.description = String(i.description || i.short_description || "");
result.exchange = String(i.listed_exchange || i.exchange || "");
result.type = String(i.type || i.security_type || "");
result.currency = String(i.currency_code || i.currency || "");
result.timezone = String(i.timezone || "");
result.pricescale = Number(i.pricescale || i.price_scale || 0);
result.minmov = Number(i.minmov || i.min_mov || 0);
result.has_intraday = !!(i.has_intraday);
result.has_daily = !!(i.has_daily);
return JSON.stringify({ok:true,data:result});
`)
}

func jsGetActiveChart() string {
	return wrapJSEval(jsPreamble + `
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"TradingView API unavailable"});
var chartCount = 1;
var chartIndex = 0;
if (typeof api.chartsCount === "function") chartCount = api.chartsCount() || 1;
if (typeof api.activeChartIndex === "function") chartIndex = api.activeChartIndex() || 0;
return JSON.stringify({ok:true,data:{chart_index:chartIndex,chart_count:chartCount}});
`)
}

func jsGetStudyInputs(studyID string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var id = %s;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
var study = null;
if (typeof chart.getStudyById === "function") study = chart.getStudyById(id);
if (!study) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"study not found: "+id});
var name = String(study.name || study.title || "");
var raw = {};
if (typeof study.getInputValues === "function") {
  raw = study.getInputValues() || {};
} else if (typeof study.inputs === "function") {
  raw = study.inputs() || {};
} else if (study.inputs) {
  raw = study.inputs || {};
}
var inputs = {};
if (Array.isArray(raw)) {
  for (var i = 0; i < raw.length; i++) {
    var item = raw[i] || {};
    var key = String(item.id || item.name || ("input_" + i));
    inputs[key] = item.value !== undefined ? item.value : item;
  }
} else {
  inputs = raw;
}
return JSON.stringify({ok:true,data:{id:String(id),name:name,inputs:inputs}});
`, jsString(studyID)))
}

func jsModifyStudyInputs(studyID string, inputs map[string]any) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var id = %s;
var newInputs = %s;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
var study = null;
if (typeof chart.getStudyById === "function") study = chart.getStudyById(id);
if (!study) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"study not found: "+id});
var merged = false;
if (typeof study.mergeUp === "function") {
  study.mergeUp(newInputs);
  merged = true;
}
if (!merged && typeof study.setInputValues === "function") {
  study.setInputValues(newInputs);
}
if (!merged && typeof study.setInputValues !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"modify study inputs unavailable"});
}
var name = String(study.name || study.title || "");
var raw = {};
if (typeof study.getInputValues === "function") {
  raw = study.getInputValues() || {};
} else if (typeof study.inputs === "function") {
  raw = study.inputs() || {};
} else if (study.inputs) {
  raw = study.inputs || {};
}
var current = {};
if (Array.isArray(raw)) {
  for (var j = 0; j < raw.length; j++) {
    var it = raw[j] || {};
    var k = String(it.id || it.name || ("input_" + j));
    current[k] = it.value !== undefined ? it.value : it;
  }
} else {
  current = raw;
}
return JSON.stringify({ok:true,data:{id:String(id),name:name,inputs:current}});
`, jsString(studyID), jsJSON(inputs)))
}

// --- Watchlist JS functions ---
// These use TradingView's internal REST API (fetch from page context)
// and React fiber props for operations without REST endpoints (flag/mark).

// jsWatchlistFetch is a shared helper that calls TV's internal symbols_list API.
const jsWatchlistFetch = `
async function _wlFetch(path, opts) {
  var resp = await fetch(path, Object.assign({credentials:"include"}, opts || {}));
  if (!resp.ok) {
    var body = "";
    try { var j = await resp.json(); body = j.detail || j.message || JSON.stringify(j); } catch(_) { body = await resp.text(); }
    throw new Error("HTTP " + resp.status + ": " + body);
  }
  return resp.json();
}
`

func jsListWatchlists() string {
	return wrapJSEvalAsync(jsWatchlistFetch + `
var raw = await _wlFetch("/api/v1/symbols_list/all/");
if (!Array.isArray(raw)) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"symbols_list/all returned non-array"});
var lists = [];
for (var i = 0; i < raw.length; i++) {
  var it = raw[i] || {};
  lists.push({
    id: String(it.id || ""),
    name: String(it.name || ""),
    type: String(it.type || ""),
    active: !!(it.active),
    count: Number((it.symbols && it.symbols.length) || 0)
  });
}
return JSON.stringify({ok:true,data:{watchlists:lists}});
`)
}

func jsGetActiveWatchlist() string {
	return wrapJSEvalAsync(jsWatchlistFetch + `
var raw = await _wlFetch("/api/v1/symbols_list/active/");
var syms = [];
if (raw.symbols && Array.isArray(raw.symbols)) {
  for (var i = 0; i < raw.symbols.length; i++) {
    var s = raw.symbols[i];
    syms.push(typeof s === "string" ? s : String(s));
  }
}
return JSON.stringify({ok:true,data:{
  id: String(raw.id || ""),
  name: String(raw.name || ""),
  type: String(raw.type || ""),
  symbols: syms
}});
`)
}

func jsSetActiveWatchlist(id string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+`
var listId = %s;
var raw = await _wlFetch("/api/v1/symbols_list/active/" + encodeURIComponent(listId) + "/", {method:"POST"});
return JSON.stringify({ok:true,data:{
  id: String(raw.id || listId),
  name: String(raw.name || ""),
  type: String(raw.type || ""),
  count: Number((raw.symbols && raw.symbols.length) || 0)
}});
`, jsString(id)))
}

func jsGetWatchlist(id string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+`
var listId = %s;
var all = await _wlFetch("/api/v1/symbols_list/all/");
var raw = null;
for (var i = 0; i < all.length; i++) {
  if (String(all[i].id) === listId) { raw = all[i]; break; }
}
if (!raw) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"watchlist not found: " + listId});
var syms = [];
if (raw.symbols && Array.isArray(raw.symbols)) {
  for (var j = 0; j < raw.symbols.length; j++) {
    var s = raw.symbols[j];
    syms.push(typeof s === "string" ? s : String(s));
  }
}
return JSON.stringify({ok:true,data:{
  id: String(raw.id || listId),
  name: String(raw.name || ""),
  type: String(raw.type || ""),
  symbols: syms
}});
`, jsString(id)))
}

func jsCreateWatchlist(name string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+`
var listName = %s;
var raw = await _wlFetch("/api/v1/symbols_list/custom/", {
  method: "POST",
  headers: {"Content-Type": "application/json"},
  body: JSON.stringify({name: listName, symbols: []})
});
return JSON.stringify({ok:true,data:{
  id: String(raw.id || ""),
  name: String(raw.name || listName),
  type: String(raw.type || "custom"),
  count: Number((raw.symbols && raw.symbols.length) || 0)
}});
`, jsString(name)))
}

func jsRenameWatchlist(id, name string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+`
var listId = %s;
var newName = %s;
var raw = await _wlFetch("/api/v1/symbols_list/custom/" + encodeURIComponent(listId) + "/rename/", {
  method: "POST",
  headers: {"Content-Type": "application/json"},
  body: JSON.stringify({name: newName})
});
return JSON.stringify({ok:true,data:{
  id: String(raw.id || listId),
  name: String(raw.name || newName),
  type: String(raw.type || "custom"),
  count: Number((raw.symbols && raw.symbols.length) || 0)
}});
`, jsString(id), jsString(name)))
}

func jsDeleteWatchlist(id string) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
var listId = %s;
var resp = await fetch("/api/v1/symbols_list/custom/" + encodeURIComponent(listId) + "/", {method: "DELETE", credentials: "include"});
if (!resp.ok && resp.status !== 404) {
  var body = "";
  try { var j = await resp.json(); body = j.detail || j.message || JSON.stringify(j); } catch(_) { body = await resp.text(); }
  throw new Error("HTTP " + resp.status + ": " + body);
}
return JSON.stringify({ok:true,data:{status:"deleted"}});
`, jsString(id)))
}

func jsAddWatchlistSymbols(id string, symbols []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+`
var listId = %s;
var syms = %s;
var updated = await _wlFetch("/api/v1/symbols_list/custom/" + encodeURIComponent(listId) + "/append/", {
  method: "POST",
  headers: {"Content-Type": "application/json"},
  body: JSON.stringify(syms)
});
var result = [];
if (Array.isArray(updated)) {
  for (var i = 0; i < updated.length; i++) {
    result.push(typeof updated[i] === "string" ? updated[i] : String(updated[i]));
  }
}
return JSON.stringify({ok:true,data:{id:listId,name:"",type:"",symbols:result}});
`, jsString(id), jsJSON(symbols)))
}

func jsRemoveWatchlistSymbols(id string, symbols []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+`
var listId = %s;
var syms = %s;
var updated = await _wlFetch("/api/v1/symbols_list/custom/" + encodeURIComponent(listId) + "/remove/", {
  method: "POST",
  headers: {"Content-Type": "application/json"},
  body: JSON.stringify(syms)
});
var result = [];
if (Array.isArray(updated)) {
  for (var i = 0; i < updated.length; i++) {
    result.push(typeof updated[i] === "string" ? updated[i] : String(updated[i]));
  }
}
return JSON.stringify({ok:true,data:{id:listId,name:"",type:"",symbols:result}});
`, jsString(id), jsJSON(symbols)))
}

func jsFlagSymbol(id, symbol string) string {
	// Flag/mark uses React fiber props since there is no REST endpoint.
	return wrapJSEvalAsync(fmt.Sprintf(`
var listId = %s;
var sym = %s;
var el = document.querySelector("[data-name='symbol-list-wrap']");
if (!el) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist widget not found"});
var fiberKey = null;
var keys = Object.keys(el);
for (var i = 0; i < keys.length; i++) {
  if (keys[i].indexOf("__reactFiber") === 0) { fiberKey = keys[i]; break; }
}
if (!fiberKey) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"React fiber not found"});
var fiber = el[fiberKey];
var depth = 0;
while (fiber && depth < 12) {
  if (fiber.memoizedProps && typeof fiber.memoizedProps.markSymbol === "function") {
    await fiber.memoizedProps.markSymbol(sym);
    return JSON.stringify({ok:true,data:{status:"toggled"}});
  }
  fiber = fiber["return"];
  depth++;
}
return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"markSymbol unavailable"});
`, jsString(id), jsString(symbol)))
}

// --- Navigation JS functions ---

// jsExecAction is a JS helper that tries all known executeActionById paths.
const jsExecAction = `
function _execAction(id) {
  if (api && typeof api.executeActionById === "function") { api.executeActionById(id); return true; }
  if (chart && typeof chart.executeActionById === "function") { chart.executeActionById(id); return true; }
  if (api && typeof api.executeAction === "function") { api.executeAction(id); return true; }
  return false;
}
`

func jsZoom(direction string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+jsExecAction+`
var dir = %s;
if (!chart && !api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
var actionId = dir === "in" ? "chartZoomIn" : "chartZoomOut";
if (!_execAction(actionId)) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"zoom unavailable"});
}
return JSON.stringify({ok:true,data:{status:"executed",direction:dir}});
`, jsString(direction)))
}

func jsScroll(bars int) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+jsExecAction+`
var bars = %d;
if (!chart && !api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.scrollChartByBar === "function") {
  chart.scrollChartByBar(bars);
} else {
  var id = bars > 0 ? "chartScrollRight" : "chartScrollLeft";
  var n = Math.abs(bars);
  for (var i = 0; i < n; i++) { if (!_execAction(id)) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"scroll unavailable"}); }
}
return JSON.stringify({ok:true,data:{status:"executed",bars:bars}});
`, bars))
}

func jsScrollToRealtime() string {
	return wrapJSEval(jsPreamble + jsExecAction + `
if (!chart && !api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
var done = false;
if (typeof chart.scrollToRealtime === "function") { try { chart.scrollToRealtime(); done = true; } catch(_) {} }
if (!done) { done = _execAction("chartScrollToLast"); }
if (!done) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"scrollToRealtime unavailable"});
return JSON.stringify({ok:true,data:{status:"executed"}});
`)
}

func jsGoToDate(timestamp int64) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
var ts = %d;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
var done = false;
if (typeof chart.goToDate === "function") { try { chart.goToDate(ts); done = true; } catch(_) {} }
if (!done && typeof chart.setVisibleRange === "function") {
  try { await chart.setVisibleRange({from:ts, to:ts + 86400}); done = true; } catch(_) {}
}
if (!done) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"goToDate unavailable"});
return JSON.stringify({ok:true,data:{status:"executed",timestamp:ts}});
`, timestamp))
}

func jsGetVisibleRange() string {
	return wrapJSEval(jsPreamble + `
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.getVisibleRange !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getVisibleRange unavailable"});
}
var r = chart.getVisibleRange();
if (!r) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"getVisibleRange returned null"});
return JSON.stringify({ok:true,data:{from:Number(r.from || 0),to:Number(r.to || 0)}});
`)
}

func jsSetVisibleRange(from, to float64) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
var from = %v;
var to = %v;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.setVisibleRange !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setVisibleRange unavailable"});
}
try {
  await chart.setVisibleRange({from:from, to:to});
} catch(e) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setVisibleRange: " + String(e && e.message || e)});
}
return JSON.stringify({ok:true,data:{from:from,to:to}});
`, from, to))
}

func jsResetScales() string {
	return wrapJSEval(jsPreamble + jsExecAction + `
if (!chart && !api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
var reset = false;
if (typeof chart.resetScales === "function") { try { chart.resetScales(); reset = true; } catch(_) {} }
if (!reset) { reset = _execAction("chartResetView"); }
if (!reset) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"resetScales unavailable"});
return JSON.stringify({ok:true,data:{status:"executed"}});
`)
}

// --- ChartAPI JS functions ---
// jsChartApiPreamble extends jsPreamble with chartApi() resolution.
// Tries multiple access paths and sets `capi` to the chartApi() singleton.
const jsChartApiPreamble = jsPreamble + `
var capi = null;
if (api && typeof api.chartApi === "function") { try { capi = api.chartApi(); } catch(_) {} }
if (!capi && chart && typeof chart.chartApi === "function") { try { capi = chart.chartApi(); } catch(_) {} }
if (!capi && api && api._chartApi) { capi = api._chartApi; }
if (!capi && chart && chart._chartApi) { capi = chart._chartApi; }
if (!capi && api) {
  var _keys = Object.keys(api);
  for (var _i = 0; _i < _keys.length; _i++) {
    var _v = api[_keys[_i]];
    if (_v && typeof _v === "object" && typeof _v.quoteCreateSession === "function") { capi = _v; break; }
  }
}
if (!capi && chart) {
  var _ckeys = Object.keys(chart);
  for (var _ci = 0; _ci < _ckeys.length; _ci++) {
    var _cv = chart[_ckeys[_ci]];
    if (_cv && typeof _cv === "object" && typeof _cv.quoteCreateSession === "function") { capi = _cv; break; }
  }
}
`

func jsProbeChartApi() string {
	return wrapJSEval(jsChartApiPreamble + `
if (!capi) return JSON.stringify({ok:true,data:{found:false,access_paths:[],methods:[]}});
var paths = [];
if (api && typeof api.chartApi === "function") paths.push("api.chartApi()");
if (chart && typeof chart.chartApi === "function") paths.push("chart.chartApi()");
if (api && api._chartApi) paths.push("api._chartApi");
if (chart && chart._chartApi) paths.push("chart._chartApi");
if (paths.length === 0) {
  if (api) {
    var _ak = Object.keys(api);
    for (var _ai = 0; _ai < _ak.length; _ai++) {
      var _av = api[_ak[_ai]];
      if (_av && typeof _av === "object" && typeof _av.quoteCreateSession === "function") {
        paths.push("api[" + JSON.stringify(_ak[_ai]) + "]");
        break;
      }
    }
  }
  if (paths.length === 0 && chart) {
    var _ck = Object.keys(chart);
    for (var _ci2 = 0; _ci2 < _ck.length; _ci2++) {
      var _cv2 = chart[_ck[_ci2]];
      if (_cv2 && typeof _cv2 === "object" && typeof _cv2.quoteCreateSession === "function") {
        paths.push("chart[" + JSON.stringify(_ck[_ci2]) + "]");
        break;
      }
    }
  }
  if (paths.length === 0) paths.push("key-scan");
}
var methods = [];
var seen = {};
var obj = capi;
while (obj && obj !== Object.prototype) {
  var mk = Object.getOwnPropertyNames(obj);
  for (var mi = 0; mi < mk.length; mi++) {
    var mn = mk[mi];
    if (mn === "constructor" || seen[mn]) continue;
    seen[mn] = true;
    try { if (typeof capi[mn] === "function") methods.push(mn); } catch(_) {}
  }
  obj = Object.getPrototypeOf(obj);
}
methods.sort();
return JSON.stringify({ok:true,data:{found:true,access_paths:paths,methods:methods}});
`)
}

func jsProbeChartApiDeep() string {
	return wrapJSEval(jsChartApiPreamble + `
if (!capi) return JSON.stringify({ok:true,data:{found:false}});
var diag = {};
var targets = ["resolveSymbol","quoteCreateSession","quoteDeleteSession","quoteAddSymbols",
  "quoteRemoveSymbols","quoteSetFields","requestFirstBarTime","switchTimezone",
  "chartCreateSession","chartDeleteSession","createSession","removeSession","connect","connected"];
for (var ti = 0; ti < targets.length; ti++) {
  var tn = targets[ti];
  if (typeof capi[tn] === "function") {
    diag[tn] = {exists:true,arity:capi[tn].length,src:capi[tn].toString().substring(0,120)};
  } else if (typeof capi[tn] !== "undefined") {
    diag[tn] = {exists:true,type:typeof capi[tn],value:String(capi[tn])};
  } else {
    diag[tn] = {exists:false};
  }
}
var ownKeys = Object.keys(capi);
var state = {};
for (var oi = 0; oi < ownKeys.length; oi++) {
  var ok2 = ownKeys[oi];
  var ov = capi[ok2];
  if (typeof ov === "function") continue;
  if (typeof ov === "string" || typeof ov === "number" || typeof ov === "boolean") {
    state[ok2] = ov;
  } else if (ov === null || ov === undefined) {
    state[ok2] = null;
  } else if (typeof ov === "object") {
    state[ok2] = "{" + Object.keys(ov).length + " keys}";
  }
}
var sessions = {};
if (capi._sessions) {
  var skeys = Object.keys(capi._sessions);
  for (var ski = 0; ski < skeys.length; ski++) {
    var sv = capi._sessions[skeys[ski]];
    var smethods = [];
    if (sv && typeof sv === "object") {
      var svk = Object.keys(sv);
      for (var svi = 0; svi < svk.length && svi < 20; svi++) {
        smethods.push(svk[svi] + ":" + typeof sv[svk[svi]]);
      }
    }
    sessions[skeys[ski]] = smethods;
  }
}
return JSON.stringify({ok:true,data:{methods:diag,state:state,sessions:sessions}});
`)
}

// jsChartApiSessionHelper resolves an existing chart session ID from capi._sessions
// or from the chart widget's internal state.
const jsChartApiSessionHelper = `
function _findChartSession() {
  if (capi && capi._sessions) {
    var sk = Object.keys(capi._sessions);
    for (var si = 0; si < sk.length; si++) {
      if (sk[si].indexOf("cs_") === 0) return sk[si];
    }
  }
  if (chart) {
    var ck = Object.keys(chart);
    for (var ci = 0; ci < ck.length; ci++) {
      var cv = chart[ck[ci]];
      if (cv && typeof cv === "object" && typeof cv._sessionId === "string" && cv._sessionId.indexOf("cs_") === 0) {
        return cv._sessionId;
      }
    }
    if (typeof chart._chartSession === "object" && chart._chartSession && chart._chartSession._sessionId) {
      return chart._chartSession._sessionId;
    }
  }
  if (api) {
    var ak = Object.keys(api);
    for (var ai = 0; ai < ak.length; ai++) {
      var av = api[ak[ai]];
      if (av && typeof av === "object" && typeof av._sessionId === "string" && av._sessionId.indexOf("cs_") === 0) {
        return av._sessionId;
      }
    }
  }
  return null;
}
`

func jsResolveSymbol(symbol string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsChartApiPreamble+jsChartApiSessionHelper+`
var sym = %s;
if (capi && typeof capi.resolveSymbol === "function") {
  var csid = _findChartSession();
  if (csid) {
    var rid = "sresol_" + Math.random().toString(36).substring(2, 10);
    var info = await new Promise(function(resolve) {
      capi.resolveSymbol(csid, rid, sym, function(data) { resolve(data); });
    });
    if (info) {
      var i = typeof info === "string" ? JSON.parse(info) : info;
      return JSON.stringify({ok:true,data:{
        symbol: String(i.symbol || i.pro_name || i.name || sym),
        name: String(i.name || i.full_name || i.pro_name || ""),
        description: String(i.description || i.short_description || ""),
        exchange: String(i.listed_exchange || i.exchange || ""),
        type: String(i.type || i.security_type || ""),
        currency: String(i.currency_code || i.currency || ""),
        timezone: String(i.timezone || ""),
        pricescale: Number(i.pricescale || i.price_scale || 0),
        minmov: Number(i.minmov || i.min_mov || 0),
        has_intraday: !!(i.has_intraday),
        has_daily: !!(i.has_daily),
        session: String(i.session || ""),
        session_holidays: String(i.session_holidays || "")
      }});
    }
  }
}
if (chart && typeof chart.symbolExt === "function") {
  var cur = "";
  if (typeof chart.symbol === "function") cur = String(chart.symbol() || "");
  if (cur === sym || !capi) {
    var i = chart.symbolExt();
    if (i) {
      return JSON.stringify({ok:true,data:{
        symbol: String(i.symbol || sym),
        name: String(i.name || i.full_name || ""),
        description: String(i.description || i.short_description || ""),
        exchange: String(i.listed_exchange || i.exchange || ""),
        type: String(i.type || i.security_type || ""),
        currency: String(i.currency_code || i.currency || ""),
        timezone: String(i.timezone || ""),
        pricescale: Number(i.pricescale || i.price_scale || 0),
        minmov: Number(i.minmov || i.min_mov || 0),
        has_intraday: !!(i.has_intraday),
        has_daily: !!(i.has_daily),
        session: String(i.session || ""),
        session_holidays: String(i.session_holidays || "")
      }});
    }
  }
}
return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"resolveSymbol unavailable"});
`, jsString(symbol)))
}

func jsSwitchTimezone(tz string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var tz = %s;
var switched = false;
if (chart && typeof chart.setTimezone === "function") {
  chart.setTimezone(tz);
  switched = true;
}
if (!switched) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"switchTimezone unavailable"});
return JSON.stringify({ok:true,data:{timezone:tz}});
`, jsString(tz)))
}

// --- Replay JS functions ---
// These use api._replayApi as the primary control surface (higher-level wrapper),
// with chart._replayManager as fallback for low-level operations.

// jsReplayApiPreamble extends jsPreamble with replayApi resolution.
// Sets `rapi` to the replay API singleton via api._replayApi.
const jsReplayApiPreamble = jsPreamble + `
var rapi = null;
if (api && api._replayApi) { rapi = api._replayApi; }
`

func jsProbeReplayManager() string {
	return wrapJSEval(jsReplayApiPreamble + `
if (!rapi) return JSON.stringify({ok:true,data:{found:false,access_paths:[],methods:[],state:{}}});
var paths = ["api._replayApi"];
var methods = [];
var seen = {};
var obj = rapi;
while (obj && obj !== Object.prototype) {
  var mk = Object.getOwnPropertyNames(obj);
  for (var mi = 0; mi < mk.length; mi++) {
    var mn = mk[mi];
    if (mn === "constructor" || seen[mn]) continue;
    seen[mn] = true;
    try { if (typeof rapi[mn] === "function") methods.push(mn); } catch(_) {}
  }
  obj = Object.getPrototypeOf(obj);
}
methods.sort();
var ownKeys = Object.keys(rapi);
var state = {};
for (var oi = 0; oi < ownKeys.length; oi++) {
  var ok2 = ownKeys[oi];
  var ov = rapi[ok2];
  if (typeof ov === "function") continue;
  if (typeof ov === "string" || typeof ov === "number" || typeof ov === "boolean") {
    state[ok2] = ov;
  } else if (ov === null || ov === undefined) {
    state[ok2] = null;
  } else if (typeof ov === "object") {
    state[ok2] = "{" + Object.keys(ov).length + " keys}";
  }
}
return JSON.stringify({ok:true,data:{found:true,access_paths:paths,methods:methods,state:state}});
`)
}

func jsProbeReplayManagerDeep() string {
	return wrapJSEval(jsReplayApiPreamble + `
if (!rapi) return JSON.stringify({ok:true,data:{found:false}});
var diag = {};
var targets = ["selectDate","selectFirstAvailableDate","selectRandomDate","replayMode",
  "leaveReplay","showReplayToolbar","hideReplayToolbar","doStep","toggleAutoplay",
  "stopReplay","changeAutoplayDelay","changeReplayResolution","goToRealtime",
  "isReplayStarted","isReplayAvailable","isAutoplayStarted","isReadyToPlay",
  "autoplayDelay","autoplayDelayWV","autoplayDelayList","getReplayDepth",
  "getReplaySelectedDate","currentDate","currentReplayResolution","replayTimingMode",
  "buy","sell","closePosition","currency","position","realizedPL","destroy",
  "isReplayModeEnabled","replaySelectedDate","startReplay","enableReplayMode"];
for (var ti = 0; ti < targets.length; ti++) {
  var tn = targets[ti];
  if (typeof rapi[tn] === "function") {
    diag[tn] = {exists:true,arity:rapi[tn].length,src:rapi[tn].toString().substring(0,400)};
  } else if (typeof rapi[tn] !== "undefined") {
    diag[tn] = {exists:true,type:typeof rapi[tn],value:String(rapi[tn])};
  } else {
    diag[tn] = {exists:false};
  }
}
var ownKeys = Object.keys(rapi);
var state = {};
for (var oi = 0; oi < ownKeys.length; oi++) {
  var ok2 = ownKeys[oi];
  var ov = rapi[ok2];
  if (typeof ov === "function") continue;
  if (typeof ov === "string" || typeof ov === "number" || typeof ov === "boolean") {
    state[ok2] = ov;
  } else if (ov === null || ov === undefined) {
    state[ok2] = null;
  } else if (typeof ov === "object") {
    var sub = {};
    var sk = Object.keys(ov);
    for (var si = 0; si < sk.length && si < 10; si++) {
      sub[sk[si]] = typeof ov[sk[si]];
    }
    state[ok2] = sub;
  }
}
return JSON.stringify({ok:true,data:{methods:diag,state:state}});
`)
}

func jsScanReplayActivation() string {
	return wrapJSEval(jsReplayApiPreamble + `
var results = {};
if (!rapi) {
  results.found = false;
  return JSON.stringify({ok:true,data:results});
}
results.found = true;
function _unwrap(r) {
  if (r === null || r === undefined) return {raw_type:"null",value:null};
  var t = typeof r;
  if (t === "string" || t === "number" || t === "boolean") return {raw_type:t,value:r};
  if (t === "object") {
    var isWV = (typeof r.value === "function" && typeof r.subscribe === "function");
    if (isWV) {
      try { return {raw_type:"WatchedValue",value:r.value(),has_value_fn:true,has_subscribe:true}; } catch(e) { return {raw_type:"WatchedValue",error:String(e)}; }
    }
    if ("_value" in r) {
      return {raw_type:"object_with_value",value:r._value};
    }
    // safe key dump for unknown objects
    try { var ks = Object.keys(r).slice(0,10); return {raw_type:"object",keys:ks}; } catch(_) { return {raw_type:"object",value:String(r)}; }
  }
  return {raw_type:t,value:String(r)};
}
function _callSafe(name) {
  if (typeof rapi[name] !== "function") return {available:false};
  try { var r = rapi[name](); return {available:true,result:_unwrap(r)}; } catch(e) { return {available:true,error:String(e.message||e)}; }
}
results.is_replay_available = _callSafe("isReplayAvailable");
results.is_replay_started = _callSafe("isReplayStarted");
results.is_ready_to_play = _callSafe("isReadyToPlay");
results.is_autoplay_started = _callSafe("isAutoplayStarted");
results.current_date = _callSafe("currentDate");
results.autoplay_delay = _callSafe("autoplayDelay");
results.current_resolution = _callSafe("currentReplayResolution");
results.replay_depth = _callSafe("getReplayDepth");
results.selected_date = _callSafe("getReplaySelectedDate");
results.is_toolbar_visible = _callSafe("isReplayToolbarVisible");
results.autoplay_delay_wv = _callSafe("autoplayDelayWV");
// Inspect _replayUIController internal state
if (rapi._replayUIController) {
  var rc = rapi._replayUIController;
  var rci = {};
  if (typeof rc.isReplayModeEnabled === "function") {
    try {
      var wv = rc.isReplayModeEnabled();
      rci.is_replay_mode_enabled = _unwrap(wv);
    } catch(e) { rci.is_replay_mode_enabled = {error:String(e)}; }
  }
  if (typeof rc.isReplayStarted === "function") {
    try {
      var wv2 = rc.isReplayStarted();
      rci.is_replay_started_internal = _unwrap(wv2);
    } catch(e) { rci.is_replay_started_internal = {error:String(e)}; }
  }
  if (typeof rc.replaySelectedDate === "function") {
    try {
      var wv3 = rc.replaySelectedDate();
      rci.replay_selected_date_internal = _unwrap(wv3);
    } catch(e) { rci.replay_selected_date_internal = {error:String(e)}; }
  }
  if (typeof rc.readyToPlay === "function") {
    try {
      var wv4 = rc.readyToPlay();
      rci.ready_to_play_internal = _unwrap(wv4);
    } catch(e) { rci.ready_to_play_internal = {error:String(e)}; }
  }
  if (typeof rc.isAutoplayStarted === "function") {
    try {
      var wv5 = rc.isAutoplayStarted();
      rci.is_autoplay_started_internal = _unwrap(wv5);
    } catch(e) { rci.is_autoplay_started_internal = {error:String(e)}; }
  }
  results._replayUIController = rci;
}
// DOM replay button
var btn = document.getElementById("header-toolbar-replay");
if (btn) {
  results.replay_button = {disabled:btn.disabled, text:btn.innerText.substring(0,50)};
}
return JSON.stringify({ok:true,data:results});
`)
}

func jsGetReplayStatus() string {
	return wrapJSEval(jsReplayApiPreamble + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
var s = {};
function _call(name) {
  if (typeof rapi[name] !== "function") return null;
  try {
    var r = rapi[name]();
    if (r === null || r === undefined) return null;
    if (typeof r === "string" || typeof r === "number" || typeof r === "boolean") return r;
    if (typeof r === "object" && typeof r.value === "function") return r.value();
    if (typeof r === "object" && "_value" in r) return r._value;
    return String(r);
  } catch(_) { return null; }
}
s.is_replay_available = !!_call("isReplayAvailable");
// isReplayStarted() WV can be stale; prefer _replayUIController.isReplayModeEnabled()
s.is_replay_started = false;
if (rapi._replayUIController && typeof rapi._replayUIController.isReplayModeEnabled === "function") {
  try {
    var _rme = rapi._replayUIController.isReplayModeEnabled();
    if (typeof _rme === "boolean") s.is_replay_started = _rme;
    else if (_rme && typeof _rme.value === "function") s.is_replay_started = !!_rme.value();
    else if (_rme && "_value" in _rme) s.is_replay_started = !!_rme._value;
  } catch(_) { s.is_replay_started = !!_call("isReplayStarted"); }
} else {
  s.is_replay_started = !!_call("isReplayStarted");
}
s.is_autoplay_started = !!_call("isAutoplayStarted");
s.is_ready_to_play = !!_call("isReadyToPlay");
s.replay_point = _call("getReplaySelectedDate");
s.server_time = _call("currentDate");
s.autoplay_delay = Number(_call("autoplayDelay") || 0);
s.depth = _call("getReplayDepth");
s.is_replay_finished = false;
s.is_replay_connected = false;
// Try chart.replayStatus() WatchedValue for additional state
if (chart && typeof chart.replayStatus === "function") {
  try {
    var wv = chart.replayStatus();
    if (wv && typeof wv.value === "function") s.replay_status_value = wv.value();
    else if (wv && "_value" in wv) s.replay_status_value = wv._value;
  } catch(_) {}
}
return JSON.stringify({ok:true,data:s});
`)
}

func jsActivateReplay(date float64) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsReplayApiPreamble+`
var date = %v;
// selectDate expects milliseconds; convert if value looks like seconds (< 1e12)
if (date < 1e12) date = date * 1000;
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.selectDate !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"selectDate unavailable"});
try {
  await rapi.selectDate(date);
} catch(e) {
  return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"selectDate failed: " + String(e.message||e)});
}
var started = false;
if (typeof rapi.isReplayStarted === "function") {
  try {
    var wv = rapi.isReplayStarted();
    if (typeof wv === "boolean") started = wv;
    else if (wv && typeof wv.value === "function") started = !!wv.value();
    else if (wv && "_value" in wv) started = !!wv._value;
  } catch(_) {}
}
return JSON.stringify({ok:true,data:{status:"activated",date:date,is_replay_started:started}});
`, date))
}

func jsActivateReplayAuto() string {
	return wrapJSEvalAsync(jsReplayApiPreamble + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.selectFirstAvailableDate === "function") {
  try {
    await rapi.selectFirstAvailableDate();
    var started = false;
    if (typeof rapi.isReplayStarted === "function") {
      try {
        var wv = rapi.isReplayStarted();
        if (typeof wv === "boolean") started = wv;
        else if (wv && typeof wv.value === "function") started = !!wv.value();
        else if (wv && "_value" in wv) started = !!wv._value;
      } catch(_) {}
    }
    var date = null;
    if (typeof rapi.getReplaySelectedDate === "function") {
      try {
        var dv = rapi.getReplaySelectedDate();
        if (typeof dv === "number") date = dv;
        else if (dv && typeof dv.value === "function") date = dv.value();
        else if (dv && "_value" in dv) date = dv._value;
      } catch(_) {}
    }
    return JSON.stringify({ok:true,data:{status:"activated",method:"selectFirstAvailableDate",date:date,is_replay_started:started}});
  } catch(e) {
    return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"selectFirstAvailableDate failed: " + String(e.message||e)});
  }
}
return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"selectFirstAvailableDate unavailable"});
`)
}

func jsDeactivateReplay() string {
	return wrapJSEval(jsReplayApiPreamble + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
// stopReplay() calls requestCloseReplay(true) which skips confirmation dialogs
if (typeof rapi.stopReplay === "function") { rapi.stopReplay(); return JSON.stringify({ok:true,data:{status:"deactivated"}}); }
if (typeof rapi.leaveReplay === "function") { rapi.leaveReplay({skipConfirm:true}); return JSON.stringify({ok:true,data:{status:"deactivated"}}); }
return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"stopReplay/leaveReplay unavailable"});
`)
}

func jsStartReplay(point float64) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsReplayApiPreamble+`
var point = %v;
// selectDate expects milliseconds; convert if value looks like seconds (< 1e12)
if (point < 1e12) point = point * 1000;
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.selectDate === "function") {
  try { await rapi.selectDate(point); } catch(e) {
    return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"selectDate failed: " + String(e.message||e)});
  }
  return JSON.stringify({ok:true,data:{status:"started",point:point}});
}
return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"selectDate unavailable"});
`, point))
}

func jsStopReplay() string {
	return wrapJSEval(jsReplayApiPreamble + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.stopReplay === "function") { rapi.stopReplay(); return JSON.stringify({ok:true,data:{status:"stopped"}}); }
if (typeof rapi.leaveReplay === "function") { rapi.leaveReplay(); return JSON.stringify({ok:true,data:{status:"stopped_via_leave"}}); }
return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"stopReplay unavailable"});
`)
}

func jsReplayStep() string {
	return wrapJSEval(jsReplayApiPreamble + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.doStep !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"doStep unavailable"});
rapi.doStep();
return JSON.stringify({ok:true,data:{status:"stepped"}});
`)
}

func jsStartAutoplay() string {
	return wrapJSEval(jsReplayApiPreamble + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.toggleAutoplay !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"toggleAutoplay unavailable"});
var already = false;
if (typeof rapi.isAutoplayStarted === "function") {
  try {
    var wv = rapi.isAutoplayStarted();
    if (typeof wv === "boolean") already = wv;
    else if (wv && typeof wv.value === "function") already = !!wv.value();
    else if (wv && "_value" in wv) already = !!wv._value;
    else already = false;
  } catch(_) {}
}
if (!already) rapi.toggleAutoplay();
return JSON.stringify({ok:true,data:{status:"autoplay_started"}});
`)
}

func jsStopAutoplay() string {
	return wrapJSEval(jsReplayApiPreamble + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.toggleAutoplay !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"toggleAutoplay unavailable"});
var running = false;
if (typeof rapi.isAutoplayStarted === "function") {
  try {
    var wv = rapi.isAutoplayStarted();
    if (typeof wv === "boolean") running = wv;
    else if (wv && typeof wv.value === "function") running = !!wv.value();
    else if (wv && "_value" in wv) running = !!wv._value;
    else running = false;
  } catch(_) {}
}
if (running) rapi.toggleAutoplay();
return JSON.stringify({ok:true,data:{status:"autoplay_stopped"}});
`)
}

func jsResetReplay() string {
	return wrapJSEval(jsReplayApiPreamble + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.goToRealtime === "function") { rapi.goToRealtime(); return JSON.stringify({ok:true,data:{status:"reset"}}); }
if (typeof rapi.leaveReplay === "function") { rapi.leaveReplay(); return JSON.stringify({ok:true,data:{status:"reset_via_leave"}}); }
return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"reset unavailable"});
`)
}

func jsChangeAutoplayDelay(delay float64) string {
	return wrapJSEval(fmt.Sprintf(jsReplayApiPreamble+`
var delay = %v;
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.changeAutoplayDelay !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"changeAutoplayDelay unavailable"});
rapi.changeAutoplayDelay(delay);
var current = delay;
if (typeof rapi.autoplayDelay === "function") { try { current = Number(rapi.autoplayDelay()); } catch(_) {} }
return JSON.stringify({ok:true,data:{status:"changed",delay:current}});
`, delay))
}

func jsRemoveStudy(studyID string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var id = %s;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.removeEntity === "function") {
  chart.removeEntity(id);
  return JSON.stringify({ok:true,data:{status:"removed"}});
}
if (typeof chart.removeStudy === "function") {
  chart.removeStudy(id);
  return JSON.stringify({ok:true,data:{status:"removed"}});
}
return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"removeStudy unavailable"});
`, jsString(studyID)))
}

// --- Backtesting Strategy API JS functions ---
// These use TradingView's _backtestingStrategyApi singleton which provides
// strategy management, input control, and report data access.

// jsBacktestingApiPreamble extends jsPreamble with _backtestingStrategyApi resolution.
const jsBacktestingApiPreamble = jsPreamble + `
var bsa = api ? api._backtestingStrategyApi : null;
`

// jsBacktestingWVHelper is a shared JS helper that unwraps WatchedValues.
const jsBacktestingWVHelper = `
function _wv(v) {
  if (v === null || v === undefined) return null;
  if (typeof v === "string" || typeof v === "number" || typeof v === "boolean") return v;
  if (typeof v === "object" && typeof v.value === "function") { try { return v.value(); } catch(_) { return null; } }
  if (typeof v === "object" && "_value" in v) return v._value;
  return v;
}
`

func jsProbeBacktestingApi() string {
	return wrapJSEval(jsBacktestingApiPreamble + `
if (!bsa) return JSON.stringify({ok:true,data:{found:false,access_paths:[],methods:[],state:{}}});
var paths = ["api._backtestingStrategyApi"];
var methods = [];
var seen = {};
var obj = bsa;
while (obj && obj !== Object.prototype) {
  var mk = Object.getOwnPropertyNames(obj);
  for (var mi = 0; mi < mk.length; mi++) {
    var mn = mk[mi];
    if (mn === "constructor" || seen[mn]) continue;
    seen[mn] = true;
    try { if (typeof bsa[mn] === "function") methods.push(mn); } catch(_) {}
  }
  obj = Object.getPrototypeOf(obj);
}
methods.sort();
var ownKeys = Object.keys(bsa);
var state = {};
for (var oi = 0; oi < ownKeys.length; oi++) {
  var ok2 = ownKeys[oi];
  var ov = bsa[ok2];
  if (typeof ov === "function") continue;
  if (typeof ov === "string" || typeof ov === "number" || typeof ov === "boolean") {
    state[ok2] = ov;
  } else if (ov === null || ov === undefined) {
    state[ok2] = null;
  } else if (typeof ov === "object") {
    state[ok2] = "{" + Object.keys(ov).length + " keys}";
  }
}
return JSON.stringify({ok:true,data:{found:true,access_paths:paths,methods:methods,state:state}});
`)
}

func jsListStrategies() string {
	return wrapJSEval(jsBacktestingApiPreamble + jsBacktestingWVHelper + `
if (!bsa) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"backtesting API unavailable"});
var raw = _wv(bsa.allStrategies);
if (!raw) raw = _wv(bsa._allStrategies);
if (!raw) return JSON.stringify({ok:true,data:{strategies:[]}});
var strategies = [];
if (Array.isArray(raw)) {
  for (var i = 0; i < raw.length; i++) {
    var s = raw[i];
    if (!s) continue;
    var info = {};
    info.id = typeof s.id === "function" ? String(s.id()) : String(s.id || s.entityId || "");
    info.name = typeof s.name === "function" ? String(s.name()) : String(s.name || s.title || "");
    if (typeof s.isVisible === "function") info.visible = !!s.isVisible();
    strategies.push(info);
  }
} else {
  strategies.push({raw_type: typeof raw, value: String(raw).substring(0, 200)});
}
return JSON.stringify({ok:true,data:{strategies:strategies}});
`)
}

func jsGetActiveStrategy() string {
	return wrapJSEval(jsBacktestingApiPreamble + jsBacktestingWVHelper + `
if (!bsa) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"backtesting API unavailable"});
var active = _wv(bsa.activeStrategy);
if (!active) active = _wv(bsa._activeStrategy);
if (!active) return JSON.stringify({ok:true,data:{strategy:null,inputs:null,meta:null,status:null}});
var strategy = {};
strategy.id = typeof active.id === "function" ? String(active.id()) : String(active.id || active.entityId || "");
strategy.name = typeof active.name === "function" ? String(active.name()) : String(active.name || active.title || "");
var inputs = _wv(bsa.activeStrategyInputsValues);
if (!inputs) inputs = _wv(bsa._activeStrategyInputsValues);
var named = _wv(bsa.activeStrategyNamedInputs);
if (!named) named = _wv(bsa._activeStrategyNamedInputs);
var meta = _wv(bsa.activeStrategyMetaInfo);
if (!meta) meta = _wv(bsa._activeStrategyMetaInfo);
var status = _wv(bsa.activeStrategyStatus);
if (!status) status = _wv(bsa._activeStrategyStatus);
return JSON.stringify({ok:true,data:{strategy:strategy,inputs:inputs,named_inputs:named,meta:meta,status:status}});
`)
}

func jsSetActiveStrategy(strategyID string) string {
	return wrapJSEval(fmt.Sprintf(jsBacktestingApiPreamble+`
var id = %s;
if (!bsa) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"backtesting API unavailable"});
if (typeof bsa.setActiveStrategy !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setActiveStrategy unavailable"});
bsa.setActiveStrategy(id);
return JSON.stringify({ok:true,data:{status:"set",strategy_id:id}});
`, jsString(strategyID)))
}

func jsSetStrategyInput(name string, value any) string {
	return wrapJSEval(fmt.Sprintf(jsBacktestingApiPreamble+`
var name = %s;
var value = %s;
if (!bsa) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"backtesting API unavailable"});
if (typeof bsa.setStrategyInput !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setStrategyInput unavailable"});
bsa.setStrategyInput(name, value);
return JSON.stringify({ok:true,data:{status:"set",name:name,value:value}});
`, jsString(name), jsJSON(value)))
}

func jsGetStrategyReport() string {
	return wrapJSEval(jsBacktestingApiPreamble + jsBacktestingWVHelper + `
if (!bsa) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"backtesting API unavailable"});
var report = _wv(bsa.activeStrategyReportData);
if (!report) report = _wv(bsa._activeStrategyReportData);
if (!report) report = _wv(bsa._reportData);
var status = _wv(bsa.activeStrategyStatus);
if (!status) status = _wv(bsa._activeStrategyStatus);
if (!status) status = _wv(bsa._status);
var isDeep = bsa._isDeepBacktesting || false;
return JSON.stringify({ok:true,data:{report:report,status:status,is_deep_backtesting:isDeep}});
`)
}

func jsGetStrategyDateRange() string {
	return wrapJSEval(jsBacktestingApiPreamble + `
if (!bsa) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"backtesting API unavailable"});
if (typeof bsa.getChartDateRange !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getChartDateRange unavailable"});
var range = bsa.getChartDateRange();
return JSON.stringify({ok:true,data:{date_range:range}});
`)
}

func jsStrategyGotoDate(timestamp float64, belowBar bool) string {
	return wrapJSEval(fmt.Sprintf(jsBacktestingApiPreamble+`
var ts = %v;
var below = %t;
if (!bsa) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"backtesting API unavailable"});
if (typeof bsa.gotoDate !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"gotoDate unavailable"});
bsa.gotoDate(ts, below);
return JSON.stringify({ok:true,data:{status:"navigated",timestamp:ts,below_bar:below}});
`, timestamp, belowBar))
}

// --- Alerts REST API JS functions ---
// These use TradingView's internal getAlertsRestApi() singleton which wraps
// the pricealerts.tradingview.com REST service for alert CRUD and fire management.

// jsAlertsApiPreamble extends jsPreamble with getAlertsRestApi() resolution.
// The alerts REST API is a webpack-internal singleton (module 418369) not exposed
// on window.TradingViewApi. We access it via webpack require captured on
// window.__tvAgentWpRequire, falling back to re-extracting it from the chunk array.
const jsAlertsApiPreamble = jsPreamble + `
var aapi = null;
// Ensure alerts chunk is loaded
if (api && typeof api.alerts === "function") { try { await api.alerts(); } catch(_) {} }
// Get or extract webpack require
var _wpReq = window.__tvAgentWpRequire || null;
if (!_wpReq) {
  var _ca = window.webpackChunktradingview;
  if (_ca && Array.isArray(_ca)) {
    try { _ca.push([["__aapi_" + Date.now()], {}, function(r) { _wpReq = r; }]); } catch(_) {}
    if (_wpReq) window.__tvAgentWpRequire = _wpReq;
  }
}
// Resolve the singleton via webpack module cache
if (_wpReq && _wpReq.c) {
  var _mc = _wpReq.c;
  var _mkeys = Object.keys(_mc);
  for (var _mi = 0; _mi < _mkeys.length; _mi++) {
    try {
      var _exp = _mc[_mkeys[_mi]].exports;
      if (_exp && typeof _exp.getAlertsRestApi === "function") { aapi = _exp.getAlertsRestApi(); break; }
    } catch(_) {}
  }
}
`

func jsScanAlertsAccess() string {
	return wrapJSEvalAsync(jsPreamble + `
var results = {};

// Step 1: Trigger lazy load via api.alerts()
if (api && typeof api.alerts === "function") {
  try { await api.alerts(); } catch(_) {}
}

// Step 2: Find webpack chunk array by scanning window for arrays with overridden push
var chunkArr = null;
var chunkName = null;
var wkeys = Object.getOwnPropertyNames(window);
for (var wi = 0; wi < wkeys.length; wi++) {
  try {
    var wv = window[wkeys[wi]];
    if (Array.isArray(wv) && wv.push !== Array.prototype.push) {
      chunkArr = wv;
      chunkName = wkeys[wi];
      break;
    }
  } catch(_) {}
}
results.chunk_array_name = chunkName;
results.chunk_array_found = !!chunkArr;

// Step 3: Extract webpack require via chunk push trick
// Cache it on window so it survives across evals
var wpRequire = window.__tvAgentWpRequire || null;
if (!wpRequire && chunkArr) {
  try {
    chunkArr.push([["__alertprobe_" + Date.now()], {}, function(req) { wpRequire = req; }]);
    if (wpRequire) window.__tvAgentWpRequire = wpRequire;
  } catch(e) { results.chunk_push_error = String(e).substring(0,200); }
}
results.has_wpRequire = !!wpRequire;

if (wpRequire) {
  // Step 4: Try module 609177 (from alerts() source)
  try {
    var mod = wpRequire(609177);
    if (mod) {
      var modKeys = Object.keys(mod);
      results.module_609177_keys = modKeys;
      if (typeof mod.getAlertsRestApi === "function") {
        results.found_getAlertsRestApi = "wpRequire(609177).getAlertsRestApi";
        var restApi = mod.getAlertsRestApi();
        if (restApi) {
          var methods = [];
          var seen = {};
          var p = restApi;
          while (p && p !== Object.prototype) {
            var names = Object.getOwnPropertyNames(p);
            for (var ni = 0; ni < names.length; ni++) {
              if (names[ni] === "constructor" || seen[names[ni]]) continue;
              seen[names[ni]] = true;
              try { if (typeof restApi[names[ni]] === "function") methods.push(names[ni]); } catch(_) {}
            }
            p = Object.getPrototypeOf(p);
          }
          methods.sort();
          results.restApi_methods = methods;
        }
      }
    }
  } catch(e) { results.module_609177_error = String(e).substring(0,200); }

  // Step 5: Scan module cache for getAlertsRestApi or createAlert
  if (wpRequire.c) {
    var cache = wpRequire.c;
    var cacheKeys = Object.keys(cache);
    results.module_cache_size = cacheKeys.length;
    var cacheHits = [];
    for (var ci = 0; ci < cacheKeys.length; ci++) {
      try {
        var cmod = cache[cacheKeys[ci]];
        if (!cmod || !cmod.exports) continue;
        var exp = cmod.exports;
        var ekeys = Object.keys(exp);
        for (var ei = 0; ei < ekeys.length; ei++) {
          var ek = ekeys[ei];
          if (ek === "getAlertsRestApi" || ek === "createAlert" || ek === "AlertsRestApi") {
            cacheHits.push({moduleId: cacheKeys[ci], key: ek, type: typeof exp[ek]});
          }
        }
      } catch(_) {}
    }
    results.cache_alert_hits = cacheHits;
  }

  // Step 6: Call getAlertsRestApi from module 418369 and inspect the singleton
  try {
    var alertMod = wpRequire(418369);
    if (alertMod && typeof alertMod.getAlertsRestApi === "function") {
      results.found_getAlertsRestApi = true;
      var restApi = alertMod.getAlertsRestApi();
      if (restApi) {
        results.restApi_type = typeof restApi;
        var methods = [];
        var seen = {};
        var rp = restApi;
        while (rp && rp !== Object.prototype) {
          var rnames = Object.getOwnPropertyNames(rp);
          for (var ri = 0; ri < rnames.length; ri++) {
            var rn = rnames[ri];
            if (rn === "constructor" || seen[rn]) continue;
            seen[rn] = true;
            try { if (typeof restApi[rn] === "function") methods.push(rn); } catch(_) {}
          }
          rp = Object.getPrototypeOf(rp);
        }
        methods.sort();
        results.restApi_methods = methods;
        // Own state
        var state = {};
        var rok = Object.keys(restApi);
        for (var si = 0; si < rok.length; si++) {
          var sv = restApi[rok[si]];
          if (typeof sv === "function") continue;
          if (typeof sv === "string" || typeof sv === "number" || typeof sv === "boolean") state[rok[si]] = sv;
          else if (sv === null || sv === undefined) state[rok[si]] = null;
          else if (typeof sv === "object") state[rok[si]] = "{" + Object.keys(sv).length + " keys}";
        }
        results.restApi_state = state;
      } else {
        results.restApi_null = true;
      }
    }
  } catch(e) { results.module_418369_error = String(e).substring(0, 200); }
}

return JSON.stringify({ok:true,data:results});
`)
}

func jsProbeAlertsRestApi() string {
	return wrapJSEvalAsync(jsAlertsApiPreamble + `
if (!aapi) return JSON.stringify({ok:true,data:{found:false,access_paths:[],methods:[],state:{}}});
var paths = ["webpack:getAlertsRestApi()"];
var methods = [];
var seen = {};
var obj = aapi;
while (obj && obj !== Object.prototype) {
  var mk = Object.getOwnPropertyNames(obj);
  for (var mi = 0; mi < mk.length; mi++) {
    var mn = mk[mi];
    if (mn === "constructor" || seen[mn]) continue;
    seen[mn] = true;
    try { if (typeof aapi[mn] === "function") methods.push(mn); } catch(_) {}
  }
  obj = Object.getPrototypeOf(obj);
}
methods.sort();
var ownKeys = Object.keys(aapi);
var state = {};
for (var oi = 0; oi < ownKeys.length; oi++) {
  var ok2 = ownKeys[oi];
  var ov = aapi[ok2];
  if (typeof ov === "function") continue;
  if (typeof ov === "string" || typeof ov === "number" || typeof ov === "boolean") {
    state[ok2] = ov;
  } else if (ov === null || ov === undefined) {
    state[ok2] = null;
  } else if (typeof ov === "object") {
    state[ok2] = "{" + Object.keys(ov).length + " keys}";
  }
}
return JSON.stringify({ok:true,data:{found:true,access_paths:paths,methods:methods,state:state}});
`)
}

func jsProbeAlertsRestApiDeep() string {
	return wrapJSEvalAsync(jsAlertsApiPreamble + `
if (!aapi) return JSON.stringify({ok:true,data:{found:false}});
var diag = {};
var targets = ["listAlerts","getAlerts","createAlert","modifyRestartAlert","deleteAlerts",
  "stopAlerts","restartAlerts","cloneAlerts","listFires","deleteFires","deleteAllFires",
  "deleteFiresByFilter","getOfflineFires","clearOfflineFires","getOfflineFireControls",
  "clearOfflineFireControls","setAlertLog","getAlertLog"];
for (var ti = 0; ti < targets.length; ti++) {
  var tn = targets[ti];
  if (typeof aapi[tn] === "function") {
    diag[tn] = {exists:true,arity:aapi[tn].length,src:aapi[tn].toString().substring(0,400)};
  } else if (typeof aapi[tn] !== "undefined") {
    diag[tn] = {exists:true,type:typeof aapi[tn],value:String(aapi[tn])};
  } else {
    diag[tn] = {exists:false};
  }
}
var ownKeys = Object.keys(aapi);
var state = {};
for (var oi = 0; oi < ownKeys.length; oi++) {
  var ok2 = ownKeys[oi];
  var ov = aapi[ok2];
  if (typeof ov === "function") continue;
  if (typeof ov === "string" || typeof ov === "number" || typeof ov === "boolean") {
    state[ok2] = ov;
  } else if (ov === null || ov === undefined) {
    state[ok2] = null;
  } else if (typeof ov === "object") {
    var sub = {};
    var sk = Object.keys(ov);
    for (var si = 0; si < sk.length && si < 10; si++) {
      sub[sk[si]] = typeof ov[sk[si]];
    }
    state[ok2] = sub;
  }
}
return JSON.stringify({ok:true,data:{methods:diag,state:state}});
`)
}

// --- Alerts CRUD JS functions ---
// All use wrapJSEvalAsync + jsAlertsApiPreamble since these hit the REST backend.

func jsListAlerts() string {
	return wrapJSEvalAsync(jsAlertsApiPreamble + `
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.listAlerts !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"listAlerts unavailable"});
var result = await aapi.listAlerts();
return JSON.stringify({ok:true,data:{alerts:result}});
`)
}

func jsGetAlerts(ids []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var rawIds = %s;
var ids = rawIds.map(function(id) { var n = Number(id); return isNaN(n) ? id : n; });
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.getAlerts !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getAlerts unavailable"});
var result = await aapi.getAlerts({alert_ids: ids});
return JSON.stringify({ok:true,data:{alerts:result}});
`, jsJSON(ids)))
}

func jsCreateAlert(params map[string]any) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var params = %s;
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.createAlert !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"createAlert unavailable"});
var result = await aapi.createAlert(params);
return JSON.stringify({ok:true,data:{alert:result}});
`, jsJSON(params)))
}

func jsModifyAlert(params map[string]any) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var params = %s;
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.modifyRestartAlert !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"modifyRestartAlert unavailable"});
var result = await aapi.modifyRestartAlert(params);
return JSON.stringify({ok:true,data:{alert:result}});
`, jsJSON(params)))
}

func jsDeleteAlerts(ids []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var ids = %s; ids = ids.map(function(id) { var n = Number(id); return isNaN(n) ? id : n; });
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.deleteAlerts !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"deleteAlerts unavailable"});
var result = await aapi.deleteAlerts({alert_ids: ids});
return JSON.stringify({ok:true,data:{status:"deleted",result:result}});
`, jsJSON(ids)))
}

func jsStopAlerts(ids []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var ids = %s; ids = ids.map(function(id) { var n = Number(id); return isNaN(n) ? id : n; });
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.stopAlerts !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"stopAlerts unavailable"});
var result = await aapi.stopAlerts({alert_ids: ids});
return JSON.stringify({ok:true,data:{status:"stopped",result:result}});
`, jsJSON(ids)))
}

func jsRestartAlerts(ids []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var ids = %s; ids = ids.map(function(id) { var n = Number(id); return isNaN(n) ? id : n; });
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.restartAlerts !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"restartAlerts unavailable"});
var result = await aapi.restartAlerts({alert_ids: ids});
return JSON.stringify({ok:true,data:{status:"restarted",result:result}});
`, jsJSON(ids)))
}

func jsCloneAlerts(ids []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var ids = %s; ids = ids.map(function(id) { var n = Number(id); return isNaN(n) ? id : n; });
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.cloneAlerts !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"cloneAlerts unavailable"});
var result = await aapi.cloneAlerts({alert_ids: ids});
return JSON.stringify({ok:true,data:{status:"cloned",result:result}});
`, jsJSON(ids)))
}

func jsListFires() string {
	return wrapJSEvalAsync(jsAlertsApiPreamble + `
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.listFires !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"listFires unavailable"});
var result = await aapi.listFires();
return JSON.stringify({ok:true,data:{fires:result}});
`)
}

func jsDeleteFires(ids []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var ids = %s; ids = ids.map(function(id) { var n = Number(id); return isNaN(n) ? id : n; });
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.deleteFires !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"deleteFires unavailable"});
var result = await aapi.deleteFires({fire_ids: ids});
return JSON.stringify({ok:true,data:{status:"deleted",result:result}});
`, jsJSON(ids)))
}

func jsDeleteAllFires() string {
	return wrapJSEvalAsync(jsAlertsApiPreamble + `
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.deleteAllFires !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"deleteAllFires unavailable"});
var result = await aapi.deleteAllFires();
return JSON.stringify({ok:true,data:{status:"deleted",result:result}});
`)
}
