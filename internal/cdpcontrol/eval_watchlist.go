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

// --- Notes JS functions ---

func jsListNotes(symbol string) string {
	filterSnippet := ""
	if symbol != "" {
		filterSnippet = fmt.Sprintf(`
var _sym = %s;
if (_sym) {
  var filtered = [];
  for (var i = 0; i < notes.length; i++) {
    if (notes[i].symbol === _sym || notes[i].symbol_full === _sym) filtered.push(notes[i]);
  }
  notes = filtered;
}`, jsString(symbol))
	}
	return wrapJSEvalAsync(jsWatchlistFetch + `
var notes = await _wlFetch("/textnotes/getall/");
if (!Array.isArray(notes)) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"textnotes/getall returned non-array"});
` + filterSnippet + `
var result = [];
for (var i = 0; i < notes.length; i++) {
  var n = notes[i] || {};
  result.push({
    id: Number(n.id || 0),
    title: String(n.title || ""),
    description: String(n.description || ""),
    created: Number(n.created || 0),
    modified: Number(n.modified || 0),
    symbol: String(n.symbol || ""),
    symbol_full: String(n.symbol_full || ""),
    snapshot: String(n.snapshot || ""),
    snapshot_page: String(n.snapshot_page || ""),
    snapshot_thumb: String(n.snapshot_thumb || "")
  });
}
return JSON.stringify({ok:true,data:{notes:result}});
`)
}

func jsCreateNote(symbolFull, description, title, snapshotUID string) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
var symbolFull = %s;
var desc = %s;
var title = %s;
var snapUID = %s;
var fd = new FormData();
fd.append("symbol", symbolFull);
fd.append("symbol_full", symbolFull);
fd.append("description", desc);
fd.append("title", title || "");
if (snapUID) fd.append("snapshot_uid", snapUID);
var resp = await fetch("/textnotes/add/", {method:"POST", credentials:"include", body:fd});
if (!resp.ok) {
  var body = "";
  try { var j = await resp.json(); body = j.detail || j.message || JSON.stringify(j); } catch(_) { body = await resp.text(); }
  throw new Error("HTTP " + resp.status + ": " + body);
}
var n = await resp.json();
return JSON.stringify({ok:true,data:{
  id: Number(n.id || 0),
  title: String(n.title || ""),
  description: String(n.description || ""),
  created: Number(n.created || 0),
  modified: Number(n.modified || 0),
  symbol: String(n.symbol || ""),
  symbol_full: String(n.symbol_full || ""),
  snapshot: String(n.snapshot || ""),
  snapshot_page: String(n.snapshot_page || ""),
  snapshot_thumb: String(n.snapshot_thumb || "")
}});
`, jsString(symbolFull), jsString(description), jsString(title), jsString(snapshotUID)))
}

func jsEditNote(noteID int, description, title, snapshotUID string) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
var noteId = %d;
var desc = %s;
var title = %s;
var snapUID = %s;
var fd = new FormData();
fd.append("id", String(noteId));
fd.append("description", desc);
fd.append("title", title || "");
if (snapUID) fd.append("snapshot_uid", snapUID);
var resp = await fetch("/textnotes/edit/", {method:"POST", credentials:"include", body:fd});
if (!resp.ok) {
  var body = "";
  try { var j = await resp.json(); body = j.detail || j.message || JSON.stringify(j); } catch(_) { body = await resp.text(); }
  throw new Error("HTTP " + resp.status + ": " + body);
}
var n = await resp.json();
return JSON.stringify({ok:true,data:{
  id: Number(n.id || 0),
  title: String(n.title || ""),
  description: String(n.description || ""),
  created: Number(n.created || 0),
  modified: Number(n.modified || 0),
  symbol: String(n.symbol || ""),
  symbol_full: String(n.symbol_full || ""),
  snapshot: String(n.snapshot || ""),
  snapshot_page: String(n.snapshot_page || ""),
  snapshot_thumb: String(n.snapshot_thumb || "")
}});
`, noteID, jsString(description), jsString(title), jsString(snapshotUID)))
}

func jsDeleteNote(noteID int) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
var noteId = %d;
var fd = new FormData();
fd.append("id_note", String(noteId));
var resp = await fetch("/textnotes/remove/", {method:"POST", credentials:"include", body:fd});
if (!resp.ok && resp.status !== 404) {
  var body = "";
  try { var j = await resp.json(); body = j.detail || j.message || JSON.stringify(j); } catch(_) { body = await resp.text(); }
  throw new Error("HTTP " + resp.status + ": " + body);
}
return JSON.stringify({ok:true,data:{status:"deleted"}});
`, noteID))
}

func jsTakeServerSnapshot() string {
	return wrapJSEvalAsync(jsPreamble + `
if (!api || !api._chartWidgetCollection) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"_chartWidgetCollection not found"});
}
var cwc = api._chartWidgetCollection;
if (typeof cwc.clientSnapshot !== "function") {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"clientSnapshot not found"});
}

// Get TV's fetch wrapper (module 131890) which adds X-Requested-With and X-Language headers.
var wpReq = window.__tvAgentWpRequire;
if (!wpReq) {
  var chunkArr = window.webpackChunktradingview;
  if (chunkArr && chunkArr.push) {
    chunkArr.push([["__tvSnapProbe"], {}, function(r) { wpReq = r; window.__tvAgentWpRequire = r; }]);
  }
}
var tvFetch = (wpReq && wpReq(131890) && wpReq(131890).fetch) || window.fetch;

// Render chart to canvas, encode as base64, upload via images JSON field.
// Using "images" instead of "preparedImage" blob because blob serialization from
// CDP-evaluated JS does not trigger server-side thumbnail generation.
var canvas = await cwc.clientSnapshot();
var dataUrl = canvas.toDataURL("image/png");
var imagesJSON = JSON.stringify({
  panes:[{content:dataUrl,contentWidth:canvas.width,contentHeight:canvas.height,
    leftAxis:{contentWidth:0},rightAxis:{contentWidth:0}}],
  timeAxis:{contentWidth:0,contentHeight:0,lhsStub:{contentWidth:0},rhsStub:{contentWidth:0}}
});
var fd = new FormData();
fd.append("previews[]", "thumb");
fd.append("images", imagesJSON);
var resp = await tvFetch("/snapshot/", {method:"POST", credentials:"same-origin", body:fd});
if (!resp.ok) {
  var body = "";
  try { body = await resp.text(); } catch(_) {}
  throw new Error("snapshot upload HTTP " + resp.status + ": " + body);
}
var uid = (await resp.text()).trim();
return JSON.stringify({ok:true,data:{uid:String(uid),url:"https://www.tradingview.com/x/" + String(uid) + "/"}});
`)
}

// --- Study Template JS functions ---
