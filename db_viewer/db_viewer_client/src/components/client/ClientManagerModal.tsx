import React, { useState, useEffect, useRef, useCallback } from 'react';
import { api, ClientConfigResponse, SaveClientRequest, TestConnectionRequest } from '../../api/client';

interface Props {
  open: boolean;
  onClose: () => void;
  onSaved: () => void;
}

type Step = 'list' | 'connection' | 'tables';

interface FormData {
  name: string;
  display_name: string;
  host: string;
  port: number;
  service_name: string;
  username: string;
  password: string;
  tables: string[];
}

const emptyForm: FormData = {
  name: '', display_name: '', host: '', port: 1521,
  service_name: '', username: '', password: '', tables: [],
};

const ClientManagerModal: React.FC<Props> = ({ open, onClose, onSaved }) => {
  const [step, setStep] = useState<Step>('list');
  const [direction, setDirection] = useState<'forward' | 'back'>('forward');
  const [clients, setClients] = useState<ClientConfigResponse[]>([]);
  const [form, setForm] = useState<FormData>({ ...emptyForm });
  const [isEdit, setIsEdit] = useState(false);
  const [editingName, setEditingName] = useState('');
  const [testStatus, setTestStatus] = useState<'idle' | 'testing' | 'success' | 'failed'>('idle');
  const [testError, setTestError] = useState('');
  const [allTables, setAllTables] = useState<string[]>([]);
  const [moreMenu, setMoreMenu] = useState<string | null>(null);
  const [morePos, setMorePos] = useState<{ top: number; left: number }>({ top: 0, left: 0 });
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const moreRef = useRef<HTMLDivElement>(null);
  const [dragItem, setDragItem] = useState<string | null>(null);
  const [dragOverTrash, setDragOverTrash] = useState(false);

  const loadClients = useCallback(async () => {
    try {
      const list = await api.listManagedClients();
      setClients(list || []);
    } catch { setClients([]); }
  }, []);

  useEffect(() => {
    if (open) {
      loadClients();
      setStep('list');
    }
  }, [open, loadClients]);

  // Close more menu on outside click
  useEffect(() => {
    if (!moreMenu) return;
    const handler = (e: MouseEvent) => {
      if (moreRef.current && !moreRef.current.contains(e.target as Node)) setMoreMenu(null);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [moreMenu]);

  if (!open) return null;

  const goTo = (next: Step, dir: 'forward' | 'back') => {
    setDirection(dir);
    setStep(next);
  };

  const startAdd = () => {
    setForm({ ...emptyForm });
    setIsEdit(false);
    setEditingName('');
    setTestStatus('idle');
    setTestError('');
    goTo('connection', 'forward');
  };

  const startEdit = (c: ClientConfigResponse) => {
    setForm({
      name: c.name, display_name: c.display_name, host: c.host,
      port: c.port, service_name: c.service_name, username: c.username,
      password: '', tables: c.tables || [],
    });
    setIsEdit(true);
    setEditingName(c.name);
    setTestStatus('idle');
    setTestError('');
    setMoreMenu(null);
    goTo('connection', 'forward');
  };

  const startClone = (c: ClientConfigResponse) => {
    setForm({
      name: c.name + '_copy', display_name: (c.display_name || c.name) + ' Copy',
      host: c.host, port: c.port, service_name: c.service_name,
      username: c.username, password: '', tables: [...(c.tables || [])],
    });
    setIsEdit(false);
    setEditingName('');
    setTestStatus('idle');
    setTestError('');
    setMoreMenu(null);
    goTo('connection', 'forward');
  };

  const handleDelete = async (name: string) => {
    try {
      await api.deleteClient(name);
      await loadClients();
      onSaved();
    } catch (e: any) {
      alert(e.message);
    }
    setDeleteConfirm(null);
    setMoreMenu(null);
  };

  const handleTestConnection = async () => {
    setTestStatus('testing');
    setTestError('');
    try {
      const connReq: TestConnectionRequest = {
        host: form.host, port: form.port, service_name: form.service_name,
        username: form.username, password: form.password,
      };
      const res = await api.testConnection(connReq);
      if (res.success) {
        setTestStatus('success');
      } else {
        setTestStatus('failed');
        setTestError(res.error || 'Connection failed');
      }
    } catch (e: any) {
      setTestStatus('failed');
      setTestError(e.message);
    }
  };

  const handleContinueToTables = async () => {
    // Fetch all tables from the connection
    try {
      let tables: string[];
      if (isEdit && testStatus !== 'success') {
        // Use existing connection
        const res = await api.listAllTablesForClient(editingName);
        tables = res.tables || [];
      } else {
        const connReq: TestConnectionRequest = {
          host: form.host, port: form.port, service_name: form.service_name,
          username: form.username, password: form.password,
        };
        const res = await api.listTablesFromConnection(connReq);
        tables = res.tables || [];
      }
      setAllTables(tables);
    } catch {
      setAllTables([]);
    }
    goTo('tables', 'forward');
  };

  const canContinue = isEdit || testStatus === 'success';

  const handleSave = async () => {
    setSaving(true);
    try {
      const body: SaveClientRequest = {
        name: form.name, display_name: form.display_name, host: form.host,
        port: form.port, service_name: form.service_name, username: form.username,
        password: form.password, tables: form.tables,
      };
      if (isEdit) {
        await api.updateClient(editingName, body);
      } else {
        await api.createClient(body);
      }
      await loadClients();
      onSaved();
      goTo('list', 'back');
    } catch (e: any) {
      alert(e.message);
    }
    setSaving(false);
  };

  const handleMoreClick = (e: React.MouseEvent, name: string) => {
    e.stopPropagation();
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    setMorePos({ top: rect.bottom + 2, left: rect.left });
    setMoreMenu(prev => prev === name ? null : name);
  };

  // Drag from available to visible
  const handleDragStartAvailable = (e: React.DragEvent, table: string) => {
    e.dataTransfer.setData('text/plain', table);
    e.dataTransfer.effectAllowed = 'copy';
    setDragItem(table);
  };

  const handleDropOnVisible = (e: React.DragEvent) => {
    e.preventDefault();
    const table = e.dataTransfer.getData('text/plain');
    if (table && !form.tables.includes(table)) {
      setForm(f => ({ ...f, tables: [...f.tables, table] }));
    }
    setDragItem(null);
  };

  // Drag within visible to reorder
  const handleDragStartVisible = (e: React.DragEvent, idx: number) => {
    e.dataTransfer.setData('text/plain', String(idx));
    e.dataTransfer.effectAllowed = 'move';
    setDragItem(form.tables[idx]);
  };

  const handleDropOnVisibleItem = (e: React.DragEvent, targetIdx: number) => {
    e.preventDefault();
    const sourceIdx = parseInt(e.dataTransfer.getData('text/plain'), 10);
    if (isNaN(sourceIdx)) return;
    const newTables = [...form.tables];
    const [moved] = newTables.splice(sourceIdx, 1);
    newTables.splice(targetIdx, 0, moved);
    setForm(f => ({ ...f, tables: newTables }));
    setDragItem(null);
  };

  // Drag out to trash
  const handleDropOnTrash = (e: React.DragEvent) => {
    e.preventDefault();
    const sourceIdx = parseInt(e.dataTransfer.getData('text/plain'), 10);
    if (!isNaN(sourceIdx)) {
      setForm(f => ({ ...f, tables: f.tables.filter((_, i) => i !== sourceIdx) }));
    }
    setDragItem(null);
    setDragOverTrash(false);
  };

  const addSpecialToken = (token: string) => {
    setForm(f => ({ ...f, tables: [...f.tables, token] }));
  };

  const slideClass = direction === 'forward' ? 'cm-slide-forward' : 'cm-slide-back';

  return (
    <div className="cm-overlay" onClick={onClose}>
      <div className="cm-modal" onClick={e => e.stopPropagation()}>
        <div className="cm-container">
          {/* Step 1: Client List */}
          {step === 'list' && (
            <div className={`cm-panel ${slideClass}`} key="list">
              <div className="cm-panel-header">
                <span className="cm-panel-title">Manage Clients</span>
                <button className="cm-close-btn" onClick={onClose}>✕</button>
              </div>
              <div className="cm-panel-body">
                <div className="cm-client-list">
                  {clients.map(c => (
                    <div key={c.name} className="cm-client-row">
                      <div className="cm-client-info">
                        <span className="cm-client-name">{c.display_name || c.name}</span>
                        <span className="cm-client-sub">{c.host}:{c.port}/{c.service_name}</span>
                      </div>
                      <span className="cm-client-more" onClick={e => handleMoreClick(e, c.name)}>···</span>
                    </div>
                  ))}
                  {clients.length === 0 && (
                    <div className="cm-empty">No clients configured</div>
                  )}
                </div>
                <div className="cm-panel-actions">
                  <button className="cm-btn cm-btn-primary" onClick={startAdd}>+ Add New Client</button>
                </div>
              </div>
            </div>
          )}

          {/* Step 2: Connection Info */}
          {step === 'connection' && (
            <div className={`cm-panel ${slideClass}`} key="connection">
              <div className="cm-panel-header">
                <button className="cm-back-btn" onClick={() => goTo('list', 'back')}>← Back</button>
                <span className="cm-panel-title">{isEdit ? 'Edit Client' : 'New Client'}</span>
              </div>
              <div className="cm-panel-body">
                <div className="cm-form">
                  <label className="cm-label">Display Name
                    <input className="cm-input" value={form.display_name} onChange={e => setForm(f => ({ ...f, display_name: e.target.value }))} />
                  </label>
                  <label className="cm-label">Internal Key
                    <input className="cm-input" value={form.name} disabled={isEdit} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} />
                  </label>
                  <label className="cm-label">Host
                    <input className="cm-input" value={form.host} onChange={e => setForm(f => ({ ...f, host: e.target.value }))} />
                  </label>
                  <label className="cm-label">Port
                    <input className="cm-input" type="number" value={form.port} onChange={e => setForm(f => ({ ...f, port: Number(e.target.value) }))} />
                  </label>
                  <label className="cm-label">Service Name
                    <input className="cm-input" value={form.service_name} onChange={e => setForm(f => ({ ...f, service_name: e.target.value }))} />
                  </label>
                  <label className="cm-label">Username
                    <input className="cm-input" value={form.username} onChange={e => setForm(f => ({ ...f, username: e.target.value }))} />
                  </label>
                  <label className="cm-label">Password
                    <input className="cm-input" type="password" value={form.password} placeholder={isEdit ? '(unchanged)' : ''} onChange={e => setForm(f => ({ ...f, password: e.target.value }))} />
                  </label>
                  <div className="cm-test-row">
                    <button className={`cm-btn cm-btn-test ${testStatus === 'success' ? 'cm-test-ok' : testStatus === 'failed' ? 'cm-test-fail' : ''}`} onClick={handleTestConnection} disabled={testStatus === 'testing'}>
                      {testStatus === 'testing' ? 'Testing…' : 'Test Connection'}
                    </button>
                    {testStatus === 'success' && <span className="cm-test-status cm-test-ok">● Connected</span>}
                    {testStatus === 'failed' && <span className="cm-test-status cm-test-fail">✗ {testError}</span>}
                  </div>
                </div>
                <div className="cm-panel-footer">
                  <button className="cm-btn cm-btn-ghost" onClick={() => goTo('list', 'back')}>Cancel</button>
                  <button className="cm-btn cm-btn-primary" disabled={!canContinue} onClick={handleContinueToTables}>Continue →</button>
                </div>
              </div>
            </div>
          )}

          {/* Step 3: Table Selection */}
          {step === 'tables' && (
            <div className={`cm-panel ${slideClass}`} key="tables">
              <div className="cm-panel-header">
                <button className="cm-back-btn" onClick={() => goTo('connection', 'back')}>← Back</button>
                <span className="cm-panel-title">Select Tables</span>
              </div>
              <div className="cm-panel-body">
                <div className="cm-two-col">
                  <div className="cm-col">
                    <div className="cm-col-header">Available</div>
                    <div className="cm-col-list">
                      {allTables.filter(t => !form.tables.includes(t)).map(t => (
                        <div key={t} className="cm-col-item" draggable onDragStart={e => handleDragStartAvailable(e, t)}>
                          {t}
                        </div>
                      ))}
                    </div>
                    <div className="cm-col-actions">
                      <button className="cm-btn cm-btn-small" onClick={() => addSpecialToken('<SPACE>')}>+ Space</button>
                      <button className="cm-btn cm-btn-small" onClick={() => addSpecialToken('<COMMENTARY>')}>+ Commentary</button>
                    </div>
                  </div>
                  <div className="cm-col" onDragOver={e => e.preventDefault()} onDrop={handleDropOnVisible}>
                    <div className="cm-col-header">Visible in App</div>
                    <div className="cm-col-list">
                      {form.tables.map((t, idx) => (
                        <div
                          key={`${t}-${idx}`}
                          className={`cm-col-item cm-col-item-visible ${t === '<SPACE>' ? 'cm-special' : ''} ${t.startsWith('<COMMENTARY>') ? 'cm-special' : ''}`}
                          draggable
                          onDragStart={e => handleDragStartVisible(e, idx)}
                          onDragOver={e => e.preventDefault()}
                          onDrop={e => handleDropOnVisibleItem(e, idx)}
                        >
                          {t === '<SPACE>' ? '── Space ──' : t.startsWith('<COMMENTARY>') ? `── ${t.replace('<COMMENTARY>', '').trim() || 'Commentary'} ──` : t}
                        </div>
                      ))}
                      {form.tables.length === 0 && (
                        <div className="cm-empty-col">Drag tables here</div>
                      )}
                    </div>
                  </div>
                </div>
                {dragItem && (
                  <div
                    className={`cm-trash-zone ${dragOverTrash ? 'cm-trash-active' : ''}`}
                    onDragOver={e => { e.preventDefault(); setDragOverTrash(true); }}
                    onDragLeave={() => setDragOverTrash(false)}
                    onDrop={handleDropOnTrash}
                  >
                    🗑 Drop here to remove
                  </div>
                )}
                <div className="cm-panel-footer">
                  <button className="cm-btn cm-btn-ghost" onClick={() => goTo('list', 'back')}>Cancel</button>
                  <button className="cm-btn cm-btn-primary" disabled={saving} onClick={handleSave}>
                    {saving ? 'Saving…' : 'Save →'}
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* More menu popover */}
        {moreMenu && (
          <div ref={moreRef} className="sidebar-context-menu" style={{ top: morePos.top, left: morePos.left, position: 'fixed' }}>
            <div className="sidebar-context-item" onClick={() => { const c = clients.find(x => x.name === moreMenu); if (c) startEdit(c); }}>Edit</div>
            <div className="sidebar-context-item" onClick={() => { const c = clients.find(x => x.name === moreMenu); if (c) startClone(c); }}>Clone</div>
            <div className="sidebar-context-divider" />
            <div className="sidebar-context-item cm-delete-item" onClick={() => setDeleteConfirm(moreMenu)}>Delete</div>
          </div>
        )}

        {/* Delete confirmation */}
        {deleteConfirm && (
          <div className="cm-confirm-overlay" onClick={() => setDeleteConfirm(null)}>
            <div className="cm-confirm-box" onClick={e => e.stopPropagation()}>
              <p>Delete client <strong>{deleteConfirm}</strong>?</p>
              <p style={{ fontSize: 12, color: '#888' }}>All related presets and data will be removed.</p>
              <div className="cm-confirm-actions">
                <button className="cm-btn cm-btn-ghost" onClick={() => setDeleteConfirm(null)}>Cancel</button>
                <button className="cm-btn cm-btn-danger" onClick={() => handleDelete(deleteConfirm)}>Delete</button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ClientManagerModal;
