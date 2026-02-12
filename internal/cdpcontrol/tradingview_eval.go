package cdpcontrol

import "fmt"

// jsPreamble is the common setup that resolves the TradingView API and active chart.
const jsPreamble = `
var api = window.TradingViewApi;
var chart = api && typeof api.activeChart === "function" ? api.activeChart() : null;`

// jsProbeObjectHelper provides _probeObj(obj, paths) — shared method enumeration
// and state collection used by all probe functions.
const jsProbeObjectHelper = `
function _probeObj(obj, paths) {
  if (!obj) return {found:false,access_paths:[],methods:[],state:{}};
  var methods = []; var seen = {};
  var p = obj;
  while (p && p !== Object.prototype) {
    var mk = Object.getOwnPropertyNames(p);
    for (var mi = 0; mi < mk.length; mi++) {
      var mn = mk[mi];
      if (mn === "constructor" || seen[mn]) continue;
      seen[mn] = true;
      try { if (typeof obj[mn] === "function") methods.push(mn); } catch(_) {}
    }
    p = Object.getPrototypeOf(p);
  }
  methods.sort();
  var state = {};
  var ownKeys = Object.keys(obj);
  for (var oi = 0; oi < ownKeys.length; oi++) {
    var k = ownKeys[oi]; var v = obj[k];
    if (typeof v === "function") continue;
    if (typeof v === "string" || typeof v === "number" || typeof v === "boolean") state[k] = v;
    else if (v === null || v === undefined) state[k] = null;
    else if (typeof v === "object") state[k] = "{" + Object.keys(v).length + " keys}";
  }
  return {found:true,access_paths:paths,methods:methods,state:state};
}
`

// jsWatchedValueHelper provides _wv(v) and _callWV(obj, name) — shared
// WatchedValue unwrapping used across replay, backtesting, and drawing functions.
const jsWatchedValueHelper = `
function _wv(v) {
  if (v === null || v === undefined) return null;
  if (typeof v === "string" || typeof v === "number" || typeof v === "boolean") return v;
  if (typeof v === "object" && typeof v.value === "function") { try { return v.value(); } catch(_) { return null; } }
  if (typeof v === "object" && "_value" in v) return v._value;
  return v;
}
function _callWV(obj, name) {
  if (typeof obj[name] !== "function") return null;
  try { return _wv(obj[name]()); } catch(_) { return null; }
}
`

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

// jsResolutionToSeconds converts a TradingView resolution string to approximate
// bar duration in seconds. Used by scroll-based navigation as a fallback.
const jsResolutionToSeconds = `
function _resToSec(res) {
  if (!res) return 86400;
  var s = String(res).toUpperCase();
  if (s === "D" || s === "1D") return 86400;
  if (s === "W" || s === "1W") return 604800;
  if (s === "M" || s === "1M") return 2592000;
  var m = s.match(/^(\d+)([DWMS]?)$/);
  if (!m) return 86400;
  var n = parseInt(m[1], 10);
  var u = m[2];
  if (u === "D") return n * 86400;
  if (u === "W") return n * 604800;
  if (u === "M") return n * 2592000;
  if (u === "S") return n;
  return n * 60;
}
`

func jsGoToDate(timestamp int64) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+jsResolutionToSeconds+`
var ts = %d;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
var done = false;
// Try native goToDate first
if (typeof chart.goToDate === "function") { try { chart.goToDate(ts); done = true; } catch(_) {} }
// Try native setVisibleRange
if (!done && typeof chart.setVisibleRange === "function") {
  try { await chart.setVisibleRange({from:ts, to:ts + 86400}); done = true; } catch(_) {}
}
// Fallback: scroll-based navigation using resolution and visible range
if (!done && typeof chart.getVisibleRange === "function" && typeof chart.scrollChartByBar === "function") {
  try {
    var r = chart.getVisibleRange();
    if (r && r.from && r.to) {
      var mid = (r.from + r.to) / 2;
      var res = typeof chart.resolution === "function" ? chart.resolution() : "D";
      var barSec = _resToSec(res);
      var offset = Math.round((ts - mid) / barSec);
      if (offset !== 0) { chart.scrollChartByBar(offset); done = true; }
      else { done = true; }
    }
  } catch(_) {}
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
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+jsResolutionToSeconds+`
var from = %v;
var to = %v;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
var done = false;
// Try native setVisibleRange first
if (typeof chart.setVisibleRange === "function") {
  try { await chart.setVisibleRange({from:from, to:to}); done = true; } catch(_) {}
}
// Fallback: scroll to the midpoint of the requested range
if (!done && typeof chart.getVisibleRange === "function" && typeof chart.scrollChartByBar === "function") {
  try {
    var r = chart.getVisibleRange();
    if (r && r.from && r.to) {
      var targetMid = (from + to) / 2;
      var curMid = (r.from + r.to) / 2;
      var res = typeof chart.resolution === "function" ? chart.resolution() : "D";
      var barSec = _resToSec(res);
      var offset = Math.round((targetMid - curMid) / barSec);
      if (offset !== 0) chart.scrollChartByBar(offset);
      done = true;
    }
  } catch(_) {}
}
if (!done) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setVisibleRange unavailable"});
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
	return wrapJSEval(jsChartApiPreamble + jsProbeObjectHelper + `
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
var r = _probeObj(capi, paths);
return JSON.stringify({ok:true,data:{found:true,access_paths:paths,methods:r.methods}});
`)
}

