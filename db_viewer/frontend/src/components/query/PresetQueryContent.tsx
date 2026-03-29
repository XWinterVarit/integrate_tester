import React, { useState } from 'react';
import { PresetQuery } from '../../types';

interface PresetQueryContentProps {
  preset: PresetQuery;
  onExecute: (query: string, args: Record<string, string>, preset: PresetQuery | null) => void;
  onClose: () => void;
}

const PresetQueryContent: React.FC<PresetQueryContentProps> = ({ preset, onExecute, onClose }) => {
  const [argValues, setArgValues] = useState<Record<string, string>>(() => {
    const defaults: Record<string, string> = {};
    preset.arguments.forEach((a) => { defaults[a.name] = ''; });
    return defaults;
  });

  const getFinalQuery = () => {
    let q = preset.query;
    for (const [key, val] of Object.entries(argValues)) {
      q = q.replace(new RegExp(`:${key}`, 'g'), `'${val}'`);
    }
    return q;
  };

  const handleExecute = () => {
    onExecute(preset.query, argValues, preset);
    onClose();
  };

  return (
    <div style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 12 }}>
      {preset.arguments.length > 0 && (
        <div className="args-section">
          <div style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', color: 'var(--text-secondary)', marginBottom: 8 }}>
            Parameters
          </div>
          {preset.arguments.map((arg) => (
            <div key={arg.name} className="arg-row">
              <label>{arg.name} ({arg.type})</label>
              <input
                value={argValues[arg.name] || ''}
                onChange={(e) => setArgValues({ ...argValues, [arg.name]: e.target.value })}
                placeholder={arg.description}
                onKeyDown={(e) => { if (e.key === 'Enter') handleExecute(); }}
                style={{ flex: 1 }}
              />
            </div>
          ))}
        </div>
      )}
      <div>
        <div style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', color: 'var(--text-secondary)', marginBottom: 6 }}>
          Query Preview
        </div>
        <div className="final-query">{getFinalQuery()}</div>
      </div>
      <div style={{ display: 'flex', gap: 8 }}>
        <button onClick={handleExecute}>Execute</button>
        <button className="secondary" onClick={onClose}>Cancel</button>
      </div>
    </div>
  );
};

export default PresetQueryContent;
