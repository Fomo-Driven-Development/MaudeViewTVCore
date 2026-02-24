package cdpcontrol

import "fmt"

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

// jsBacktestingWVHelper aliases jsWatchedValueHelper for backward compatibility.

func jsScanBacktestingAccess() string {
	return wrapJSEvalAsync(jsPreamble + `
var results = {};

// 1. Direct checks on api and chart
results.api_direct   = !!(api && api._backtestingStrategyApi);
results.chart_direct = !!(chart && chart._backtestingStrategyApi);

// 2. Scan api object for any "backtest"/"strategy" properties
var apiKeys = [];
if (api) {
  var ak = [];
  try { ak = Object.getOwnPropertyNames(Object.getPrototypeOf(api)).concat(Object.keys(api)); } catch(_) { ak = Object.keys(api); }
  for (var i = 0; i < ak.length; i++) {
    var kl = ak[i].toLowerCase();
    if (kl.indexOf("backtest") !== -1 || kl.indexOf("strategy") !== -1) {
      var t = "unknown"; try { t = typeof api[ak[i]]; } catch(_) {}
      apiKeys.push({key: ak[i], type: t, truthy: !!(api[ak[i]])});
    }
  }
}
results.api_backtest_keys = apiKeys;

// 3. Same scan on chart
var chartKeys = [];
if (chart) {
  var ck = [];
  try { ck = Object.getOwnPropertyNames(Object.getPrototypeOf(chart)).concat(Object.keys(chart)); } catch(_) { ck = Object.keys(chart); }
  for (var j = 0; j < ck.length; j++) {
    var cl = ck[j].toLowerCase();
    if (cl.indexOf("backtest") !== -1 || cl.indexOf("strategy") !== -1) {
      var ct = "unknown"; try { ct = typeof chart[ck[j]]; } catch(_) {}
      chartKeys.push({key: ck[j], type: ct, truthy: !!(chart[ck[j]])});
    }
  }
}
results.chart_backtest_keys = chartKeys;

// 4. Webpack module cache scan
var wpRequire = window.__tvAgentWpRequire || null;
if (!wpRequire) {
  var wkeys = Object.getOwnPropertyNames(window);
  for (var wi = 0; wi < wkeys.length; wi++) {
    try {
      var wv = window[wkeys[wi]];
      if (Array.isArray(wv) && wv.push !== Array.prototype.push) {
        wv.push([["__bsa_" + Date.now()], {}, function(r) { wpRequire = r; }]);
        if (wpRequire) { window.__tvAgentWpRequire = wpRequire; break; }
      }
    } catch(_) {}
  }
}
var cacheHits = [];
if (wpRequire && wpRequire.c) {
  var mc = wpRequire.c; var mkeys = Object.keys(mc);
  results.module_cache_size = mkeys.length;
  for (var mi = 0; mi < mkeys.length; mi++) {
    try {
      var exp = mc[mkeys[mi]].exports;
      if (!exp || typeof exp !== "object") continue;
      var ekeys = Object.keys(exp);
      for (var ei = 0; ei < ekeys.length; ei++) {
        var ekl = ekeys[ei].toLowerCase();
        if (ekl.indexOf("backtest") !== -1 || (ekl.indexOf("strategy") !== -1 && ekl.indexOf("api") !== -1)) {
          cacheHits.push({moduleId: mkeys[mi], key: ekeys[ei], type: typeof exp[ekeys[ei]]});
        }
      }
    } catch(_) {}
  }
}
results.cache_hits = cacheHits;
return JSON.stringify({ok:true,data:results});
`)
}

func jsProbeBacktestingApi() string {
	return wrapJSEvalAsync(jsBacktestingApiPreamble + jsProbeObjectHelper + `
var r = _probeObj(bsa, ["api.backtestingStrategyApi()"]);
return JSON.stringify({ok:true,data:r});
`)
}