func jsDeepHealthCheck() string {
	return wrapJSEvalAsync(jsPreamble + `
var r = {
  tradingview_api: !!(api && typeof api.activeChart === "function"),
  chart_widget: !!(api && api._chartWidgetCollection && typeof api._chartWidgetCollection.images === "function"),
  chart_api: false,
  webpack_require: false,
  alerts_api: false,
  watchlist_rest: false,
  replay_api: !!(api && api._replayApi),
  backtesting_api: !!(api && api._backtestingStrategyApi),
  pine_editor_dom: !!(document.querySelector('button[data-name="pine-dialog-button"]') || document.querySelector('button[aria-label="Pine"]')),
  monaco_webpack: !!(window.__tvMonacoNs && window.__tvMonacoNs.editor),
  load_chart: !!(api && api._loadChartService),
  save_chart: false
};
// chart_api — look for chartApi() or quoteCreateSession
if (api && typeof api.chartApi === "function") { try { r.chart_api = !!api.chartApi(); } catch(_) {} }
if (!r.chart_api) {
  var ch = r.tradingview_api ? api.activeChart() : null;
  if (ch && typeof ch.chartApi === "function") { try { r.chart_api = !!ch.chartApi(); } catch(_) {} }
}
// save_chart — check via _loadChartService._saveChartService
if (api && api._loadChartService) {
  var ls = api._loadChartService;
  var lsKeys = Object.keys(ls);
  for (var li = 0; li < lsKeys.length; li++) {
    var lv = ls[lsKeys[li]];
    if (lv && typeof lv === "object" && typeof lv.saveChartSilently === "function") { r.save_chart = true; break; }
  }
}
// webpack_require
var _wpReq = window.__tvAgentWpRequire || null;
if (!_wpReq) {
  var _ca = window.webpackChunktradingview;
  if (_ca && Array.isArray(_ca)) {
    try { _ca.push([["__dhc_" + Date.now()], {}, function(req) { _wpReq = req; }]); } catch(_) {}
    if (_wpReq) window.__tvAgentWpRequire = _wpReq;
  }
}
r.webpack_require = !!(_wpReq && _wpReq.c);
// alerts_api — scan webpack modules for getAlertsRestApi
if (r.webpack_require) {
  if (api && typeof api.alerts === "function") { try { await api.alerts(); } catch(_) {} }
  var _mc = _wpReq.c;
  var _mkeys = Object.keys(_mc);
  for (var _mi = 0; _mi < _mkeys.length; _mi++) {
    try {
      var _exp = _mc[_mkeys[_mi]].exports;
      if (_exp && typeof _exp.getAlertsRestApi === "function") { r.alerts_api = true; break; }
    } catch(_) {}
  }
}
// watchlist_rest — check for fetch-based watchlist API (basic DOM check)
r.watchlist_rest = !!(api && typeof api.getWatchedListWidget === "function");
if (!r.watchlist_rest) {
  r.watchlist_rest = !!document.querySelector('[data-name="base-watchlist-menu"]');
}
return JSON.stringify({ok:true,data:r});
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
	return wrapJSEval(jsReplayApiPreamble + jsProbeObjectHelper + `
var r = _probeObj(rapi, ["api._replayApi"]);
return JSON.stringify({ok:true,data:r});
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
	return wrapJSEval(jsReplayApiPreamble + jsWatchedValueHelper + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
var s = {};
s.is_replay_available = !!_callWV(rapi, "isReplayAvailable");
// isReplayStarted() WV can be stale; prefer _replayUIController.isReplayModeEnabled()
s.is_replay_started = false;
if (rapi._replayUIController && typeof rapi._replayUIController.isReplayModeEnabled === "function") {
  try {
    s.is_replay_started = !!_wv(rapi._replayUIController.isReplayModeEnabled());
  } catch(_) { s.is_replay_started = !!_callWV(rapi, "isReplayStarted"); }
} else {
  s.is_replay_started = !!_callWV(rapi, "isReplayStarted");
}
s.is_autoplay_started = !!_callWV(rapi, "isAutoplayStarted");
s.is_ready_to_play = !!_callWV(rapi, "isReadyToPlay");
s.replay_point = _callWV(rapi, "getReplaySelectedDate");
s.server_time = _callWV(rapi, "currentDate");
s.autoplay_delay = Number(_callWV(rapi, "autoplayDelay") || 0);
s.depth = _callWV(rapi, "getReplayDepth");
s.is_replay_finished = false;
s.is_replay_connected = false;
// Try chart.replayStatus() WatchedValue for additional state
if (chart && typeof chart.replayStatus === "function") {
  try { s.replay_status_value = _wv(chart.replayStatus()); } catch(_) {}
}
return JSON.stringify({ok:true,data:s});
`)
}

func jsActivateReplay(date float64) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsReplayApiPreamble+jsWatchedValueHelper+`
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
var started = !!_callWV(rapi, "isReplayStarted");
return JSON.stringify({ok:true,data:{status:"activated",date:date,is_replay_started:started}});
`, date))
}

func jsActivateReplayAuto() string {
	return wrapJSEvalAsync(jsReplayApiPreamble + jsWatchedValueHelper + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.selectFirstAvailableDate === "function") {
  try {
    await rapi.selectFirstAvailableDate();
    var started = !!_callWV(rapi, "isReplayStarted");
    var date = _callWV(rapi, "getReplaySelectedDate");
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
	return wrapJSEval(jsReplayApiPreamble + jsWatchedValueHelper + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.toggleAutoplay !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"toggleAutoplay unavailable"});
if (!_callWV(rapi, "isAutoplayStarted")) rapi.toggleAutoplay();
return JSON.stringify({ok:true,data:{status:"autoplay_started"}});
`)
}

func jsStopAutoplay() string {
	return wrapJSEval(jsReplayApiPreamble + jsWatchedValueHelper + `
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.toggleAutoplay !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"toggleAutoplay unavailable"});
if (!!_callWV(rapi, "isAutoplayStarted")) rapi.toggleAutoplay();
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

// jsBacktestingWVHelper aliases jsWatchedValueHelper for backward compatibility.
const jsBacktestingWVHelper = jsWatchedValueHelper

func jsProbeBacktestingApi() string {
	return wrapJSEval(jsBacktestingApiPreamble + jsProbeObjectHelper + `
var r = _probeObj(bsa, ["api._backtestingStrategyApi"]);
return JSON.stringify({ok:true,data:r});
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
function _coerceIds(arr) { return arr.map(function(id) { var n = Number(id); return isNaN(n) ? id : n; }); }
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
  // Scan module cache for getAlertsRestApi or createAlert
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
    // Resolve singleton via first cache hit
    results.found_getAlertsRestApi = false;
    for (var hi = 0; hi < cacheHits.length; hi++) {
      if (cacheHits[hi].key === "getAlertsRestApi" && cacheHits[hi].type === "function") {
        try {
          var hitMod = cache[cacheHits[hi].moduleId].exports;
          var restApi = hitMod.getAlertsRestApi();
          if (restApi) {
            results.found_getAlertsRestApi = true;
            results.restApi_type = typeof restApi;
          }
        } catch(_) {}
        break;
      }
    }
  }
}

return JSON.stringify({ok:true,data:results});
`)
}

func jsProbeAlertsRestApi() string {
	return wrapJSEvalAsync(jsAlertsApiPreamble + jsProbeObjectHelper + `
var r = _probeObj(aapi, ["webpack:getAlertsRestApi()"]);
return JSON.stringify({ok:true,data:r});
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
var ids = _coerceIds(%s);
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
var ids = _coerceIds(%s);
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.deleteAlerts !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"deleteAlerts unavailable"});
var result = await aapi.deleteAlerts({alert_ids: ids});
return JSON.stringify({ok:true,data:{status:"deleted",result:result}});
`, jsJSON(ids)))
}

func jsStopAlerts(ids []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var ids = _coerceIds(%s);
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.stopAlerts !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"stopAlerts unavailable"});
var result = await aapi.stopAlerts({alert_ids: ids});
return JSON.stringify({ok:true,data:{status:"stopped",result:result}});
`, jsJSON(ids)))
}

func jsRestartAlerts(ids []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var ids = _coerceIds(%s);
if (!aapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"alerts API unavailable"});
if (typeof aapi.restartAlerts !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"restartAlerts unavailable"});
var result = await aapi.restartAlerts({alert_ids: ids});
return JSON.stringify({ok:true,data:{status:"restarted",result:result}});
`, jsJSON(ids)))
}

func jsCloneAlerts(ids []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsAlertsApiPreamble+`
var ids = _coerceIds(%s);
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
var ids = _coerceIds(%s);
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

// --- Drawing/Shape JS functions ---
// These use the TradingView chart widget API for shape CRUD,
// drawing toggles, tool selection, z-order, and state export/import.

func jsListDrawings() string {
	return wrapJSEval(jsPreamble + `
if (!chart || typeof chart.getAllShapes !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getAllShapes unavailable"});
}
var items = chart.getAllShapes() || [];
var shapes = [];
for (var i = 0; i < items.length; i++) {
  var it = items[i] || {};
  shapes.push({id:String(it.id || it.entityId || ""), name:String(it.name || it.title || "")});
}
return JSON.stringify({ok:true,data:{shapes:shapes}});
`)
}

