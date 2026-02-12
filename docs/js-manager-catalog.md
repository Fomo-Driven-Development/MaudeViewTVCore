# TradingView JS Manager Catalog

> Extracted from 489 JS bundles captured 2026-02-11.
> These are browser-side singletons accessible via CDP JS evaluation (`evalOnAnyChart` / `evalOnChart`).
> No REST API — all accessed through runtime JavaScript.

## Automation Potential Legend

- **HIGH** — Direct data access or chart control, strong candidate for controller endpoints
- **MEDIUM** — Useful for specific workflows, may require UI state setup
- **LOW** — Primarily UI cosmetics or internal plumbing

---

## 1. Chart Data API

### `chartApi()` / `_getChartApi()`

**Automation Potential: HIGH**

The most extensive runtime API. Controls chart sessions, studies, series, quotes, and symbol resolution.

**Session Management:**

| Method | Description |
|--------|-------------|
| `connect()` | Connect to chart API |
| `disconnect()` | Disconnect from API |
| `isConnected()` | Check connection status |
| `connected` | Property: connection status |
| `chartCreateSession(options)` | Create chart session |
| `chartDeleteSession(sessionId)` | Delete chart session |
| `createSession(options)` | Create session |
| `removeSession(sessionId)` | Remove session |
| `serverTime` | Property: current server time |

**Study Operations:**

| Method | Description |
|--------|-------------|
| `canCreateStudy()` | Check if study creation is allowed |
| `createStudy(name, forceOverlay, locked, inputs, overrides)` | Create a study |
| `modifyStudy(sessionId, studyId, name, inputs, forcedOverlay, overrides)` | Modify study |
| `removeStudy(sessionId, studyId)` | Remove study |
| `notifyStudy(sessionId, studyId)` | Notify study of changes |
| `getStudyCounter()` | Get study count |
| `getChildStudyCounter()` | Get child study count |
| `getFundamentalCounter()` | Get fundamental study count |

**Series & Pointsets:**

| Method | Description |
|--------|-------------|
| `createSeries(options)` | Create data series |
| `modifySeries(options)` | Modify data series |
| `removeSeries(seriesId)` | Remove data series |
| `createPointset(options)` | Create pointset |
| `modifyPointset(options)` | Modify pointset |
| `removePointset(pointsetId)` | Remove pointset |

**Quote Operations:**

| Method | Description |
|--------|-------------|
| `quoteCreateSession()` | Create quote session |
| `quoteDeleteSession(sessionId)` | Delete quote session |
| `quoteAddSymbols(symbols)` | Add symbols to quote feed |
| `quoteRemoveSymbols(symbols)` | Remove symbols from quote feed |
| `quoteFastSymbols(symbols)` | Get fast quote data |
| `quoteSetFields(fields)` | Set quote fields to receive |
| `quoteHibernateAll()` | Hibernate all quote feeds |

**Symbol & Data:**

| Method | Description |
|--------|-------------|
| `resolveSymbol(symbol)` | Resolve symbol info |
| `requestMetadata(options)` | Request symbol metadata |
| `requestFirstBarTime(options)` | Get first available bar time |
| `requestMoreData(options)` | Request additional historical data |
| `requestMoreTickmarks(options)` | Request additional tickmarks |
| `sendRequest(request)` | Send generic request |
| `switchTimezone(tz)` | Switch chart timezone |
| `setFutureTickmarksMode(mode)` | Set future tickmarks mode |

**Properties:**

| Property | Description |
|----------|-------------|
| `availableCurrencies` | List of available currencies |
| `availableMetrics` | List of available metrics |
| `availablePriceSources` | List of price sources |
| `availableUnits` | List of available units |
| `defaultResolutions` | Default resolution list |
| `lastSymbolResolveInfo` | Last resolved symbol info |

**Bundles:** `14100.db2cea9b5ceb02841ead.js`, `studies.800ff56286c7fd944d59.js`, `symbol-search-dialog.978b821d44bbe4aa418d.js`, `main_chart.0c2c1527a95a1ad708f8.js`

---

## 2. Replay Manager

### `_replayManager`

