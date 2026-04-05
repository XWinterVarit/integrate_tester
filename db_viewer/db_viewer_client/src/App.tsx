import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import ReactDOM from 'react-dom';
import { api } from './api/client';
import { ViewMode, FloatingWindow as FloatingWindowType, PresetFilter, PresetQuery, Row } from './types';
import { parseFilterColumns } from './utils/filterColumns';
import Sidebar from './components/layout/Sidebar';
import Toolbar from './components/layout/Toolbar';
import DataTable from './components/table/DataTable';
import TransposeView from './components/table/TransposeView';
import FloatingWindow from './components/floating/FloatingWindow';
import { RowJsonContent, ColumnInfoContent, FieldEditContent, TableInfoContent } from './components/floating/FloatingContent';
import DeleteConfirm from './components/floating/DeleteConfirm';
import InsertForm from './components/floating/InsertForm';
import FilterDropdown from './components/filter/FilterDropdown';
import PresetQueryPanel from './components/query/PresetQueryPanel';
import PresetQueryContent from './components/query/PresetQueryContent';
import ExportButton from './components/export/ExportButton';
import Toast, { ToastMessage } from './components/ui/Toast';
import FieldDescEditor from './components/field/FieldDescEditor';
import ClientManagerModal from './components/client/ClientManagerModal';
import AboutModal from './components/client/AboutModal';

// Per-table state that should be remembered when switching tables
interface TableState {
  activeFilter: PresetFilter | null;
  activePresetQuery: PresetQuery | null;
  activePresetArgs: Record<string, string>;
  sortCol: string;
  sortDir: string;
  where: string;
  viewMode: ViewMode;
}

