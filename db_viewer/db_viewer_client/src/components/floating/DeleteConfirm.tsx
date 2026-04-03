import React, { useState, useEffect } from 'react';
import { api } from '../../api/client';
import { Row } from '../../types';

interface DeleteConfirmProps {
  client: string;
  table: string;
  row: Row;
  onDeleted?: () => void;
}

const DeleteConfirm: React.FC<DeleteConfirmProps> = ({ client, table, row, onDeleted }) => {
  const [deleteQuery, setDeleteQuery] = useState('');
  const [deleting, setDeleting] = useState(false);
  const [message, setMessage] = useState('');

  useEffect(() => {
    const rowid = String(row['ROWID'] || '');
    if (rowid) {
      api.buildDeleteQuery(client, table, rowid)
        .then((res) => setDeleteQuery(res.query))
        .catch(() => setDeleteQuery('-- error building query --'));
    }
  }, [client, table, row]);

  const handleDelete = async () => {
    setDeleting(true);
    setMessage('');
    try {
      await api.deleteRow(client, table, { rowid: String(row['ROWID'] || '') });
      setMessage('✓ Deleted successfully');
      setTimeout(() => { if (onDeleted) onDeleted(); }, 800);
    } catch (e: any) {
      setMessage(`Error: ${e.message}`);
    } finally {
      setDeleting(false);
    }
  };

  // Show row data summary
  const rowEntries = Object.entries(row).filter(
    ([k]) => k !== 'ROWID' && k !== '__blob_columns'
  );

  return (
    <div className="delete-confirm">
      <div className="delete-warning">
        ⚠️ Are you sure you want to delete this row?
      </div>

      <div className="delete-row-summary">
        <div className="delete-summary-label">Row Data:</div>
        <div className="delete-row-data">
          {rowEntries.map(([k, v]) => (
            <div key={k} className="delete-row-field">
              <span className="delete-field-name">{k}:</span>
              <span className="delete-field-value">
                {v === null || v === undefined ? (
                  <span className="null-value">null</span>
                ) : String(v)}
              </span>
            </div>
          ))}
        </div>
      </div>

      <div className="delete-query-preview">
        <div className="delete-query-label">DELETE Query:</div>
        <pre className="delete-query-sql">{deleteQuery || '-- loading... --'}</pre>
      </div>

      <div className="delete-actions">
        <button className="delete-btn-confirm" onClick={handleDelete} disabled={deleting}>
          {deleting ? 'Deleting...' : '🗑 Confirm Delete'}
        </button>
        {message && (
          <span className={`delete-message ${message.startsWith('✓') ? 'success' : 'error'}`}>
            {message}
          </span>
        )}
      </div>
    </div>
  );
};

export default DeleteConfirm;
