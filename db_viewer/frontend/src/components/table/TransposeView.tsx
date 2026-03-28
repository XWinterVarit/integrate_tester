import React from 'react';
import { Row } from '../../types';
import { FilteredColumn, parseFilterColumns } from '../../utils/filterColumns';
import { useScrollShadow } from '../../hooks/useScrollShadow';

interface TransposeViewProps {
  rows: Row[];
  allColumns: string[];
  filterColumns: string[] | null;
  onColumnClick: (colName: string) => void;
  onFieldClick: (row: Row, colName: string) => void;
  onRowClick: (row: Row) => void;
}

const TransposeView: React.FC<TransposeViewProps> = ({
  rows, allColumns, filterColumns, onColumnClick, onFieldClick, onRowClick,
}) => {
  const { scrollRef, wrapperClass } = useScrollShadow();

  if (rows.length === 0) {
    return <div className="empty-state">No data to display</div>;
  }

  const items: FilteredColumn[] = filterColumns
    ? parseFilterColumns(filterColumns, allColumns)
    : allColumns.map((c) => ({ type: 'column' as const, name: c }));

  return (
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
                  <td colSpan={rows.length + 1}>{item.text}</td>
                </tr>
              );
            }
            const colName = item.name!;
            return (
              <tr key={colName}>
                <td className="field-label">
                  <span onClick={() => onColumnClick(colName)} style={{ cursor: 'pointer' }}>
                    {colName}
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
  );
};

export default TransposeView;
