const BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || res.statusText);
  }
  return res.json();
}

export const api = {
  getClients: () =>
    request<{ name: string; schema: string }[]>('/api/clients'),

  getTables: (client: string) =>
    request<string[]>(`/api/clients/${client}/tables`),

  getRows: (client: string, table: string, params: Record<string, string>) => {
    const qs = new URLSearchParams(params).toString();
    return request<Record<string, any>[]>(`/api/clients/${client}/tables/${table}/rows?${qs}`);
  },

  executeQuery: (client: string, table: string, body: { query: string; args: Record<string, string>; limit: number; offset?: number; sort?: string; sort_dir?: string }) =>
    request<Record<string, any>[]>(`/api/clients/${client}/tables/${table}/query`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  getColumns: (client: string, table: string) =>
    request<Record<string, any>[]>(`/api/clients/${client}/tables/${table}/columns`),

  getConstraints: (client: string, table: string) =>
    request<Record<string, any>[]>(`/api/clients/${client}/tables/${table}/constraints`),

  getIndexes: (client: string, table: string) =>
    request<Record<string, any>[]>(`/api/clients/${client}/tables/${table}/indexes`),

  getTableSize: (client: string, table: string) =>
    request<Record<string, any>[]>(`/api/clients/${client}/tables/${table}/size`),

  getRowCount: (client: string, table: string) =>
    request<{ count: number }>(`/api/clients/${client}/tables/${table}/count`),

  getFilters: (client: string, table: string) =>
    request<{ name: string; details: string; columns: string[] }[]>(`/api/clients/${client}/tables/${table}/filters`),

  getPresetQueries: (client: string, table: string) =>
    request<{ index: number; name: string; query: string; arguments: { name: string; type: string; description: string }[] }[]>(
      `/api/clients/${client}/tables/${table}/preset-queries`
    ),

  resolvePresetQuery: (client: string, table: string, index: number, args: Record<string, string>) =>
    request<{ resolved_query: string }>(`/api/clients/${client}/tables/${table}/preset-queries/${index}/resolve`, {
      method: 'POST',
      body: JSON.stringify({ args }),
    }),

  exportTable: (client: string, table: string, params: Record<string, string>) => {
    const qs = new URLSearchParams(params).toString();
    return `${BASE_URL}/api/clients/${client}/tables/${table}/export?${qs}`;
  },

  updateCell: (client: string, table: string, body: { column: string; value: string; rowid: string }) =>
    request<{ status: string }>(`/api/clients/${client}/tables/${table}/rows/update`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  blobDownloadUrl: (client: string, table: string, params: { column: string; rowid: string }) => {
    const qs = new URLSearchParams(params).toString();
    return `${BASE_URL}/api/clients/${client}/tables/${table}/blob?${qs}`;
  },
  uploadBlob: (client: string, table: string, params: { column: string; rowid: string }, data: ArrayBuffer) => {
    const qs = new URLSearchParams(params).toString();
    return request<{ ok: boolean }>(`/api/clients/${client}/tables/${table}/blob?${qs}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/octet-stream' },
      body: data,
    });
  },

  deleteRow: (client: string, table: string, body: { rowid: string }) =>
    request<{ status: string }>(`/api/clients/${client}/tables/${table}/rows/delete`, {
      method: 'DELETE',
      body: JSON.stringify(body),
    }),

  insertRow: (client: string, table: string, body: { columns: string[]; values: string[] }) =>
    request<{ status: string }>(`/api/clients/${client}/tables/${table}/rows/insert`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  buildDeleteQuery: (client: string, table: string, rowid: string) =>
    request<{ query: string }>(`/api/clients/${client}/tables/${table}/rows/delete-query?rowid=${encodeURIComponent(rowid)}`),

  buildUpdateQuery: (client: string, table: string, params: { column: string; value: string; rowid: string }) => {
    const qs = new URLSearchParams(params).toString();
    return request<{ query: string }>(`/api/clients/${client}/tables/${table}/rows/update-query?${qs}`);
  },

  buildInsertQuery: (client: string, table: string, body: { columns: string[]; values: string[] }) =>
    request<{ query: string }>(`/api/clients/${client}/tables/${table}/rows/insert-query`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  touchRecentFilter: (key: string) =>
    request<{ status: string }>('/api/recent/filter', {
      method: 'POST',
      body: JSON.stringify({ key }),
    }),

  touchRecentQuery: (key: string) =>
    request<{ status: string }>('/api/recent/query', {
      method: 'POST',
      body: JSON.stringify({ key }),
    }),
};
