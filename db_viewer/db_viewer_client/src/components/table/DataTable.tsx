import React, { useState, useCallback, useRef } from 'react';
import ReactDOM from 'react-dom';
import { Row } from '../../types';
import { FilteredColumn } from '../../utils/filterColumns';
import { useScrollShadow } from '../../hooks/useScrollShadow';

interface TooltipState {
  meta: Record<string, any>;
  colName: string;
  x: number;
  y: number;
}

interface DataTableProps {
  rows: Row[];
  columns: string[];
  filterItems: FilteredColumn[] | null;
  pageOffset: number;
  columnMeta?: Record<string, any>[];
  sortCol?: string;
  sortDir?: string;
  onRowClick: (row: Row) => void;
  onColumnClick: (colName: string) => void;
  onSortClick: (colName: string) => void;
  onFieldClick: (row: Row, colName: string) => void;
  onDeleteRow: (row: Row) => void;
  onCloneRow: (row: Row) => void;
}

const DataTable: React.FC<DataTableProps> = ({
  rows, columns, filterItems, pageOffset, columnMeta, sortCol, sortDir, onRowClick, onColumnClick, onSortClick, onFieldClick, onDeleteRow, onCloneRow,
}) => {
  const [colWidths, setColWidths] = useState<Record<string, number>>({});
  const resizingRef = useRef<{ col: string; startX: number; startW: number } | null>(null);
  const justResizedRef = useRef(false);
  const { scrollRef, wrapperClass } = useScrollShadow();
  const [openMenuIdx, setOpenMenuIdx] = useState<number | null>(null);
  const [menuPos, setMenuPos] = useState<{ top: number; left: number }>({ top: 0, left: 0 });
  const [tooltip, setTooltip] = useState<TooltipState | null>(null);

  const handleThMouseEnter = useCallback((e: React.MouseEvent, meta: Record<string, any>, colName: string) => {
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    setTooltip({ meta, colName, x: rect.left, y: rect.bottom + 4 });
  }, []);

  const handleThMouseLeave = useCallback(() => setTooltip(null), []);

  const handleResizeStart = useCallback((e: React.MouseEvent, colName: string, thEl: HTMLTableCellElement) => {
    e.preventDefault();
    e.stopPropagation();
    const startW = thEl.offsetWidth;
    resizingRef.current = { col: colName, startX: e.clientX, startW };

    const onMouseMove = (ev: MouseEvent) => {
      if (!resizingRef.current) return;
      const diff = ev.clientX - resizingRef.current.startX;
      const newW = Math.max(50, resizingRef.current.startW + diff);
      setColWidths((prev) => ({ ...prev, [resizingRef.current!.col]: newW }));
    };

    const onMouseUp = () => {
      resizingRef.current = null;
      justResizedRef.current = true;
      document.removeEventListener('mousemove', onMouseMove);
      document.removeEventListener('mouseup', onMouseUp);
      requestAnimationFrame(() => { justResizedRef.current = false; });
    };

    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
  }, []);

  // Close menu when clicking outside
  const menuRef = useRef<HTMLDivElement | null>(null);
  React.useEffect(() => {
    if (openMenuIdx === null) return;
    const handler = (e: MouseEvent) => {
      const target = e.target as Node;
      if (menuRef.current && !menuRef.current.contains(target)) {
        // If clicking another row-menu-btn, let its onClick handle the toggle
        if ((target as HTMLElement).closest?.('.row-menu-btn')) return;
        setOpenMenuIdx(null);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [openMenuIdx]);

  if (rows.length === 0) {
    return <div className="empty-state">No data to display</div>;
  }

  const displayItems: FilteredColumn[] = filterItems
    ? filterItems
    : columns.map((c) => ({ type: 'column' as const, name: c }));

  return (
    <>
    <div className={wrapperClass}>
    <div className="data-view" ref={scrollRef}>
      <table className="data-table" style={{ tableLayout: 'auto', width: 'max-content', minWidth: '100%' }}>
        <thead>
          <tr>
            <th className="row-action-header">#</th>
            {displayItems.map((item, i) => {
              if (item.type === 'space') {
                return <th key={`space-${i}`} className="filter-space-col"></th>;
              }
              if (item.type === 'commentary') {
                return (
                  <th key={`comment-${i}`} className="filter-commentary-col">
                    <span className="commentary-label">{item.text}</span>
                  </th>
                );
              }
              const colName = item.name!;
              const width = colWidths[colName];
              const meta = columnMeta?.find((c) => c.COLUMN_NAME === colName);
              return (
                <th
                  key={colName}
                  onClick={() => { if (!justResizedRef.current) onSortClick(colName); }}
                  style={{ cursor: 'pointer', width: width ? `${width}px` : undefined }}
                  className="col-header-with-tooltip"
                  onMouseEnter={meta ? (e) => handleThMouseEnter(e, meta, colName) : undefined}
                  onMouseLeave={meta ? handleThMouseLeave : undefined}
                >
                  <span className="col-header-text">{colName}</span>
                  {sortCol === colName && (
                    <span className="col-sort-indicator">{sortDir === 'desc' ? ' ↓' : ' ↑'}</span>
                  )}
                  <span
                    className="col-info-icon"
                    onClick={(e) => { e.stopPropagation(); onColumnClick(colName); }}
                  >
                    ℹ
                  </span>
                  <span
                    className="col-resize-handle"
                    onMouseDown={(e) => {
                      const th = e.currentTarget.parentElement as HTMLTableCellElement;
                      handleResizeStart(e, colName, th);
                    }}
                  />
                </th>
              );
            })}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => {
            const rowNo = pageOffset + i + 1;
            return (
              <tr key={i}>
                <td className="row-action-cell">
                  <span className="row-action-num">{rowNo}</span>
                  <button
                    className="row-menu-btn"
                    title="Row actions"
                    onClick={(e) => {
                      e.stopPropagation();
                      if (openMenuIdx === i) {
                        setOpenMenuIdx(null);
                      } else {
                        const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
                        setMenuPos({ top: rect.bottom + 2, left: rect.left });
                        setOpenMenuIdx(i);
                      }
                    }}
                  >
                    ⋯
                  </button>
                  {openMenuIdx === i && ReactDOM.createPortal(
                    <div className="row-menu-dropdown" ref={menuRef} style={{ top: menuPos.top, left: menuPos.left }}>
                      <div className="row-menu-item" onClick={() => { setOpenMenuIdx(null); onRowClick(row); }}>View Row Data</div>
                      <div className="row-menu-item" onClick={() => { setOpenMenuIdx(null); onCloneRow(row); }}>Clone Row</div>
                      <div className="row-menu-item row-menu-item-danger" onClick={() => { setOpenMenuIdx(null); onDeleteRow(row); }}>Delete Row</div>
                    </div>,
                    document.body
                  )}
                </td>
                {displayItems.map((item, j) => {
                  if (item.type === 'space') {
                    return <td key={`space-${j}`} className="filter-space-col"></td>;
                  }
                  if (item.type === 'commentary') {
                    return <td key={`comment-${j}`} className="filter-commentary-col"></td>;
                  }
                  const colName = item.name!;
                  const val = row[colName];
                  const isNull = val === null || val === undefined;
                  const blobCols: string[] = row['__blob_columns'] || [];
                  const isBlob = blobCols.includes(colName);
                  return (
                    <td
                      key={colName}
                      title={isNull ? 'null' : isBlob ? '[BLOB data]' : String(val)}
                      className={`clickable-cell${isNull ? ' null-value' : ''}${isBlob ? ' blob-value' : ''}`}
                      onClick={() => onFieldClick(row, colName)}
                    >
                      {isNull ? 'null' : isBlob ? (
                        <span className="blob-truncated">
                          {String(val) === '' ? '(empty)' : '[BLOB]'}
                          {String(val).endsWith('...') ? <span className="blob-truncate-indicator"> ***truncate***</span> : ''}
                        </span>
                      ) : String(val)}
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
    </div>
    {tooltip && ReactDOM.createPortal(
      <div
        className="col-header-tooltip"
        style={{ display: 'block', position: 'fixed', left: tooltip.x, top: tooltip.y, zIndex: 9999 }}
      >
        <div className="col-tooltip-name">{tooltip.colName}</div>
        <div className="col-tooltip-row">
          <span className="col-tooltip-label">Type</span>
          <span className="col-tooltip-value">
            {(() => {
              const t = tooltip.meta.DATA_TYPE as string;
              const p = tooltip.meta.DATA_PRECISION;
              const s = tooltip.meta.DATA_SCALE;
              const l = tooltip.meta.DATA_LENGTH;
              if (p != null) return s != null ? `${t}(${p},${s})` : `${t}(${p})`;
              if (/CHAR|RAW|BINARY/i.test(t) && l != null) return `${t}(${l})`;
              return t;
            })()}
          </span>
        </div>
        <div className="col-tooltip-row">
          <span className="col-tooltip-label">Mandatory</span>
          <span className={`col-tooltip-value ${tooltip.meta.NULLABLE === 'N' ? 'col-tooltip-yes' : ''}`}>
            {tooltip.meta.NULLABLE === 'N' ? 'Yes' : 'No'}
          </span>
        </div>
      </div>,
      document.body
    )}
    </>
  );
};

export default DataTable;