**Automation Potential: HIGH**

Full chart replay control — start, stop, step, autoplay. 24 methods/properties.

**Playback Control:**

| Method | Description |
|--------|-------------|
| `startReplay(point)` | Start replay at specific point |
| `stopReplay()` | Stop active replay |
| `finishReplay()` | Finish replay session |
| `resetReplay()` | Reset replay state |
| `doReplayStep()` | Step forward one bar |
| `startAutoplay()` | Start autoplay mode |
| `replaySelectedPoint(point)` | Replay from specific point |

**Configuration:**

| Method | Description |
|--------|-------------|
| `changeAutoplayDelay(delay)` | Change autoplay speed |
| `autoplayDelay` | Property: current autoplay delay |
| `getDepth()` | Get replay depth |

**State:**

| Property | Description |
|----------|-------------|
| `isReplayStarted` | Whether replay is active |
| `isReplayFinished` | Whether replay has finished |
| `isReplayConnected` | Whether replay session is connected |
| `isAutoplayStarted` | Whether autoplay is running |
| `replayPoint` | Current replay point |
| `replayStatusWV` | Watched value of replay status |
| `serverTime` | Server time in replay |

**Model Management:**

| Method | Description |
|--------|-------------|
| `addModel(model)` | Add replay model |
| `removeModel(model)` | Remove replay model |
| `setModels(models)` | Set all replay models |
| `models` | Property: collection of models |

**Lifecycle:**

| Method | Description |
|--------|-------------|
| `destroy()` | Destroy replay manager |
| `disconnectionSessionIfExists()` | Disconnect existing session |
| `onEndOfDataReached` | Callback: end of data |
| `onPointTooDeep` | Callback: point too deep |

**Bundle:** `replay.bece7214ab451e682eb5.js`

---

## 3. Hotlists Manager

### `hotlistsManager()`

**Automation Potential: HIGH**

Market movers data — top gainers, most active, top losers by exchange.

| Method | Description |
|--------|-------------|
| `getOneHotlist(null, exchange, group)` | Fetch symbols for a specific hotlist |
| `getMarkets()` | Get market/hotlist organization |
| `availableExchanges` | Property: list of supported exchanges |
| `groupsTitlesMap` | Property: group ID to title mapping |
| `compilations` | Property: set of compilation hotlists |
| `getExchangeName(exchange)` | Get exchange display name |
| `getExchangeFullName(exchange)` | Get full exchange name |
| `getExchangeFlag(exchange)` | Get country flag for exchange |

**Bundles:** `show-watchlists-dialog.3506536237399bfa292c.js`, `42108.775c5fdbb74c22bfb475.js`, `15913.a89f8c19738ee319a7a4.js`

---

## 4. Alerts REST API

### `getAlertsRestApi()`

**Automation Potential: HIGH**

Full alerts CRUD and fire management. Wraps the `pricealerts.tradingview.com` REST service.

**Alert CRUD:**

| Method | Description |
|--------|-------------|
| `createAlert(params)` | Create new alert |
| `modifyRestartAlert(params)` | Update and restart alert |
| `deleteAlerts(alertIds)` | Delete alerts |
| `stopAlerts(alertIds)` | Pause alerts |
| `restartAlerts(alertIds)` | Resume alerts |
| `cloneAlerts(alertIds)` | Clone alerts |
| `listAlerts()` | List all active alerts |
| `getAlerts(alertIds)` | Get specific alerts |

**Fire (Trigger) Management:**

| Method | Description |
|--------|-------------|
| `listFires()` | Get triggered alert events |
| `deleteFires(fireIds)` | Delete specific fires |
| `deleteAllFires()` | Clear all fire history |
| `deleteFiresByFilter(filter)` | Delete fires by filter |
| `getOfflineFires()` | Get offline-triggered events |
| `getOfflineFireControls()` | Get notification controls |
| `clearOfflineFires()` | Clear offline fire history |
| `clearOfflineFireControls()` | Clear offline fire controls |

**Bundle:** `alerts-rest-api.6aff0f1cea74fa1ac07c.js`

---

## 5. Accounts Manager

