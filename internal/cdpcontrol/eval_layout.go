package cdpcontrol

import "fmt"

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

// jsLayoutMatchHelper provides _findCurrentLayout(list, currentId) to match by url, short_url, or id.
// Also provides _readFavorite(svc, entry) which checks the live favorites map first, then falls back
// to the static entry.favorite boolean. The favorites map is the source of truth because
// _handleFavorite() updates the map but doesn't mutate the chartList entry snapshots.

func jsGetLayoutFavorite() string {
	return wrapJSEvalAsync(jsPreamble + jsLayoutMatchHelper + `
if (!api || !api._loadChartService) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"_loadChartService unavailable"});
var svc = api._loadChartService;
if (typeof svc.refreshChartList === "function") {
  try { var r = svc.refreshChartList(); if (r && typeof r.then === "function") await r; } catch(_) {}
}
var st = typeof svc.state === "function" ? svc.state() : null;
var stVal = st && typeof st.value === "function" ? st.value() : st;
if (!stVal || !stVal.chartList) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chartList not available"});
var currentId = typeof api.layoutId === "function" ? String(api.layoutId() || "") : "";
if (!currentId) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"cannot determine current layout"});
var found = _findCurrentLayout(stVal.chartList, currentId);
if (!found) return JSON.stringify({ok:false,error_code:"NOT_FOUND",error_message:"current layout not found in chartList (id=" + currentId + ")"});
var isFav = _readFavorite(svc, found);
return JSON.stringify({ok:true,data:{layout_id:currentId,layout_name:String(found.name||""),is_favorite:isFav}});
`)
}

func jsToggleLayoutFavorite() string {
	return wrapJSEvalAsync(jsPreamble + jsLayoutMatchHelper + `
if (!api || !api._loadChartService) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"_loadChartService unavailable"});
var svc = api._loadChartService;
if (typeof svc.refreshChartList === "function") {
  try { var r = svc.refreshChartList(); if (r && typeof r.then === "function") await r; } catch(_) {}
}
var st = typeof svc.state === "function" ? svc.state() : null;
var stVal = st && typeof st.value === "function" ? st.value() : st;
if (!stVal || !stVal.chartList) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"chartList not available"});
var currentId = typeof api.layoutId === "function" ? String(api.layoutId() || "") : "";
if (!currentId) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"cannot determine current layout"});
var found = _findCurrentLayout(stVal.chartList, currentId);
if (!found) return JSON.stringify({ok:false,error_code:"NOT_FOUND",error_message:"current layout not found in chartList (id=" + currentId + ")"});
var beforeFav = _readFavorite(svc, found);
if (typeof found.favoriteAction !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"favoriteAction not available on chart entry"});
}
try {
  var result = found.favoriteAction();
  if (result && typeof result.then === "function") await result;
} catch(e) {
  return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"favoriteAction failed: " + String(e && e.message || e)});
}
var afterFav = _readFavorite(svc, found);
return JSON.stringify({ok:true,data:{layout_id:currentId,layout_name:String(found.name||""),is_favorite:afterFav,was_favorite:beforeFav}});
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

func jsGetPaneInfo() string {
	return wrapJSEval(jsPreamble + `
if (!api) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"TradingViewApi unavailable"});
var gridTemplate = typeof api.layout === "function" ? String(api.layout() || "s") : "s";
var chartCount = typeof api.chartsCount === "function" ? Number(api.chartsCount() || 1) : 1;
var activeIndex = typeof api.activeChartIndex === "function" ? Number(api.activeChartIndex() || 0) : 0;
var panes = [];
if (api._chartWidgetCollection && typeof api._chartWidgetCollection.images === "function") {
  var imgs = api._chartWidgetCollection.images();
  if (imgs && Array.isArray(imgs.charts)) {
    for (var i = 0; i < imgs.charts.length; i++) {
      var c = imgs.charts[i] || {};
      var cm = c.meta || {};
      panes.push({
        index: i,
        symbol: String(cm.symbol || ""),
        exchange: String(cm.exchange || ""),
        resolution: String(cm.resolution || "")
      });
    }
  }
}
if (panes.length === 0 && chart) {
  var sym = typeof chart.symbol === "function" ? String(chart.symbol() || "") : "";
  var res = typeof chart.resolution === "function" ? String(chart.resolution() || "") : "";
  panes.push({index: 0, symbol: sym, exchange: "", resolution: res});
}
return JSON.stringify({ok:true,data:{grid_template:gridTemplate,chart_count:chartCount,active_index:activeIndex,panes:panes}});
`)
}

// --- Indicator Dialog JS functions (DOM-based) ---
// These drive the TradingView Indicators dialog via "/" shortcut, DOM scraping, and CDP events.
