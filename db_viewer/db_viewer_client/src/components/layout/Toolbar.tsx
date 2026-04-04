import React, { useState } from 'react';
import { ViewMode } from '../../types';

interface ToolbarProps {
  columns: string[];
  where: string;
  sortCol: string;
  sortDir: string;
  limit: number;
  viewMode: ViewMode;
  onWhereChange: (v: string) => void;
  onSortColChange: (v: string) => void;
  onSortDirChange: (v: string) => void;
  onLimitChange: (v: number) => void;
  onViewModeChange: (v: ViewMode) => void;
  onRefresh: () => void;
  onShowTableInfo: () => void;
}

const LIMIT_OPTIONS = [10, 20, 50, 100, 500];

const Toolbar: React.FC<ToolbarProps> = ({
  columns, where, sortCol, sortDir, limit, viewMode,
  onWhereChange, onSortColChange, onSortDirChange,
  onLimitChange, onViewModeChange, onRefresh, onShowTableInfo,
}) => {
  const [localWhere, setLocalWhere] = useState(where);

  const handleWhereKey = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      onWhereChange(localWhere);
    }
  };

  // Sync local state when parent resets values (e.g. table change)
  React.useEffect(() => { setLocalWhere(where); }, [where]);

  return (
    <div className="toolbar">
      <div className="toolbar-group">
        <label>Where</label>
        <input
          value={localWhere}
          onChange={(e) => setLocalWhere(e.target.value)}
          onKeyDown={handleWhereKey}
          placeholder="e.g. NAME = 'aaa' AND AGE > 30"
          style={{ width: 260 }}
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
        <select value={limit} onChange={(e) => onLimitChange(Number(e.target.value))} style={{ width: 70 }}>
          {LIMIT_OPTIONS.map(opt => (
            <option key={opt} value={opt}>{opt}</option>
          ))}
        </select>
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
