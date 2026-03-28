# DB Viewer

A web application for browsing Oracle database tables. Built with a **Go** backend (using `go-ora/v2`) and a **React + TypeScript** frontend with an Apple-style UI.

---

## Features

- **Left pane**: lists configured tables per client connection
- **Right pane**: displays rows with row view / transpose view toggle
- **Preset column filters**: with `<SPACE>`, `<COMMENTARY>`, `<THE REST>` support
- **Preset queries**: searchable dropdown, editable arguments, final query preview before execution
- **Floating windows**: moveable, resizable, pop-out to new browser window — for row JSON, column info (with edit), and table info (constraints, indexes, size)
- **Export**: current view or full table as CSV/JSON
- **Toolbar**: select columns, sort, limit, refresh
- **Recent usage sorting**: preset filters and queries sorted by most recently used (in-memory)

---

## Project Structure

```
db_viewer/
├── backend/
│   ├── cmd/
│   │   └── main.go              # Entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── loader.go        # YAML config loader
│   │   ├── handler/
│   │   │   ├── client.go        # Client/table list handlers
│   │   │   ├── export.go        # CSV/JSON export handler
│   │   │   ├── json.go          # JSON response helpers
│   │   │   ├── query.go         # Query & preset handlers
│   │   │   ├── recent.go        # Recent usage handlers
│   │   │   ├── router.go        # Route registration
│   │   │   └── table.go         # Table data handlers
│   │   ├── middleware/
│   │   │   └── cors.go          # CORS middleware
│   │   ├── model/
│   │   │   ├── config.go        # Config structs
│   │   │   ├── request.go       # Request DTOs
│   │   │   └── response.go      # Response DTOs
│   │   ├── repository/
│   │   │   └── oracle.go        # Oracle DB queries
│   │   ├── service/
│   │   │   └── table.go         # Business logic
│   │   └── tracker/
│   │       └── recent.go        # Recent usage tracker
│   └── config.yml               # Example config
├── frontend/
│   ├── public/
│   │   └── index.html
│   ├── src/
│   │   ├── api/
│   │   │   └── client.ts        # API client
│   │   ├── components/
│   │   │   ├── export/
│   │   │   │   └── ExportButton.tsx
│   │   │   ├── filter/
│   │   │   │   └── FilterDropdown.tsx
│   │   │   ├── floating/
│   │   │   │   ├── FloatingContent.tsx
│   │   │   │   └── FloatingWindow.tsx
│   │   │   ├── layout/
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   └── Toolbar.tsx
│   │   │   ├── query/
│   │   │   │   └── PresetQueryPanel.tsx
│   │   │   └── table/
│   │   │       ├── DataTable.tsx
│   │   │       └── TransposeView.tsx
│   │   ├── hooks/
│   │   │   └── useApi.ts
│   │   ├── styles/
│   │   │   └── global.css
│   │   ├── types/
│   │   │   └── index.ts
│   │   ├── utils/
│   │   │   └── filterColumns.ts
│   │   ├── App.tsx
│   │   └── index.tsx
│   ├── package.json
│   └── tsconfig.json
├── requirement.txt
└── README.md
```

---

## Prerequisites

- **Go 1.23+**
- **Node.js 18+** and **npm**
- **Oracle Database** accessible from the backend

---

## Configuration

Edit `backend/config.yml` to match your Oracle database:

```yaml
server:
  port: 8080
  cors_origin: "http://localhost:3000"

clients:
  - name: "my_db"
    user: "db_user"
    password: "db_password"
    host: "oracle-host"
    port: 1521
    service: "ORCL"
    schema: "MY_SCHEMA"
    tables:
      - name: "EMPLOYEES"
        preset_filters:
          - name: "Personal Info"
            details: "View personal information"
            columns:
              - "EMPLOYEE_ID"
              - "FIRST_NAME"
              - "LAST_NAME"
              - "<SPACE>"
              - "<COMMENTARY> Contact"
              - "EMAIL"
              - "<THE REST>"
        preset_queries:
          - name: "Search by Name"
            query: "SELECT * FROM {THIS_TABLE} WHERE FIRST_NAME LIKE '%' || :NAME || '%'"
            arguments:
              - name: "NAME"
                type: "string"
                description: "Enter first name to search"
```

### Config Notes

- `{THIS_TABLE}` is automatically replaced with `SCHEMA.TABLE_NAME`
- If no tables are configured, the left pane will be empty (by design — this tool is for testers, not developers)
- Each table always has a default "no filter" and "query all" even without explicit presets

---

## How to Run

### 1. Start the Backend

```bash
cd db_viewer/backend

# Run directly
go run ./cmd/main.go

# Or with a custom config path
go run ./cmd/main.go /path/to/config.yml
```

The backend starts on `http://localhost:8080` by default.

### 2. Start the Frontend

```bash
cd db_viewer/frontend

# Install dependencies
npm install

# Start dev server
npm start
```

The frontend starts on `http://localhost:3000` and proxies API calls to the backend.

### 3. Open the App

Navigate to **http://localhost:3000** in your browser.

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/clients` | List configured clients |
| GET | `/api/clients/{client}/tables` | List tables for a client |
| GET | `/api/clients/{client}/tables/{table}/rows` | Get rows (with select, sort, limit params) |
| POST | `/api/clients/{client}/tables/{table}/query` | Execute a custom query |
| GET | `/api/clients/{client}/tables/{table}/columns` | Get column metadata |
| GET | `/api/clients/{client}/tables/{table}/constraints` | Get table constraints |
| GET | `/api/clients/{client}/tables/{table}/indexes` | Get table indexes |
| GET | `/api/clients/{client}/tables/{table}/size` | Get table size info |
| GET | `/api/clients/{client}/tables/{table}/filters` | Get preset column filters |
| GET | `/api/clients/{client}/tables/{table}/preset-queries` | Get preset queries |
| POST | `/api/clients/{client}/tables/{table}/preset-queries/{index}/resolve` | Resolve a preset query |
| GET | `/api/clients/{client}/tables/{table}/export` | Export table data (CSV/JSON) |
| PUT | `/api/clients/{client}/tables/{table}/rows/update` | Update a cell value |
| POST | `/api/recent/filter` | Track recent filter usage |
| POST | `/api/recent/query` | Track recent query usage |
