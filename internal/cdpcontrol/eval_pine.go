package cdpcontrol

import "fmt"

func jsProbePineDOM() string {
	return wrapJSEval(`
var result = {buttons:[], bottom_tabs:[], toolbar:[], monaco:null, console_panel:null};
// Scan right sidebar buttons
var sidebar = document.querySelectorAll('[data-name]');
for (var i = 0; i < sidebar.length; i++) {
  var el = sidebar[i];
  var name = el.getAttribute('data-name') || '';
  if (name.toLowerCase().indexOf('pine') !== -1 || name.toLowerCase().indexOf('editor') !== -1 || name.toLowerCase().indexOf('script') !== -1) {
    result.buttons.push({data_name: name, tag: el.tagName, visible: el.offsetParent !== null, text: (el.textContent || '').trim().substring(0, 50)});
  }
}
// Scan bottom panel tabs
var bottomTabs = document.querySelectorAll('#bottom-area button, #bottom-area [role="tab"], .bottom-widgetbar-content button');
for (var i = 0; i < bottomTabs.length; i++) {
  var el = bottomTabs[i];
  var txt = (el.textContent || '').trim();
  var dn = el.getAttribute('data-name') || '';
  if (txt || dn) {
    result.bottom_tabs.push({data_name: dn, text: txt.substring(0, 50), tag: el.tagName, visible: el.offsetParent !== null});
  }
}
// Scan Pine toolbar buttons (save, add to chart, etc.)
var toolbarBtns = document.querySelectorAll('[class*="pine"] button, [data-name*="pine"] button, .tv-script-editor button');
for (var i = 0; i < toolbarBtns.length; i++) {
  var el = toolbarBtns[i];
  var dn = el.getAttribute('data-name') || '';
  var ariaLabel = el.getAttribute('aria-label') || '';
  var txt = (el.textContent || '').trim();
  result.toolbar.push({data_name: dn, aria_label: ariaLabel, text: txt.substring(0, 50), tag: el.tagName, visible: el.offsetParent !== null});
}
// Check for Monaco editor
var monacoEl = document.querySelector('.monaco-editor');
if (monacoEl) {
  result.monaco = {found: true, visible: monacoEl.offsetParent !== null, classes: monacoEl.className.substring(0, 100)};
}
// Check for Pine console
var consoleEl = document.querySelector('[class*="console"], [data-name*="console"]');
if (consoleEl) {
  result.console_panel = {found: true, visible: consoleEl.offsetParent !== null, tag: consoleEl.tagName};
}
return JSON.stringify({ok:true,data:result});
`)
}

// jsPineLocateToggleBtn returns JS that finds the button to click for
// open/close and returns its center coordinates. The Go caller then dispatches
// a trusted CDP Input.dispatchMouseEvent at those coordinates.

func jsPineLocateToggleBtn() string {
	return wrapJSEval(`
var monacoEl = document.querySelector('.monaco-editor');
var isOpen = !!(monacoEl && monacoEl.offsetParent !== null);

var btn = null;
if (isOpen) {
  // Find the Close button inside the Pine editor panel
  var panel = monacoEl;
  for (var up = 0; up < 10 && panel; up++) {
    panel = panel.parentElement;
    if (!panel) break;
    var closeBtns = panel.querySelectorAll('button');
    for (var bi = 0; bi < closeBtns.length; bi++) {
      var b = closeBtns[bi];
      if (!b.offsetParent) continue;
      var cls = b.className || '';
      var txt = (b.textContent || '').trim().toLowerCase();
      if (cls.indexOf('closeButton') !== -1 || txt === 'close') {
        btn = b; break;
      }
    }
    if (btn) break;
  }
} else {
  // Find the sidebar Pine button
  btn = document.querySelector('button[data-name="pine-dialog-button"]')
     || document.querySelector('button[aria-label="Pine"]');
  if (!btn) {
    var allBtns = document.querySelectorAll('[role="toolbar"] button, [class*="toolbar"] button');
    for (var i = 0; i < allBtns.length; i++) {
      if ((allBtns[i].textContent || '').trim() === 'Pine') { btn = allBtns[i]; break; }
    }
  }
}
if (!btn) return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"Pine editor toggle button not found in DOM"});
var rect = btn.getBoundingClientRect();
var x = rect.x + rect.width / 2;
var y = rect.y + rect.height / 2;
return JSON.stringify({ok:true,data:{action:isOpen?"close":"open",x:x,y:y}});
`)
}

