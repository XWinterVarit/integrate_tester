import React from 'react';
import { ViewMode } from '../../types';

interface ToolbarProps {
  columns: string[];
  selectCols: string;
  sortCol: string;
  sortDir: string;
  limit: number;
  viewMode: ViewMode;
  onSelectColsChange: (v: string) => void;
  onSortColChange: (v: string) => void;
  onSortDirChange: (v: string) => void;
  onLimitChange: (v: number) => void;
  onViewModeChange: (v: ViewMode) => void;
  onRefresh: () => void;
  onShowTableInfo: () => void;
}

const Toolbar: React.FC<ToolbarProps> = ({
  columns, selectCols, sortCol, sortDir, limit, viewMode,
  onSelectColsChange, onSortColChange, onSortDirChange,
  onLimitChange, onViewModeChange, onRefresh, onShowTableInfo,
}) => {
  return (
    <div className="toolbar">
      <div className="toolbar-group">
        <label>Select</label>
        <input
          value={selectCols}
          onChange={(e) => onSelectColsChange(e.target.value)}
          placeholder="* or COL1, COL2"
          style={{ width: 160 }}
        />
      </div>

      <div className="toolbar-group">
        <label>Sort</label>
        <select value={sortCol} onChange={(e) => onSortColChange(e.target.value)}>
          <option value="">None</option>
          {columns.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
        <select value={sortDir} onChange={(e) => onSortDirChange(e.target.value)} style={{ width: 70 }}>
          <option value="asc">ASC</option>
          <option value="desc">DESC</option>
        </select>
      </div>

      <div className="toolbar-group">
        <label>Limit</label>
        <input
          type="number"
          value={limit}
          onChange={(e) => onLimitChange(Number(e.target.value))}
          style={{ width: 70 }}
          min={1}
        />
      </div>

      <div className="view-toggle">
        <button
          className={viewMode === 'row' ? 'active' : ''}
          onClick={() => onViewModeChange('row')}
        >
          Rows
        </button>
        <button
          className={viewMode === 'transpose' ? 'active' : ''}
          onClick={() => onViewModeChange('transpose')}
        >
          Transpose
        </button>
      </div>

      <button onClick={onRefresh}>Refresh</button>
      <button className="secondary" onClick={onShowTableInfo}>Table Info</button>
    </div>
  );
};

export default Toolbar;
