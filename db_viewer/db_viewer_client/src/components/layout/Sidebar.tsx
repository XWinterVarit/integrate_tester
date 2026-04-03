import React, { useState, useCallback, useRef, useEffect } from 'react';

interface SidebarProps {
  clients: { name: string; schema: string }[];
  tables: string[];
  selectedClient: string;
  selectedTable: string;
  onSelectClient: (name: string) => void;
  onSelectTable: (name: string) => void;
  onShowTableInfo?: (table: string) => void;
  onShowFieldDesc?: (table: string) => void;
}

const MIN_WIDTH = 160;
const MAX_WIDTH = 400;
const DEFAULT_WIDTH = 240;

const Sidebar: React.FC<SidebarProps> = ({
  clients, tables, selectedClient, selectedTable,
  onSelectClient, onSelectTable, onShowTableInfo, onShowFieldDesc,
}) => {
  const [width, setWidth] = useState(() => {
    const saved = localStorage.getItem('sidebar-width');
    return saved ? Math.max(MIN_WIDTH, Math.min(MAX_WIDTH, Number(saved))) : DEFAULT_WIDTH;
  });
  const [menuTable, setMenuTable] = useState<string | null>(null);
  const [menuPos, setMenuPos] = useState<{ top: number; left: number }>({ top: 0, left: 0 });
  const menuRef = useRef<HTMLDivElement>(null);
  const isResizing = useRef(false);

  // Resize logic
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    isResizing.current = true;
    const startX = e.clientX;
    const startWidth = width;

    const onMouseMove = (ev: MouseEvent) => {
      const newWidth = Math.max(MIN_WIDTH, Math.min(MAX_WIDTH, startWidth + ev.clientX - startX));
      setWidth(newWidth);
    };
    const onMouseUp = () => {
      isResizing.current = false;
      document.removeEventListener('mousemove', onMouseMove);
      document.removeEventListener('mouseup', onMouseUp);
      setWidth((w) => {
        localStorage.setItem('sidebar-width', String(w));
        return w;
      });
    };
    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
  }, [width]);

  // Close context menu on outside click
  useEffect(() => {
    if (!menuTable) return;
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuTable(null);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [menuTable]);

  const handleMoreClick = (e: React.MouseEvent, table: string) => {
    e.preventDefault();
    e.stopPropagation();
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    setMenuPos({ top: rect.bottom + 2, left: rect.left });
    setMenuTable((prev) => (prev === table ? null : table));
  };

  return (
    <>
      <div className="sidebar" style={{ width, minWidth: width }}>
        <div className="sidebar-header">Connections</div>
        {clients.map((c) => (
          <div
            key={c.name}
            className={`sidebar-item ${selectedClient === c.name ? 'active' : ''}`}
            onClick={() => onSelectClient(c.name)}
          >
            {c.name}
            <span style={{ fontSize: 11, color: 'var(--text-secondary)', marginLeft: 6 }}>
              {c.schema}
            </span>
          </div>
        ))}
        {selectedClient && (
          <>
            <div className="sidebar-header">Tables</div>
            {tables.length === 0 && (
              <div style={{ padding: '10px 16px', color: 'var(--text-secondary)', fontSize: 13 }}>
                No tables configured
              </div>
            )}
            {tables.map((t, idx) => {
              if (t === '<SPACE>') {
                return <div key={`space-${idx}`} className="sidebar-space" />;
              }
              if (t.startsWith('<COMMENTARY>')) {
                const label = t.replace('<COMMENTARY>', '').trim();
                return (
                  <div key={`comment-${idx}`} className="sidebar-commentary">
                    {label}
                  </div>
                );
              }
              const tableUrl = `${window.location.pathname}?client=${encodeURIComponent(selectedClient)}&table=${encodeURIComponent(t)}`;
              return (
                <a
                  key={t}
                  href={tableUrl}
                  className={`sidebar-item sidebar-table-item ${selectedTable === t ? 'active' : ''}`}
                  onClick={(e) => {
                    e.preventDefault();
                    onSelectTable(t);
                  }}
                >
                  <span className="sidebar-table-name">{t}</span>
                  <span
                    className="sidebar-more-icon"
                    onClick={(e) => handleMoreClick(e, t)}
                  >
                    ···
                  </span>
                </a>
              );
            })}
          </>
        )}
      </div>
      <div className="pane-resize-handle" onMouseDown={handleMouseDown} />

      {menuTable && (
        <div
          ref={menuRef}
          className="sidebar-context-menu"
          style={{ top: menuPos.top, left: menuPos.left }}
        >
          <div
            className="sidebar-context-item"
            onClick={() => {
              onShowTableInfo?.(menuTable);
              setMenuTable(null);
            }}
          >
            Table Info
          </div>
          <div
            className="sidebar-context-item"
            onClick={() => {
              onShowFieldDesc?.(menuTable);
              setMenuTable(null);
            }}
          >
            Field Desc
          </div>
        </div>
      )}
    </>
  );
};

export default Sidebar;
