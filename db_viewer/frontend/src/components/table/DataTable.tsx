import React, { useState, useCallback, useRef } from 'react';
import { Row } from '../../types';
import { FilteredColumn } from '../../utils/filterColumns';
import { useScrollShadow } from '../../hooks/useScrollShadow';

interface DataTableProps {
  rows: Row[];
  columns: string[];
  filterItems: FilteredColumn[] | null;
  pageOffset: number;
  onRowClick: (row: Row) => void;
  onColumnClick: (colName: string) => void;
  onSortClick: (colName: string) => void;
  onFieldClick: (row: Row, colName: string) => void;
  onDeleteRow: (row: Row) => void;
  onCloneRow: (row: Row) => void;
}

const DataTable: React.FC<DataTableProps> = ({
  rows, columns, filterItems, pageOffset, onRowClick, onColumnClick, onSortClick, onFieldClick, onDeleteRow, onCloneRow,
}) => {
  const [colWidths, setColWidths] = useState<Record<string, number>>({});
  const resizingRef = useRef<{ col: string; startX: number; startW: number } | null>(null);
  const justResizedRef = useRef(false);
  const { scrollRef, wrapperClass } = useScrollShadow();
  const [openMenuIdx, setOpenMenuIdx] = useState<number | null>(null);
  const [menuPos, setMenuPos] = useState<{ top: number; left: number }>({ top: 0, left: 0 });

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
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
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
              return (
                <th
                  key={colName}
                  onClick={() => { if (!justResizedRef.current) onSortClick(colName); }}
                  style={{ cursor: 'pointer', position: 'relative', width: width ? `${width}px` : undefined }}
                >
                  <span className="col-header-text">{colName}</span>
                  <span
                    className="col-info-icon"
                    title="Column info"
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
                  {openMenuIdx === i && (
                    <div className="row-menu-dropdown" ref={menuRef} style={{ top: menuPos.top, left: menuPos.left }}>
                      <div className="row-menu-item" onClick={() => { setOpenMenuIdx(null); onRowClick(row); }}>View Row Data</div>
                      <div className="row-menu-item" onClick={() => { setOpenMenuIdx(null); onCloneRow(row); }}>Clone Row</div>
                      <div className="row-menu-item row-menu-item-danger" onClick={() => { setOpenMenuIdx(null); onDeleteRow(row); }}>Delete Row</div>
                    </div>
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
  );
};

export default DataTable;
