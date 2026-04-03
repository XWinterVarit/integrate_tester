import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { api } from '../../api/client';
import { Row } from '../../types';

interface ColumnMeta {
  COLUMN_NAME: string;
  DATA_TYPE: string;
  DATA_LENGTH: number;
  NULLABLE: string;
  DATA_DEFAULT: string | null;
}

interface InsertFormProps {
  client: string;
  table: string;
  columnMeta: ColumnMeta[];
  columns: string[];
  prefillRow?: Row | null;
  onInserted?: () => void;
}

const BLOB_UPLOAD_LIMIT = 10 * 1024 * 1024; // 10 MB

function formatMBInsert(bytes: number): string {
  return (bytes / 1048576).toFixed(2) + ' MB';
}

const InsertForm: React.FC<InsertFormProps> = ({
  client, table, columnMeta, columns, prefillRow, onInserted,
}) => {
  const orderedMeta = useMemo(() => columns.length > 0
    ? columns
        .map((c) => columnMeta.find((m) => m.COLUMN_NAME === c))
        .filter((m): m is ColumnMeta => m !== undefined)
    : columnMeta, [columns, columnMeta]);

  const [values, setValues] = useState<Record<string, string>>({});
  const [blobFiles, setBlobFiles] = useState<Record<string, File | null>>({});
  const [insertQuery, setInsertQuery] = useState('');
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');

  useEffect(() => {
    const init: Record<string, string> = {};
    for (const m of orderedMeta) {
      if (prefillRow && prefillRow[m.COLUMN_NAME] !== null && prefillRow[m.COLUMN_NAME] !== undefined) {
        init[m.COLUMN_NAME] = String(prefillRow[m.COLUMN_NAME]);
      } else {
        init[m.COLUMN_NAME] = '';
      }
    }
    setValues(init);
  }, [orderedMeta, prefillRow]);

  const buildPreview = useCallback(async () => {
    const cols: string[] = [];
    const vals: string[] = [];
    for (const m of orderedMeta) {
      const v = values[m.COLUMN_NAME];
      if (v !== undefined && v !== '') {
        cols.push(m.COLUMN_NAME);
        vals.push(v);
      }
    }
    if (cols.length === 0) {
      setInsertQuery('');
      return;
    }
    try {
      const res = await api.buildInsertQuery(client, table, { columns: cols, values: vals });
      setInsertQuery(res.query);
    } catch {
      setInsertQuery('-- error building query preview --');
    }
  }, [client, table, orderedMeta, values]);

  useEffect(() => {
    const timer = setTimeout(buildPreview, 300);
    return () => clearTimeout(timer);
  }, [buildPreview]);

  const isBlobCol = (m: ColumnMeta) => m.DATA_TYPE === 'BLOB' || m.DATA_TYPE === 'RAW' || m.DATA_TYPE === 'LONG RAW';

  const hasBlobTooLarge = orderedMeta.some((m) => {
    if (!isBlobCol(m)) return false;
    const f = blobFiles[m.COLUMN_NAME];
    return f ? f.size > BLOB_UPLOAD_LIMIT : false;
  });

  const handleSave = async () => {
    const cols: string[] = [];
    const vals: string[] = [];
    for (const m of orderedMeta) {
      if (isBlobCol(m)) continue; // blobs handled separately after insert
      const v = values[m.COLUMN_NAME];
      if (v !== undefined && v !== '') {
        cols.push(m.COLUMN_NAME);
        vals.push(v);
      }
    }
    if (cols.length === 0) {
      setMessage('No values to insert');
      return;
    }
    setSaving(true);
    setMessage('');
    try {
      const result = await api.insertRow(client, table, { columns: cols, values: vals });
      // Upload blob files if any
      const rowid = (result as any)?.rowid;
      if (rowid) {
        for (const m of orderedMeta) {
          if (!isBlobCol(m)) continue;
          const f = blobFiles[m.COLUMN_NAME];
          if (f) {
            const buf = await f.arrayBuffer();
            await api.uploadBlob(client, table, { column: m.COLUMN_NAME, rowid }, buf);
          }
        }
      }
      setMessage('✓ Inserted successfully');
      setTimeout(() => { if (onInserted) onInserted(); }, 800);
    } catch (e: any) {
      setMessage(`Error: ${e.message}`);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="insert-form">
      <div className="insert-form-fields">
        {orderedMeta.map((m) => {
          const isNullable = m.NULLABLE === 'Y';
          return (
            <div className="insert-field-card" key={m.COLUMN_NAME}>
              <div className="insert-field-header">
                <span className="insert-field-name">{m.COLUMN_NAME}</span>
                <span className="insert-field-meta">
                  <span className="insert-type-badge">{m.DATA_TYPE}</span>
                  <span className="insert-size-badge">({m.DATA_LENGTH})</span>
                  {isNullable
                    ? <span className="nullable-yes">Optional</span>
                    : <span className="nullable-no">Required</span>
                  }
                </span>
              </div>
              {isBlobCol(m) ? (
                <div className="insert-blob-field">
                  <input
                    type="file"
                    style={{ fontSize: 12, marginBottom: 4 }}
                    onChange={(e) => setBlobFiles((prev) => ({ ...prev, [m.COLUMN_NAME]: e.target.files?.[0] ?? null }))}
                  />
                  {blobFiles[m.COLUMN_NAME] && (() => {
                    const f = blobFiles[m.COLUMN_NAME]!;
                    const tooLarge = f.size > BLOB_UPLOAD_LIMIT;
                    return (
                      <div style={{ fontSize: 11, color: tooLarge ? 'var(--danger)' : 'var(--text-secondary)' }}>
                        {f.name} — {formatMBInsert(f.size)}{tooLarge ? ' (exceeds 10 MB limit)' : ''}
                      </div>
                    );
                  })()}
                </div>
              ) : (
                <input
                  type="text"
                  className="insert-input"
                  placeholder={isNullable ? 'NULL (leave empty)' : 'Required'}
                  value={values[m.COLUMN_NAME] || ''}
                  onChange={(e) => setValues((prev) => ({ ...prev, [m.COLUMN_NAME]: e.target.value }))}
                />
              )}
              {m.DATA_DEFAULT && (
                <div className="insert-field-default">Default: {m.DATA_DEFAULT}</div>
              )}
            </div>
          );
        })}
      </div>

      <div className="insert-query-preview">
        <div className="insert-query-label">Final INSERT Query</div>
        <pre className="insert-query-sql">{insertQuery || '-- fill in values above --'}</pre>
      </div>

      <div className="insert-actions">
        <button className="secondary" onClick={handleSave} disabled={saving || hasBlobTooLarge}>
          {saving ? 'Inserting...' : 'Insert Row'}
        </button>
        {message && (
          <span className={`insert-message ${message.startsWith('✓') ? 'success' : 'error'}`}>
            {message}
          </span>
        )}
      </div>
    </div>
  );
};

export default InsertForm;
