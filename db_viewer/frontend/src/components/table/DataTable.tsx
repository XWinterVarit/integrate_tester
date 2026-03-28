import React from 'react';
import { Row } from '../../types';

interface DataTableProps {
  rows: Row[];
  columns: string[];
  onRowClick: (row: Row) => void;
  onColumnClick: (colName: string) => void;
  onSortClick: (colName: string) => void;
}

const DataTable: React.FC<DataTableProps> = ({
  rows, columns, onRowClick, onColumnClick, onSortClick,
}) => {
  if (rows.length === 0) {
    return <div className="empty-state">No data to display</div>;
  }

  return (
    <div className="data-view">
      <table className="data-table">
        <thead>
          <tr>
            {columns.map((col) => (
              <th key={col} onClick={() => onSortClick(col)}>
                <span onClick={(e) => { e.stopPropagation(); onColumnClick(col); }}
                  style={{ cursor: 'pointer' }}>
                  {col}
                </span>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} onClick={() => onRowClick(row)}>
              {columns.map((col) => (
                <td key={col} title={String(row[col] ?? '')}>
                  {String(row[col] ?? '')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default DataTable;
