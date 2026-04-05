import React, { useState, useEffect, useRef, useCallback } from 'react';
import { PresetQuery, PresetQueryArg } from '../../types';
import { api } from '../../api/client';
import { EditorView, keymap, placeholder as cmPlaceholder } from '@codemirror/view';
import { EditorState } from '@codemirror/state';
import { sql, PLSQL } from '@codemirror/lang-sql';
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { syntaxHighlighting, defaultHighlightStyle, bracketMatching } from '@codemirror/language';
import { closeBrackets } from '@codemirror/autocomplete';
import { searchKeymap } from '@codemirror/search';

interface QueryEditorPopupProps {
  initial: PresetQuery | null;
  client: string;
  table: string;
  onSave: (data: { name: string; query: string; arguments: PresetQueryArg[] }) => Promise<void>;
  onClose: () => void;
}

const QueryEditorPopup: React.FC<QueryEditorPopupProps> = ({ initial, client, table, onSave, onClose }) => {
  const [name, setName] = useState(initial?.name || '');
  const [query, setQuery] = useState(initial?.query || '');
  const [args, setArgs] = useState<PresetQueryArg[]>(initial?.arguments || []);
  const [syntaxValid, setSyntaxValid] = useState<boolean | null>(null);
  const [syntaxError, setSyntaxError] = useState('');
  const [undefinedArgs, setUndefinedArgs] = useState<string[]>([]);
  const [saveError, setSaveError] = useState('');
  const [saving, setSaving] = useState(false);
  const [argPopup, setArgPopup] = useState<{ index: number; arg: PresetQueryArg } | null>(null);
  const lastSyncRef = useRef(query);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const editorRef = useRef<HTMLDivElement | null>(null);
  const viewRef = useRef<EditorView | null>(null);

  // Debounced validation: sync every 3 seconds after edit
  const validate = useCallback(async (q: string, a: PresetQueryArg[]) => {
    try {
      const result = await api.validateQuery(client, table, { query: q, arguments: a });
      setSyntaxValid(result.valid);
      setSyntaxError(result.error || '');
      setUndefinedArgs(result.undefined_args || []);
    } catch {
      setSyntaxValid(false);
      setSyntaxError('Validation request failed');
    }
  }, [client, table]);

  useEffect(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      if (query !== lastSyncRef.current || true) {
        lastSyncRef.current = query;
        validate(query, args);
      }
    }, 3000);
    return () => { if (timerRef.current) clearTimeout(timerRef.current); };
  }, [query, args, validate]);

  // Initial validation
  useEffect(() => {
    if (query) validate(query, args);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // CodeMirror setup
  useEffect(() => {
    if (!editorRef.current) return;
    const theme = EditorView.theme({
      '&': { height: '100%', fontSize: '13px' },
      '.cm-scroller': { overflow: 'auto', fontFamily: 'SF Mono, Menlo, Monaco, Consolas, monospace' },
      '.cm-content': { padding: '8px 0' },
      '.cm-gutters': { background: 'var(--bg-secondary, #f5f5f7)', border: 'none' },
      '.cm-activeLineGutter': { background: 'var(--bg-tertiary, #e8e8ed)' },
      '.cm-activeLine': { background: 'rgba(0,0,0,0.03)' },
      '&.cm-focused': { outline: 'none' },
    });
    const state = EditorState.create({
      doc: query,
      extensions: [
        keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap]),
        history(),
        bracketMatching(),
        closeBrackets(),
        sql({ dialect: PLSQL, upperCaseKeywords: true }),
        syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
        cmPlaceholder('SELECT * FROM {THIS_TABLE} WHERE ...'),
        theme,
        EditorView.lineWrapping,
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            setQuery(update.state.doc.toString());
          }
        }),
      ],
    });
    const view = new EditorView({ state, parent: editorRef.current });
    viewRef.current = view;
    return () => { view.destroy(); };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleSave = async () => {
    if (!name.trim() || !query.trim()) return;
    if (syntaxValid === false) {
      setSaveError('Cannot save: query has syntax errors. Please fix the query first.');
      return;
    }
    setSaveError('');
    setSaving(true);
    try {
      await onSave({ name: name.trim(), query: query.trim(), arguments: args });
    } catch (err: any) {
      setSaveError(err?.message || 'Save failed');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteArg = (idx: number) => {
    setArgs((prev) => prev.filter((_, i) => i !== idx));
  };

  const handleAddArg = () => {
    setArgPopup({ index: -1, arg: { name: '', type: 'string', description: '' } });
  };

  const handleEditArg = (idx: number) => {
    setArgPopup({ index: idx, arg: { ...args[idx] } });
  };

  const handleArgSave = (arg: PresetQueryArg) => {
    if (!argPopup) return;
    if (argPopup.index === -1) {
      setArgs((prev) => [...prev, arg]);
    } else {
      setArgs((prev) => prev.map((a, i) => i === argPopup.index ? arg : a));
    }
    setArgPopup(null);
  };

  return (
    <div className="modal-overlay">
      <div className="modal-editor modal-editor-wide" onClick={(e) => e.stopPropagation()}>
        <div className="modal-editor-header">
          <button className="modal-close-btn" onClick={onClose}>✕</button>
          <span className="modal-editor-title">{initial ? 'Edit Query' : 'New Query'}</span>
        </div>

        <div className="modal-editor-fields">
          <label>Name</label>
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Query name" autoFocus />
        </div>

        <div className="query-editor-layout">
          {/* Left: SQL Editor */}
          <div className="query-editor-sql">
            <div className="query-editor-sql-label">SQL</div>
            <div ref={editorRef} className="query-editor-cm" />
          </div>

          {/* Right: Syntax + Arguments */}
          <div className="query-editor-right">
            {/* Syntax status */}
            <div className="query-editor-syntax">
              <div className="query-editor-syntax-label">Syntax</div>
              {syntaxValid === null ? (
                <span style={{ color: 'var(--text-secondary)', fontSize: 12 }}>Checking...</span>
              ) : syntaxValid ? (
                <span style={{ color: 'var(--success)', fontSize: 13, fontWeight: 600 }}>✅ Valid</span>
              ) : (
                <span style={{ color: 'var(--danger)', fontSize: 12 }}>❌ {syntaxError}</span>
              )}
            </div>

            {/* Arguments */}
            <div className="query-editor-args">
              <div className="query-editor-args-label">Arguments</div>
              <div className="query-editor-args-list">
                {/* THIS_TABLE always present */}
                <div className="query-editor-arg-row locked">
                  <span className="arg-name">{'{THIS_TABLE}'}</span>
                  <span className="arg-lock">🔒</span>
                </div>
                {args.map((a, idx) => (
                  <div key={a.name} className={`query-editor-arg-row${undefinedArgs.includes(a.name) ? '' : ''}`}>
                    <span className="arg-name">:{a.name}</span>
                    <span className="arg-type">{a.type}</span>
                    <span className="arg-actions">
                      <button className="icon-btn" onClick={() => handleEditArg(idx)} title="Edit">✏️</button>
                      <button className="icon-btn" onClick={() => handleDeleteArg(idx)} title="Delete">🗑</button>
                    </span>
                  </div>
                ))}
              </div>
              <button className="secondary small" onClick={handleAddArg} style={{ marginTop: 8 }}>＋ Add Argument</button>
            </div>
          </div>
        </div>

        <div className="modal-editor-footer">
          {saveError && (
            <span style={{ color: 'var(--danger)', fontSize: 12, marginRight: 'auto' }}>❌ {saveError}</span>
          )}
          <button className="secondary" onClick={onClose}>Cancel</button>
          <button onClick={handleSave} disabled={!name.trim() || !query.trim() || saving}>
            {saving ? 'Saving…' : 'Save'}
          </button>
        </div>
      </div>

      {/* Argument sub-popup */}
      {argPopup && (
        <div className="modal-overlay" style={{ zIndex: 1100 }} onClick={() => setArgPopup(null)}>
          <div className="modal-box" onClick={(e) => e.stopPropagation()} style={{ minWidth: 320 }}>
            <h4 style={{ margin: '0 0 12px' }}>Argument</h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <label style={{ fontSize: 12, fontWeight: 600 }}>Name</label>
              <input
                value={argPopup.arg.name}
                onChange={(e) => setArgPopup({ ...argPopup, arg: { ...argPopup.arg, name: e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, '') } })}
                placeholder="ARGUMENT_NAME"
                autoFocus
              />
              <label style={{ fontSize: 12, fontWeight: 600 }}>Type</label>
              <select
                value={argPopup.arg.type}
                onChange={(e) => setArgPopup({ ...argPopup, arg: { ...argPopup.arg, type: e.target.value } })}
              >
                <option value="string">string</option>
                <option value="number">number</option>
                <option value="date">date</option>
              </select>
              <label style={{ fontSize: 12, fontWeight: 600 }}>Description</label>
              <input
                value={argPopup.arg.description}
                onChange={(e) => setArgPopup({ ...argPopup, arg: { ...argPopup.arg, description: e.target.value } })}
                placeholder="Help text for the user"
              />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 16 }}>
              <button className="secondary" onClick={() => setArgPopup(null)}>Cancel</button>
              <button onClick={() => handleArgSave(argPopup.arg)} disabled={!argPopup.arg.name}>Save</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default QueryEditorPopup;
