import React, { useState, useCallback, useRef } from 'react';
import ReactDOM from 'react-dom';
import { Row } from '../../types';
import { FilteredColumn, parseFilterColumns } from '../../utils/filterColumns';
import { useScrollShadow } from '../../hooks/useScrollShadow';

interface TransposeViewProps {
  rows: Row[];
  allColumns: string[];
  filterColumns: string[] | null;
  columnMeta?: Record<string, any>[];
  sortCol?: string;
  sortDir?: string;
  onColumnClick: (colName: string) => void;
  onSortClick: (colName: string) => void;
  onFieldClick: (row: Row, colName: string) => void;
  onRowClick: (row: Row) => void;
  onDeleteRow: (row: Row) => void;
  onCloneRow: (row: Row) => void;
}

interface TooltipState {
  meta: Record<string, any>;
  colName: string;
  x: number;
  y: number;
}

const TransposeView: React.FC<TransposeViewProps> = ({
  rows, allColumns, filterColumns, columnMeta, sortCol, sortDir, onColumnClick, onSortClick, onFieldClick, onRowClick, onDeleteRow, onCloneRow,
}) => {
  const { scrollRef, wrapperClass } = useScrollShadow();
  const [tooltip, setTooltip] = useState<TooltipState | null>(null);
  const [openMenuIdx, setOpenMenuIdx] = useState<number | null>(null);
  const [menuPos, setMenuPos] = useState<{ top: number; left: number }>({ top: 0, left: 0 });
  const menuRef = useRef<HTMLDivElement | null>(null);

  React.useEffect(() => {
    if (openMenuIdx === null) return;
    const handler = (e: MouseEvent) => {
      const target = e.target as Node;
      if (menuRef.current && !menuRef.current.contains(target)) {
        if ((target as HTMLElement).closest?.('.row-menu-btn')) return;
        setOpenMenuIdx(null);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [openMenuIdx]);
  const [colWidths, setColWidths] = useState<Record<number, number>>({});
  const resizingRef = useRef<{ idx: number; startX: number; startW: number } | null>(null);

  const handleResizeStart = useCallback((e: React.MouseEvent, idx: number, thEl: HTMLTableCellElement) => {
    e.preventDefault();
    e.stopPropagation();
    const startW = thEl.offsetWidth;
    resizingRef.current = { idx, startX: e.clientX, startW };

    const onMouseMove = (ev: MouseEvent) => {
      if (!resizingRef.current) return;
      const diff = ev.clientX - resizingRef.current.startX;
      const newW = Math.max(50, resizingRef.current.startW + diff);
      setColWidths((prev) => ({ ...prev, [resizingRef.current!.idx]: newW }));
    };

    const onMouseUp = () => {
      resizingRef.current = null;
      document.removeEventListener('mousemove', onMouseMove);
      document.removeEventListener('mouseup', onMouseUp);
    };

    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
  }, []);

  if (rows.length === 0) {
    return <div className="empty-state">No data to display</div>;
  }

  const items: FilteredColumn[] = filterColumns
    ? parseFilterColumns(filterColumns, allColumns)
    : allColumns.map((c) => ({ type: 'column' as const, name: c }));

  const handleMouseEnter = (e: React.MouseEvent, meta: Record<string, any>, colName: string) => {
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    setTooltip({ meta, colName, x: rect.left, y: rect.bottom + 4 });
  };

  const handleMouseLeave = () => setTooltip(null);

  return (
    <>
      <div className={wrapperClass}>
        <div className="data-view" ref={scrollRef}>
          <table className="transpose-table">
            <thead>
              <tr>
                <th className="field-label" style={{ minWidth: 140 }}>Column</th>
                {rows.map((row, ri) => (
                  <th key={ri} style={{ textAlign: 'center', minWidth: 120, width: colWidths[ri] ? `${colWidths[ri]}px` : undefined, position: 'relative' }}>
                    <span style={{ fontSize: 11, color: 'var(--text-secondary)' }}>Row {ri + 1}</span>
                    <button
                      className="row-menu-btn"
                      title="Row actions"
                      style={{ marginLeft: 6 }}
                      onClick={(e) => {
                        e.stopPropagation();
                        if (openMenuIdx === ri) {
                          setOpenMenuIdx(null);
                        } else {
                          const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
                          setMenuPos({ top: rect.bottom + 2, left: rect.left });
                          setOpenMenuIdx(ri);
                        }
                      }}
                    >
                      ⋯
                    </button>
                    {openMenuIdx === ri && ReactDOM.createPortal(
                      <div className="row-menu-dropdown" ref={menuRef} style={{ top: menuPos.top, left: menuPos.left }}>
                        <div className="row-menu-item" onClick={() => { setOpenMenuIdx(null); onRowClick(row); }}>View Row Data</div>
                        <div className="row-menu-item" onClick={() => { setOpenMenuIdx(null); onCloneRow(row); }}>Clone Row</div>
                        <div className="row-menu-item row-menu-item-danger" onClick={() => { setOpenMenuIdx(null); onDeleteRow(row); }}>Delete Row</div>
                      </div>,
                      document.body
                    )}
                    <span
                      className="col-resize-handle"
                      onMouseDown={(e) => {
                        const th = e.currentTarget.parentElement as HTMLTableCellElement;
                        handleResizeStart(e, ri, th);
                      }}
                    />
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {items.map((item, i) => {
                if (item.type === 'space') {
                  return (
                    <tr key={`space-${i}`} className="space-row">
                      <td colSpan={rows.length + 1}></td>
                    </tr>
                  );
                }
                if (item.type === 'commentary') {
                  return (
                    <tr key={`comment-${i}`} className="commentary-row">
                      <td className="field-label" style={{ fontWeight: 600 }}>{item.text}</td>
                      {rows.map((_, ri) => <td key={ri}></td>)}
                    </tr>
                  );
                }
                const colName = item.name!;
                const meta = columnMeta?.find((c) => c.COLUMN_NAME === colName);
                return (
                  <tr key={colName}>
                    <td
                      className="field-label"
                      onMouseEnter={meta ? (e) => handleMouseEnter(e, meta, colName) : undefined}
                      onMouseLeave={meta ? handleMouseLeave : undefined}
                    >
                      <span onClick={() => onSortClick(colName)} style={{ cursor: 'pointer' }}>
                        {colName}
                        {sortCol === colName && (
                          <span className="col-sort-indicator">{sortDir === 'desc' ? ' →' : ' ←'}</span>
                        )}
                      </span>
                      <span
                        className="col-info-icon"
                        onClick={(e) => { e.stopPropagation(); onColumnClick(colName); }}
                      >
                        ℹ
                      </span>
                    </td>
                    {rows.map((row, ri) => {
                      const val = row[colName];
                      const isNull = val === null || val === undefined;
                      const blobCols: string[] = row['__blob_columns'] || [];
                      const isBlob = blobCols.includes(colName);
                      return (
                        <td
                          key={ri}
                          className={`field-value clickable-cell${isNull ? ' null-value' : ''}${isBlob ? ' blob-value' : ''}`}
                          onClick={() => onFieldClick(row, colName)}
                        >
                          {isNull ? 'null' : isBlob ? (
                            <span className="blob-truncated">{String(val).substring(0, 40)}{String(val).length > 40 ? '…' : ''}</span>
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

export default TransposeView;
