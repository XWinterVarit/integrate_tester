import React, { useState, useRef, useEffect, useCallback } from 'react';
import ReactDOM from 'react-dom';
import { PresetQuery } from '../../types';
import { api } from '../../api/client';
import QueryEditorPopup from './QueryEditorPopup';

interface PresetQueryPanelProps {
  presets: PresetQuery[];
  client: string;
  table: string;
  activePreset: PresetQuery | null;
  onExecute: (query: string, args: Record<string, string>, preset: PresetQuery | null) => void;
  onClear: () => void;
  onOpenPopup: (preset: PresetQuery) => void;
  onRefresh: () => void;
}

const PresetQueryPanel: React.FC<PresetQueryPanelProps> = ({
  presets, client, table, activePreset, onExecute, onClear, onOpenPopup, onRefresh,
}) => {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [moreMenu, setMoreMenu] = useState<string | null>(null);
  const [morePos, setMorePos] = useState<{ top: number; left: number }>({ top: 0, left: 0 });
  const [editorOpen, setEditorOpen] = useState(false);
  const [editingQuery, setEditingQuery] = useState<PresetQuery | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const moreRef = useRef<HTMLDivElement | null>(null);

  const defaultQuery: PresetQuery = {
    name: 'Select All',
    query: `SELECT * FROM ${table}`,
    arguments: [],
  };

  useEffect(() => {
    if (!open) { setMoreMenu(null); return; }
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
        setMoreMenu(null);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  // Close more-popover on outside click
  useEffect(() => {
    if (!moreMenu) return;
    const handler = (e: MouseEvent) => {
      if (moreRef.current && !moreRef.current.contains(e.target as Node)) {
        setMoreMenu(null);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [moreMenu]);

  const filtered = presets.filter(
    (p) => p.name.toLowerCase().includes(search.toLowerCase())
  );

  const handleSelect = (preset: PresetQuery) => {
    setOpen(false);
    setSearch('');
    if (preset.arguments.length === 0) {
      onExecute(preset.query, {}, preset);
    } else {
      onOpenPopup(preset);
    }
  };

  const handleSelectAll = () => {
    setOpen(false);
    setSearch('');
    onExecute(defaultQuery.query, {}, null);
  };

  const handleClearPreset = () => {
    onClear();
    onExecute(defaultQuery.query, {}, null);
  };

  const handleDelete = useCallback(async (name: string) => {
    try {
      await api.deletePresetQuery(client, table, name);
      if (activePreset?.name === name) onClear();
      onRefresh();
    } catch {}
    setConfirmDelete(null);
    setMoreMenu(null);
    setOpen(false);
  }, [client, table, activePreset, onClear, onRefresh]);

  const handleEdit = (preset: PresetQuery) => {
    setEditingQuery(preset);
    setEditorOpen(true);
    setMoreMenu(null);
    setOpen(false);
  };

  const handleDuplicate = (preset: PresetQuery) => {
    setEditingQuery({ ...preset, name: `${preset.name} (copy)` });
    setEditorOpen(true);
    setMoreMenu(null);
    setOpen(false);
  };

  const handleAddNew = () => {
    setEditingQuery(null);
    setEditorOpen(true);
    setOpen(false);
  };

  const handleEditorSave = async (data: { name: string; query: string; arguments: { name: string; type: string; description: string }[] }) => {
    await api.savePresetQuery(client, table, data); // throws on error → QueryEditorPopup catches and shows it
    setEditorOpen(false);
    setEditingQuery(null);
    onRefresh();
  };

  return (
    <>
      <div className="preset-dropdown" style={{ display: 'inline-block' }} ref={containerRef}>
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
            {/* Sticky top: search */}
            <div className="preset-dropdown-header">
              <input
                className="search-input"
                placeholder="Search presets..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                autoFocus
              />
            </div>

            {/* Scrollable list */}
            <div className="preset-dropdown-body">
              {/* Select All default */}
              <div
                className={`preset-dropdown-item${!activePreset ? ' active' : ''}`}
                onClick={handleSelectAll}
                style={{ fontStyle: 'italic', color: 'var(--text-secondary)' }}
              >
                Select All
              </div>

              {filtered.map((p) => (
                <div
                  key={p.name}
                  className={`preset-dropdown-item preset-dropdown-item-with-more${activePreset?.name === p.name ? ' active' : ''}`}
                  onClick={() => handleSelect(p)}
                >
                  <span style={{ flex: 1 }}>{p.name}</span>
                  <div
                    className="more-icon"
                    onClick={(e) => {
                      e.stopPropagation();
                      const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
                      setMorePos({ top: rect.bottom + 2, left: rect.right - 110 });
                      setMoreMenu(moreMenu === p.name ? null : p.name);
                    }}
                  >
                    ···
                  </div>
                  {moreMenu === p.name && ReactDOM.createPortal(
                    <div className="more-popover" ref={moreRef} style={{ position: 'fixed', top: morePos.top, left: morePos.left, zIndex: 9999 }}>
                      <div className="more-popover-item" onClick={(e) => { e.stopPropagation(); handleEdit(p); }}>✏️ Edit</div>
                      <div className="more-popover-item" onClick={(e) => { e.stopPropagation(); handleDuplicate(p); }}>📄 Duplicate</div>
                      <div className="more-popover-item danger" onClick={(e) => { e.stopPropagation(); setConfirmDelete(p.name); setMoreMenu(null); }}>🗑 Delete</div>
                    </div>,
                    document.body
                  )}
                </div>
              ))}
            </div>

            {/* Sticky bottom: add new */}
            <div className="preset-dropdown-footer">
              <div className="preset-dropdown-item add-new" onClick={handleAddNew}>
                ＋ Add New Query
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Delete confirmation */}
      {confirmDelete && (
        <div className="modal-overlay" onClick={() => setConfirmDelete(null)}>
          <div className="modal-box" onClick={(e) => e.stopPropagation()}>
            <p>Delete query <strong>{confirmDelete}</strong>?</p>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button className="secondary" onClick={() => setConfirmDelete(null)}>Cancel</button>
              <button style={{ background: 'var(--danger)' }} onClick={() => handleDelete(confirmDelete)}>Delete</button>
            </div>
          </div>
        </div>
      )}

      {/* Query editor popup */}
      {editorOpen && (
        <QueryEditorPopup
          initial={editingQuery}
          client={client}
          table={table}
          onSave={handleEditorSave}
          onClose={() => { setEditorOpen(false); setEditingQuery(null); }}
        />
      )}
    </>
  );
};

export default PresetQueryPanel;
