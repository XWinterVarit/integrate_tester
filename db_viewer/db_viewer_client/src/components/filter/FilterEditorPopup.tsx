import React, { useState, useRef, useCallback } from 'react';
import { PresetFilter } from '../../types';

interface FilterEditorPopupProps {
  initial: PresetFilter | null;
  allColumns: string[];
  onSave: (data: { name: string; details: string; columns: string[] }) => void;
  onClose: () => void;
}

const FilterEditorPopup: React.FC<FilterEditorPopupProps> = ({ initial, allColumns, onSave, onClose }) => {
  const [name, setName] = useState(initial?.name || '');
  const [details, setDetails] = useState(initial?.details || '');
  const [activeItems, setActiveItems] = useState<string[]>(initial?.columns || []);
  const [dragIdx, setDragIdx] = useState<number | null>(null);
  const [dragOverIdx, setDragOverIdx] = useState<number | null>(null);
  const [showTrash, setShowTrash] = useState(false);
  const rightRef = useRef<HTMLDivElement | null>(null);

  // Available = all real columns not already in active (excluding special tags)
  const realActive = new Set(activeItems.filter((c) => !c.startsWith('<')));
  const available = allColumns.filter((c) => !realActive.has(c));

  const addToActive = (col: string) => {
    setActiveItems((prev) => [...prev, col]);
  };

  const removeFromActive = (idx: number) => {
    setActiveItems((prev) => prev.filter((_, i) => i !== idx));
  };

  const addSpecial = (tag: string) => {
    if (tag === '<COMMENTARY>') {
      const label = prompt('Commentary text:');
      if (!label) return;
      setActiveItems((prev) => [...prev, `<COMMENTARY> ${label}`]);
    } else {
      setActiveItems((prev) => [...prev, tag]);
    }
  };

  // Drag reorder within active panel
  const handleDragStart = useCallback((idx: number) => {
    setDragIdx(idx);
    setShowTrash(true);
  }, []);

  // Top/bottom half detection: dragOverIdx = idx (insert before) or idx+1 (insert after)
  const handleDragOver = useCallback((e: React.DragEvent, idx: number) => {
    e.preventDefault();
    if (dragIdx === null) return;
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    setDragOverIdx(e.clientY > rect.top + rect.height / 2 ? idx + 1 : idx);
  }, [dragIdx]);

  // insertBefore = gap index (0..N)
  const handleDrop = useCallback((e: React.DragEvent, insertBefore: number) => {
    e.preventDefault();
    e.stopPropagation();
    if (dragIdx === null) {
      setDragIdx(null);
      setDragOverIdx(null);
      setShowTrash(false);
      return;
    }
    setActiveItems((prev) => {
      const items = [...prev];
      const [moved] = items.splice(dragIdx, 1);
      const adjustedIdx = dragIdx < insertBefore ? insertBefore - 1 : insertBefore;
      items.splice(adjustedIdx, 0, moved);
      return items;
    });
    setDragIdx(null);
    setDragOverIdx(null);
    setShowTrash(false);
  }, [dragIdx]);

  const handleDragEnd = useCallback(() => {
    setDragIdx(null);
    setDragOverIdx(null);
    setShowTrash(false);
  }, []);

  const handleTrashDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    if (dragIdx !== null) {
      removeFromActive(dragIdx);
    }
    setDragIdx(null);
    setShowTrash(false);
  }, [dragIdx]);

  const handleSave = () => {
    if (!name.trim()) return;
    if (activeItems.length === 0) return;
    onSave({ name: name.trim(), details: details.trim(), columns: activeItems });
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-editor" onClick={(e) => e.stopPropagation()}>
        <div className="modal-editor-header">
          <button className="modal-close-btn" onClick={onClose}>✕</button>
          <span className="modal-editor-title">{initial ? 'Edit Filter' : 'New Filter'}</span>
        </div>

        <div className="modal-editor-fields">
          <label>Name</label>
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Filter name" autoFocus />
          <label>Description</label>
          <input value={details} onChange={(e) => setDetails(e.target.value)} placeholder="Optional description" />
        </div>

        <div className="filter-editor-panels">
          {/* Left: Available Fields */}
          <div className="filter-editor-panel">
            <div className="filter-editor-panel-title">Available Fields</div>
            <div className="filter-editor-list">
              {available.map((col) => (
                <div key={col} className="filter-editor-item" onClick={() => addToActive(col)}>
                  {col}
                </div>
              ))}
            </div>
            <div className="filter-editor-specials">
              <button className="secondary small" onClick={() => addSpecial('<SPACE>')}>+ &lt;SPACE&gt;</button>
              <button className="secondary small" onClick={() => addSpecial('<COMMENTARY>')}>+ &lt;COMMENTARY&gt;</button>
              <button className="secondary small" onClick={() => addSpecial('<THE REST>')}>+ &lt;THE REST&gt;</button>
            </div>
          </div>

          {/* Right: Active Filters */}
          <div className="filter-editor-panel">
            <div className="filter-editor-panel-title">Active Filters</div>
            <div className="filter-editor-list" ref={rightRef}>
              {activeItems.map((item, idx) => (
                <div
                  key={`${item}-${idx}`}
                  className={`filter-editor-item active-item${dragOverIdx === idx && dragIdx !== idx ? ' drag-over' : ''}${dragIdx === idx ? ' dragging' : ''}`}
                  draggable
                  onDragStart={() => handleDragStart(idx)}
                  onDragOver={(e) => handleDragOver(e, idx)}
                  onDrop={(e) => handleDrop(e, dragOverIdx ?? idx)}
                  onDragEnd={handleDragEnd}
                >
                  <span className="drag-handle">⠿</span>
                  <span className="item-label">{item}</span>
                  <span className="remove-btn" onClick={(e) => { e.stopPropagation(); removeFromActive(idx); }}>✕</span>
                </div>
              ))}
              {/* End sentinel: drop zone after the last item */}
              {dragIdx !== null && (
                <div
                  className={`filter-editor-drop-end${dragOverIdx === activeItems.length ? ' active' : ''}`}
                  onDragOver={e => { e.preventDefault(); setDragOverIdx(activeItems.length); }}
                  onDrop={e => handleDrop(e, activeItems.length)}
                />
              )}
            </div>
            {showTrash && (
              <div
                className="filter-editor-trash"
                onDragOver={(e) => e.preventDefault()}
                onDrop={handleTrashDrop}
              >
                🗑 Drop here to remove
              </div>
            )}
          </div>
        </div>

        <div className="modal-editor-footer">
          <button className="secondary" onClick={onClose}>Cancel</button>
          <button onClick={handleSave} disabled={!name.trim() || activeItems.length === 0}>Save</button>
        </div>
      </div>
    </div>
  );
};

export default FilterEditorPopup;
