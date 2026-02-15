package cdpcontrol

import "fmt"

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