func jsListStrategies() string {
	return wrapJSEvalAsync(jsBacktestingApiPreamble + jsBacktestingWVHelper + `
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
	return wrapJSEvalAsync(jsBacktestingApiPreamble + jsBacktestingWVHelper + `
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
	return wrapJSEvalAsync(fmt.Sprintf(jsBacktestingApiPreamble+`
var id = %s;
if (!bsa) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"backtesting API unavailable"});
if (typeof bsa.setActiveStrategy !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setActiveStrategy unavailable"});
bsa.setActiveStrategy(id);
return JSON.stringify({ok:true,data:{status:"set",strategy_id:id}});
`, jsString(strategyID)))
}

func jsSetStrategyInput(name string, value any) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsBacktestingApiPreamble+`
var name = %s;
var value = %s;
if (!bsa) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"backtesting API unavailable"});
if (typeof bsa.setStrategyInput !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"setStrategyInput unavailable"});
bsa.setStrategyInput(name, value);
return JSON.stringify({ok:true,data:{status:"set",name:name,value:value}});
`, jsString(name), jsJSON(value)))
}

func jsGetStrategyReport() string {
	return wrapJSEvalAsync(jsBacktestingApiPreamble + jsBacktestingWVHelper + `
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
	return wrapJSEvalAsync(jsBacktestingApiPreamble + `
if (!bsa) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"backtesting API unavailable"});
if (typeof bsa.getChartDateRange !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"getChartDateRange unavailable"});
var range = bsa.getChartDateRange();
return JSON.stringify({ok:true,data:{date_range:range}});
`)
}

func jsStrategyGotoDate(timestamp float64, belowBar bool) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsBacktestingApiPreamble+`
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

