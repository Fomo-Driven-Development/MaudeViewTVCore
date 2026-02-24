package cdpcontrol

import "fmt"

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

func jsGetChartType() string {
	return wrapJSEval(jsPreamble + `
var ct = null;
if (chart && typeof chart.chartType === "function") {
  try { ct = chart.chartType(); } catch(_) {}
}
if (ct === null || ct === undefined) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chartType getter unavailable"});
return JSON.stringify({ok:true,data:{chart_type:ct}});
`)
}

func jsSetChartType(chartType int) string {
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var requested = %d;
if (!chart || typeof chart.setChartType !== "function")
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setChartType unavailable"});
chart.setChartType(requested);
return JSON.stringify({ok:true,data:{}});
`, chartType))
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

// jsResolutionToSeconds converts a TradingView resolution string to approximate
// bar duration in seconds. Used by scroll-based navigation as a fallback.

// jsGoToFillDate waits for the "Go to" dialog to appear and fills the date textbox.
// The dialog must already be open (via Alt+G keyboard shortcut).

func jsGoToFillDate(dateStr string) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
var dateStr = %s;
// Wait for dialog to appear (up to 3s)
var deadline = Date.now() + 3000;
var dialog = null;
var tb = null;
while (Date.now() < deadline) {
  tb = document.querySelector('input[placeholder*="YYYY"]');
  if (tb) { dialog = tb.closest('[role="dialog"]') || tb.parentElement; break; }
  dialog = document.querySelector('[role="dialog"]');
  if (dialog) {
    tb = dialog.querySelector('input[placeholder*="YYYY"]') || dialog.querySelector('input[type="text"]');
    if (tb) break;
  }
  await new Promise(function(r){ setTimeout(r, 100); });
}
if (!dialog || !tb) return JSON.stringify({ok:false, error_code:"EVAL_FAILURE", error_message:"Go to dialog did not appear"});

// Ensure "Date" tab is selected (not "Custom range")
var tabs = dialog.querySelectorAll('[role="tab"]');
for (var i = 0; i < tabs.length; i++) {
  if (tabs[i].textContent.trim() === 'Date' && tabs[i].getAttribute('aria-selected') !== 'true') {
    tabs[i].click();
    await new Promise(function(r){ setTimeout(r, 200); });
  }
}

// Use native setter to bypass React controlled input
var nativeSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
nativeSetter.call(tb, dateStr);
tb.dispatchEvent(new Event('input', {bubbles: true}));
tb.dispatchEvent(new Event('change', {bubbles: true}));
await new Promise(function(r){ setTimeout(r, 300); });

// Focus the textbox so Enter key submits the form
tb.focus();
return JSON.stringify({ok:true, data:{status:"filled", date:dateStr}});
`, jsString(dateStr)))
}

// jsGoToWaitClose waits for the "Go to" dialog to close and reads the visible range.

