import React, { useState, useRef, useEffect } from 'react';
import { PresetFilter } from '../../types';

interface FilterDropdownProps {
  filters: PresetFilter[];
  activeFilter: PresetFilter | null;
  onSelect: (filter: PresetFilter | null) => void;
}

const FilterDropdown: React.FC<FilterDropdownProps> = ({ filters, activeFilter, onSelect }) => {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const containerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  const filtered = filters.filter(
    (f) => f.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="preset-dropdown" ref={containerRef}>
      <button className="secondary" onClick={() => setOpen(!open)}>
        {activeFilter ? `Filter: ${activeFilter.name}` : 'Column Filter'}
      </button>
      {open && (
        <div className="preset-dropdown-menu">
          <input
            className="search-input"
            placeholder="Search filters..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            autoFocus
          />
          <div
            className="preset-dropdown-item"
            onClick={() => { onSelect(null); setOpen(false); }}
            style={{ fontStyle: 'italic', color: 'var(--text-secondary)' }}
          >
            No Filter (show all)
          </div>
          {filtered.map((f) => (
            <div
              key={f.name}
              className="preset-dropdown-item"
              onClick={() => { onSelect(f); setOpen(false); }}
            >
              <div style={{ fontWeight: 500 }}>{f.name}</div>
              {f.details && (
                <div style={{ fontSize: 11, color: 'var(--text-secondary)' }}>{f.details}</div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default FilterDropdown;