// jsPineWaitForOpen polls until the Pine editor is visible with rendered content.

func jsPineWaitForOpen() string {
	return wrapJSEvalAsync(`
var deadline = Date.now() + 5000;
var isVisible = false;
var monacoReady = false;
while (Date.now() < deadline) {
  var el = document.querySelector('.monaco-editor');
  if (el && el.offsetParent !== null) {
    isVisible = true;
    var vl = el.querySelector('.view-lines');
    if (vl && vl.children.length > 0) {
      // Also verify no spinner overlay is blocking
      var dialog = el.closest('[class*="wrap-"]') || el.closest('[class*="dialog"]');
      var sp = dialog ? dialog.querySelector('.tv-spinner--shown') : null;
      if (!sp || sp.offsetParent === null) { monacoReady = true; break; }
    }
  }
  await new Promise(function(r) { setTimeout(r, 200); });
}
return JSON.stringify({ok:true,data:{status:"opened",is_visible:isVisible,monaco_ready:monacoReady}});
`)
}

// jsPineWaitForClose polls until the Pine editor disappears.

func jsPineWaitForClose() string {
	return wrapJSEvalAsync(`
var deadline = Date.now() + 3000;
while (Date.now() < deadline) {
  var el = document.querySelector('.monaco-editor');
  if (!el || el.offsetParent === null) break;
  await new Promise(function(r) { setTimeout(r, 200); });
}
var stillVisible = (function() { var el = document.querySelector('.monaco-editor'); return !!(el && el.offsetParent !== null); })();
return JSON.stringify({ok:true,data:{status:"closed",is_visible:stillVisible,monaco_ready:false}});
`)
}

// jsPineMonacoPreamble returns JS code that discovers the Monaco editor namespace
// via the webpack module cache (since TradingView doesn't expose monaco globally)
// and caches it on window.__tvMonacoNs. After this preamble, the variable `monacoNs`
// is set to the monaco namespace (or null if not found).

func jsPineStatus() string {
	return wrapJSEval(`
var monacoEl = document.querySelector('.monaco-editor');
var isVisible = !!(monacoEl && monacoEl.offsetParent !== null);
var monacoReady = false;
if (isVisible && monacoEl) {
  // Check for rendered editor content in DOM
  var viewLines = monacoEl.querySelector('.view-lines');
  var hasContent = !!(viewLines && viewLines.children.length > 0);
  // Check for stale loading screen overlay (TradingView bug on reopen)
  var dialog = monacoEl.closest('[class*="wrap-"]') || monacoEl.closest('[class*="dialog"]');
  var hasSpinner = false;
  if (dialog) {
    var sp = dialog.querySelector('.tv-spinner--shown');
    hasSpinner = !!(sp && sp.offsetParent !== null);
  }
  monacoReady = hasContent && !hasSpinner;
}
return JSON.stringify({ok:true,data:{status:isVisible?"open":"closed",is_visible:isVisible,monaco_ready:monacoReady}});
`)
}

func jsPineGetSource() string {
	return wrapJSEval(jsPineMonacoPreamble + `
var monacoEl = document.querySelector('.monaco-editor');
if (!monacoEl || monacoEl.offsetParent === null) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"Pine editor not visible — call POST /pine/toggle first"});
}
var source = "";
if (monacoNs) {
  try {
    var models = monacoNs.editor.getModels();
    if (models && models.length > 0) {
      source = models[0].getValue() || "";
    }
  } catch(_) {}
}
// Fallback: read from the visible code lines in DOM
if (!source) {
  try {
    var lines = monacoEl.querySelectorAll('.view-line');
    if (lines.length > 0) {
      var parts = [];
      for (var i = 0; i < lines.length; i++) {
        parts.push(lines[i].textContent || "");
      }
      source = parts.join("\\n");
    }
  } catch(_) {}
}
if (!source) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"could not read source from Monaco editor"});
var scriptName = "";
var m = source.match(/(?:indicator|strategy|library)\s*\(\s*(?:"([^"]+)"|'([^']+)')/);
if (m) scriptName = m[1] || m[2] || "";
return JSON.stringify({ok:true,data:{
  status:"open",
  is_visible:true,
  monaco_ready:true,
  script_name:scriptName,
  script_source:source,
  source_length:source.length,
  line_count:source.split("\\n").length
}});
`)
}