### `_accountsManager`

**Automation Potential: HIGH**

Full account/broker account management for paper and live trading.

| Method | Description |
|--------|-------------|
| `createAccount(params)` | Create trading account |
| `deleteAccount(accountId)` | Delete trading account |
| `resetAccount(accountId)` | Reset account state |
| `setCurrentAccount(accountId)` | Switch active account |
| `changeAccountSettings(settings)` | Change account settings |
| `currentAccountId` | Property: active account ID |
| `currentAccountApi` | Property: current account API |
| `currentAccountMetainfo` | Property: account metadata |
| `accountsMetainfo` | Property: all accounts metadata |
| `accountSettingsInfo` | Property: account settings |

**Bundle:** `trading.6d7bb86f9c28feed3042.js`

---

## 6. Current Account API

### `currentAccountApi()`

**Automation Potential: HIGH**

Trading capability queries for the active broker account. Returns capability objects for each operation.

| Method | Description |
|--------|-------------|
| `placeOrderCapability()` | Can place orders |
| `modifyOrderCapability()` | Can modify orders |
| `cancelOrderCapability()` | Can cancel orders |
| `closePositionCapability()` | Can close positions |
| `closeIndividualPositionCapability()` | Can close individual positions |
| `reversePositionCapability()` | Can reverse positions |
| `editPositionBracketsCapability()` | Can edit position brackets |
| `editIndividualPositionBracketsCapability()` | Can edit individual brackets |
| `isTradableCapability()` | Symbol is tradable |
| `ordersHistoryCapability()` | Can view order history |
| `orderDialogOptionsCapability()` | Order dialog options |
| `positionDialogOptionsCapability()` | Position dialog options |
| `priceFormatterCapability()` | Price formatting |
| `spreadFormatterCapability()` | Spread formatting |
| `symbolInfoCapability()` | Symbol info access |
| `symbolSpecificTradingOptionsCapability()` | Symbol-specific options |
| `connectWarningMessageCapability()` | Connection warnings |
| `customPageableTableDataCapabilities()` | Custom table data |
| `unhideSymbolSearchGroupsCapability()` | Unhide symbol groups |

**Bundle:** Multiple trading bundles

---

## 7. Deep Backtesting Manager

### `_deepBacktestingManager`

**Automation Potential: HIGH**

Controls deep backtesting / strategy testing sessions.

| Method | Description |
|--------|-------------|
| `requestData(params)` | Request backtesting data |
| `setReportData(data)` | Set report data |
| `resetState()` | Reset backtesting state |
| `disconnect()` | Disconnect session |
| `destroy()` | Destroy manager |
| `activeStrategyReportData` | Property: current report data |
| `activeStrategyStatus` | Property: strategy status |

**Bundle:** `backtesting-strategy-facade.f1f45d5d53028df782df.js` (inferred)

---

## 8. Context Menu Manager

### `ContextMenuManager`

**Automation Potential: MEDIUM**

Programmatic context menu creation and display. Could trigger chart actions without mouse interaction.

| Method | Description |
|--------|-------------|
| `createMenu(items, options, detail?, onClose?)` | Create context menu |
| `showMenu(items, position, options?, detail?, onClose?)` | Show menu at position |
| `hideAll()` | Hide all open menus |
| `getShown()` | Get currently shown menu |
| `setCustomItemsProcessor(processor)` | Set custom items processor |
| `setCustomRendererFactory(factory)` | Set custom renderer factory |

**Known Menu Names:**
`ChartContextMenu`, `PriceScaleContextMenu`, `PriceScaleLabelContextMenu`, `DataWindowWidgetSeriesContextMenu`, `DataWindowWidgetStudyContextMenu`, `DOMWidgetContextMenu`, `ObjectTreeContextMenu`, `LegendPropertiesContextMenu`, `LineToolFloatingToolbarMoreMenu`, `NotesItemContextMenu`, `ContextPlusMenu`, `ScalePlusMenu`, `CrosshairMenuView`, `TimeScaleContextMenu`, `TimezoneMenuContextMenu`, `TradingOrderContextMenu`, `SymbolIntervalChartSyncingMenu`

