package cdpcontrol

import "fmt"

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

func jsReplayStep(count int) string {
	return wrapJSEval(jsReplayApiPreamble + fmt.Sprintf(`
if (!rapi) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"replay API unavailable"});
if (typeof rapi.doStep !== "function") return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"doStep unavailable"});
var n = %d;
for (var i = 0; i < n; i++) rapi.doStep();
return JSON.stringify({ok:true,data:{status:"stepped",count:n}});
`, count))
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