func jsPineSetSource(source string) string {
	return wrapJSEval(fmt.Sprintf(jsPineMonacoPreamble+`
var newSource = %s;
var monacoEl = document.querySelector('.monaco-editor');
if (!monacoEl || monacoEl.offsetParent === null) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"Pine editor not visible — call POST /pine/toggle first"});
}
var setOk = false;
if (monacoNs) {
  try {
    var models = monacoNs.editor.getModels();
    if (models && models.length > 0) {
      models[0].setValue(newSource);
      setOk = true;
    }
  } catch(_) {}
}
if (!setOk) return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"could not write source to Monaco editor — Monaco namespace not found"});
var scriptName = "";
var m = newSource.match(/(?:indicator|strategy|library)\s*\(\s*(?:"([^"]+)"|'([^']+)')/);
if (m) scriptName = m[1] || m[2] || "";
return JSON.stringify({ok:true,data:{
  status:"set",
  is_visible:true,
  monaco_ready:true,
  script_name:scriptName,
  source_length:newSource.length,
  line_count:newSource.split("\\n").length
}});
`, jsString(source)))
}

// jsPineFocusEditor ensures the Monaco editor is visible and focused.
// Called before sending trusted CDP key events for save/add-to-chart.

func jsPineFocusEditor() string {
	return wrapJSEval(`
var monacoEl = document.querySelector('.monaco-editor');
if (!monacoEl || monacoEl.offsetParent === null) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"Pine editor not visible — call POST /pine/toggle first"});
}
// Focus the Monaco textarea so keyboard shortcuts are received
var textarea = monacoEl.querySelector('textarea.inputarea');
if (textarea) textarea.focus();
return JSON.stringify({ok:true,data:{status:"focused",is_visible:true,monaco_ready:true}});
`)
}

// jsPineWaitForSave polls briefly after Ctrl+S to let the save complete.

func jsPineWaitForSave() string {
	return wrapJSEvalAsync(`
await new Promise(function(r) { setTimeout(r, 1500); });
return JSON.stringify({ok:true,data:{status:"saved",is_visible:true,monaco_ready:true}});
`)
}

// jsPineWaitForAddToChart waits for TradingView to process Ctrl+Enter.
// If a "Cannot add a script with unsaved changes" confirmation dialog appears,
// it clicks "Save and add to chart" to proceed.

func jsPineWaitForAddToChart() string {
	return wrapJSEvalAsync(`
var deadline = Date.now() + 3000;
while (Date.now() < deadline) {
  // Check for the confirmation dialog about unsaved changes
  var btns = document.querySelectorAll('button');
  for (var i = 0; i < btns.length; i++) {
    var txt = (btns[i].textContent || '').trim();
    if (txt === 'Save and add to chart') {
      btns[i].click();
      await new Promise(function(r) { setTimeout(r, 2000); });
      return JSON.stringify({ok:true,data:{status:"added",is_visible:true,monaco_ready:true}});
    }
  }
  await new Promise(function(r) { setTimeout(r, 200); });
}
return JSON.stringify({ok:true,data:{status:"added",is_visible:true,monaco_ready:true}});
`)
}

func jsPineGetConsole() string {
	return wrapJSEval(`
var messages = [];
// Try reading from Pine console DOM elements
var consoleSelectors = [
  '[class*="console"] [class*="message"]',
  '[class*="console"] [class*="row"]',
  '.tv-script-editor__console [class*="message"]',
  '[data-name*="console"] [class*="message"]'
];
for (var si = 0; si < consoleSelectors.length; si++) {
  var els = document.querySelectorAll(consoleSelectors[si]);
  if (els.length > 0) {
    for (var i = 0; i < els.length; i++) {
      var el = els[i];
      var text = (el.textContent || '').trim();
      if (!text) continue;
      var type = 'info';
      var cls = el.className || '';
      if (cls.indexOf('error') !== -1) type = 'error';
      else if (cls.indexOf('warn') !== -1) type = 'warning';
      messages.push({type:type, message:text});
    }
    break;
  }
}
return JSON.stringify({ok:true,data:{messages:messages}});
`)
}

// jsPineBriefWait waits the given milliseconds then returns the current Pine editor status.

func jsPineBriefWait(ms int) string {
	return wrapJSEvalAsync(fmt.Sprintf(`
await new Promise(function(r) { setTimeout(r, %d); });
var monacoEl = document.querySelector('.monaco-editor');
var isVisible = !!(monacoEl && monacoEl.offsetParent !== null);
var monacoReady = false;
if (isVisible && monacoEl) {
  var viewLines = monacoEl.querySelector('.view-lines');
  monacoReady = !!(viewLines && viewLines.children.length > 0);
}
return JSON.stringify({ok:true,data:{status:isVisible?"open":"closed",is_visible:isVisible,monaco_ready:monacoReady}});
`, ms))
}