func jsProbeIndicatorDialogDOM() string {
	return wrapJSEvalAsync(`
// Wait for search results to render — poll for content to appear
var dlg = document.querySelector('[data-name="indicators-dialog"]');
var contentArea = dlg ? dlg.querySelector('[data-role="dialog-content"]') : null;
var pollDeadline = Date.now() + 3000;
while (Date.now() < pollDeadline) {
  if (contentArea && contentArea.children.length > 0) break;
  if (contentArea && (contentArea.textContent || '').trim().length > 0) break;
  await new Promise(function(r) { setTimeout(r, 300); });
}
var result = {search_input:null, search_value:null, dialog_container:null, result_items:[], content_children:[], deep_scan:[], sidebar:[], all_visible_data_names:[], close_buttons:[]};

// Find the search input — the main landmark
var inputs = document.querySelectorAll('input');
for (var i = 0; i < inputs.length; i++) {
  var inp = inputs[i];
  if (inp.offsetParent !== null) {
    var ph = (inp.placeholder || '').toLowerCase();
    if (ph.indexOf('search') !== -1) {
      result.search_input = {
        placeholder: inp.placeholder,
        value: inp.value || '',
        class: (inp.className || '').substring(0, 200),
        id: inp.id || '',
        parent_class: (inp.parentElement ? inp.parentElement.className : '').substring(0, 200),
        grandparent_class: (inp.parentElement && inp.parentElement.parentElement ? inp.parentElement.parentElement.className : '').substring(0, 200)
      };
    }
  }
}

// Walk up from the search input to find the dialog container
if (result.search_input) {
  var el = null;
  var allInputs = document.querySelectorAll('input');
  for (var ii = 0; ii < allInputs.length; ii++) {
    if (allInputs[ii].offsetParent && (allInputs[ii].placeholder || '').toLowerCase().indexOf('search') !== -1) {
      el = allInputs[ii]; break;
    }
  }
  if (el) {
    for (var up = 0; up < 15 && el; up++) {
      el = el.parentElement;
      if (!el) break;
      var cls = el.className || '';
      var w = el.offsetWidth || 0;
      var h = el.offsetHeight || 0;
      if (w > 400 && h > 300) {
        result.dialog_container = {
          tag: el.tagName,
          class: cls.substring(0, 300),
          data_name: el.getAttribute('data-name') || '',
          data_dialog_name: el.getAttribute('data-dialog-name') || '',
          role: el.getAttribute('role') || '',
          width: w, height: h, depth: up
        };
        // Scan children with data-role
        var allChildren = el.querySelectorAll('*');
        var resultCount = 0;
        for (var ci = 0; ci < allChildren.length && resultCount < 10; ci++) {
          var child = allChildren[ci];
          if (!child.offsetParent) continue;
          var ccls = child.className || '';
          var dataRole = child.getAttribute('data-role') || '';
          if (dataRole) {
            result.result_items.push({tag: child.tagName, class: ccls.substring(0, 150), data_role: dataRole, text: (child.textContent || '').trim().substring(0, 80)});
            resultCount++;
          }
        }
        // Look for elements with "title" in class
        var titleEls = el.querySelectorAll('[class*="title"]');
        for (var ti = 0; ti < titleEls.length && ti < 5; ti++) {
          var te = titleEls[ti];
          if (te.offsetParent) {
            result.result_items.push({tag: te.tagName, class: (te.className || '').substring(0, 150), text: (te.textContent || '').trim().substring(0, 80), type: 'title-class'});
          }
        }
        // Look for tab/category elements
        var tabs = el.querySelectorAll('[role="tab"], [data-value], [class*="tab"]');
        for (var ti2 = 0; ti2 < tabs.length; ti2++) {
          var tab = tabs[ti2];
          if (tab.offsetParent) {
            result.sidebar.push({tag: tab.tagName, class: (tab.className || '').substring(0, 150), text: (tab.textContent || '').trim().substring(0, 50), role: tab.getAttribute('role') || '', data_value: tab.getAttribute('data-value') || ''});
          }
        }
        // Inspect dialog-content area children
        var contentArea = el.querySelector('[data-role="dialog-content"]');
        if (contentArea) {
          var cc = contentArea.children;
          for (var cci = 0; cci < cc.length && cci < 15; cci++) {
            var cchild = cc[cci];
            result.content_children.push({
              tag: cchild.tagName,
              class: (cchild.className || '').substring(0, 200),
              child_count: cchild.children ? cchild.children.length : 0,
              text: (cchild.textContent || '').trim().substring(0, 100),
              data_name: cchild.getAttribute('data-name') || '',
              visible: !!(cchild.offsetParent),
              height: cchild.offsetHeight || 0
            });
          }
        }
        break;
      }
    }
  }
}

// Deep scan: list ALL visible elements inside the dialog that have text content
if (dlg) {
  var allDesc = dlg.querySelectorAll('*');
  for (var di = 0; di < allDesc.length && result.deep_scan.length < 25; di++) {
    var desc = allDesc[di];
    if (!desc.offsetParent) continue;
    // Only include leaf-ish elements (not containers with tons of nested text)
    var ownText = '';
    for (var ni = 0; ni < desc.childNodes.length; ni++) {
      if (desc.childNodes[ni].nodeType === 3) ownText += desc.childNodes[ni].textContent;
    }
    ownText = ownText.trim();
    if (ownText.length > 0 && ownText.length < 100) {
      result.deep_scan.push({
        tag: desc.tagName,
        class: (desc.className || '').substring(0, 150),
        text: ownText,
        parent_class: (desc.parentElement ? desc.parentElement.className : '').substring(0, 100)
      });
    }
  }
}

// Collect all visible data-name attributes (for reference)
var dataNameEls = document.querySelectorAll('[data-name]');
for (var i = 0; i < dataNameEls.length; i++) {
  var dn = dataNameEls[i];
  if (dn.offsetParent !== null) {
    var name = dn.getAttribute('data-name');
    if (name && result.all_visible_data_names.indexOf(name) === -1 && result.all_visible_data_names.length < 30) {
      result.all_visible_data_names.push(name);
    }
  }
}

return JSON.stringify({ok:true,data:result});
`)
}

func jsWaitForIndicatorDialog() string {
	return wrapJSEvalAsync(`
var deadline = Date.now() + 5000;
var found = false;
var ix = 0, iy = 0;
while (Date.now() < deadline) {
  var inp = document.getElementById('indicators-dialog-search-input');
  if (!inp) {
    var dlg = document.querySelector('[data-name="indicators-dialog"]');
    if (dlg && dlg.offsetParent !== null) inp = dlg.querySelector('input');
  }
  if (inp && inp.offsetParent !== null) {
    found = true;
    var rect = inp.getBoundingClientRect();
    ix = rect.x + rect.width / 2;
    iy = rect.y + rect.height / 2;
    break;
  }
  await new Promise(function(r) { setTimeout(r, 200); });
}
return JSON.stringify({ok:true,data:{dialog_found:found,input_x:ix,input_y:iy}});
`)
}

// jsSetIndicatorSearchValue types the search query using document.execCommand
// which creates a trusted text insertion that fires all native events (beforeinput,
// input) and is handled correctly by React's controlled components.

