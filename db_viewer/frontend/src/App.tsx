import React, { useState, useEffect, useCallback } from 'react';
import { api } from './api/client';
import { ViewMode, FloatingWindow as FloatingWindowType, PresetFilter, PresetQuery, Row } from './types';
import Sidebar from './components/layout/Sidebar';
import Toolbar from './components/layout/Toolbar';
import DataTable from './components/table/DataTable';
import TransposeView from './components/table/TransposeView';
import FloatingWindow from './components/floating/FloatingWindow';
import { RowJsonContent, ColumnInfoContent, TableInfoContent } from './components/floating/FloatingContent';
import FilterDropdown from './components/filter/FilterDropdown';
import PresetQueryPanel from './components/query/PresetQueryPanel';
import ExportButton from './components/export/ExportButton';

const App: React.FC = () => {
  // Client & table selection
  const [clients, setClients] = useState<{ name: string; schema: string }[]>([]);
  const [selectedClient, setSelectedClient] = useState('');
  const [tables, setTables] = useState<string[]>([]);
  const [selectedTable, setSelectedTable] = useState('');

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
  const [transposeRowIndex, setTransposeRowIndex] = useState(0);

  // Filters & presets
  const [filters, setFilters] = useState<PresetFilter[]>([]);
  const [activeFilter, setActiveFilter] = useState<PresetFilter | null>(null);
  const [presetQueries, setPresetQueries] = useState<PresetQuery[]>([]);

  // Floating windows
  const [floatingWindows, setFloatingWindows] = useState<FloatingWindowType[]>([]);
  let floatingCounter = 0;

  // Load clients on mount
  useEffect(() => {
    api.getClients().then(setClients).catch(() => {});
  }, []);

  // Load tables when client changes
  useEffect(() => {
    if (!selectedClient) return;
    setSelectedTable('');
    setRows([]);
    setAllColumns([]);
    api.getTables(selectedClient).then(setTables).catch(() => setTables([]));
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
      setTransposeRowIndex(0);
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
    const currentRow = rows[transposeRowIndex] || rows[0] || null;
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

  // Visible columns based on active filter
  const visibleColumns = activeFilter
    ? activeFilter.columns.filter((c) => !c.startsWith('<')).filter((c) => allColumns.includes(c))
    : allColumns;

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
        onSelectTable={setSelectedTable}
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
                onRowClick={handleRowClick}
                onColumnClick={handleColumnClick}
                onSortClick={handleSortClick}
              />
            ) : (
              <TransposeView
                rows={rows}
                allColumns={allColumns}
                filterColumns={activeFilter?.columns || null}
                selectedRowIndex={transposeRowIndex}
                onRowIndexChange={setTransposeRowIndex}
                onColumnClick={handleColumnClick}
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
              client={selectedClient}
              table={selectedTable}
              row={win.content.currentRow}
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
