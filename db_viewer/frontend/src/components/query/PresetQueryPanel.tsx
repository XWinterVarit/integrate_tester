import React, { useState } from 'react';
import { PresetQuery } from '../../types';
interface PresetQueryPanelProps {
  presets: PresetQuery[];
  table: string;
  activePreset: PresetQuery | null;
  onExecute: (query: string, args: Record<string, string>, preset: PresetQuery | null) => void;
  onClear: () => void;
  onOpenPopup: (preset: PresetQuery) => void;
}
const PresetQueryPanel: React.FC<PresetQueryPanelProps> = ({ presets, table, activePreset, onExecute, onClear, onOpenPopup }) => {
  const defaultQuery: PresetQuery = {
    index: -1,
    name: 'Select All',
    query: `SELECT * FROM ${table}`,
    arguments: [],
  };
  const allPresets = [defaultQuery, ...presets];
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const filtered = allPresets.filter(
    (p) => p.name.toLowerCase().includes(search.toLowerCase())
  );
  const handleSelect = (preset: PresetQuery) => {
    setOpen(false);
    setSearch('');
    if (preset.index === -1) {
      // Select All — execute immediately, no popup needed
      onExecute(preset.query, {}, null);
    } else {
      onOpenPopup(preset);
    }
  };
  const handleClearPreset = () => {
    onClear();
    onExecute(defaultQuery.query, {}, null);
  };
  return (
    <div>
      <div className="preset-dropdown" style={{ display: 'inline-block' }}>
        <button className="secondary" onClick={() => setOpen(!open)}>
          Preset Queries
        </button>
        {activePreset && (
          <button
            className="secondary"
            onClick={handleClearPreset}
            style={{ marginLeft: 4, fontSize: 11 }}
            title="Clear active preset query"
          >
            ✕ Clear
          </button>
        )}
        {open && (
          <div className="preset-dropdown-menu">
            <input
              className="search-input"
              placeholder="Search presets..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              autoFocus
            />
            {filtered.map((p) => (
              <div
                key={p.index}
                className={`preset-dropdown-item${activePreset && activePreset.index === p.index ? ' active' : ''}`}
                onClick={() => handleSelect(p)}
              >
                {p.name}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
export default PresetQueryPanel;
