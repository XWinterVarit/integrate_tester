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
  description?: string;
}

export const ColumnInfoContent: React.FC<ColumnInfoContentProps> = ({
  columnName, columnMeta, constraints, description,
}) => {
  const relatedConstraints = constraints.filter(
    (c) => String(c.COLUMNS || '').split(',').map((s: string) => s.trim()).includes(columnName)
  );

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

      {description && (
        <div style={{ marginTop: 8, borderTop: '1px solid var(--border)', paddingTop: 8 }}>
          <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: 0.5, marginBottom: 4 }}>
            Description
          </div>
          <div style={{ fontSize: 13, whiteSpace: 'pre-wrap', lineHeight: 1.5, color: 'var(--text-primary)' }}>
            {description}
          </div>
        </div>
      )}
    </div>
  );
};

interface FieldEditContentProps {
  columnName: string;
  value: any;
  client: string;
  table: string;
  row: Record<string, any>;
  onSaved?: (columnName: string, newValue: string, row: Record<string, any>) => void;
}

const BLOB_DISPLAY_LIMIT = 500 * 1024; // 500 KB — show full content below this
const BLOB_SAVE_LIMIT = 10 * 1024 * 1024; // 10 MB — disable save above this

function base64ByteLength(b64: string): number {
  const s = b64.endsWith('...') ? b64.slice(0, -3) : b64;
  const padding = (s.match(/=+$/) || [''])[0].length;
  return Math.floor(s.length * 3 / 4) - padding;
}

function formatMB(bytes: number): string {
  return (bytes / 1048576).toFixed(2) + ' MB';
}