func jsGoToWaitClose() string {
	return wrapJSEvalAsync(jsPreamble + `
var deadline = Date.now() + 5000;
while (Date.now() < deadline) {
  var d = document.querySelector('[role="dialog"]');
  var tb = document.querySelector('input[placeholder*="YYYY"]');
  if (!d && !tb) break;
  if (d && d.offsetParent === null && (!tb || tb.offsetParent === null)) break;
  await new Promise(function(r){ setTimeout(r, 200); });
}
// Settle for data load
await new Promise(function(r){ setTimeout(r, 500); });
// Read visible range
var r = chart && typeof chart.getVisibleRange === "function" ? chart.getVisibleRange() : null;
var from = r ? Number(r.from || 0) : 0;
var to = r ? Number(r.to || 0) : 0;
return JSON.stringify({ok:true, data:{status:"navigated", from:from, to:to}});
`)
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

func jsSetTimeFrame(preset, resolution string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsPreamble+`
var preset = %s;
var res = %s;
if (!chart) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chart unavailable"});
if (typeof chart.setTimeFrame !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setTimeFrame unavailable"});
var p = preset.toUpperCase();
if (p === "1Y") p = "12M";
else if (p === "5Y") p = "60M";
// setTimeFrame -> _chartWidget.loadRange(e) -> model.loadRange(e)
// Undo system (Zs) expects: {val: {type,value}, res: intervalString}
// Oi wrapper extracts .val for areEqualTimeFrames; redo uses .val and .res
// Default resolution per preset when none provided
var defaultRes = {"1D":"1","5D":"5","1M":"30","3M":"60","6M":"120","YTD":"1D","12M":"1D","60M":"1W","ALL":"1M"};
var curRes = res || defaultRes[p] || (typeof chart.resolution === "function" ? String(chart.resolution()||"D") : "D");
var tf = {val: {type:"period-back", value:p}, res: curRes};
try { chart.setTimeFrame(tf); } catch(e) {
  return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"setTimeFrame failed: "+String(e.message||e)});
}
// Brief settle for data load
await new Promise(function(r){ setTimeout(r, 500); });
var finalRes = typeof chart.resolution === "function" ? String(chart.resolution()||"") : "";
var r = typeof chart.getVisibleRange === "function" ? chart.getVisibleRange() : null;
var from = r ? Number(r.from||0) : 0;
var to = r ? Number(r.to||0) : 0;
return JSON.stringify({ok:true,data:{preset:preset,resolution:finalRes,from:from,to:to}});
`, jsString(preset), jsString(resolution)))
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

// jsGetChartToggles reads the active state of the Log Scale, Auto Scale, and
// Extended Hours toolbar buttons via DOM inspection.

func jsGetChartToggles() string {
	return wrapJSEval(`
var result = {};
function readBtn(names) {
  for (var i = 0; i < names.length; i++) {
    var btns = document.querySelectorAll('button[data-name="' + names[i] + '"]');
    for (var j = 0; j < btns.length; j++) {
      var b = btns[j];
      if (b.offsetParent === null) continue;
      var pressed = b.getAttribute("aria-pressed");
      if (pressed !== null) return pressed === "true";
      var cls = b.className || "";
      if (cls.indexOf("isActive") !== -1 || cls.indexOf("active") !== -1) return true;
      if (cls.indexOf("isChecked") !== -1 || cls.indexOf("checked") !== -1) return true;
      var active = b.getAttribute("data-active");
      if (active !== null) return active === "true";
      return false;
    }
  }
  return null;
}
result.log_scale = readBtn(["logarithm", "log-scale", "logScale"]);
result.auto_scale = readBtn(["auto", "auto-scale", "autoScale"]);
result.extended_hours = readBtn(["extended-hours", "extendedHours", "sessions"]);
return JSON.stringify({ok:true,data:result});
`)
}

// --- ChartAPI JS functions ---
// jsChartApiPreamble extends jsPreamble with chartApi() resolution.
// Tries multiple access paths and sets `capi` to the chartApi() singleton.

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
  backtesting_api: false,
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
// backtesting_api — call api.backtestingStrategyApi() (returns a Promise)
if (api && typeof api.backtestingStrategyApi === "function") {
  try { r.backtesting_api = !!(await api.backtestingStrategyApi()); } catch(_) {}
} else if (api && api._backtestingStrategyApi) {
  r.backtesting_api = true;
}
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
// hotlists_manager — scan webpack modules for hotlistsManager export
r.hotlists_manager = false;
if (r.webpack_require && _wpReq && _wpReq.c) {
  var _hmc = _wpReq.c;
  var _hmkeys = Object.keys(_hmc);
  for (var _hmi = 0; _hmi < _hmkeys.length; _hmi++) {
    try {
      var _hexp = _hmc[_hmkeys[_hmi]].exports;
      if (_hexp && typeof _hexp.hotlistsManager === "function") { r.hotlists_manager = true; break; }
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