func jsSetIndicatorSearchValue(query string) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
var query = %s;
var inp = document.getElementById('indicators-dialog-search-input');
if (!inp) {
  var dlg = document.querySelector('[data-name="indicators-dialog"]');
  if (dlg) inp = dlg.querySelector('input');
}
if (!inp) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"search input not found"});
inp.focus();
inp.select();
document.execCommand('insertText', false, query);
await new Promise(function(r) { setTimeout(r, 800); });
return JSON.stringify({ok:true,data:{status:"typed",value:inp.value}});
`, jsString(query)))
}

func jsScrapeIndicatorResults() string {
	return wrapJSEvalAsync(`
// Wait for search results to render inside the dialog
var dlg = document.querySelector('[data-name="indicators-dialog"]');
if (!dlg) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"indicators dialog not found"});
var contentArea = dlg.querySelector('[data-role="dialog-content"]');
var pollEnd = Date.now() + 3000;
while (Date.now() < pollEnd) {
  if (contentArea && contentArea.querySelector('[data-role="list-item"]')) break;
  await new Promise(function(r) { setTimeout(r, 300); });
}
var results = [];
var rows = dlg.querySelectorAll('[data-role="list-item"]');
for (var i = 0; i < rows.length && results.length < 50; i++) {
  var row = rows[i];
  if (!row.offsetParent) continue;
  // Extract indicator name from the title element
  var nameEl = row.querySelector('[class*="title-"]');
  var name = nameEl ? (nameEl.textContent || '').trim() : '';
  if (!name) name = (row.textContent || '').trim().split('\n')[0].trim();
  if (!name) continue;
  // Extract author
  var authorEl = row.querySelector('[class*="author"], [class*="user-"]');
  var author = authorEl ? (authorEl.textContent || '').replace(/^by\s*/i, '').trim() : '';
  // Extract boosts count
  var boostEl = row.querySelector('[class*="boosts"], [class*="likes"], [class*="count-"]');
  var boosts = 0;
  if (boostEl) {
    var btext = (boostEl.textContent || '').replace(/[^0-9.kKmM]/g, '');
    var num = parseFloat(btext);
    if (!isNaN(num)) {
      if (btext.indexOf('K') !== -1 || btext.indexOf('k') !== -1) num *= 1000;
      else if (btext.indexOf('M') !== -1 || btext.indexOf('m') !== -1) num *= 1000000;
      boosts = Math.round(num);
    }
  }
  // Check favorite star state
  var starEl = row.querySelector('[class*="star"], [class*="favorite"], [class*="fav"]');
  var isFav = false;
  if (starEl) {
    var starCls = starEl.className || '';
    isFav = starCls.indexOf('active') !== -1 || starCls.indexOf('filled') !== -1 || starCls.indexOf('checked') !== -1
         || starEl.getAttribute('aria-checked') === 'true' || starEl.getAttribute('aria-pressed') === 'true'
         || starEl.getAttribute('data-active') === 'true';
  }
  results.push({name: name, author: author, boosts: boosts, is_favorite: isFav, index: results.length});
}
return JSON.stringify({ok:true,data:{results:results,total_count:results.length}});
`)
}

func jsClickIndicatorResult(index int) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
var targetIndex = %d;
var dlg = document.querySelector('[data-name="indicators-dialog"]');
if (!dlg) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"indicators dialog not found"});
// Wait for results to load
var pollEnd = Date.now() + 3000;
while (Date.now() < pollEnd) {
  if (dlg.querySelector('[data-role="list-item"]')) break;
  await new Promise(function(r) { setTimeout(r, 300); });
}
var rows = dlg.querySelectorAll('[data-role="list-item"]');
var visible = [];
for (var i = 0; i < rows.length; i++) {
  if (rows[i].offsetParent) visible.push(rows[i]);
}
if (targetIndex < 0 || targetIndex >= visible.length) {
  return JSON.stringify({ok:false,error_code:"VALIDATION",error_message:"index " + targetIndex + " out of range, found " + visible.length + " results"});
}
var row = visible[targetIndex];
var nameEl = row.querySelector('[class*="title-"]');
var name = nameEl ? (nameEl.textContent || '').trim() : (row.textContent || '').trim().split('\n')[0].trim();
row.click();
await new Promise(function(r) { setTimeout(r, 500); });
return JSON.stringify({ok:true,data:{status:"added",index:targetIndex,name:name}});
`, index))
}