func jsGetDrawing(id string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var id = %s;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.getShapeById !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getShapeById unavailable"});
var shape = chart.getShapeById(id);
if (!shape) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"shape not found: "+id});
var props = {};
try {
  if (typeof shape.getProperties === "function") props = shape.getProperties() || {};
  else if (typeof shape.properties === "function") props = shape.properties() || {};
  else if (shape.properties) props = shape.properties || {};
} catch(_) {}
return JSON.stringify({ok:true,data:{id:id,properties:props}});
`, jsString(id)))
}

func jsCreateDrawing(point string, options string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
var point = %s;
var opts = %s;
if (!chart || typeof chart.createShape !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"createShape unavailable"});
}
var id = await chart.createShape(point, opts);
if (!id) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"createShape returned null"});
return JSON.stringify({ok:true,data:{id:String(id)}});
`, point, options))
}

func jsCreateMultipointDrawing(points string, options string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
var points = %s;
var opts = %s;
if (!chart || typeof chart.createMultipointShape !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"createMultipointShape unavailable"});
}
var id = await chart.createMultipointShape(points, opts);
if (!id) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"createMultipointShape returned null"});
return JSON.stringify({ok:true,data:{id:String(id)}});
`, points, options))
}

func jsCloneDrawing(id string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var id = %s;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.cloneLineTool !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"cloneLineTool unavailable"});
var newId = chart.cloneLineTool(id);
if (!newId) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"cloneLineTool returned null"});
return JSON.stringify({ok:true,data:{id:String(newId)}});
`, jsString(id)))
}

func jsRemoveDrawing(id string, disableUndo bool) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var id = %s;
var disableUndo = %t;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.removeEntity !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"removeEntity unavailable"});
chart.removeEntity(id, {disableUndo: disableUndo});
return JSON.stringify({ok:true,data:{status:"removed"}});
`, jsString(id), disableUndo))
}

func jsRemoveAllDrawings() string {
	return wrapJSEval(jsPreamble + `
if (!chart || typeof chart.removeAllShapes !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"removeAllShapes unavailable"});
}
chart.removeAllShapes();
return JSON.stringify({ok:true,data:{status:"removed"}});
`)
}

func jsGetDrawingToggles() string {
	return wrapJSEval(jsPreamble + jsWatchedValueHelper + `
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"API unavailable"});
var result = {};
function _wvRead(fn, fallbackProp) {
  var wv = null;
  if (typeof api[fn] === "function") { try { wv = api[fn](); } catch(_) {} }
  if (!wv && fallbackProp && api[fallbackProp]) wv = api[fallbackProp];
  return _wv(wv);
}
result.hide_all = _wvRead("hideAllDrawingTools", "_hideDrawingsWatchedValue");
result.lock_all = _wvRead("lockAllDrawingTools", "_lockDrawingsWatchedValue");
result.magnet_enabled = _wvRead("magnetEnabled", "_magnetEnabledWV");
result.magnet_mode = _wvRead("magnetMode", "_magnetModeWV");
return JSON.stringify({ok:true,data:result});
`)
}

func jsSetHideDrawings(val bool) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var val = %t;
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"API unavailable"});
var wv = null;
if (typeof api.hideAllDrawingTools === "function") { try { wv = api.hideAllDrawingTools(); } catch(_) {} }
if (!wv && api._hideDrawingsWatchedValue) wv = api._hideDrawingsWatchedValue;
if (!wv || typeof wv.setValue !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"hideAllDrawingTools WV unavailable"});
wv.setValue(val);
return JSON.stringify({ok:true,data:{status:"set",value:val}});
`, val))
}

func jsSetLockDrawings(val bool) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var val = %t;
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"API unavailable"});
var wv = null;
if (typeof api.lockAllDrawingTools === "function") { try { wv = api.lockAllDrawingTools(); } catch(_) {} }
if (!wv && api._lockDrawingsWatchedValue) wv = api._lockDrawingsWatchedValue;
if (!wv || typeof wv.setValue !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"lockAllDrawingTools WV unavailable"});
wv.setValue(val);
return JSON.stringify({ok:true,data:{status:"set",value:val}});
`, val))
}

func jsSetMagnet(enabled bool, mode int) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var enabled = %t;
var mode = %d;
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"API unavailable"});
var ewv = null;
if (typeof api.magnetEnabled === "function") { try { ewv = api.magnetEnabled(); } catch(_) {} }
if (!ewv && api._magnetEnabledWV) ewv = api._magnetEnabledWV;
if (!ewv || typeof ewv.setValue !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"magnetEnabled WV unavailable"});
ewv.setValue(enabled);
if (mode >= 0) {
  var mwv = null;
  if (typeof api.magnetMode === "function") { try { mwv = api.magnetMode(); } catch(_) {} }
  if (!mwv && api._magnetModeWV) mwv = api._magnetModeWV;
  if (mwv && typeof mwv.setValue === "function") mwv.setValue(mode);
}
return JSON.stringify({ok:true,data:{status:"set",enabled:enabled,mode:mode}});
`, enabled, mode))
}

func jsSetDrawingVisibility(id string, visible bool) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var id = %s;
var vis = %t;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.setEntityVisibility !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setEntityVisibility unavailable"});
chart.setEntityVisibility(id, vis);
return JSON.stringify({ok:true,data:{status:"set",id:id,visible:vis}});
`, jsString(id), visible))
}

func jsGetDrawingTool() string {
	return wrapJSEval(jsPreamble + jsWatchedValueHelper + `
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"API unavailable"});
if (typeof api.selectedLineTool !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"selectedLineTool unavailable"});
var tool = _wv(api.selectedLineTool());
return JSON.stringify({ok:true,data:{tool:tool}});
`)
}

func jsSetDrawingTool(tool string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
var tool = %s;
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"API unavailable"});
if (typeof api.selectLineTool !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"selectLineTool unavailable"});
await api.selectLineTool(tool);
return JSON.stringify({ok:true,data:{status:"set",tool:tool}});
`, jsString(tool)))
}

