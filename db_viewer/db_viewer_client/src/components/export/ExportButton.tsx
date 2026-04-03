import React, { useState } from 'react';
import { api } from '../../api/client';

interface ExportButtonProps {
  client: string;
  table: string;
  queryParams: Record<string, string>;
}

const ExportButton: React.FC<ExportButtonProps> = ({ client, table, queryParams }) => {
  const [open, setOpen] = useState(false);

  const handleExport = (type: string, format: string) => {
    const params = { ...queryParams, type, format };
    const url = api.exportTable(client, table, params);
    window.open(url, '_blank');
    setOpen(false);
  };

  return (
    <div className="preset-dropdown" style={{ display: 'inline-block' }}>
      <button className="secondary" onClick={() => setOpen(!open)}>
        Export
      </button>
      {open && (
        <div className="preset-dropdown-menu">
          <div className="preset-dropdown-item" onClick={() => handleExport('current', 'csv')}>
            Current View → CSV
          </div>
          <div className="preset-dropdown-item" onClick={() => handleExport('current', 'json')}>
            Current View → JSON
          </div>
          <div className="preset-dropdown-item" onClick={() => handleExport('full', 'csv')}>
            Full Table → CSV
          </div>
          <div className="preset-dropdown-item" onClick={() => handleExport('full', 'json')}>
            Full Table → JSON
          </div>
        </div>
      )}
    </div>
  );
};

export default ExportButton;