// jsPineWaitForNewScript waits for a new script template to load in the editor after a chord shortcut.

func jsPineWaitForNewScript() string {
	return jsPineBriefWait(2000)
}

// jsPineClickFirstScriptResult clicks the first result row in the "Open my script" dialog,
// waits for the script to load, then returns close button coords so the caller can CDP-click it.

func jsPineClickFirstScriptResult() string {
	return wrapJSEvalAsync(`
// The "Open my script" dialog has result rows with class containing "itemInfo-".
// The itemInfo div has an onclick handler that loads the script.
var deadline = Date.now() + 3000;
var clicked = false;
var itemEl = null;
while (Date.now() < deadline && !clicked) {
  var items = document.querySelectorAll('[class*="itemInfo-"]');
  if (items.length > 0) {
    itemEl = items[0];
    items[0].click();
    clicked = true;
    break;
  }
  await new Promise(function(r) { setTimeout(r, 200); });
}
// Wait for editor to reload the script
await new Promise(function(r) { setTimeout(r, 1200); });
// Find the dialog close button: walk up from the clicked item to find container,
// then find the close button within it.
var closeX = 0, closeY = 0;
if (itemEl) {
  var container = itemEl;
  for (var i = 0; i < 10 && container; i++) {
    var closeBtn = container.querySelector('[class*="close-"]');
    if (closeBtn && closeBtn.tagName === 'BUTTON') {
      var r = closeBtn.getBoundingClientRect();
      closeX = r.x + r.width / 2;
      closeY = r.y + r.height / 2;
      break;
    }
    container = container.parentElement;
  }
}
var monacoEl = document.querySelector('.monaco-editor');
var isVisible = !!(monacoEl && monacoEl.offsetParent !== null);
var monacoReady = false;
if (isVisible && monacoEl) {
  var vl = monacoEl.querySelector('.view-lines');
  monacoReady = !!(vl && vl.children.length > 0);
}
return JSON.stringify({ok:true,data:{status:isVisible?"open":"closed",is_visible:isVisible,monaco_ready:monacoReady,close_x:closeX,close_y:closeY}});
`)
}

// jsPineFindReplace uses the Monaco API to find all occurrences and replace them,
// preserving undo history via pushEditOperations.

func jsPineFindReplace(find, replace string) string {
	return wrapJSEval(fmt.Sprintf(jsPineMonacoPreamble+`
var findStr = %s;
var replaceStr = %s;
var monacoEl = document.querySelector('.monaco-editor');
if (!monacoEl || monacoEl.offsetParent === null) {
  return JSON.stringify({ok:false,error_code:"API_UNAVAILABLE",error_message:"Pine editor not visible — call POST /pine/toggle first"});
}
if (!monacoNs) {
  return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"Monaco namespace not found"});
}
var models = monacoNs.editor.getModels();
if (!models || models.length === 0) {
  return JSON.stringify({ok:false,error_code:"EVAL_FAILURE",error_message:"No Monaco models found"});
}
var model = models[0];
var matches = model.findMatches(findStr, false, false, true, null, false);
var matchCount = matches.length;
if (matchCount === 0) {
  var src = model.getValue();
  var scriptName = "";
  var m = src.match(/(?:indicator|strategy|library)\s*\(\s*(?:"([^"]+)"|'([^']+)')/);
  if (m) scriptName = m[1] || m[2] || "";
  return JSON.stringify({ok:true,data:{status:"no_matches",is_visible:true,monaco_ready:true,match_count:0,script_name:scriptName,source_length:src.length,line_count:src.split("\\n").length}});
}
var edits = [];
for (var i = 0; i < matches.length; i++) {
  edits.push({range: matches[i].range, text: replaceStr});
}
model.pushEditOperations([], edits, function() { return null; });
var newSource = model.getValue();
var scriptName = "";
var m = newSource.match(/(?:indicator|strategy|library)\s*\(\s*(?:"([^"]+)"|'([^']+)')/);
if (m) scriptName = m[1] || m[2] || "";
return JSON.stringify({ok:true,data:{status:"replaced",is_visible:true,monaco_ready:true,match_count:matchCount,script_name:scriptName,source_length:newSource.length,line_count:newSource.split("\\n").length}});
`, jsString(find), jsString(replace)))
}

// --- Layout management JS functions ---
