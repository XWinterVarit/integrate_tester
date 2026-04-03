import React, { useState, useRef, useEffect, useCallback } from 'react';
import { PresetFilter } from '../../types';
import { api } from '../../api/client';
import FilterEditorPopup from './FilterEditorPopup';

interface FilterDropdownProps {
  filters: PresetFilter[];
  activeFilter: PresetFilter | null;
  onSelect: (filter: PresetFilter | null) => void;
  client: string;
  table: string;
  allColumns: string[];
  onRefresh: () => void;
}

const FilterDropdown: React.FC<FilterDropdownProps> = ({
  filters, activeFilter, onSelect, client, table, allColumns, onRefresh,
}) => {
  const [open, setOpen] = useState(false);
  const [moreMenu, setMoreMenu] = useState<string | null>(null);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editingFilter, setEditingFilter] = useState<PresetFilter | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const moreRef = useRef<HTMLDivElement | null>(null);

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

  const handleDelete = useCallback(async (name: string) => {
    try {
      await api.deletePresetFilter(client, table, name);
      if (activeFilter?.name === name) onSelect(null);
      onRefresh();
    } catch {}
    setConfirmDelete(null);
    setMoreMenu(null);
    setOpen(false);
  }, [client, table, activeFilter, onSelect, onRefresh]);

  const handleEdit = (filter: PresetFilter) => {
    setEditingFilter(filter);
    setEditorOpen(true);
    setMoreMenu(null);
    setOpen(false);
  };

  const handleDuplicate = (filter: PresetFilter) => {
    setEditingFilter({ ...filter, name: `${filter.name} (copy)` });
    setEditorOpen(true);
    setMoreMenu(null);
    setOpen(false);
  };

  const handleAddNew = () => {
    setEditingFilter(null);
    setEditorOpen(true);
    setOpen(false);
  };

  const handleEditorSave = async (data: { name: string; details: string; columns: string[] }) => {
    try {
      await api.savePresetFilter(client, table, data);
      setEditorOpen(false);
      setEditingFilter(null);
      onRefresh();
    } catch {}
  };

  return (
    <>
      <div className="preset-dropdown" ref={containerRef}>
        <button className="secondary" onClick={() => setOpen(!open)}>
          {activeFilter ? `Filter: ${activeFilter.name}` : 'Column Filter'}
        </button>
        {open && (
          <div className="preset-dropdown-menu">
            {/* Default / No filter */}
            <div
              className={`preset-dropdown-item${!activeFilter ? ' active' : ''}`}
              onClick={() => { onSelect(null); setOpen(false); }}
              style={{ fontStyle: 'italic', color: 'var(--text-secondary)' }}
            >
              Default (show all)
            </div>

            {filters.map((f) => (
              <div
                key={f.name}
                className={`preset-dropdown-item preset-dropdown-item-with-more${activeFilter?.name === f.name ? ' active' : ''}`}
                onClick={() => { onSelect(f); setOpen(false); }}
              >
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 500 }}>{f.name}</div>
                  {f.details && (
                    <div style={{ fontSize: 11, color: 'var(--text-secondary)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{f.details}</div>
                  )}
                </div>
                <div
                  className="more-icon"
                  onClick={(e) => { e.stopPropagation(); setMoreMenu(moreMenu === f.name ? null : f.name); }}
                >
                  ···
                </div>
                {moreMenu === f.name && (
                  <div className="more-popover" ref={moreRef}>
                    <div className="more-popover-item" onClick={(e) => { e.stopPropagation(); handleEdit(f); }}>✏️ Edit</div>
                    <div className="more-popover-item" onClick={(e) => { e.stopPropagation(); handleDuplicate(f); }}>📄 Duplicate</div>
                    <div className="more-popover-item danger" onClick={(e) => { e.stopPropagation(); setConfirmDelete(f.name); setMoreMenu(null); }}>🗑 Delete</div>
                  </div>
                )}
              </div>
            ))}

            <div className="preset-dropdown-divider" />
            <div className="preset-dropdown-item add-new" onClick={handleAddNew}>
              ＋ Add New Filter
            </div>
          </div>
        )}
      </div>

      {/* Delete confirmation */}
      {confirmDelete && (
        <div className="modal-overlay" onClick={() => setConfirmDelete(null)}>
          <div className="modal-box" onClick={(e) => e.stopPropagation()}>
            <p>Delete filter <strong>{confirmDelete}</strong>?</p>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button className="secondary" onClick={() => setConfirmDelete(null)}>Cancel</button>
              <button style={{ background: 'var(--danger)' }} onClick={() => handleDelete(confirmDelete)}>Delete</button>
            </div>
          </div>
        </div>
      )}

      {/* Filter editor popup */}
      {editorOpen && (
        <FilterEditorPopup
          initial={editingFilter}
          allColumns={allColumns}
          onSave={handleEditorSave}
          onClose={() => { setEditorOpen(false); setEditingFilter(null); }}
        />
      )}
    </>
  );
};

export default FilterDropdown;