func jsSetDrawingZOrder(id string, action string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var id = %s;
var action = %s;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.getShapeById !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getShapeById unavailable"});
var shapeApi = chart.getShapeById(id);
if (!shapeApi) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"shape not found: "+id});
// bringToFront and sendToBack are available directly on the shape API.
// bringForward and sendBackward require selection+shapesGroupController which
// is unreliable in CDP eval context, so we map them to front/back as best effort.
var methods = {
  "bring_forward": "bringToFront",
  "bring_to_front": "bringToFront",
  "send_backward": "sendToBack",
  "send_to_back": "sendToBack"
};
var methodName = methods[action];
if (!methodName) return JSON.stringify({ok:false,error_code:"VALIDATION",error_message:"unknown z-order action: "+action});
if (typeof shapeApi[methodName] !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:methodName+" unavailable on shape API"});
shapeApi[methodName]();
var newZ = typeof shapeApi.zorder === "function" ? shapeApi.zorder() : null;
return JSON.stringify({ok:true,data:{status:"executed",action:action,zorder:newZ}});
`, jsString(id), jsString(action)))
}

func jsExportDrawingsState() string {
	return wrapJSEval(jsPreamble + `
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.getLineToolsState !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getLineToolsState unavailable"});
var state = chart.getLineToolsState();
// Convert Map fields to plain objects for JSON serialization.
// getLineToolsState returns sources and groups as Map instances
// which JSON.stringify serializes as {}.
function mapToObj(m) {
  if (!(m instanceof Map)) return m;
  var obj = {};
  m.forEach(function(v, k) { obj[k] = v; });
  return obj;
}
if (state.sources instanceof Map) state.sources = mapToObj(state.sources);
if (state.groups instanceof Map) state.groups = mapToObj(state.groups);
return JSON.stringify({ok:true,data:{state:state}});
`)
}

func jsImportDrawingsState(state string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
var dto = %s;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.applyLineToolsState !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"applyLineToolsState unavailable"});
// applyLineToolsState expects sources and groups as Map instances.
// JSON round-trip converts them to plain objects — convert back.
function objToMap(o) {
  if (!o || typeof o !== "object" || o instanceof Map) return o;
  var m = new Map();
  var keys = Object.keys(o);
  for (var i = 0; i < keys.length; i++) { m.set(keys[i], o[keys[i]]); }
  return m;
}
if (dto.sources && !(dto.sources instanceof Map)) dto.sources = objToMap(dto.sources);
if (dto.groups && !(dto.groups instanceof Map)) dto.groups = objToMap(dto.groups);
// groupsToValidate and lineToolsToValidate should remain Arrays (already are after JSON parse)
await chart.applyLineToolsState(dto);
return JSON.stringify({ok:true,data:{status:"imported"}});
`, state))
}

func jsTakeSnapshot(format, quality string, hideResolution bool) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"TradingView API unavailable"});
if (typeof api.takeClientScreenshot !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"takeClientScreenshot unavailable"});

var fmt = %s;
var qual = parseFloat(%s) || 0.92;
var hideRes = %v;
var opts = {};
if (hideRes) opts.hideResolution = true;

var canvas = await api.takeClientScreenshot(opts);
if (!canvas || typeof canvas.toDataURL !== "function") {
  return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"screenshot returned non-canvas"});
}

var mime = fmt === "jpeg" ? "image/jpeg" : "image/png";
var dataUrl = canvas.toDataURL(mime, qual);
var w = canvas.width || 0;
var h = canvas.height || 0;

var meta = {layout:"",theme:"",charts:[]};
try {
  if (api._chartWidgetCollection && typeof api._chartWidgetCollection.images === "function") {
    var imgs = api._chartWidgetCollection.images();
    if (imgs) {
      meta.layout = String(imgs.layout || "");
      meta.theme = String(imgs.theme || "");
      if (Array.isArray(imgs.charts)) {
        for (var i = 0; i < imgs.charts.length; i++) {
          var c = imgs.charts[i] || {};
          var cm = c.meta || {};
          var entry = {
            meta: {
              symbol: String(cm.symbol || ""),
              exchange: String(cm.exchange || ""),
              resolution: String(cm.resolution || ""),
              description: String(cm.description || "")
            }
          };
          if (Array.isArray(c.ohlc)) entry.ohlc = c.ohlc.map(String);
          if (c.quotes && typeof c.quotes === "object") {
            var q = {};
            var qk = Object.keys(c.quotes);
            for (var j = 0; j < qk.length; j++) q[qk[j]] = String(c.quotes[qk[j]]);
            entry.quotes = q;
          }
          meta.charts.push(entry);
        }
      }
    }
  }
} catch(_) {}

return JSON.stringify({ok:true,data:{data_url:dataUrl,width:w,height:h,metadata:meta}});
`, jsString(format), jsString(quality), hideResolution))
}

// --- Pine Editor JS functions (DOM-based) ---
// These use DOM button clicks and Monaco direct access instead of internal APIs.
// All are session-level operations (evalOnAnyChart).

func jsProbePineDOM() string {
	return wrapJSEval(`
var result = {buttons:[], bottom_tabs:[], toolbar:[], monaco:null, console_panel:null};
// Scan right sidebar buttons
var sidebar = document.querySelectorAll('[data-name]');
for (var i = 0; i < sidebar.length; i++) {
  var el = sidebar[i];
  var name = el.getAttribute('data-name') || '';
  if (name.toLowerCase().indexOf('pine') !== -1 || name.toLowerCase().indexOf('editor') !== -1 || name.toLowerCase().indexOf('script') !== -1) {
    result.buttons.push({data_name: name, tag: el.tagName, visible: el.offsetParent !== null, text: (el.textContent || '').trim().substring(0, 50)});
  }
}
// Scan bottom panel tabs
var bottomTabs = document.querySelectorAll('#bottom-area button, #bottom-area [role="tab"], .bottom-widgetbar-content button');
for (var i = 0; i < bottomTabs.length; i++) {
  var el = bottomTabs[i];
  var txt = (el.textContent || '').trim();
  var dn = el.getAttribute('data-name') || '';
  if (txt || dn) {
    result.bottom_tabs.push({data_name: dn, text: txt.substring(0, 50), tag: el.tagName, visible: el.offsetParent !== null});
  }
}
// Scan Pine toolbar buttons (save, add to chart, etc.)
var toolbarBtns = document.querySelectorAll('[class*="pine"] button, [data-name*="pine"] button, .tv-script-editor button');
for (var i = 0; i < toolbarBtns.length; i++) {
  var el = toolbarBtns[i];
  var dn = el.getAttribute('data-name') || '';
  var ariaLabel = el.getAttribute('aria-label') || '';
  var txt = (el.textContent || '').trim();
  result.toolbar.push({data_name: dn, aria_label: ariaLabel, text: txt.substring(0, 50), tag: el.tagName, visible: el.offsetParent !== null});
}
// Check for Monaco editor
var monacoEl = document.querySelector('.monaco-editor');
if (monacoEl) {
  result.monaco = {found: true, visible: monacoEl.offsetParent !== null, classes: monacoEl.className.substring(0, 100)};
}
// Check for Pine console
var consoleEl = document.querySelector('[class*="console"], [data-name*="console"]');
if (consoleEl) {
  result.console_panel = {found: true, visible: consoleEl.offsetParent !== null, tag: consoleEl.tagName};
}
return JSON.stringify({ok:true,data:result});
`)
}

// jsPineLocateToggleBtn returns JS that finds the button to click for
// open/close and returns its center coordinates. The Go caller then dispatches
// a trusted CDP Input.dispatchMouseEvent at those coordinates.
func jsPineLocateToggleBtn() string {
	return wrapJSEval(`
var monacoEl = document.querySelector('.monaco-editor');
var isOpen = !!(monacoEl && monacoEl.offsetParent !== null);

