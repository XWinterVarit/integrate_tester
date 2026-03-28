import React from 'react';

interface SidebarProps {
  clients: { name: string; schema: string }[];
  tables: string[];
  selectedClient: string;
  selectedTable: string;
  onSelectClient: (name: string) => void;
  onSelectTable: (name: string) => void;
}

const Sidebar: React.FC<SidebarProps> = ({
  clients, tables, selectedClient, selectedTable,
  onSelectClient, onSelectTable,
}) => {
  return (
    <div className="sidebar">
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
          {tables.map((t) => (
            <div
              key={t}
              className={`sidebar-item ${selectedTable === t ? 'active' : ''}`}
              onClick={() => onSelectTable(t)}
            >
              {t}
            </div>
          ))}
        </>
      )}
    </div>
  );
};

export default Sidebar;