func jsClickIndicatorCategory(category string) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
var target = %s.toLowerCase();
var dlg = document.querySelector('[data-name="indicators-dialog"]');
if (!dlg) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"indicators dialog not found"});
var found = false;
// Sidebar items have class "sidebarItem-*" with "title-*" inside
var items = dlg.querySelectorAll('[class*="sidebarItem-"]');
for (var i = 0; i < items.length; i++) {
  var item = items[i];
  if (!item.offsetParent) continue;
  var txt = (item.textContent || '').trim().toLowerCase();
  if (txt === target || txt.indexOf(target) === 0) {
    item.click();
    found = true;
    break;
  }
}
if (!found) {
  return JSON.stringify({ok:false,error_code:"VALIDATION",error_message:"category not found: " + target});
}
await new Promise(function(r) { setTimeout(r, 500); });
return JSON.stringify({ok:true,data:{status:"navigated",category:target}});
`, jsString(category)))
}

func jsLocateIndicatorFavoriteStar(index int) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
var targetIndex = %d;
var dlg = document.querySelector('[data-name="indicators-dialog"]');
if (!dlg) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"indicators dialog not found"});
var pollEnd = Date.now() + 3000;
while (Date.now() < pollEnd) {
  if (dlg.querySelector('[data-role="list-item"]')) break;
  await new Promise(function(r) { setTimeout(r, 300); });
}
var rows = dlg.querySelectorAll('[data-role="list-item"]');
var visible = [];
for (var i = 0; i < rows.length; i++) {
  if (rows[i].offsetParent) visible.push(rows[i]);
}
if (targetIndex < 0 || targetIndex >= visible.length) {
  return JSON.stringify({ok:false,error_code:"VALIDATION",error_message:"index " + targetIndex + " out of range, found " + visible.length + " results"});
}
var row = visible[targetIndex];
var nameEl = row.querySelector('[class*="title-"]');
var name = nameEl ? (nameEl.textContent || '').trim() : (row.textContent || '').trim().split('\n')[0].trim();
// Find the star/favorite button — look for buttons with star/fav in class or aria-label
var starEl = row.querySelector('[class*="star"], [class*="favorite"], [class*="fav"], button[class*="star"], button[class*="fav"]');
if (!starEl) {
  var btns = row.querySelectorAll('button, [role="button"]');
  for (var bi = 0; bi < btns.length; bi++) {
    var b = btns[bi];
    var cls = (b.className || '').toLowerCase();
    var ariaLabel = (b.getAttribute('aria-label') || '').toLowerCase();
    if (cls.indexOf('star') !== -1 || cls.indexOf('fav') !== -1 || ariaLabel.indexOf('fav') !== -1 || ariaLabel.indexOf('star') !== -1) {
      starEl = b; break;
    }
  }
}
if (!starEl) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"favorite star button not found on row " + targetIndex});
}
var rect = starEl.getBoundingClientRect();
var x = rect.x + rect.width / 2;
var y = rect.y + rect.height / 2;
var starCls = starEl.className || '';
var isFav = starCls.indexOf('active') !== -1 || starCls.indexOf('filled') !== -1 || starCls.indexOf('checked') !== -1
         || starEl.getAttribute('aria-checked') === 'true' || starEl.getAttribute('aria-pressed') === 'true';
return JSON.stringify({ok:true,data:{name:name,index:targetIndex,is_favorite:isFav,x:x,y:y}});
`, index))
}

func jsWaitForIndicatorDialogClosed() string {
	return wrapJSEvalAsync(`
var deadline = Date.now() + 3000;
while (Date.now() < deadline) {
  var dlg = document.querySelector('[data-name="indicators-dialog"]');
  if (!dlg || dlg.offsetParent === null) break;
  await new Promise(function(r) { setTimeout(r, 200); });
}
return JSON.stringify({ok:true,data:{status:"dismissed"}});
`)
}