var btn = null;
if (isOpen) {
  // Find the Close button inside the Pine editor panel
  var panel = monacoEl;
  for (var up = 0; up < 10 && panel; up++) {
    panel = panel.parentElement;
    if (!panel) break;
    var closeBtns = panel.querySelectorAll('button');
    for (var bi = 0; bi < closeBtns.length; bi++) {
      var b = closeBtns[bi];
      if (!b.offsetParent) continue;
      var cls = b.className || '';
      var txt = (b.textContent || '').trim().toLowerCase();
      if (cls.indexOf('closeButton') !== -1 || txt === 'close') {
        btn = b; break;
      }
    }
    if (btn) break;
  }
} else {
  // Find the sidebar Pine button
  btn = document.querySelector('button[data-name="pine-dialog-button"]')
     || document.querySelector('button[aria-label="Pine"]');
  if (!btn) {
    var allBtns = document.querySelectorAll('[role="toolbar"] button, [class*="toolbar"] button');
    for (var i = 0; i < allBtns.length; i++) {
      if ((allBtns[i].textContent || '').trim() === 'Pine') { btn = allBtns[i]; break; }
    }
  }
}
if (!btn) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"Pine editor toggle button not found in DOM"});
var rect = btn.getBoundingClientRect();
var x = rect.x + rect.width / 2;
var y = rect.y + rect.height / 2;
return JSON.stringify({ok:true,data:{action:isOpen?"close":"open",x:x,y:y}});
`)
}

// jsPineWaitForOpen polls until the Pine editor is visible with rendered content.
func jsPineWaitForOpen() string {
	return wrapJSEvalAsync(`
var deadline = Date.now() + 5000;
var isVisible = false;
var monacoReady = false;
while (Date.now() < deadline) {
  var el = document.querySelector('.monaco-editor');
  if (el && el.offsetParent !== null) {
    isVisible = true;
    var vl = el.querySelector('.view-lines');
    if (vl && vl.children.length > 0) {
      // Also verify no spinner overlay is blocking
      var dialog = el.closest('[class*="wrap-"]') || el.closest('[class*="dialog"]');
      var sp = dialog ? dialog.querySelector('.tv-spinner--shown') : null;
      if (!sp || sp.offsetParent === null) { monacoReady = true; break; }
    }
  }
  await new Promise(function(r) { setTimeout(r, 200); });
}
return JSON.stringify({ok:true,data:{status:"opened",is_visible:isVisible,monaco_ready:monacoReady}});
`)
}

// jsPineWaitForClose polls until the Pine editor disappears.
func jsPineWaitForClose() string {
	return wrapJSEvalAsync(`
var deadline = Date.now() + 3000;
while (Date.now() < deadline) {
  var el = document.querySelector('.monaco-editor');
  if (!el || el.offsetParent === null) break;
  await new Promise(function(r) { setTimeout(r, 200); });
}
var stillVisible = (function() { var el = document.querySelector('.monaco-editor'); return !!(el && el.offsetParent !== null); })();
return JSON.stringify({ok:true,data:{status:"closed",is_visible:stillVisible,monaco_ready:false}});
`)
}

// jsPineMonacoPreamble returns JS code that discovers the Monaco editor namespace
// via the webpack module cache (since TradingView doesn't expose monaco globally)
// and caches it on window.__tvMonacoNs. After this preamble, the variable `monacoNs`
// is set to the monaco namespace (or null if not found).
const jsPineMonacoPreamble = `
// Discover Monaco namespace from webpack cache
var monacoNs = window.__tvMonacoNs || null;
if (!monacoNs) {
  // Ensure webpack require is available
  if (!window.__tvAgentWpRequire) {
    var chunkArr = window.webpackChunktradingview;
    if (chunkArr) {
      chunkArr.push([["__tvMonacoDisc"], {}, function(r) { window.__tvAgentWpRequire = r; }]);
    }
  }
  if (window.__tvAgentWpRequire) {
    var cache = window.__tvAgentWpRequire.c;
    for (var modId in cache) {
      var mod = cache[modId];
      if (mod && mod.exports && mod.exports.editor && typeof mod.exports.editor.getModels === 'function') {
        monacoNs = mod.exports;
        window.__tvMonacoNs = monacoNs;
        break;
      }
    }
  }
}
// Also try global monaco as fallback
if (!monacoNs && typeof monaco !== 'undefined' && monaco.editor) {
  monacoNs = monaco;
  window.__tvMonacoNs = monacoNs;
}
`

func jsPineStatus() string {
	return wrapJSEval(`
var monacoEl = document.querySelector('.monaco-editor');
var isVisible = !!(monacoEl && monacoEl.offsetParent !== null);
var monacoReady = false;
if (isVisible && monacoEl) {
  // Check for rendered editor content in DOM
  var viewLines = monacoEl.querySelector('.view-lines');
  var hasContent = !!(viewLines && viewLines.children.length > 0);
  // Check for stale loading screen overlay (TradingView bug on reopen)
  var dialog = monacoEl.closest('[class*="wrap-"]') || monacoEl.closest('[class*="dialog"]');
  var hasSpinner = false;
  if (dialog) {
    var sp = dialog.querySelector('.tv-spinner--shown');
    hasSpinner = !!(sp && sp.offsetParent !== null);
  }
  monacoReady = hasContent && !hasSpinner;
}
return JSON.stringify({ok:true,data:{status:isVisible?"open":"closed",is_visible:isVisible,monaco_ready:monacoReady}});
`)
}

func jsPineGetSource() string {
	return wrapJSEval(jsPineMonacoPreamble + `
var monacoEl = document.querySelector('.monaco-editor');
if (!monacoEl || monacoEl.offsetParent === null) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"Pine editor not visible — call POST /pine/toggle first"});
}
var source = "";
if (monacoNs) {
  try {
    var models = monacoNs.editor.getModels();
    if (models && models.length > 0) {
      source = models[0].getValue() || "";
    }
  } catch(_) {}
}
// Fallback: read from the visible code lines in DOM
if (!source) {
  try {
    var lines = monacoEl.querySelectorAll('.view-line');
    if (lines.length > 0) {
      var parts = [];
      for (var i = 0; i < lines.length; i++) {
        parts.push(lines[i].textContent || "");
      }
      source = parts.join("\\n");
    }
  } catch(_) {}
}
if (!source) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"could not read source from Monaco editor"});
var scriptName = "";
var m = source.match(/(?:indicator|strategy|library)\s*\(\s*(?:"([^"]+)"|'([^']+)')/);
if (m) scriptName = m[1] || m[2] || "";
return JSON.stringify({ok:true,data:{
  status:"open",
  is_visible:true,
  monaco_ready:true,
  script_name:scriptName,
  script_source:source,
  source_length:source.length,
  line_count:source.split("\\n").length
}});
`)
}

func jsPineSetSource(source string) string {
	return wrapJSEval(fmt.Sprintf(jsPineMonacoPreamble+`
var newSource = %s;
var monacoEl = document.querySelector('.monaco-editor');
if (!monacoEl || monacoEl.offsetParent === null) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"Pine editor not visible — call POST /pine/toggle first"});
}
var setOk = false;
if (monacoNs) {
  try {
    var models = monacoNs.editor.getModels();
    if (models && models.length > 0) {
      models[0].setValue(newSource);
      setOk = true;
    }
  } catch(_) {}
}
if (!setOk) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"could not write source to Monaco editor — Monaco namespace not found"});
var scriptName = "";
var m = newSource.match(/(?:indicator|strategy|library)\s*\(\s*(?:"([^"]+)"|'([^']+)')/);
if (m) scriptName = m[1] || m[2] || "";
return JSON.stringify({ok:true,data:{
  status:"set",
  is_visible:true,
  monaco_ready:true,
  script_name:scriptName,
  source_length:newSource.length,
  line_count:newSource.split("\\n").length
}});
`, jsString(source)))
}

// jsPineFocusEditor ensures the Monaco editor is visible and focused.
// Called before sending trusted CDP key events for save/add-to-chart.
func jsPineFocusEditor() string {
	return wrapJSEval(`
var monacoEl = document.querySelector('.monaco-editor');
if (!monacoEl || monacoEl.offsetParent === null) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"Pine editor not visible — call POST /pine/toggle first"});
}
// Focus the Monaco textarea so keyboard shortcuts are received
var textarea = monacoEl.querySelector('textarea.inputarea');
if (textarea) textarea.focus();
return JSON.stringify({ok:true,data:{status:"focused",is_visible:true,monaco_ready:true}});
`)
}

// jsPineWaitForSave polls briefly after Ctrl+S to let the save complete.
func jsPineWaitForSave() string {
	return wrapJSEvalAsync(`
