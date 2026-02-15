package cdpcontrol

import "fmt"

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

func jsGetCurrency() string {
	return wrapJSEval(jsPreamble + `
var cw = chart._chartWidget;
var ms = cw.model().mainSeries();
var cur = ms.currency ? ms.currency() : null;
var isConverted = typeof ms.isConvertedToOtherCurrency === "function" ? ms.isConvertedToOtherCurrency() : false;
var native = "";
try { var si = ms.symbolInfo; if (si) native = si.currency_code || si.currency || ""; } catch(_){}
return JSON.stringify({ok:true,data:{currency:cur||"",is_converted:!!isConverted,native_currency:native}});
`)
}

func jsSetCurrency(code string) string {
	if code == "null" {
		return wrapJSEval(jsPreamble + `
var cw = chart._chartWidget;
var ms = cw.model().mainSeries();
ms.setCurrency(null);
return JSON.stringify({ok:true,data:{status:"ok"}});
`)
	}
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var cw = chart._chartWidget;
var ms = cw.model().mainSeries();
ms.setCurrency(%s);
return JSON.stringify({ok:true,data:{status:"ok"}});
`, jsString(code)))
}

func jsGetAvailableCurrencies() string {
	return wrapJSEvalAsync(jsPreamble + `
var list = await api._chartApiInstance.availableCurrencies();
var result = [];
for (var i = 0; i < list.length; i++) {
  var c = list[i];
  result.push({code:c.code||"",description:c.description||"",id:c.id||"",kind:c.kind||""});
}
return JSON.stringify({ok:true,data:{currencies:result}});
`)
}

func jsGetUnit() string {
	return wrapJSEval(jsPreamble + `
var cw = chart._chartWidget;
var ms = cw.model().mainSeries();
var u = ms.unit ? ms.unit() : null;
var isConverted = typeof ms.isConvertedToOtherUnit === "function" ? ms.isConvertedToOtherUnit() : false;
return JSON.stringify({ok:true,data:{unit:u||"",is_converted:!!isConverted}});
`)
}

func jsSetUnit(id string) string {
	if id == "null" {
		return wrapJSEval(jsPreamble + `
var cw = chart._chartWidget;
var ms = cw.model().mainSeries();
ms.setUnit(null);
return JSON.stringify({ok:true,data:{status:"ok"}});
`)
	}
	return wrapJSEval(fmt.Sprintf(jsPreamble+`
var cw = chart._chartWidget;
var ms = cw.model().mainSeries();
ms.setUnit(%s);
return JSON.stringify({ok:true,data:{status:"ok"}});
`, jsString(id)))
}

func jsGetAvailableUnits() string {
	return wrapJSEvalAsync(jsPreamble + `
var cats = await api._chartApiInstance.availableUnits();
var result = [];
var keys = Object.keys(cats);
for (var i = 0; i < keys.length; i++) {
  var catKey = keys[i];
  var items = cats[catKey];
  if (!Array.isArray(items)) continue;
  for (var j = 0; j < items.length; j++) {
    var u = items[j];
    result.push({id:u.id||"",name:u.name||"",description:u.description||"",type:catKey});
  }
}
return JSON.stringify({ok:true,data:{units:result}});
`)
}

// --- Colored Watchlist JS functions ---

func jsProbeHotlistsManager() string {
	return wrapJSEvalAsync(jsHotlistsPreamble + jsProbeObjectHelper + `
var r = _probeObj(hmgr, ["webpack:hotlistsManager()"]);
return JSON.stringify({ok:true,data:r});
`)
}

func jsProbeHotlistsManagerDeep() string {
	return wrapJSEvalAsync(jsHotlistsPreamble + `
if (!hmgr) return JSON.stringify({ok:true,data:{found:false}});
var r = {found:true, methods:{}, properties:{}};
// Enumerate methods with arity
var p = hmgr;
while (p && p !== Object.prototype) {
  var mk = Object.getOwnPropertyNames(p);
  for (var mi = 0; mi < mk.length; mi++) {
    var mn = mk[mi];
    if (mn === "constructor") continue;
    try {
      if (typeof hmgr[mn] === "function") {
        r.methods[mn] = {arity: hmgr[mn].length};
      }
    } catch(_) {}
  }
  p = Object.getPrototypeOf(p);
}
// Enumerate data properties
var ownKeys = Object.keys(hmgr);
for (var oi = 0; oi < ownKeys.length; oi++) {
  var k = ownKeys[oi];
  var v = hmgr[k];
  if (typeof v === "function") continue;
  if (typeof v === "string" || typeof v === "number" || typeof v === "boolean") {
    r.properties[k] = {type: typeof v, value: v};
  } else if (v === null || v === undefined) {
    r.properties[k] = {type: "null", value: null};
  } else if (Array.isArray(v)) {
    r.properties[k] = {type: "array", length: v.length, sample: v.slice(0, 5)};
  } else if (typeof v === "object") {
    var objKeys = Object.keys(v);
    r.properties[k] = {type: "object", keys: objKeys.length, sample_keys: objKeys.slice(0, 10)};
  }
}
// Try known accessors
try {
  var ae = typeof hmgr.availableExchanges === "function" ? hmgr.availableExchanges() : hmgr.availableExchanges;
  if (ae) r.available_exchanges_sample = Array.isArray(ae) ? ae.slice(0, 10) : ae;
} catch(_) {}
try {
  var gtm = typeof hmgr.groupsTitlesMap === "function" ? hmgr.groupsTitlesMap() : hmgr.groupsTitlesMap;
  if (gtm) r.groups_titles_map = gtm;
} catch(_) {}
return JSON.stringify({ok:true,data:r});
`)
}

func jsGetHotlistMarkets() string {
	return wrapJSEvalAsync(jsHotlistsPreamble + `
if (!hmgr) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"hotlistsManager not found"});
if (typeof hmgr.getMarkets !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getMarkets not found on hotlistsManager"});
var markets = hmgr.getMarkets();
return JSON.stringify({ok:true,data:{markets:markets}});
`)
}

func jsGetHotlistExchanges() string {
	return wrapJSEvalAsync(jsHotlistsPreamble + `
if (!hmgr) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"hotlistsManager not found"});
var ae = typeof hmgr.availableExchanges === "function" ? hmgr.availableExchanges() : hmgr.availableExchanges;
if (!ae || !Array.isArray(ae)) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"availableExchanges not found"});
var gtm = typeof hmgr.groupsTitlesMap === "function" ? hmgr.groupsTitlesMap() : hmgr.groupsTitlesMap;
var result = [];
for (var i = 0; i < ae.length; i++) {
  var ex = ae[i];
  var entry = {exchange: ex, name: "", full_name: "", flag: "", groups: []};
  try { if (typeof hmgr.getExchangeName === "function") entry.name = hmgr.getExchangeName(ex) || ""; } catch(_) {}
  try { if (typeof hmgr.getExchangeFullName === "function") entry.full_name = hmgr.getExchangeFullName(ex) || ""; } catch(_) {}
  try { if (typeof hmgr.getExchangeFlag === "function") entry.flag = hmgr.getExchangeFlag(ex) || ""; } catch(_) {}
  if (gtm && typeof gtm === "object") {
    var gkeys = Object.keys(gtm);
    for (var gi = 0; gi < gkeys.length; gi++) {
      entry.groups.push({id: gkeys[gi], title: String(gtm[gkeys[gi]] || gkeys[gi])});
    }
  }
  result.push(entry);
}
return JSON.stringify({ok:true,data:{exchanges:result}});
`)
}

func jsGetOneHotlist(exchange, group string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsHotlistsPreamble+`
if (!hmgr) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"hotlistsManager not found"});
var exchange = %s;
var group = %s;
// Validate exchange
var ae = typeof hmgr.availableExchanges === "function" ? hmgr.availableExchanges() : hmgr.availableExchanges;
if (ae && Array.isArray(ae) && ae.indexOf(exchange) === -1) {
  return JSON.stringify({ok:false,error_code:"VALIDATION",error_message:"exchange " + exchange + " not in availableExchanges"});
}
// Try three calling patterns: promise, sync, callback
var items = null;
try {
  var result = hmgr.getOneHotlist(null, exchange, group);
  if (result && typeof result.then === "function") {
    items = await result;
  } else if (Array.isArray(result)) {
    items = result;
  } else if (result && typeof result === "object") {
    items = result;
  }
} catch(e) {
  // Callback pattern fallback
  try {
    items = await new Promise(function(resolve, reject) {
      hmgr.getOneHotlist(function(data) { resolve(data); }, exchange, group);
      setTimeout(function() { reject(new Error("timeout")); }, 5000);
    });
  } catch(e2) {
    return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"getOneHotlist failed: " + String(e) + " / " + String(e2)});
  }
}
if (!items) return JSON.stringify({ok:true,data:{exchange:exchange,group:group,symbols:[]}});
var arr = Array.isArray(items) ? items : (items.data ? items.data : [items]);
var symbols = [];
for (var si = 0; si < arr.length; si++) {
  var item = arr[si];
  if (typeof item === "string") { symbols.push({symbol:item}); continue; }
  if (!item || typeof item !== "object") continue;
  var sym = item.s || item.symbol || item.name || "";
  var extra = {};
  var ikeys = Object.keys(item);
  for (var ik = 0; ik < ikeys.length; ik++) {
    var iky = ikeys[ik];
    if (iky === "s" || iky === "symbol" || iky === "name") continue;
    var iv = item[iky];
    if (typeof iv === "string" || typeof iv === "number" || typeof iv === "boolean" || iv === null) extra[iky] = iv;
  }
  symbols.push({symbol:String(sym),extra:Object.keys(extra).length > 0 ? extra : undefined});
}
return JSON.stringify({ok:true,data:{exchange:exchange,group:group,symbols:symbols}});
`, jsString(exchange), jsString(group)))
}

func jsProbeDataWindow() string {
	return wrapJSEval(jsPreamble + `
