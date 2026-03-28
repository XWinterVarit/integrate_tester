import React, { useState } from 'react';
import { api } from '../../api/client';

interface RowJsonContentProps {
  row: Record<string, any>;
}

export const RowJsonContent: React.FC<RowJsonContentProps> = ({ row }) => (
  <pre>{JSON.stringify(row, null, 2)}</pre>
);

interface ColumnInfoContentProps {
  columnName: string;
  columnMeta: Record<string, any> | null;
  constraints: Record<string, any>[];
  client: string;
  table: string;
  row: Record<string, any> | null;
}

export const ColumnInfoContent: React.FC<ColumnInfoContentProps> = ({
  columnName, columnMeta, constraints, client, table, row,
}) => {
  const [editValue, setEditValue] = useState(row ? String(row[columnName] ?? '') : '');
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');

  const relatedConstraints = constraints.filter(
    (c) => String(c.COLUMNS || '').split(',').map((s: string) => s.trim()).includes(columnName)
  );

  const handleSave = async () => {
    if (!row) return;
    setSaving(true);
    setMessage('');
    try {
      const firstCol = Object.keys(row)[0];
      await api.updateCell(client, table, {
        column: columnName,
        value: editValue,
        where_column: firstCol,
        where_value: String(row[firstCol]),
      });
      setMessage('Saved successfully');
    } catch (e: any) {
      setMessage(`Error: ${e.message}`);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>
      <div style={{ marginBottom: 12 }}>
        <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: 0.5 }}>
          Column
        </div>
        <div style={{ fontSize: 16, fontWeight: 600 }}>{columnName}</div>
      </div>

      {columnMeta && (
        <div className="table-info-grid" style={{ marginBottom: 12 }}>
          <div className="info-card">
            <div className="label">Type</div>
            <div className="value" style={{ fontSize: 14 }}>{columnMeta.DATA_TYPE}</div>
          </div>
          <div className="info-card">
            <div className="label">Length</div>
            <div className="value" style={{ fontSize: 14 }}>{columnMeta.DATA_LENGTH}</div>
          </div>
          <div className="info-card">
            <div className="label">Nullable</div>
            <div className="value" style={{ fontSize: 14 }}>{columnMeta.NULLABLE}</div>
          </div>
          <div className="info-card">
            <div className="label">Default</div>
            <div className="value" style={{ fontSize: 14 }}>{columnMeta.DATA_DEFAULT ?? 'None'}</div>
          </div>
        </div>
      )}

      {relatedConstraints.length > 0 && (
        <div style={{ marginBottom: 12 }}>
          <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: 4 }}>
            Constraints
          </div>
          {relatedConstraints.map((c, i) => (
            <div key={i} style={{ fontSize: 12, padding: '4px 0' }}>
              <strong>{c.CONSTRAINT_NAME}</strong> ({c.CONSTRAINT_TYPE}) — {c.STATUS}
            </div>
          ))}
        </div>
      )}

      {row && (
        <div style={{ marginTop: 12 }}>
          <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: 4 }}>
            Edit Value
          </div>
          <input
            className="edit-cell-input"
            value={editValue}
            onChange={(e) => setEditValue(e.target.value)}
          />
          <div style={{ marginTop: 8, display: 'flex', gap: 8, alignItems: 'center' }}>
            <button onClick={handleSave} disabled={saving}>
              {saving ? 'Saving...' : 'Save'}
            </button>
            {message && (
              <span style={{ fontSize: 12, color: message.startsWith('Error') ? 'var(--danger)' : 'var(--success)' }}>
                {message}
              </span>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

interface TableInfoContentProps {
  size: Record<string, any>[];
  constraints: Record<string, any>[];
  indexes: Record<string, any>[];
}

export const TableInfoContent: React.FC<TableInfoContentProps> = ({ size, constraints, indexes }) => {
  const sizeInfo = size[0] || {};

  return (
    <div>
      <div className="table-info-grid" style={{ marginBottom: 16 }}>
        <div className="info-card">
          <div className="label">Rows</div>
          <div className="value">{sizeInfo.ROW_COUNT ?? '—'}</div>
        </div>
        <div className="info-card">
          <div className="label">Size (bytes)</div>
          <div className="value">{sizeInfo.BYTES ?? '—'}</div>
        </div>
        <div className="info-card">
          <div className="label">Blocks</div>
          <div className="value">{sizeInfo.BLOCKS ?? '—'}</div>
        </div>
      </div>

      <div style={{ marginBottom: 16 }}>
        <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: 8 }}>
          Constraints
        </div>
        {constraints.length === 0 ? (
          <div style={{ fontSize: 12, color: 'var(--text-secondary)' }}>None</div>
        ) : (
          <table className="data-table" style={{ fontSize: 12 }}>
            <thead>
              <tr>
                <th>Name</th><th>Type</th><th>Columns</th><th>Status</th>
              </tr>
            </thead>
            <tbody>
              {constraints.map((c, i) => (
                <tr key={i}>
                  <td>{c.CONSTRAINT_NAME}</td>
                  <td>{c.CONSTRAINT_TYPE}</td>
                  <td>{c.COLUMNS}</td>
                  <td>{c.STATUS}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div>
        <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: 8 }}>
          Indexes
        </div>
        {indexes.length === 0 ? (
          <div style={{ fontSize: 12, color: 'var(--text-secondary)' }}>None</div>
        ) : (
          <table className="data-table" style={{ fontSize: 12 }}>
            <thead>
              <tr>
                <th>Name</th><th>Type</th><th>Unique</th><th>Columns</th>
              </tr>
            </thead>
            <tbody>
              {indexes.map((idx, i) => (
                <tr key={i}>
                  <td>{idx.INDEX_NAME}</td>
                  <td>{idx.INDEX_TYPE}</td>
                  <td>{idx.UNIQUENESS}</td>
                  <td>{idx.COLUMNS}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