await new Promise(function(r) { setTimeout(r, 1500); });
return JSON.stringify({ok:true,data:{status:"saved",is_visible:true,monaco_ready:true}});
`)
}

// jsPineWaitForAddToChart waits for TradingView to process Ctrl+Enter.
// If a "Cannot add a script with unsaved changes" confirmation dialog appears,
// it clicks "Save and add to chart" to proceed.
func jsPineWaitForAddToChart() string {
	return wrapJSEvalAsync(`
var deadline = Date.now() + 3000;
while (Date.now() < deadline) {
  // Check for the confirmation dialog about unsaved changes
  var btns = document.querySelectorAll('button');
  for (var i = 0; i < btns.length; i++) {
    var txt = (btns[i].textContent || '').trim();
    if (txt === 'Save and add to chart') {
      btns[i].click();
      await new Promise(function(r) { setTimeout(r, 2000); });
      return JSON.stringify({ok:true,data:{status:"added",is_visible:true,monaco_ready:true}});
    }
  }
  await new Promise(function(r) { setTimeout(r, 200); });
}
return JSON.stringify({ok:true,data:{status:"added",is_visible:true,monaco_ready:true}});
`)
}

func jsPineGetConsole() string {
	return wrapJSEval(`
var messages = [];
// Try reading from Pine console DOM elements
var consoleSelectors = [
  '[class*="console"] [class*="message"]',
  '[class*="console"] [class*="row"]',
  '.tv-script-editor__console [class*="message"]',
  '[data-name*="console"] [class*="message"]'
];
for (var si = 0; si < consoleSelectors.length; si++) {
  var els = document.querySelectorAll(consoleSelectors[si]);
  if (els.length > 0) {
    for (var i = 0; i < els.length; i++) {
      var el = els[i];
      var text = (el.textContent || '').trim();
      if (!text) continue;
      var type = 'info';
      var cls = el.className || '';
      if (cls.indexOf('error') !== -1) type = 'error';
      else if (cls.indexOf('warn') !== -1) type = 'warning';
      messages.push({type:type, message:text});
    }
    break;
  }
}
return JSON.stringify({ok:true,data:{messages:messages}});
`)
}

// jsPineBriefWait waits the given milliseconds then returns the current Pine editor status.
func jsPineBriefWait(ms int) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
await new Promise(function(r) { setTimeout(r, %d); });
var monacoEl = document.querySelector('.monaco-editor');
var isVisible = !!(monacoEl && monacoEl.offsetParent !== null);
var monacoReady = false;
if (isVisible && monacoEl) {
  var viewLines = monacoEl.querySelector('.view-lines');
  monacoReady = !!(viewLines && viewLines.children.length > 0);
}
return JSON.stringify({ok:true,data:{status:isVisible?"open":"closed",is_visible:isVisible,monaco_ready:monacoReady}});
`, ms))
}

// jsPineWaitForNewScript waits for a new script template to load in the editor after a chord shortcut.
func jsPineWaitForNewScript() string {
	return jsPineBriefWait(2000)
}

// jsPineClickFirstScriptResult clicks the first result row in the "Open my script" dialog,
// waits for the script to load, then returns close button coords so the caller can CDP-click it.
func jsPineClickFirstScriptResult() string {
	return wrapJSEvalAsync(`
// The "Open my script" dialog has result rows with class containing "itemInfo-".
// The itemInfo div has an onclick handler that loads the script.
var deadline = Date.now() + 3000;
var clicked = false;
var itemEl = null;
while (Date.now() < deadline && !clicked) {
  var items = document.querySelectorAll('[class*="itemInfo-"]');
  if (items.length > 0) {
    itemEl = items[0];
    items[0].click();
    clicked = true;
    break;
  }
  await new Promise(function(r) { setTimeout(r, 200); });
}
// Wait for editor to reload the script
await new Promise(function(r) { setTimeout(r, 1200); });
// Find the dialog close button: walk up from the clicked item to find container,
// then find the close button within it.
var closeX = 0, closeY = 0;
if (itemEl) {
  var container = itemEl;
  for (var i = 0; i < 10 && container; i++) {
    var closeBtn = container.querySelector('[class*="close-"]');
    if (closeBtn && closeBtn.tagName === 'BUTTON') {
      var r = closeBtn.getBoundingClientRect();
      closeX = r.x + r.width / 2;
      closeY = r.y + r.height / 2;
      break;
    }
    container = container.parentElement;
  }
}
var monacoEl = document.querySelector('.monaco-editor');
var isVisible = !!(monacoEl && monacoEl.offsetParent !== null);
var monacoReady = false;
if (isVisible && monacoEl) {
  var vl = monacoEl.querySelector('.view-lines');
  monacoReady = !!(vl && vl.children.length > 0);
}
return JSON.stringify({ok:true,data:{status:isVisible?"open":"closed",is_visible:isVisible,monaco_ready:monacoReady,close_x:closeX,close_y:closeY}});
`)
}

// jsPineFindReplace uses the Monaco API to find all occurrences and replace them,
// preserving undo history via pushEditOperations.
func jsPineFindReplace(find, replace string) string {
	return wrapJSEval(fmt.Sprintf(jsPineMonacoPreamble+`
var findStr = %s;
var replaceStr = %s;
var monacoEl = document.querySelector('.monaco-editor');
if (!monacoEl || monacoEl.offsetParent === null) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"Pine editor not visible — call POST /pine/toggle first"});
}
if (!monacoNs) {
  return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"Monaco namespace not found"});
}
var models = monacoNs.editor.getModels();
if (!models || models.length === 0) {
  return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"No Monaco models found"});
}
var model = models[0];
var matches = model.findMatches(findStr, false, false, true, null, false);
var matchCount = matches.length;
if (matchCount === 0) {
  var src = model.getValue();
  var scriptName = "";
  var m = src.match(/(?:indicator|strategy|library)\s*\(\s*(?:"([^"]+)"|'([^']+)')/);
  if (m) scriptName = m[1] || m[2] || "";
  return JSON.stringify({ok:true,data:{status:"no_matches",is_visible:true,monaco_ready:true,match_count:0,script_name:scriptName,source_length:src.length,line_count:src.split("\\n").length}});
}
var edits = [];
for (var i = 0; i < matches.length; i++) {
  edits.push({range: matches[i].range, text: replaceStr});
}
model.pushEditOperations([], edits, function() { return null; });
var newSource = model.getValue();
var scriptName = "";
var m = newSource.match(/(?:indicator|strategy|library)\s*\(\s*(?:"([^"]+)"|'([^']+)')/);
if (m) scriptName = m[1] || m[2] || "";
return JSON.stringify({ok:true,data:{status:"replaced",is_visible:true,monaco_ready:true,match_count:matchCount,script_name:scriptName,source_length:newSource.length,line_count:newSource.split("\\n").length}});
`, jsString(find), jsString(replace)))
}