func jsCheckIndicatorFavoriteState(index int) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
var targetIndex = %d;
await new Promise(function(r) { setTimeout(r, 500); });
var dlg = document.querySelector('[data-name="indicators-dialog"]');
if (!dlg) return JSON.stringify({ok:true,data:{name:"",is_favorite:false}});
var rows = dlg.querySelectorAll('[data-role="list-item"]');
var visible = [];
for (var i = 0; i < rows.length; i++) {
  if (rows[i].offsetParent) visible.push(rows[i]);
}
if (targetIndex < 0 || targetIndex >= visible.length) {
  return JSON.stringify({ok:true,data:{name:"",is_favorite:false}});
}
var row = visible[targetIndex];
var nameEl = row.querySelector('[class*="title-"]');
var name = nameEl ? (nameEl.textContent || '').trim() : (row.textContent || '').trim().split('\n')[0].trim();
var starEl = row.querySelector('[class*="star"], [class*="favorite"], [class*="fav"]');
var isFav = false;
if (starEl) {
  var starCls = starEl.className || '';
  isFav = starCls.indexOf('active') !== -1 || starCls.indexOf('filled') !== -1 || starCls.indexOf('checked') !== -1
       || starEl.getAttribute('aria-checked') === 'true' || starEl.getAttribute('aria-pressed') === 'true';
}
return JSON.stringify({ok:true,data:{name:name,is_favorite:isFav}});
`, index))
}

// --- Currency / Unit JS eval functions ---

func jsListStudyTemplates() string {
	return wrapJSEvalAsync(jsWatchlistFetch + `
var raw = await _wlFetch("/api/v1/study-templates");
var custom = [];
var standard = [];
var fundamentals = [];
if (raw.custom && Array.isArray(raw.custom)) {
  for (var i = 0; i < raw.custom.length; i++) {
    var c = raw.custom[i];
    custom.push({id:Number(c.id||0),name:String(c.name||""),meta_info:c.meta_info||null,favorite_date:String(c.favorite_date||"")});
  }
}
if (raw.standard && Array.isArray(raw.standard)) {
  for (var i = 0; i < raw.standard.length; i++) {
    var s = raw.standard[i];
    standard.push({id:Number(s.id||0),name:String(s.name||""),meta_info:s.meta_info||null,favorite_date:String(s.favorite_date||"")});
  }
}
if (raw.fundamentals && Array.isArray(raw.fundamentals)) {
  for (var i = 0; i < raw.fundamentals.length; i++) {
    var f = raw.fundamentals[i];
    fundamentals.push({id:Number(f.id||0),name:String(f.name||""),meta_info:f.meta_info||null,favorite_date:String(f.favorite_date||"")});
  }
}
return JSON.stringify({ok:true,data:{custom:custom,standard:standard,fundamentals:fundamentals}});
`)
}

func jsGetStudyTemplate(id int) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+`
var tid = %d;
var raw = await _wlFetch("/api/v1/study-templates/" + tid);
return JSON.stringify({ok:true,data:{id:Number(raw.id||tid),name:String(raw.name||""),meta_info:raw.meta_info||null,favorite_date:String(raw.favorite_date||"")}});
`, id))
}

func jsApplyStudyTemplateByName(name string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+jsPreamble+`
var targetName = %s;
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"TradingView API unavailable"});
var raw = await _wlFetch("/api/v1/study-templates");
var match = null;
var lower = targetName.toLowerCase();
var all = (raw.custom || []).concat(raw.standard || []).concat(raw.fundamentals || []);
for (var i = 0; i < all.length; i++) {
  if ((all[i].name || "").toLowerCase() === lower) { match = all[i]; break; }
}
if (!match) return JSON.stringify({ok:false,error_code:"VALIDATION",error_message:"template not found: " + targetName});
var detail = await _wlFetch("/api/v1/study-templates/" + match.id);
var chart = api.activeChart();
if (!chart || typeof chart.applyStudyTemplate !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"applyStudyTemplate unavailable"});
}
var tplData = typeof detail.content === "string" ? JSON.parse(detail.content) : detail.content;
if (!tplData) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"template content is empty"});
chart.applyStudyTemplate(tplData);
return JSON.stringify({ok:true,data:{id:Number(match.id),name:String(match.name),status:"applied"}});
`, jsString(name)))
}

// jsHotlistsPreamble extends jsPreamble with hotlistsManager() resolution.
// The hotlists manager is a webpack-internal singleton — accessed via webpack
// require extraction + module cache scan. No lazy-load trigger needed.
