package cdpcontrol

import "fmt"

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

func jsListColoredWatchlists() string {
	return wrapJSEvalAsync(jsWatchlistFetch + `
var raw = await _wlFetch("/api/v1/symbols_list/colored/");
if (!Array.isArray(raw)) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"colored watchlists returned non-array"});
var lists = [];
for (var i = 0; i < raw.length; i++) {
  var it = raw[i] || {};
  var syms = [];
  if (it.symbols && Array.isArray(it.symbols)) {
    for (var j = 0; j < it.symbols.length; j++) {
      syms.push(typeof it.symbols[j] === "string" ? it.symbols[j] : String(it.symbols[j]));
    }
  }
  lists.push({
    id: String(it.id || ""),
    name: String(it.name || ""),
    color: String(it.color || ""),
    symbols: syms,
    modified: String(it.modified || "")
  });
}
return JSON.stringify({ok:true,data:{colored_watchlists:lists}});
`)
}

// _jsColoredMutationHelper extracts symbols from a colored watchlist mutation response.
// TV may return a bare array of symbols or an object with a .symbols array.

func jsReplaceColoredWatchlist(color string, symbols []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+jsColoredMutationHelper+`
var color = %s;
var syms = %s;
var raw = await _wlFetch("/api/v1/symbols_list/colored/" + encodeURIComponent(color) + "/replace/", {
  method: "POST",
  headers: {"Content-Type": "application/json"},
  body: JSON.stringify(syms)
});
return JSON.stringify({ok:true,data:_coloredResult(raw, color)});
`, jsString(color), jsJSON(symbols)))
}

func jsAppendColoredWatchlist(color string, symbols []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+jsColoredMutationHelper+`
var color = %s;
var syms = %s;
var raw = await _wlFetch("/api/v1/symbols_list/colored/" + encodeURIComponent(color) + "/append/", {
  method: "POST",
  headers: {"Content-Type": "application/json"},
  body: JSON.stringify(syms)
});
return JSON.stringify({ok:true,data:_coloredResult(raw, color)});
`, jsString(color), jsJSON(symbols)))
}

func jsRemoveColoredWatchlist(color string, symbols []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+jsColoredMutationHelper+`
var color = %s;
var syms = %s;
var raw = await _wlFetch("/api/v1/symbols_list/colored/" + encodeURIComponent(color) + "/remove/", {
  method: "POST",
  headers: {"Content-Type": "application/json"},
  body: JSON.stringify(syms)
});
return JSON.stringify({ok:true,data:_coloredResult(raw, color)});
`, jsString(color), jsJSON(symbols)))
}

func jsBulkRemoveColoredWatchlist(symbols []string) string {
	return wrapJSEvalAsync(fmt.Sprintf(jsWatchlistFetch+`
var syms = %s;
await _wlFetch("/api/v1/symbols_list/colored/bulk_remove/", {
  method: "POST",
  headers: {"Content-Type": "application/json"},
  body: JSON.stringify(syms)
});
return JSON.stringify({ok:true,data:{status:"ok"}});
`, jsJSON(symbols)))
}

// --- Study Template JS functions ---