// --- Layout management JS functions ---

func jsListLayouts() string {
	return wrapJSEval(jsPreamble + `
if (!api || !api._loadChartService) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"_loadChartService unavailable"});
var svc = api._loadChartService;
var st = typeof svc.state === "function" ? svc.state() : null;
var stVal = st && typeof st.value === "function" ? st.value() : st;
if (!stVal || !stVal.chartList) {
  if (typeof svc.refreshChartList === "function") svc.refreshChartList();
  st = typeof svc.state === "function" ? svc.state() : null;
  stVal = st && typeof st.value === "function" ? st.value() : st;
}
if (!stVal || !stVal.chartList) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chartList not available"});
var list = stVal.chartList;
var layouts = [];
for (var i = 0; i < list.length; i++) {
  var c = list[i];
  layouts.push({
    id: Number(c.id || c.uid || 0),
    url: String(c.url || c.short_url || ""),
    name: String(c.name || ""),
    symbol: String(c.symbol || ""),
    interval: String(c.resolution || c.interval || ""),
    modified: Number(c.modified || c.timestamp || 0),
    favorite: Boolean(c.is_favorite || c.favorite || false)
  });
}
return JSON.stringify({ok:true,data:{layouts:layouts}});
`)
}

func jsLayoutStatus() string {
	return wrapJSEval(jsPreamble + `
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"TradingViewApi unavailable"});
var layoutName = typeof api.layoutName === "function" ? String(api.layoutName() || "") : "";
var layoutId = typeof api.layoutId === "function" ? String(api.layoutId() || "") : "";
var gridTemplate = typeof api.layout === "function" ? String(api.layout() || "s") : "s";
var chartCount = typeof api.chartsCount === "function" ? Number(api.chartsCount() || 1) : 1;
var activeIndex = typeof api.activeChartIndex === "function" ? Number(api.activeChartIndex() || 0) : 0;
var isFullscreen = false;
var isMaximized = false;
var hasChanges = false;
if (api._chartWidgetCollection) {
  var cwc = api._chartWidgetCollection;
  if (cwc.fullscreen && typeof cwc.fullscreen === "function") {
    var fsVal = cwc.fullscreen();
    isFullscreen = Boolean(fsVal && typeof fsVal.value === "function" ? fsVal.value() : fsVal);
  }
  isMaximized = cwc._maximizedChartDef != null;
}
var saveSvc = typeof api.getSaveChartService === "function" ? api.getSaveChartService() : null;
if (saveSvc) {
  var hc = typeof saveSvc.hasChanges === "function" ? saveSvc.hasChanges() : saveSvc.hasChanges;
  if (hc && typeof hc.value === "function") hc = hc.value();
  hasChanges = Boolean(hc);
}
return JSON.stringify({ok:true,data:{layout_name:layoutName,layout_id:layoutId,grid_template:gridTemplate,chart_count:chartCount,active_index:activeIndex,is_maximized:isMaximized,is_fullscreen:isFullscreen,has_changes:hasChanges}});
`)
}

func jsSuppressBeforeunload() string {
	return wrapJSEval(`
window.onbeforeunload = null;
var evts = typeof getEventListeners === "function" ? getEventListeners(window) : null;
if (!evts) {
  window.addEventListener("beforeunload", function(e) { e.stopImmediatePropagation(); }, true);
}
return JSON.stringify({ok:true,data:{}});
`)
}

func jsSwitchLayoutResolveURL(id int) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
if (!api || !api._loadChartService) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"_loadChartService unavailable"});
var svc = api._loadChartService;
var st = typeof svc.state === "function" ? svc.state() : null;
var stVal = st && typeof st.value === "function" ? st.value() : st;
if (!stVal || !stVal.chartList) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chartList not available"});
var targetId = %d;
var found = null;
for (var i = 0; i < stVal.chartList.length; i++) {
  if (Number(stVal.chartList[i].id) === targetId) { found = stVal.chartList[i]; break; }
}
if (!found) return JSON.stringify({ok:false,error_code:"NOT_FOUND",error_message:"layout not found: " + targetId});
var shortUrl = String(found.url || found.short_url || "");
if (!shortUrl) return JSON.stringify({ok:false,error_code:"NOT_FOUND",error_message:"no URL for layout: " + targetId});
return JSON.stringify({ok:true,data:{short_url:shortUrl,name:String(found.name||"")}});
`, id))
}

func jsSaveLayout() string {
	return wrapJSEvalAsync(jsPreamble + `
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"TradingViewApi unavailable"});
var saveSvc = typeof api.getSaveChartService === "function" ? api.getSaveChartService() : null;
if (!saveSvc || typeof saveSvc.saveChartSilently !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"saveChartSilently unavailable"});
await saveSvc.saveChartSilently(undefined, undefined, {});
var layoutName = typeof api.layoutName === "function" ? String(api.layoutName() || "") : "";
var layoutId = typeof api.layoutId === "function" ? String(api.layoutId() || "") : "";
return JSON.stringify({ok:true,data:{status:"saved",layout_name:layoutName,layout_id:layoutId}});
`)
}

func jsCloneLayout(name string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"TradingViewApi unavailable"});
var saveSvc = typeof api.getSaveChartService === "function" ? api.getSaveChartService() : null;
if (!saveSvc || !saveSvc._saveAsController || typeof saveSvc._saveAsController._doCloneCurrentLayout !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"_doCloneCurrentLayout unavailable"});
var cloneName = %s;
try {
  await saveSvc._saveAsController._doCloneCurrentLayout(cloneName);
} catch(e) {
  return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"clone failed: " + String(e && e.message || e)});
}
// Refresh the in-memory chart list so subsequent list reads see the clone
if (api._loadChartService && typeof api._loadChartService.refreshChartList === "function") {
  try { api._loadChartService.refreshChartList(); } catch(_) {}
}
var layoutName = typeof api.layoutName === "function" ? String(api.layoutName() || "") : "";
var layoutId = typeof api.layoutId === "function" ? String(api.layoutId() || "") : "";
return JSON.stringify({ok:true,data:{status:"cloned",layout_name:layoutName,layout_id:layoutId}});
`, jsString(name)))
}

