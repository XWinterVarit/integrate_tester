import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { api } from './api/client';
import { ViewMode, FloatingWindow as FloatingWindowType, PresetFilter, PresetQuery, Row } from './types';
import { parseFilterColumns } from './utils/filterColumns';
import Sidebar from './components/layout/Sidebar';
import Toolbar from './components/layout/Toolbar';
import DataTable from './components/table/DataTable';
import TransposeView from './components/table/TransposeView';
import FloatingWindow from './components/floating/FloatingWindow';
import { RowJsonContent, ColumnInfoContent, FieldEditContent, TableInfoContent } from './components/floating/FloatingContent';
import FilterDropdown from './components/filter/FilterDropdown';
import PresetQueryPanel from './components/query/PresetQueryPanel';
import ExportButton from './components/export/ExportButton';

// Per-table state that should be remembered when switching tables
interface TableState {
  activeFilter: PresetFilter | null;
  sortCol: string;
  sortDir: string;
  selectCols: string;
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
  const [error, setError] = useState('');

  // Query params
  const [selectCols, setSelectCols] = useState('');
  const [sortCol, setSortCol] = useState('');
  const [sortDir, setSortDir] = useState('asc');
  const [limit, setLimit] = useState(100);

  // View
  const [viewMode, setViewMode] = useState<ViewMode>('row');

  // Filters & presets
  const [filters, setFilters] = useState<PresetFilter[]>([]);
  const [activeFilter, setActiveFilter] = useState<PresetFilter | null>(null);
  const [presetQueries, setPresetQueries] = useState<PresetQuery[]>([]);

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
  useEffect(() => {
    api.getClients().then(setClients).catch(() => {});
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
    try {
      const params: Record<string, string> = { limit: String(limit) };
      if (selectCols) params.select = selectCols;
      if (sortCol) { params.sort = sortCol; params.sort_dir = sortDir; }

      const [rowData, colData, filterData, presetData] = await Promise.all([
        api.getRows(selectedClient, selectedTable, params),
        api.getColumns(selectedClient, selectedTable),
        api.getFilters(selectedClient, selectedTable),
        api.getPresetQueries(selectedClient, selectedTable),
      ]);

      setRows(rowData || []);
      setColumnMeta(colData || []);
      setAllColumns(
        rowData && rowData.length > 0
          ? Object.keys(rowData[0])
          : (colData || []).map((c: any) => c.COLUMN_NAME)
      );
      setFilters(filterData || []);
      setPresetQueries(presetData || []);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, [selectedClient, selectedTable, selectCols, sortCol, sortDir, limit]);

  useEffect(() => {
    loadTableData();
  }, [loadTableData]);

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

  const handlePresetExecute = async (query: string, args: Record<string, string>) => {
    if (!selectedClient || !selectedTable) return;
    setLoading(true);
    setError('');
    try {
      const result = await api.executeQuery(selectedClient, selectedTable, { query, args, limit });
      setRows(result || []);
      if (result && result.length > 0) {
        setAllColumns(Object.keys(result[0]));
      }
      api.touchRecentQuery(`${selectedClient}:${selectedTable}:preset`).catch(() => {});
    } catch (e: any) {
      setError(e.message);
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

  const queryParams: Record<string, string> = { limit: String(limit) };
  if (selectCols) queryParams.select = selectCols;
  if (sortCol) { queryParams.sort = sortCol; queryParams.sort_dir = sortDir; }

  return (
    <div className="app-layout">
      <Sidebar
        clients={clients}
        tables={tables}
        selectedClient={selectedClient}
        selectedTable={selectedTable}
        onSelectClient={setSelectedClient}
        onSelectTable={(t: string) => {
          // Save current table's state
          if (selectedTable) {
            setTableStateCache((prev) => ({
              ...prev,
              [`${selectedClient}:${selectedTable}`]: {
                activeFilter,
                sortCol,
                sortDir,
                selectCols,
                viewMode,
              },
            }));
          }
          // Restore new table's state or reset
          const cached = tableStateCache[`${selectedClient}:${t}`];
          if (cached) {
            setSortCol(cached.sortCol);
            setSortDir(cached.sortDir);
            setSelectCols(cached.selectCols);
            setActiveFilter(cached.activeFilter);
            setViewMode(cached.viewMode);
          } else {
            setSortCol('');
            setSortDir('asc');
            setSelectCols('');
            setActiveFilter(null);
            setViewMode('row');
          }
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
              selectCols={selectCols}
              sortCol={sortCol}
              sortDir={sortDir}
              limit={limit}
              viewMode={viewMode}
              onSelectColsChange={setSelectCols}
              onSortColChange={setSortCol}
              onSortDirChange={setSortDir}
              onLimitChange={setLimit}
              onViewModeChange={setViewMode}
              onRefresh={loadTableData}
              onShowTableInfo={handleShowTableInfo}
            />

            <div className="toolbar" style={{ borderTop: 'none', paddingTop: 0 }}>
              <FilterDropdown
                filters={filters}
                activeFilter={activeFilter}
                onSelect={handleFilterSelect}
              />
              <PresetQueryPanel
                presets={presetQueries}
                table={selectedTable}
                onExecute={handlePresetExecute}
              />
              <ExportButton
                client={selectedClient}
                table={selectedTable}
                queryParams={queryParams}
              />
              {loading && <span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>Loading...</span>}
              {error && <span style={{ fontSize: 12, color: 'var(--danger)' }}>{error}</span>}
            </div>

            {viewMode === 'row' ? (
              <DataTable
                rows={rows}
                columns={visibleColumns}
                filterItems={filterItems}
                onRowClick={handleRowClick}
                onColumnClick={handleColumnClick}
                onSortClick={handleSortClick}
                onFieldClick={handleFieldClick}
              />
            ) : (
              <TransposeView
                rows={rows}
                allColumns={allColumns}
                filterColumns={activeFilter?.columns || null}
                onColumnClick={handleColumnClick}
                onFieldClick={handleFieldClick}
                onRowClick={handleRowClick}
              />
            )}
          </>
        )}
      </div>

      {floatingWindows.map((win) => (
        <FloatingWindow key={win.id} window={win} onClose={closeFloating} onPopOut={popOutFloating}>
          {win.type === 'row-json' && <RowJsonContent row={win.content} />}
          {win.type === 'column-info' && (
            <ColumnInfoContent
              columnName={win.content.colName}
              columnMeta={win.content.meta}
              constraints={constraints}
            />
          )}
          {win.type === 'field-edit' && (
            <FieldEditContent
              columnName={win.content.colName}
              value={win.content.value}
              client={selectedClient}
              table={selectedTable}
              row={win.content.row}
              onSaved={(colName, newValue, originalRow) => {
                closeFloating(win.id);
                setRows((prev) =>
                  prev.map((r) => {
                    const firstCol = Object.keys(originalRow)[0];
                    if (String(r[firstCol]) === String(originalRow[firstCol])) {
                      return { ...r, [colName]: newValue };
                    }
                    return r;
                  })
                );
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
        </FloatingWindow>
      ))}
    </div>
  );
};

export default App;