const App: React.FC = () => {
  // Read initial state from URL search params (enables "Open in new tab")
  const urlParams = useMemo(() => new URLSearchParams(window.location.search), []);
  const initialClient = urlParams.get('client') || '';
  const initialTable = urlParams.get('table') || '';

  // Client & table selection
  const [clients, setClients] = useState<{ name: string; schema: string }[]>([]);
  const [selectedClient, setSelectedClient] = useState(initialClient);
  const [tables, setTables] = useState<string[]>([]);
  const [selectedTable, setSelectedTable] = useState(initialTable);

  // Per-table state cache
  const [tableStateCache, setTableStateCache] = useState<Record<string, TableState>>({});

  // Data
  const [rows, setRows] = useState<Row[]>([]);
  const [allColumns, setAllColumns] = useState<string[]>([]);
  const [columnMeta, setColumnMeta] = useState<Record<string, any>[]>([]);
  const [constraints, setConstraints] = useState<Record<string, any>[]>([]);
  const [indexes, setIndexes] = useState<Record<string, any>[]>([]);
  const [tableSize, setTableSize] = useState<Record<string, any>[]>([]);
  const [loading, setLoading] = useState(false);
  const [slowQuery, setSlowQuery] = useState(false);
  const [error, setError] = useState('');
  const [fieldDescriptions, setFieldDescriptions] = useState<Record<string, string>>({});
  const [fieldDescTarget, setFieldDescTarget] = useState<string | null>(null);
  const [showClientManager, setShowClientManager] = useState(false);
  const [showAbout, setShowAbout] = useState(false);

  // Query params
  const [where, setWhere] = useState('');
  const [sortCol, setSortCol] = useState('');
  const [sortDir, setSortDir] = useState('asc');
  const [limit, setLimit] = useState(100);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalCount, setTotalCount] = useState(0);

  // View
  const [viewMode, setViewMode] = useState<ViewMode>('row');

  // Filters & presets
  const [filters, setFilters] = useState<PresetFilter[]>([]);
  const [activeFilter, setActiveFilter] = useState<PresetFilter | null>(null);
  const [presetQueries, setPresetQueries] = useState<PresetQuery[]>([]);
  const [activePresetQuery, setActivePresetQuery] = useState<PresetQuery | null>(null);
  const [activePresetArgs, setActivePresetArgs] = useState<Record<string, string>>({});

  // Toast notifications
  const [toasts, setToasts] = useState<ToastMessage[]>([]);
  // Query indicator tooltip (portal-based to escape overflow:hidden)
  const queryIndicatorRef = useRef<HTMLDivElement>(null);
  const [queryTooltipPos, setQueryTooltipPos] = useState<{ top: number; left: number } | null>(null);
  let toastCounter = 0;
  const addToast = useCallback((type: 'success' | 'error', text: string, duration?: string) => {
    const id = Date.now() + (toastCounter++);
    setToasts((prev) => [...prev, { id, type, text, duration }]);
  }, []);
  const dismissToast = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);
  const formatDuration = (ms: number) => ms >= 1000 ? `${(ms / 1000).toFixed(1)}s` : `${ms}ms`;
  // Floating windows
  const [floatingWindows, setFloatingWindows] = useState<FloatingWindowType[]>([]);
  let floatingCounter = 0;

  // Sync URL search params when client/table changes
  useEffect(() => {
    const params = new URLSearchParams();
    if (selectedClient) params.set('client', selectedClient);
    if (selectedTable) params.set('table', selectedTable);
    const newUrl = params.toString() ? `${window.location.pathname}?${params}` : window.location.pathname;
    window.history.replaceState(null, '', newUrl);
  }, [selectedClient, selectedTable]);

  // Load clients on mount
  const refreshClients = useCallback(() => {
    api.getClients().then((data) => setClients(data ?? [])).catch(() => {});
  }, []);
  useEffect(() => {
    refreshClients();
  }, []); 

  // Track the previous client to detect real client changes vs initial mount
  const prevClientRef = React.useRef<string | null>(null);

  // Load tables when client changes
  useEffect(() => {
    if (!selectedClient) return;
    // Only clear table when the user actually switches clients, not on initial mount
    if (prevClientRef.current !== null && prevClientRef.current !== selectedClient) {
      setSelectedTable('');
    }
    prevClientRef.current = selectedClient;
    setRows([]);
    setAllColumns([]);
    api.getTables(selectedClient).then(setTables).catch(() => setTables([]));
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedClient]);

  // Load data when table changes
  const loadTableData = useCallback(async () => {
    if (!selectedClient || !selectedTable) return;
    setLoading(true);
    setError('');
    const t0 = Date.now();
    try {
      const [colData, filterData, presetData, descData] = await Promise.all([
        api.getColumns(selectedClient, selectedTable),
        api.listPresetFilters(selectedClient, selectedTable),
        api.listPresetQueries(selectedClient, selectedTable),
        api.getFieldDescriptions(selectedClient, selectedTable).catch(() => ({})),
      ]);
      setColumnMeta(colData || []);
      setFilters(filterData || []);
      setPresetQueries(presetData || []);
      setFieldDescriptions(descData || {});

      // Fetch total row count
      const countData = await api.getRowCount(selectedClient, selectedTable);
      setTotalCount(countData.count);

      const offset = (currentPage - 1) * limit;

      // If a preset query is active, re-execute it; otherwise do default fetch
      let rowData: Row[];
      if (activePresetQuery) {
        rowData = await api.executeQuery(selectedClient, selectedTable, {
          query: activePresetQuery.query,
          args: activePresetArgs,
          limit,
          offset,
          sort: sortCol || undefined,
          sort_dir: sortCol ? sortDir : undefined,
        });
      } else {
        const params: Record<string, string> = { limit: String(limit), offset: String(offset) };
        if (where) params.where = where;
        if (sortCol) { params.sort = sortCol; params.sort_dir = sortDir; }
        rowData = await api.getRows(selectedClient, selectedTable, params);
      }

      setRows(rowData || []);
      // Sort columns by table structure order (columnMeta) instead of alphabetically
      const metaOrder = (colData || []).map((c: any) => c.COLUMN_NAME as string);
      const rawCols = rowData && rowData.length > 0
        ? Object.keys(rowData[0]).filter((k) => k !== 'ROWID' && k !== '__blob_columns')
        : metaOrder;
      const orderMap = new Map(metaOrder.map((name: string, idx: number) => [name, idx]));
      const sorted = [...rawCols].sort((a, b) => (orderMap.get(a) ?? 9999) - (orderMap.get(b) ?? 9999));
      setAllColumns(sorted);
      addToast('success', `Loaded ${selectedTable}`, formatDuration(Date.now() - t0));
    } catch (e: any) {
      setError(e.message);
      addToast('error', e.message || 'Query failed', formatDuration(Date.now() - t0));
    } finally {
      setLoading(false);
    }
  }, [selectedClient, selectedTable, where, sortCol, sortDir, limit, currentPage, activePresetQuery, activePresetArgs, addToast]);

  useEffect(() => {
    loadTableData();
  }, [loadTableData]);

  useEffect(() => {
    if (!loading) { setSlowQuery(false); return; }
    const t = setTimeout(() => setSlowQuery(true), 3000);
    return () => clearTimeout(t);
  }, [loading]);

  // Floating window helpers
  const addFloating = (title: string, type: FloatingWindowType['type'], content: any) => {
    const id = `float-${Date.now()}-${floatingCounter++}`;
    setFloatingWindows((prev) => [
      ...prev,
      { id, title, content, type, x: 300 + prev.length * 30, y: 100 + prev.length * 30, width: 450, height: 400 },
    ]);
  };

  const closeFloating = (id: string) => {
    setFloatingWindows((prev) => prev.filter((w) => w.id !== id));
  };

  const popOutFloating = (win: FloatingWindowType) => {
    closeFloating(win.id);
  };

  // Handlers
  const handleRowClick = (row: Row) => {
    addFloating('Row Data', 'row-json', row);
  };

  const handleColumnClick = (colName: string) => {
    const meta = columnMeta.find((c: any) => c.COLUMN_NAME === colName) || null;
    const currentRow = rows[0] || null;
    addFloating(`Column: ${colName}`, 'column-info', { colName, meta, currentRow });
  };

  const handleSortClick = (colName: string) => {
    if (sortCol === colName) {
      setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
    } else {
      setSortCol(colName);
      setSortDir('asc');
    }
  };

  const handleFieldClick = (row: Row, colName: string) => {
    addFloating(`${colName}`, 'field-edit', { colName, value: row[colName], row });
  };

  const handleDeleteRow = (row: Row) => {
    addFloating('Delete Row', 'delete-confirm', { row });
  };

  const handleCloneRow = (row: Row) => {
    addFloating('Clone Row (Insert)', 'insert-form', { prefillRow: row });
  };

  const handleInsertNew = () => {
    addFloating('Insert New Row', 'insert-form', { prefillRow: null });
  };

  const handleShowTableInfo = async () => {
    if (!selectedClient || !selectedTable) return;
    try {
      const [sizeData, conData, idxData] = await Promise.all([
        api.getTableSize(selectedClient, selectedTable),
        api.getConstraints(selectedClient, selectedTable),
        api.getIndexes(selectedClient, selectedTable),
      ]);
      setTableSize(sizeData || []);
      setConstraints(conData || []);
      setIndexes(idxData || []);
      addFloating(`Table: ${selectedTable}`, 'table-info', { size: sizeData, constraints: conData, indexes: idxData });
    } catch (e: any) {
      setError(e.message);
    }
  };

  const handleFilterSelect = (filter: PresetFilter | null) => {
    setActiveFilter(filter);
    if (filter) {
      api.touchRecentFilter(`${selectedClient}:${selectedTable}:${filter.name}`).catch(() => {});
    }
  };

  const handlePresetExecute = async (query: string, args: Record<string, string>, preset: PresetQuery | null) => {
    if (!selectedClient || !selectedTable) return;
    setActivePresetQuery(preset);
    setActivePresetArgs(args);
    setLoading(true);
    setError('');
    const t0 = Date.now();
    try {
      setCurrentPage(1);
      const result = await api.executeQuery(selectedClient, selectedTable, { query, args, limit, offset: 0, sort: sortCol || undefined, sort_dir: sortCol ? sortDir : undefined });
      setRows(result || []);
      if (result && result.length > 0) {
        setAllColumns(Object.keys(result[0]).filter((k) => k !== 'ROWID' && k !== '__blob_columns'));
      }
      api.touchRecentQuery(`${selectedClient}:${selectedTable}:preset`).catch(() => {});
      addToast('success', 'Query executed', formatDuration(Date.now() - t0));
    } catch (e: any) {
      setError(e.message);
      addToast('error', e.message || 'Query failed', formatDuration(Date.now() - t0));
    } finally {
      setLoading(false);
    }
  };

  // Visible columns based on active filter (only real columns, no special commands)
  const visibleColumns = activeFilter
    ? activeFilter.columns.filter((c) => !c.startsWith('<')).filter((c) => allColumns.includes(c))
    : allColumns;

  // Parsed filter items for DataTable (includes SPACE/COMMENTARY entries)
  const filterItems = useMemo(() => {
    if (!activeFilter) return null;
    return parseFilterColumns(activeFilter.columns, allColumns);
  }, [activeFilter, allColumns]);

  const totalPages = Math.max(1, Math.ceil(totalCount / limit));

  const queryParams: Record<string, string> = { limit: String(limit) };
  if (where) queryParams.where = where;
  if (sortCol) { queryParams.sort = sortCol; queryParams.sort_dir = sortDir; }

  return (
    <div className="app-layout">
      <Sidebar
        clients={clients}
        tables={tables}
        selectedClient={selectedClient}
        selectedTable={selectedTable}
        onSelectClient={setSelectedClient}
        onShowTableInfo={async (t: string) => {
          try {
            const [sizeData, conData, idxData] = await Promise.all([
              api.getTableSize(selectedClient, t),
              api.getConstraints(selectedClient, t),
              api.getIndexes(selectedClient, t),
            ]);
            addFloating(`Table: ${t}`, 'table-info', { size: sizeData, constraints: conData, indexes: idxData });
          } catch (e: any) {
            addToast('error', e.message || 'Failed to load table info');
          }
        }}
        onShowFieldDesc={(t: string) => setFieldDescTarget(t)}
        onManageClients={() => setShowClientManager(true)}
        onShowAbout={() => setShowAbout(true)}
        onSelectTable={(t: string) => {
          // Save current table's state
          if (selectedTable) {
            setTableStateCache((prev) => ({
              ...prev,
              [`${selectedClient}:${selectedTable}`]: {
                activeFilter,
                activePresetQuery,
                activePresetArgs,
                sortCol,
                sortDir,
                where,
                viewMode,
              },
            }));
          }
          // Restore new table's state or reset
          const cached = tableStateCache[`${selectedClient}:${t}`];
          if (cached) {
            setSortCol(cached.sortCol);
            setSortDir(cached.sortDir);
            setWhere(cached.where);
            setActiveFilter(cached.activeFilter);
            setActivePresetQuery(cached.activePresetQuery);
            setActivePresetArgs(cached.activePresetArgs);
            setViewMode(cached.viewMode);
          } else {
            setSortCol('');
            setSortDir('asc');
            setWhere('');
            setActiveFilter(null);
            setActivePresetQuery(null);
            setActivePresetArgs({});
            setViewMode('row');
          }
          setCurrentPage(1);
          setSelectedTable(t);
        }}
      />

      <div className="main-content">
        {!selectedTable ? (
          <div className="empty-state">Select a table to view data</div>
        ) : (
          <>
            <Toolbar
              columns={allColumns}
              where={where}
              sortCol={sortCol}
              sortDir={sortDir}
              limit={limit}
              viewMode={viewMode}
              whereDisabled={!!activePresetQuery}
              onWhereChange={setWhere}
              onSortColChange={setSortCol}
              onSortDirChange={setSortDir}
              onLimitChange={(v: number) => { setLimit(v > 0 ? v : 100); setCurrentPage(1); }}
              onViewModeChange={setViewMode}
              onRefresh={loadTableData}
              onShowTableInfo={handleShowTableInfo}
            />

            <div className="toolbar toolbar-secondary" style={{ borderTop: 'none', paddingTop: 0 }}>
              <FilterDropdown
                filters={filters}
                activeFilter={activeFilter}
                onSelect={handleFilterSelect}
                client={selectedClient}
                table={selectedTable}
                allColumns={allColumns}
                onRefresh={loadTableData}
              />
              <PresetQueryPanel
                presets={presetQueries}
                client={selectedClient}
                table={selectedTable}
                activePreset={activePresetQuery}
                onExecute={handlePresetExecute}
                onClear={() => { setActivePresetQuery(null); setActivePresetArgs({}); }}
                onOpenPopup={(preset) => addFloating(`Preset: ${preset.name}`, 'preset-query', { preset })}
                onRefresh={loadTableData}
              />
              <ExportButton
                client={selectedClient}
                table={selectedTable}
                queryParams={queryParams}
              />
              <button className="secondary" style={{ marginLeft: 'auto' }} onClick={handleInsertNew}>➕ Insert</button>
              {totalPages > 1 && (
                <div className="pagination-bar">
                  <button disabled={currentPage <= 1} onClick={() => setCurrentPage(1)}>«</button>
                  <button disabled={currentPage <= 1} onClick={() => setCurrentPage((p) => p - 1)}>‹</button>
                  <span className="pagination-info">
                    Page {currentPage} of {totalPages} ({totalCount} rows)
                  </span>
                  <button disabled={currentPage >= totalPages} onClick={() => setCurrentPage((p) => p + 1)}>›</button>
                  <button disabled={currentPage >= totalPages} onClick={() => setCurrentPage(totalPages)}>»</button>
                </div>
              )}
              {loading && <span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>Loading...</span>}
              {slowQuery && <span style={{ fontSize: 12, color: '#d97706', fontWeight: 600 }}>⏳ Query is taking longer than usual…</span>}
              {error && <span style={{ fontSize: 12, color: 'var(--danger)' }}>{error}</span>}
            </div>

            {(activeFilter || activePresetQuery) && (
              <div className="active-indicators">
                {activeFilter && (
                  <span className="active-indicator filter-indicator">
                    🔍 Filter: <strong>{activeFilter.name}</strong>
                  </span>
                )}
                {activePresetQuery && (
                  <div
                    ref={queryIndicatorRef}
                    className="active-indicator query-indicator query-indicator-interactive"
                    onClick={() => addFloating(
                      `Preset: ${activePresetQuery.name}`,
                      'preset-query',
                      { preset: activePresetQuery, initialArgs: activePresetArgs }
                    )}
                    onMouseEnter={() => {
                      if (queryIndicatorRef.current && activePresetQuery.arguments.length > 0) {
                        const rect = queryIndicatorRef.current.getBoundingClientRect();
                        setQueryTooltipPos({ top: rect.bottom + 6, left: rect.left });
                      }
                    }}
                    onMouseLeave={() => setQueryTooltipPos(null)}
                    title="Click to edit query parameters"
                  >
                    📋 Query: <strong>{activePresetQuery.name}</strong>
                    <span
                      className="query-indicator-close"
                      onClick={(e) => {
                        e.stopPropagation();
                        setQueryTooltipPos(null);
                        setActivePresetQuery(null);
                        setActivePresetArgs({});
                      }}
                      title="Clear preset query"
                    >✕</span>
                  </div>
                )}
              </div>
            )}

            {viewMode === 'row' ? (
              <DataTable
                rows={rows}
                columns={visibleColumns}
                filterItems={filterItems}
                pageOffset={(currentPage - 1) * limit}
                columnMeta={columnMeta as any}
                sortCol={sortCol}
                sortDir={sortDir}
                onRowClick={handleRowClick}
                onColumnClick={handleColumnClick}
                onSortClick={handleSortClick}
                onFieldClick={handleFieldClick}
                onDeleteRow={handleDeleteRow}
                onCloneRow={handleCloneRow}
                fieldDescriptions={fieldDescriptions}
              />
            ) : (
              <TransposeView
                rows={rows}
                allColumns={allColumns}
                filterColumns={activeFilter?.columns || null}
                columnMeta={columnMeta as any}
                sortCol={sortCol}
                sortDir={sortDir}
                onColumnClick={handleColumnClick}
                onSortClick={handleSortClick}
                onFieldClick={handleFieldClick}
                onRowClick={handleRowClick}
                onDeleteRow={handleDeleteRow}
                onCloneRow={handleCloneRow}
                fieldDescriptions={fieldDescriptions}
              />
            )}
          </>
        )}
      </div>

      {floatingWindows.map((win) => (
        <FloatingWindow key={win.id} window={win} onClose={closeFloating} onPopOut={popOutFloating}
          disablePopOut={win.type === 'insert-form' || win.type === 'delete-confirm' || win.type === 'field-edit' || win.type === 'preset-query'}>
          {win.type === 'row-json' && <RowJsonContent row={win.content} />}
          {win.type === 'column-info' && (
            <ColumnInfoContent
              columnName={win.content.colName}
              columnMeta={win.content.meta}
              constraints={constraints}
              description={fieldDescriptions[win.content.colName]}
            />
          )}
          {win.type === 'field-edit' && (
            <FieldEditContent
              columnName={win.content.colName}
              value={win.content.value}
              client={selectedClient}
              table={selectedTable}
              row={win.content.row}
              onSaved={(colName, newValue, originalRow, duration) => {
                closeFloating(win.id);
                setRows((prev) =>
                  prev.map((r) => {
                    if (r['ROWID'] && String(r['ROWID']) === String(originalRow['ROWID'])) {
                      return { ...r, [colName]: newValue };
                    }
                    return r;
                  })
                );
                addToast('success', 'Saved successfully', duration);
              }}
            />
          )}
          {win.type === 'table-info' && (
            <TableInfoContent
              size={win.content.size || []}
              constraints={win.content.constraints || []}
              indexes={win.content.indexes || []}
            />
          )}
          {win.type === 'delete-confirm' && (
            <DeleteConfirm
              client={selectedClient}
              table={selectedTable}
              row={win.content.row}
              onDeleted={() => {
                closeFloating(win.id);
                loadTableData();
              }}
            />
          )}
          {win.type === 'preset-query' && (
            <PresetQueryContent
              preset={win.content.preset}
              initialArgs={win.content.initialArgs}
              onExecute={handlePresetExecute}
              onClose={() => closeFloating(win.id)}
            />
          )}
          {win.type === 'insert-form' && (
            <InsertForm
              client={selectedClient}
              table={selectedTable}
              columnMeta={columnMeta as any}
              columns={allColumns}
              prefillRow={win.content.prefillRow}
              onInserted={() => {
                closeFloating(win.id);
                loadTableData();
              }}
            />
          )}
        </FloatingWindow>
      ))}
      <Toast messages={toasts} onDismiss={dismissToast} />

      <ClientManagerModal
        open={showClientManager}
        onClose={() => { setShowClientManager(false); refreshClients(); }}
        onSaved={refreshClients}
      />
      <AboutModal
        open={showAbout}
        onClose={() => setShowAbout(false)}
      />

      {fieldDescTarget && (
        <FieldDescEditor
          client={selectedClient}
          table={fieldDescTarget}
          columns={allColumns}
          presetFilters={filters}
          onClose={() => {
            setFieldDescTarget(null);
            loadTableData();
          }}
        />
      )}

      {/* Query indicator tooltip — portal so it escapes overflow:hidden parents */}
      {queryTooltipPos && activePresetQuery && activePresetQuery.arguments.length > 0 && ReactDOM.createPortal(
        <div
          className="query-indicator-tooltip"
          style={{ position: 'fixed', top: queryTooltipPos.top, left: queryTooltipPos.left, zIndex: 9999 }}
        >
          {activePresetQuery.arguments.map((arg) => (
            <div key={arg.name} className="query-indicator-tooltip-row">
              <span className="query-tooltip-key">{arg.name}:</span>
              <span className="query-tooltip-value">{activePresetArgs[arg.name] || '(empty)'}</span>
            </div>
          ))}
        </div>,
        document.body
      )}
    </div>
  );
};

export default App;
