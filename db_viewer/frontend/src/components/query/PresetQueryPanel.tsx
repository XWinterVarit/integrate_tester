import React, { useState } from 'react';
import { PresetQuery } from '../../types';

interface PresetQueryPanelProps {
  presets: PresetQuery[];
  table: string;
  activePreset: PresetQuery | null;
  onExecute: (query: string, args: Record<string, string>, preset: PresetQuery | null) => void;
  onClear: () => void;
}

const PresetQueryPanel: React.FC<PresetQueryPanelProps> = ({ presets, table, activePreset, onExecute, onClear }) => {
  const defaultQuery: PresetQuery = {
    index: -1,
    name: 'Select All',
    query: `SELECT * FROM ${table}`,
    arguments: [],
  };
  const allPresets = [defaultQuery, ...presets];
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [selected, setSelected] = useState<PresetQuery | null>(null);
  const [argValues, setArgValues] = useState<Record<string, string>>({});

  const filtered = allPresets.filter(
    (p) => p.name.toLowerCase().includes(search.toLowerCase())
  );

  const handleSelect = (preset: PresetQuery) => {
    setSelected(preset);
    setOpen(false);
    const defaults: Record<string, string> = {};
    preset.arguments.forEach((a) => { defaults[a.name] = ''; });
    setArgValues(defaults);
  };

  const getFinalQuery = () => {
    if (!selected) return '';
    let q = selected.query;
    for (const [key, val] of Object.entries(argValues)) {
      q = q.replace(new RegExp(`:${key}`, 'g'), `'${val}'`);
    }
    return q;
  };

  const handleExecute = () => {
    if (!selected) return;
    const isDefault = selected.index === -1;
    onExecute(selected.query, argValues, isDefault ? null : selected);
    setSelected(null);
  };

  const handleClearPreset = () => {
    onClear();
    // Re-execute default select all
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

      {selected && (
        <div className="preset-query-panel" style={{ marginTop: 8 }}>
          <div style={{ fontWeight: 600, marginBottom: 8 }}>{selected.name}</div>

          {selected.arguments.length > 0 && (
            <div className="args-section">
              <div style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', color: 'var(--text-secondary)', marginBottom: 8 }}>
                Edit Parameters
              </div>
              {selected.arguments.map((arg) => (
                <div key={arg.name} className="arg-row">
                  <label>{arg.name} ({arg.type})</label>
                  <input
                    value={argValues[arg.name] || ''}
                    onChange={(e) => setArgValues({ ...argValues, [arg.name]: e.target.value })}
                    placeholder={arg.description}
                    style={{ flex: 1 }}
                  />
                  <span className="arg-desc">{arg.description}</span>
                </div>
              ))}
            </div>
          )}

          <div style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', color: 'var(--text-secondary)', marginTop: 12 }}>
            Final Query Preview
          </div>
          <div className="final-query">{getFinalQuery()}</div>

          <div style={{ marginTop: 12, display: 'flex', gap: 8 }}>
            <button onClick={handleExecute}>Execute</button>
            <button className="secondary" onClick={() => setSelected(null)}>Cancel</button>
          </div>
        </div>
      )}
    </div>
  );
};

export default PresetQueryPanel;
