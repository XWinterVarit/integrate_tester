import React, { useState, useEffect, useCallback } from 'react';
import { api } from '../../api/client';
import { PresetFilter } from '../../types';

interface FieldDescEditorProps {
  client: string;
  table: string;
  columns: string[];
  presetFilters: PresetFilter[];
  onClose: () => void;
}

const FieldDescEditor: React.FC<FieldDescEditorProps> = ({
  client, table, columns, presetFilters, onClose,
}) => {
  const [descriptions, setDescriptions] = useState<Record<string, string>>({});
  const [selectedPreset, setSelectedPreset] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setLoading(true);
    api.getFieldDescriptions(client, table)
      .then((data) => setDescriptions(data || {}))
      .catch(() => setDescriptions({}))
      .finally(() => setLoading(false));
  }, [client, table]);

  const handleChange = useCallback((col: string, value: string) => {
    setDescriptions((prev) => ({ ...prev, [col]: value }));
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      const cleaned: Record<string, string> = {};
      for (const [k, v] of Object.entries(descriptions)) {
        if (v.trim()) cleaned[k] = v;
      }
      await api.saveFieldDescriptions(client, table, cleaned);
      onClose();
    } catch {
      // stay open on error
    } finally {
      setSaving(false);
    }
  };

  const filteredColumns = selectedPreset
    ? (() => {
        const preset = presetFilters.find((f) => f.name === selectedPreset);
        if (!preset) return columns;
        const presetCols = preset.columns
          .filter((c) => !c.startsWith('<'))
          .filter((c) => columns.includes(c));
        return presetCols.length > 0 ? presetCols : columns;
      })()
    : columns;

  return (
    <div className="field-desc-overlay" onClick={onClose}>
      <div className="field-desc-modal" onClick={(e) => e.stopPropagation()}>
        <div className="field-desc-header">
          <span className="field-desc-title">Field Descriptions — {table}</span>
          <button className="field-desc-close" onClick={onClose}>✕</button>
        </div>

        <div className="field-desc-filter">
          <select
            value={selectedPreset}
            onChange={(e) => setSelectedPreset(e.target.value)}
            className="field-desc-select"
          >
            <option value="">All columns</option>
            {presetFilters.map((f) => (
              <option key={f.name} value={f.name}>{f.name}</option>
            ))}
          </select>
        </div>

        <div className="field-desc-list">
          {loading ? (
            <div style={{ padding: 20, color: 'var(--text-secondary)', textAlign: 'center' }}>Loading…</div>
          ) : (
            filteredColumns.map((col) => (
              <div key={col} className="field-desc-row">
                <label className="field-desc-label">{col}</label>
                <textarea
                  className="field-desc-textarea"
                  placeholder="Add description…"
                  value={descriptions[col] || ''}
                  onChange={(e) => handleChange(col, e.target.value)}
                  rows={2}
                />
              </div>
            ))
          )}
        </div>

        <div className="field-desc-footer">
          <button className="secondary" onClick={onClose}>Cancel</button>
          <button onClick={handleSave} disabled={saving || loading}>
            {saving ? 'Saving…' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default FieldDescEditor;
