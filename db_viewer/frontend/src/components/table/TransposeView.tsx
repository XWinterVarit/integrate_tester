import React from 'react';
import { Row } from '../../types';
import { FilteredColumn, parseFilterColumns } from '../../utils/filterColumns';

interface TransposeViewProps {
  rows: Row[];
  allColumns: string[];
  filterColumns: string[] | null;
  selectedRowIndex: number;
  onRowIndexChange: (i: number) => void;
  onColumnClick: (colName: string) => void;
}

const TransposeView: React.FC<TransposeViewProps> = ({
  rows, allColumns, filterColumns, selectedRowIndex,
  onRowIndexChange, onColumnClick,
}) => {
  if (rows.length === 0) {
    return <div className="empty-state">No data to display</div>;
  }

  const row = rows[selectedRowIndex] || rows[0];
  const items: FilteredColumn[] = filterColumns
    ? parseFilterColumns(filterColumns, allColumns)
    : allColumns.map((c) => ({ type: 'column' as const, name: c }));

  return (
    <div className="data-view">
      <div style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', display: 'flex', gap: 8, alignItems: 'center' }}>
        <button className="secondary" disabled={selectedRowIndex <= 0} onClick={() => onRowIndexChange(selectedRowIndex - 1)}>
          ← Prev
        </button>
        <span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>
          Row {selectedRowIndex + 1} of {rows.length}
        </span>
        <button className="secondary" disabled={selectedRowIndex >= rows.length - 1} onClick={() => onRowIndexChange(selectedRowIndex + 1)}>
          Next →
        </button>
      </div>
      <table className="transpose-table">
        <tbody>
          {items.map((item, i) => {
            if (item.type === 'space') {
              return <tr key={`space-${i}`} className="space-row"><td colSpan={2}></td></tr>;
            }
            if (item.type === 'commentary') {
              return (
                <tr key={`comment-${i}`} className="commentary-row">
                  <td colSpan={2}>{item.text}</td>
                </tr>
              );
            }
            const colName = item.name!;
            return (
              <tr key={colName}>
                <td className="field-label" onClick={() => onColumnClick(colName)}>
                  {colName}
                </td>
                <td className="field-value">{String(row[colName] ?? '')}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};

export default TransposeView;
