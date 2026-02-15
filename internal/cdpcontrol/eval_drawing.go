package cdpcontrol

import "fmt"

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