func jsDeleteLayout(id int) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
if (!api || !api._loadChartService) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"_loadChartService unavailable"});
var targetId = %d;
var svc = api._loadChartService;
var st = typeof svc.state === "function" ? svc.state() : null;
var stVal = st && typeof st.value === "function" ? st.value() : st;
if (!stVal || !stVal.chartList) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chartList not available"});
var found = null;
for (var i = 0; i < stVal.chartList.length; i++) {
  if (Number(stVal.chartList[i].id) === targetId) { found = stVal.chartList[i]; break; }
}
if (!found) return JSON.stringify({ok:false,error_code:"NOT_FOUND",error_message:"layout not found: " + targetId});
var isActive = typeof found.active === "function" ? found.active() : false;
if (isActive) return JSON.stringify({ok:false,error_code:"VALIDATION",error_message:"cannot delete the active layout"});
if (typeof found.deleteAction !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"deleteAction not available on chart entry"});
try {
  var result = found.deleteAction();
  if (result && typeof result.then === "function") await result;
} catch(e) {
  return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"deleteAction failed: " + String(e && e.message || e)});
}
return JSON.stringify({ok:true,data:{status:"deleted",layout_id:String(targetId)}});
`, id))
}

func jsRenameLayout(name string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"TradingViewApi unavailable"});
var cwc = api._chartWidgetCollection;
if (!cwc || !cwc.metaInfo || !cwc.metaInfo.name || typeof cwc.metaInfo.name.setValue !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"metaInfo.name.setValue unavailable"});
var newName = %s;
cwc.metaInfo.name.setValue(newName);
var saveSvc = typeof api.getSaveChartService === "function" ? api.getSaveChartService() : null;
if (saveSvc && typeof saveSvc.saveChartSilently === "function") await saveSvc.saveChartSilently(undefined, undefined, {});
var layoutName = typeof api.layoutName === "function" ? String(api.layoutName() || "") : "";
var layoutId = typeof api.layoutId === "function" ? String(api.layoutId() || "") : "";
return JSON.stringify({ok:true,data:{status:"renamed",layout_name:layoutName,layout_id:layoutId}});
`, jsString(name)))
}

func jsSetGrid(template string) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
if (!api || typeof api.setLayout !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setLayout unavailable"});
var tmpl = %s;
api.setLayout(tmpl);
var gridTemplate = typeof api.layout === "function" ? String(api.layout() || tmpl) : tmpl;
var chartCount = typeof api.chartsCount === "function" ? Number(api.chartsCount() || 1) : 1;
var activeIndex = typeof api.activeChartIndex === "function" ? Number(api.activeChartIndex() || 0) : 0;
var layoutName = typeof api.layoutName === "function" ? String(api.layoutName() || "") : "";
var layoutId = typeof api.layoutId === "function" ? String(api.layoutId() || "") : "";
var isFullscreen = false;
var isMaximized = false;
var hasChanges = false;
if (api._chartWidgetCollection) {
  var cwc = api._chartWidgetCollection;
  if (cwc.fullscreen && typeof cwc.fullscreen === "function") {
    var fsVal = cwc.fullscreen();
    isFullscreen = Boolean(fsVal && typeof fsVal.value === "function" ? fsVal.value() : fsVal);
  }
  isMaximized = cwc._maximizedChartDef != null;
}
var saveSvc = typeof api.getSaveChartService === "function" ? api.getSaveChartService() : null;
if (saveSvc) {
  var hc = typeof saveSvc.hasChanges === "function" ? saveSvc.hasChanges() : saveSvc.hasChanges;
  if (hc && typeof hc.value === "function") hc = hc.value();
  hasChanges = Boolean(hc);
}
return JSON.stringify({ok:true,data:{layout_name:layoutName,layout_id:layoutId,grid_template:gridTemplate,chart_count:chartCount,active_index:activeIndex,is_maximized:isMaximized,is_fullscreen:isFullscreen,has_changes:hasChanges}});
`, jsString(template)))
}

func jsActivateChart(index int) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
if (!api || typeof api.setActiveChart !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setActiveChart unavailable"});
var idx = %d;
api.setActiveChart(idx);
var layoutName = typeof api.layoutName === "function" ? String(api.layoutName() || "") : "";
var layoutId = typeof api.layoutId === "function" ? String(api.layoutId() || "") : "";
var gridTemplate = typeof api.layout === "function" ? String(api.layout() || "s") : "s";
var chartCount = typeof api.chartsCount === "function" ? Number(api.chartsCount() || 1) : 1;
var activeIndex = typeof api.activeChartIndex === "function" ? Number(api.activeChartIndex() || 0) : 0;
var isFullscreen = false;
var isMaximized = false;
var hasChanges = false;
if (api._chartWidgetCollection) {
  var cwc = api._chartWidgetCollection;
  if (cwc.fullscreen && typeof cwc.fullscreen === "function") {
    var fsVal = cwc.fullscreen();
    isFullscreen = Boolean(fsVal && typeof fsVal.value === "function" ? fsVal.value() : fsVal);
  }
  isMaximized = cwc._maximizedChartDef != null;
}
var saveSvc = typeof api.getSaveChartService === "function" ? api.getSaveChartService() : null;
if (saveSvc) {
  var hc = typeof saveSvc.hasChanges === "function" ? saveSvc.hasChanges() : saveSvc.hasChanges;
  if (hc && typeof hc.value === "function") hc = hc.value();
  hasChanges = Boolean(hc);
}
return JSON.stringify({ok:true,data:{layout_name:layoutName,layout_id:layoutId,grid_template:gridTemplate,chart_count:chartCount,active_index:activeIndex,is_maximized:isMaximized,is_fullscreen:isFullscreen,has_changes:hasChanges}});
`, index))
}

func jsToggleFullscreen() string {
	return wrapJSEval(jsPreamble + `
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"TradingViewApi unavailable"});
var isFS = false;
if (api._chartWidgetCollection && api._chartWidgetCollection.fullscreen && typeof api._chartWidgetCollection.fullscreen === "function") {
  var fsVal = api._chartWidgetCollection.fullscreen();
  isFS = Boolean(fsVal && typeof fsVal.value === "function" ? fsVal.value() : fsVal);
}
if (isFS) {
  if (typeof api.exitFullscreen === "function") api.exitFullscreen();
} else {
  if (typeof api.startFullscreen === "function") api.startFullscreen();
}
var layoutName = typeof api.layoutName === "function" ? String(api.layoutName() || "") : "";
var layoutId = typeof api.layoutId === "function" ? String(api.layoutId() || "") : "";
var gridTemplate = typeof api.layout === "function" ? String(api.layout() || "s") : "s";
var chartCount = typeof api.chartsCount === "function" ? Number(api.chartsCount() || 1) : 1;
var activeIndex = typeof api.activeChartIndex === "function" ? Number(api.activeChartIndex() || 0) : 0;
var isFullscreen = !isFS;
var isMaximized = false;
var hasChanges = false;
if (api._chartWidgetCollection) {
  isMaximized = api._chartWidgetCollection._maximizedChartDef != null;
}
var saveSvc = typeof api.getSaveChartService === "function" ? api.getSaveChartService() : null;
if (saveSvc) {
  var hc = typeof saveSvc.hasChanges === "function" ? saveSvc.hasChanges() : saveSvc.hasChanges;
  if (hc && typeof hc.value === "function") hc = hc.value();
  hasChanges = Boolean(hc);
}
return JSON.stringify({ok:true,data:{layout_name:layoutName,layout_id:layoutId,grid_template:gridTemplate,chart_count:chartCount,active_index:activeIndex,is_maximized:isMaximized,is_fullscreen:isFullscreen,has_changes:hasChanges}});
`)
}