**Bundles:** `17350.7e1b2a2567efeb85f3ea.js`, `chart-widget-gui.83b79c0d5965d1452951.js`, `floating-toolbars.10ce4afaf954964447ef.js`, and 11 more

---

## 9. Script Manager

### `_scriptManager`

**Automation Potential: MEDIUM**

Pine Script editor control.

| Method | Description |
|--------|-------------|
| `openScript(scriptId)` | Open script in editor |
| `setIsBlurEditor(isBlur)` | Set editor blur state |

**Bundle:** `pine-editor-full.f8219ddb2fda9cac201b.js`

---

## 10. Session Summary Manager

### `_sessionSummaryManager`

**Automation Potential: MEDIUM**

Trading session performance summary.

| Property | Description |
|----------|-------------|
| `currency` | Session currency |
| `realizedPL` | Realized profit/loss |
| `getSessionHighestProfit()` | Highest profit in session |
| `getSessionSuccessRate()` | Win rate percentage |

**Bundle:** `trading.6d7bb86f9c28feed3042.js` (inferred)

---

## 11. Dialogs Opener Manager

### `dialogsOpenerManager`

**Automation Potential: MEDIUM**

Track and control dialog open/close state.

| Method | Description |
|--------|-------------|
| `isOpened(dialogName)` | Check if dialog is open |
| `setAsOpened(dialogName)` | Mark dialog as opened |
| `setAsClosed(dialogName)` | Mark dialog as closed |

**Bundles:** `67856.6ac4692ccfab804a287e.js`, `change-interval-dialog.2a51bc3c342d5b3ca27f.js`

---

## 12. Toast Manager

### `toastManager`

**Automation Potential: MEDIUM**

Notification toast system. Useful for observing alert triggers.

| Method | Description |
|--------|-------------|
| `addToast(toast)` | Add notification toast |
| `removeToast(id)` | Remove specific toast |
| `removeGroup(group)` | Remove group of toasts |
| `subscribeOnToast(callback)` | Subscribe to toast events |
| `startAddAnimation()` | Start add animation |
| `hideAll()` | Hide all toasts |

**Toast Groups:** `"alerts"`, `"alertsFireControl"`, `"alertsWatchlistFireControl"`

**Bundles:** `17350.7e1b2a2567efeb85f3ea.js`, `alerts-notifications.0c4d5957f4d16eaa7e13.js`, `trading.6d7bb86f9c28feed3042.js`

---

## 13. Execution Points Manager

### `_executionsPointsManager`

**Automation Potential: MEDIUM**

Trade execution visualization on chart.

| Method | Description |
|--------|-------------|
| `destroy()` | Destroy manager |
| `existingPoints` | Property: execution points |
| `existingPointsChanged` | Property: change callback |

**Bundles:** `replay.bece7214ab451e682eb5.js`, `trading-custom-sources.aa127b5584653e9ffa2b.js`

---

## 14. Order Presets Manager

### `_orderPresetsManager`

**Automation Potential: MEDIUM**

Manages saved order configurations.

| Method | Description |
|--------|-------------|
| `createConsumer()` | Create order presets consumer |

**Bundle:** `trading.6d7bb86f9c28feed3042.js`

---

## 15. Subscriber Manager

### `_subscriberManager`

**Automation Potential: MEDIUM**

Manages data subscriptions for chart feeds.

| Method | Description |
|--------|-------------|
| `createSubscriber(options)` | Create data subscriber |
| `destroy()` | Destroy manager |
| `isEmpty()` | Check if no subscribers |
| `setMinAvailableResolution(res)` | Set minimum resolution |
| `setSessionId(sessionId)` | Set session ID |

**Bundle:** Chart data bundles

---

## 16. Auth Token Manager

### `_authTokenManager`

**Automation Potential: MEDIUM**

Authentication token lifecycle.

| Method | Description |
|--------|-------------|
| `get()` | Get current auth token |
| `invalidated` | Property: token invalidated status |

**Bundles:** `43201.59d3aa1221f3d1b9e3fc.js`, `backtesting-strategy-facade.f1f45d5d53028df782df.js`