if (!chart) return JSON.stringify({ok:false,error_code:"CHART_NOT_FOUND",error_message:"no active chart"});
var r = {panel_visible:false,dom_elements:[],crosshair_methods:[],legend_elements:[],chart_widget_props:[],model_props:[],data_window_state:{}};
// 1. Data window DOM
var dwSelectors = ['[data-name="data-window"]','[class*="dataWindow"]','[class*="legend"]','[class*="valuesWrapper"]','[class*="pane-legend"]'];
for (var di = 0; di < dwSelectors.length; di++) {
  var els = document.querySelectorAll(dwSelectors[di]);
  if (els.length > 0) {
    r.dom_elements.push(dwSelectors[di] + " (" + els.length + ")");
    if (dwSelectors[di].indexOf("dataWindow") >= 0 || dwSelectors[di].indexOf("data-window") >= 0) r.panel_visible = true;
  }
}
// 2. Chart widget properties — look for crosshair/cursor/datawindow/legend
try {
  var cw = chart._chartWidget || (typeof chart.getChartWidget === "function" ? chart.getChartWidget() : null);
  if (cw) {
    var cwp = cw;
    var cwSeen = {};
    while (cwp && cwp !== Object.prototype) {
      var cwKeys = Object.getOwnPropertyNames(cwp);
      for (var ci = 0; ci < cwKeys.length; ci++) {
        var ck = cwKeys[ci].toLowerCase();
        if (cwSeen[cwKeys[ci]]) continue;
        cwSeen[cwKeys[ci]] = true;
        if (ck.indexOf("crosshair") >= 0 || ck.indexOf("cursor") >= 0 || ck.indexOf("datawindow") >= 0 || ck.indexOf("legend") >= 0 || ck.indexOf("data_window") >= 0) {
          r.chart_widget_props.push(cwKeys[ci] + ":" + typeof cw[cwKeys[ci]]);
        }
      }
      cwp = Object.getPrototypeOf(cwp);
    }
  }
} catch(_) {}
// 3. Model properties
try {
  var mdl = typeof chart.model === "function" ? chart.model() : chart._model;
  if (mdl) {
    var mp = mdl;
    var mpSeen = {};
    while (mp && mp !== Object.prototype) {
      var mpKeys = Object.getOwnPropertyNames(mp);
      for (var mi = 0; mi < mpKeys.length; mi++) {
        var mk = mpKeys[mi].toLowerCase();
        if (mpSeen[mpKeys[mi]]) continue;
        mpSeen[mpKeys[mi]] = true;
        if (mk.indexOf("crosshair") >= 0 || mk.indexOf("cursor") >= 0 || mk.indexOf("ohlc") >= 0 || mk.indexOf("bar") >= 0 || mk.indexOf("datawindow") >= 0 || mk.indexOf("data_window") >= 0) {
          r.model_props.push(mpKeys[mi] + ":" + typeof mdl[mpKeys[mi]]);
        }
      }
      mp = Object.getPrototypeOf(mp);
    }
    // 4. Crosshair source
    if (typeof mdl.crosshairSource === "function") {
      try {
        var cs = mdl.crosshairSource();
        if (cs) {
          var csKeys = Object.keys(cs);
          for (var csi = 0; csi < csKeys.length; csi++) {
            r.crosshair_methods.push(csKeys[csi] + ":" + typeof cs[csKeys[csi]]);
          }
        }
      } catch(_) {}
    }
  }
} catch(_) {}
// 5. Legend elements — scan pane-legend DOM
try {
  var legends = document.querySelectorAll('[class*="pane-legend"],[class*="valuesWrapper"]');
  for (var li = 0; li < legends.length && li < 5; li++) {
    var txt = legends[li].textContent || "";
    if (txt.length > 100) txt = txt.substring(0, 100) + "...";
    r.legend_elements.push(txt.trim());
  }
} catch(_) {}
return JSON.stringify({ok:true,data:r});
`)
}
