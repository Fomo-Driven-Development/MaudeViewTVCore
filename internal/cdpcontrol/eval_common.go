package cdpcontrol

import "encoding/json"

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

const jsExecAction = `
function _execAction(id) {
  if (api && typeof api.executeActionById === "function") { api.executeActionById(id); return true; }
  if (chart && typeof chart.executeActionById === "function") { chart.executeActionById(id); return true; }
  if (api && typeof api.executeAction === "function") { api.executeAction(id); return true; }
  return false;
}
`

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

// jsGoToFillDate waits for the "Go to" dialog to appear and fills the date textbox.
// The dialog must already be open (via Alt+G keyboard shortcut).

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

const jsReplayApiPreamble = jsPreamble + `
var rapi = null;
if (api && api._replayApi) { rapi = api._replayApi; }
`

const jsBacktestingApiPreamble = jsPreamble + `
var bsa = null;
if (api) {
  if (typeof api.backtestingStrategyApi === "function") { try { bsa = await api.backtestingStrategyApi(); } catch(_) {} }
  if (!bsa && api._backtestingStrategyApi) { bsa = api._backtestingStrategyApi; }
}
`

// jsBacktestingWVHelper aliases jsWatchedValueHelper for backward compatibility.

const jsBacktestingWVHelper = jsWatchedValueHelper

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

const jsLayoutMatchHelper = `
function _findCurrentLayout(list, currentId) {
  for (var i = 0; i < list.length; i++) {
    var c = list[i];
    var url = String(c.url || c.short_url || "");
    if (url === currentId || String(c.short_url || "") === currentId || String(c.id || "") === currentId) return c;
  }
  // Fallback: match by active() flag
  for (var j = 0; j < list.length; j++) {
    var c2 = list[j];
    if (typeof c2.active === "function" && c2.active()) return c2;
  }
  return null;
}
function _readFavorite(svc, entry) {
  // The live favorites map in state is the source of truth after toggling.
  try {
    var raw = svc._state && svc._state._value ? svc._state._value : null;
    if (!raw) {
      var obs = typeof svc.state === "function" ? svc.state() : null;
      raw = obs && typeof obs.value === "function" ? obs.value() : null;
    }
    if (raw && raw.favorites && typeof raw.favorites === "object") {
      return Boolean(raw.favorites[entry.id]);
    }
  } catch(_) {}
  // Fallback to snapshot on the entry itself.
  return Boolean(entry.favorite || false);
}
`

const jsColoredMutationHelper = `
function _coloredResult(raw, color) {
  var result = [];
  var src = Array.isArray(raw) ? raw : (raw && Array.isArray(raw.symbols) ? raw.symbols : []);
  for (var i = 0; i < src.length; i++) {
    result.push(typeof src[i] === "string" ? src[i] : String(src[i]));
  }
  return {
    id: String((raw && raw.id) || ""),
    name: String((raw && raw.name) || ""),
    color: String((raw && raw.color) || color),
    symbols: result,
    modified: String((raw && raw.modified) || "")
  };
}
`

const jsHotlistsPreamble = jsPreamble + `
var hmgr = window.__tvAgentHotlistsMgr || null;
if (!hmgr) {
  var _wpReq = window.__tvAgentWpRequire || null;
  if (!_wpReq) {
    var _ca = window.webpackChunktradingview;
    if (_ca && Array.isArray(_ca)) {
      try { _ca.push([["__hmgr_" + Date.now()], {}, function(r) { _wpReq = r; }]); } catch(_) {}
      if (_wpReq) window.__tvAgentWpRequire = _wpReq;
    }
  }
  if (_wpReq && _wpReq.c) {
    var _mc = _wpReq.c;
    var _mkeys = Object.keys(_mc);
    for (var _mi = 0; _mi < _mkeys.length; _mi++) {
      try {
        var _exp = _mc[_mkeys[_mi]].exports;
        if (_exp && typeof _exp.hotlistsManager === "function") {
          hmgr = _exp.hotlistsManager();
          if (hmgr) { window.__tvAgentHotlistsMgr = hmgr; }
          break;
        }
      } catch(_) {}
    }
  }
}
`

func jsString(v string) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func jsJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func buildIIFE(async bool, body string) string {
	prefix := "(function(){\n"
	if async {
		prefix = "(async function(){\n"
	}
	return prefix + `try {
` + body + `
} catch (err) {
return JSON.stringify({ok:false,error_code:"` + CodeEvalFailure + `",error_message:String(err && err.message || err)});
}
})()`
}

func wrapJSEval(body string) string      { return buildIIFE(false, body) }
func wrapJSEvalAsync(body string) string { return buildIIFE(true, body) }