---

## 17. Session Storage Manager

### `sessionStorageManager`

**Automation Potential: LOW**

Read/write browser session storage.

| Method | Description |
|--------|-------------|
| `getItemOrDefault(key, defaultValue)` | Get value or default |
| `setItem(key, value)` | Set value |

**Known Keys:** `"goToDateTabLastPickedDate"`, `"detailsKeyStatsExpanded"`, `"detailsIncomeStatementPeriodId"`

**Bundles:** `87411.48febbddb540a6d967f3.js`, `go-to-date-dialog-impl.7d12819e125634dd2c9a.js`

---

## 18. Color Manager

### `_colorManager`

**Automation Potential: LOW**

UI color theming.

| Method | Description |
|--------|-------------|
| `getColor(colorId)` | Get color value |
| `setColor(colorId, value)` | Set color value |

**Bundles:** `65137.c6b9f4805f7b3aacd786.js`, `chart-widget-gui.83b79c0d5965d1452951.js`

---

## 19. Zone Manager (Canvas)

### `_zoneManager`

**Automation Potential: LOW**

Canvas rendering zone management.

| Method | Description |
|--------|-------------|
| `getCanvasHeight()` | Get canvas height |
| `getCanvasWidth()` | Get canvas width |
| `getDOMHeight()` | Get DOM height |
| `getDOMWidth()` | Get DOM width |
| `setDOMHeight(h)` | Set DOM height |
| `setDOMWidth(w)` | Set DOM width |
| `setLineHeight(h)` | Set line height |
| `setOuterHeight(h)` | Set outer height |
| `setPixelRatio(r)` | Set pixel ratio |
| `setZones(zones)` | Set rendering zones |
| `resolveColorZones(zones)` | Resolve color zones |
| `getId2Color()` | Get ID-to-color mapping |
| `getOuterHeight()` | Get outer height |

**Bundle:** Monaco editor bundles

---

## 20. Misc Managers

### `PriceCurrencyCache`
Caches subscription pricing, not market data. `getInstance()` singleton.

### `iconManager`
Monaco editor icon mappings. `setIcons(iconMap)`.

### `getConfigurationManager()`
Monaco editor configuration. Internal to code editor.

### `getAdapterManager()`
Debug adapter registration for Monaco. `registerDebugAdapterDescriptorFactory()`, `registerDebugAdapterFactory()`, `unregisterDebugAdapterDescriptorFactory()`.

---

## Summary by Automation Potential

### HIGH — Strong controller endpoint candidates
| Manager | Primary Use |
|---------|-------------|
| `chartApi()` / `_getChartApi()` | Chart sessions, studies, quotes, symbol resolution |
| `_replayManager` | Replay playback control |
| `hotlistsManager()` | Market movers data |
| `getAlertsRestApi()` | Alert CRUD and fire management |
| `_accountsManager` | Trading account management |
| `currentAccountApi()` | Trading capability queries |
| `_deepBacktestingManager` | Strategy backtesting |

### MEDIUM — Useful for specific workflows
| Manager | Primary Use |
|---------|-------------|
| `ContextMenuManager` | Programmatic context menu actions |
| `_scriptManager` | Pine Script editor control |
| `_sessionSummaryManager` | Trading session P&L |
| `dialogsOpenerManager` | Dialog state control |
| `toastManager` | Notification observation |
| `_executionsPointsManager` | Trade execution visualization |
| `_orderPresetsManager` | Order presets |
| `_subscriberManager` | Data feed subscriptions |
| `_authTokenManager` | Auth token access |

### LOW — UI internals
| Manager | Primary Use |
|---------|-------------|
| `sessionStorageManager` | Browser session storage |
| `_colorManager` | UI theming |
| `_zoneManager` | Canvas rendering |
| `PriceCurrencyCache` | Subscription pricing |
| `iconManager` | Editor icons |
| `getConfigurationManager()` | Editor config |
| `getAdapterManager()` | Debug adapters |

---

## Source

Extracted from 489 JS bundles in `research_data/2026-02-11/chart_rzWLrz7t/resources/js/`.
