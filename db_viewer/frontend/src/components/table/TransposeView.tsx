import React, { useState } from 'react';
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
}

interface TooltipState {
  meta: Record<string, any>;
  colName: string;
  x: number;
  y: number;
}

const TransposeView: React.FC<TransposeViewProps> = ({
  rows, allColumns, filterColumns, columnMeta, sortCol, sortDir, onColumnClick, onSortClick, onFieldClick, onRowClick,
}) => {
  const { scrollRef, wrapperClass } = useScrollShadow();
  const [tooltip, setTooltip] = useState<TooltipState | null>(null);

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
                  <th key={ri} style={{ textAlign: 'center', minWidth: 120 }}>
                    <span style={{ fontSize: 11, color: 'var(--text-secondary)' }}>Row {ri + 1}</span>
                    <button
                      className="expand-row-btn"
                      title="View row data"
                      style={{ marginLeft: 6 }}
                      onClick={() => onRowClick(row)}
                    >
                      ⤢
                    </button>
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
              {tooltip.meta.DATA_LENGTH != null
                ? `${tooltip.meta.DATA_TYPE}(${tooltip.meta.DATA_LENGTH})`
                : tooltip.meta.DATA_TYPE}
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