export const FieldEditContent: React.FC<FieldEditContentProps> = ({
  columnName, value, client, table, row, onSaved,
}) => {
  const [editValue, setEditValue] = useState(String(value ?? ''));
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [updateQuery, setUpdateQuery] = useState('');

  const rowid = String(row['ROWID'] || '');
  const blobCols: string[] = row['__blob_columns'] || [];
  const isBlob = blobCols.includes(columnName);

  React.useEffect(() => {
    if (isBlob || !rowid) return;
    api.buildUpdateQuery(client, table, { column: columnName, value: editValue, rowid })
      .then((res) => setUpdateQuery(res.query))
      .catch(() => setUpdateQuery('-- error building query --'));
  }, [client, table, columnName, editValue, rowid, isBlob]);

  const blobStr = String(value ?? '');
  const isTruncated = blobStr.endsWith('...');
  const approxBytes = blobStr ? base64ByteLength(blobStr) : 0;

  const handleSave = async () => {
    setSaving(true);
    setMessage('');
    try {
      await api.updateCell(client, table, {
        column: columnName,
        value: editValue,
        rowid,
      });
      if (onSaved) {
        onSaved(columnName, editValue, row);
      }
    } catch (e: any) {
      setMessage(`Error: ${e.message}`);
    } finally {
      setSaving(false);
    }
  };

  const handleBlobDownload = () => {
    const url = api.blobDownloadUrl(client, table, {
      column: columnName,
      rowid: String(row['ROWID'] || ''),
    });
    window.open(url, '_blank');
  };

  const handleBlobUpload = async () => {
    if (!uploadFile) return;
    setSaving(true);
    setMessage('');
    try {
      const buf = await uploadFile.arrayBuffer();
      await api.uploadBlob(client, table, { column: columnName, rowid: String(row['ROWID'] || '') }, buf);
      setMessage('✓ Uploaded successfully');
      setTimeout(() => { if (onSaved) onSaved(columnName, '', row); }, 800);
    } catch (e: any) {
      setMessage(`Error: ${e.message}`);
    } finally {
      setSaving(false);
    }
  };

  const uploadFileTooLarge = uploadFile ? uploadFile.size > BLOB_SAVE_LIMIT : false;

  return (
    <div>
      <div style={{ marginBottom: 12 }}>
        <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: 0.5 }}>
          Column
        </div>
        <div style={{ fontSize: 16, fontWeight: 600 }}>
          {columnName}
          {isBlob && <span style={{ fontSize: 11, color: 'var(--text-secondary)', marginLeft: 8 }}>(BLOB)</span>}
        </div>
      </div>
      {isBlob ? (
        <div>
          <div style={{ marginBottom: 8, fontSize: 12, color: 'var(--text-secondary)' }}>
            Size: <strong>{approxBytes > 0 ? formatMB(approxBytes) + (isTruncated ? ' (truncated preview)' : '') : '(empty)'}</strong>
          </div>
          {blobStr && (
            <div style={{ marginBottom: 12 }}>
              <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: 4 }}>
                {isTruncated ? 'Preview (truncated)' : 'Content (base64)'}
              </div>
              <div style={{
                fontFamily: 'monospace', fontSize: 12, padding: 8,
                background: 'var(--bg-secondary, #f5f5f5)', borderRadius: 4,
                wordBreak: 'break-all', maxHeight: 120, overflow: 'auto',
              }}>
                {isTruncated ? blobStr.slice(0, -3) + ' ***truncate***' : blobStr}
              </div>
            </div>
          )}
          <div style={{ display: 'flex', gap: 8, marginBottom: 12, flexWrap: 'wrap', alignItems: 'center' }}>
            <button className="secondary" onClick={handleBlobDownload}>⬇ Download</button>
          </div>
          <div style={{ borderTop: '1px solid var(--border)', paddingTop: 12 }}>
            <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: 6 }}>
              Upload new file
            </div>
            <input
              type="file"
              style={{ marginBottom: 8, fontSize: 12 }}
              onChange={(e) => setUploadFile(e.target.files?.[0] ?? null)}
            />
            {uploadFile && (
              <div style={{ fontSize: 12, marginBottom: 8, color: uploadFileTooLarge ? 'var(--danger)' : 'var(--text-secondary)' }}>
                {uploadFile.name} — {formatMB(uploadFile.size)}
                {uploadFileTooLarge && ' (exceeds 10 MB limit)'}
              </div>
            )}
            <button onClick={handleBlobUpload} disabled={saving || !uploadFile || uploadFileTooLarge}>
              {saving ? 'Uploading...' : '⬆ Upload & Save'}
            </button>
            {message && (
              <span style={{ fontSize: 12, marginLeft: 8, color: message.startsWith('✓') ? 'var(--success, green)' : 'var(--danger)' }}>
                {message}
              </span>
            )}
          </div>
        </div>
      ) : (
        <>
          <div style={{ marginBottom: 12 }}>
            <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: 4 }}>
              Value
            </div>
            <textarea
              className="edit-cell-input"
              style={{ width: '100%', minHeight: 80, resize: 'vertical', fontFamily: 'monospace', fontSize: 13 }}
              value={editValue}
              onChange={(e) => setEditValue(e.target.value)}
            />
          </div>
          <div className="delete-query-preview" style={{ marginBottom: 12 }}>
            <div className="delete-query-label">UPDATE Query:</div>
            <pre className="delete-query-sql">{updateQuery || '-- loading... --'}</pre>
          </div>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <button onClick={handleSave} disabled={saving}>
              {saving ? 'Saving...' : 'Save'}
            </button>
            {message && (
              <span style={{ fontSize: 12, color: message.startsWith('✓') ? 'var(--success, green)' : 'var(--danger)' }}>
                {message}
              </span>
            )}
          </div>
        </>
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
          <div className="label">Size (MB)</div>
          <div className="value">{sizeInfo.BYTES != null ? (sizeInfo.BYTES / 1048576).toFixed(2) + ' MB' : '—'}</div>
        </div>
        <div className="info-card">
          <div className="label">Blocks</div>
          <div className="value">{sizeInfo.BLOCKS ?? '—'}</div>
        </div>
      </div>

      {constraints.length > 0 && (
        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: 8, fontWeight: 600 }}>
            Constraints
          </div>
          <table className="data-table" style={{ fontSize: 12 }}>
            <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Columns</th>
                <th>Status</th>
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
        </div>
      )}

      {indexes.length > 0 && (
        <div>
          <div style={{ fontSize: 11, color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: 8, fontWeight: 600 }}>
            Indexes
          </div>
          <table className="data-table" style={{ fontSize: 12 }}>
            <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Uniqueness</th>
                <th>Columns</th>
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
        </div>
      )}
    </div>
  );
};
