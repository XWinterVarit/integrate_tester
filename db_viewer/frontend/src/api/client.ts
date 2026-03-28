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

  executeQuery: (client: string, table: string, body: { query: string; args: Record<string, string>; limit: number }) =>
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
