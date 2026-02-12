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

// jsWatchlistPreamble discovers the watchlist/symbol-list facade.
const jsWatchlistPreamble = `
var wl = null;
if (window.TradingViewApi && typeof window.TradingViewApi.getWatchlist === "function") {
  wl = window.TradingViewApi;
}
if (!wl) {
  var frames = document.querySelectorAll("iframe");
  for (var fi = 0; fi < frames.length; fi++) {
    try {
      var fw = frames[fi].contentWindow;
      if (fw && fw.TradingViewApi && typeof fw.TradingViewApi.getWatchlist === "function") {
        wl = fw.TradingViewApi;
        break;
      }
    } catch(_){}
  }
}
`

func jsListWatchlists() string {
	return wrapJSEvalAsync(jsWatchlistPreamble + `
if (!wl) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist API unavailable"});
var raw = null;
if (typeof wl.getWatchlist === "function") {
  try { raw = await wl.getWatchlist(); } catch(_){}
}
if (!raw && typeof wl.getSymbolLists === "function") {
  try { raw = await wl.getSymbolLists(); } catch(_){}
}
if (!raw) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getWatchlist returned null"});
var lists = [];
if (Array.isArray(raw)) {
  for (var i = 0; i < raw.length; i++) {
    var it = raw[i] || {};
    lists.push({
      id: String(it.id || it.listId || ""),
      name: String(it.name || it.title || ""),
      type: String(it.type || ""),
      active: !!(it.active || it.isActive),
      count: Number(it.count || (it.symbols && it.symbols.length) || 0)
    });
  }
} else if (typeof raw === "object") {
  var keys = Object.keys(raw);
  for (var k = 0; k < keys.length; k++) {
    var it = raw[keys[k]] || {};
    lists.push({
      id: String(it.id || it.listId || keys[k]),
      name: String(it.name || it.title || ""),
      type: String(it.type || ""),
      active: !!(it.active || it.isActive),
      count: Number(it.count || (it.symbols && it.symbols.length) || 0)
    });
  }
}
return JSON.stringify({ok:true,data:{watchlists:lists}});
`)
}

func jsGetActiveWatchlist() string {
	return wrapJSEvalAsync(jsWatchlistPreamble + `
if (!wl) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist API unavailable"});
var raw = null;
if (typeof wl.getActiveWatchlist === "function") {
  try { raw = await wl.getActiveWatchlist(); } catch(_){}
}
if (!raw && typeof wl.getActiveSymbolList === "function") {
  try { raw = await wl.getActiveSymbolList(); } catch(_){}
}
if (!raw) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getActiveWatchlist returned null"});
var syms = [];
if (raw.symbols && Array.isArray(raw.symbols)) {
  for (var i = 0; i < raw.symbols.length; i++) {
    var s = raw.symbols[i];
    syms.push(typeof s === "string" ? s : String(s.symbol || s.name || s));
  }
}
return JSON.stringify({ok:true,data:{
  id: String(raw.id || raw.listId || ""),
  name: String(raw.name || raw.title || ""),
  type: String(raw.type || ""),
  symbols: syms
}});
`)
}

func jsSetActiveWatchlist(id string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistPreamble+`
var listId = %s;
if (!wl) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist API unavailable"});
var ok = false;
if (typeof wl.setActiveWatchlist === "function") {
  try { await wl.setActiveWatchlist(listId); ok = true; } catch(_){}
}
if (!ok && typeof wl.selectSymbolList === "function") {
  try { await wl.selectSymbolList(listId); ok = true; } catch(_){}
}
if (!ok) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setActiveWatchlist unavailable"});
return JSON.stringify({ok:true,data:{id:listId,name:""}});
`, jsString(id)))
}

func jsGetWatchlist(id string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistPreamble+`
var listId = %s;
if (!wl) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist API unavailable"});
var raw = null;
if (typeof wl.getWatchlistById === "function") {
  try { raw = await wl.getWatchlistById(listId); } catch(_){}
}
if (!raw && typeof wl.getSymbolList === "function") {
  try { raw = await wl.getSymbolList(listId); } catch(_){}
}
if (!raw) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"watchlist not found: "+listId});
var syms = [];
if (raw.symbols && Array.isArray(raw.symbols)) {
  for (var i = 0; i < raw.symbols.length; i++) {
    var s = raw.symbols[i];
    syms.push(typeof s === "string" ? s : String(s.symbol || s.name || s));
  }
}
return JSON.stringify({ok:true,data:{
  id: String(raw.id || raw.listId || listId),
  name: String(raw.name || raw.title || ""),
  type: String(raw.type || ""),
  symbols: syms
}});
`, jsString(id)))
}

func jsCreateWatchlist(name string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistPreamble+`
var listName = %s;
if (!wl) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist API unavailable"});
var raw = null;
if (typeof wl.createWatchlist === "function") {
  try { raw = await wl.createWatchlist(listName); } catch(_){}
}
if (!raw && typeof wl.createSymbolList === "function") {
  try { raw = await wl.createSymbolList(listName); } catch(_){}
}
if (!raw) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"createWatchlist unavailable"});
return JSON.stringify({ok:true,data:{
  id: String(raw.id || raw.listId || ""),
  name: String(raw.name || raw.title || listName),
  type: String(raw.type || ""),
  count: Number(raw.count || (raw.symbols && raw.symbols.length) || 0)
}});
`, jsString(name)))
}

func jsRenameWatchlist(id, name string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistPreamble+`
var listId = %s;
var newName = %s;
if (!wl) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist API unavailable"});
var ok = false;
if (typeof wl.renameWatchlist === "function") {
  try { await wl.renameWatchlist(listId, newName); ok = true; } catch(_){}
}
if (!ok && typeof wl.renameSymbolList === "function") {
  try { await wl.renameSymbolList(listId, newName); ok = true; } catch(_){}
}
if (!ok) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"renameWatchlist unavailable"});
return JSON.stringify({ok:true,data:{id:listId,name:newName,type:"",count:0}});
`, jsString(id), jsString(name)))
}

func jsDeleteWatchlist(id string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistPreamble+`
var listId = %s;
if (!wl) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist API unavailable"});
var ok = false;
if (typeof wl.deleteWatchlist === "function") {
  try { await wl.deleteWatchlist(listId); ok = true; } catch(_){}
}
if (!ok && typeof wl.removeSymbolList === "function") {
  try { await wl.removeSymbolList(listId); ok = true; } catch(_){}
}
if (!ok) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"deleteWatchlist unavailable"});
return JSON.stringify({ok:true,data:{status:"deleted"}});
`, jsString(id)))
}

func jsAddWatchlistSymbols(id string, symbols []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistPreamble+`
var listId = %s;
var syms = %s;
if (!wl) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist API unavailable"});
for (var i = 0; i < syms.length; i++) {
  var added = false;
  if (typeof wl.addSymbol === "function") {
    try { await wl.addSymbol(listId, syms[i]); added = true; } catch(_){}
  }
  if (!added && typeof wl.addSymbolToList === "function") {
    try { await wl.addSymbolToList(listId, syms[i]); added = true; } catch(_){}
  }
  if (!added) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"addSymbol unavailable"});
}
var raw = null;
if (typeof wl.getWatchlistById === "function") {
  try { raw = await wl.getWatchlistById(listId); } catch(_){}
}
if (!raw && typeof wl.getSymbolList === "function") {
  try { raw = await wl.getSymbolList(listId); } catch(_){}
}
var result = [];
if (raw && raw.symbols && Array.isArray(raw.symbols)) {
  for (var j = 0; j < raw.symbols.length; j++) {
    var s = raw.symbols[j];
    result.push(typeof s === "string" ? s : String(s.symbol || s.name || s));
  }
}
return JSON.stringify({ok:true,data:{
  id: String(raw && (raw.id || raw.listId) || listId),
  name: String(raw && (raw.name || raw.title) || ""),
  type: String(raw && raw.type || ""),
  symbols: result
}});
`, jsString(id), jsJSON(symbols)))
}

func jsRemoveWatchlistSymbols(id string, symbols []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistPreamble+`
var listId = %s;
var syms = %s;
if (!wl) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist API unavailable"});
for (var i = 0; i < syms.length; i++) {
  var removed = false;
  if (typeof wl.removeSymbol === "function") {
    try { await wl.removeSymbol(listId, syms[i]); removed = true; } catch(_){}
  }
  if (!removed && typeof wl.removeSymbolFromList === "function") {
    try { await wl.removeSymbolFromList(listId, syms[i]); removed = true; } catch(_){}
  }
  if (!removed) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"removeSymbol unavailable"});
}
var raw = null;
if (typeof wl.getWatchlistById === "function") {
  try { raw = await wl.getWatchlistById(listId); } catch(_){}
}
if (!raw && typeof wl.getSymbolList === "function") {
  try { raw = await wl.getSymbolList(listId); } catch(_){}
}
var result = [];
if (raw && raw.symbols && Array.isArray(raw.symbols)) {
  for (var j = 0; j < raw.symbols.length; j++) {
    var s = raw.symbols[j];
    result.push(typeof s === "string" ? s : String(s.symbol || s.name || s));
  }
}
return JSON.stringify({ok:true,data:{
  id: String(raw && (raw.id || raw.listId) || listId),
  name: String(raw && (raw.name || raw.title) || ""),
  type: String(raw && raw.type || ""),
  symbols: result
}});
`, jsString(id), jsJSON(symbols)))
}

func jsFlagSymbol(id, symbol string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistPreamble+`
var listId = %s;
var sym = %s;
if (!wl) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"watchlist API unavailable"});
var ok = false;
if (typeof wl.flagSymbol === "function") {
  try { await wl.flagSymbol(listId, sym); ok = true; } catch(_){}
}
if (!ok && typeof wl.toggleFlagSymbol === "function") {
  try { await wl.toggleFlagSymbol(listId, sym); ok = true; } catch(_){}
}
if (!ok) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"flagSymbol unavailable"});
return JSON.stringify({ok:true,data:{status:"toggled"}});
`, jsString(id), jsString(symbol)))
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
