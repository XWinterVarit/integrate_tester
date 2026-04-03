export interface ClientInfo {
  name: string;
  schema: string;
}

export interface PresetFilter {
  name: string;
  details: string;
  columns: string[];
}

export interface PresetQueryArg {
  name: string;
  type: string;
  description: string;
}

export interface PresetQuery {
  index: number;
  name: string;
  query: string;
  arguments: PresetQueryArg[];
}

export interface ColumnInfo {
  COLUMN_NAME: string;
  DATA_TYPE: string;
  DATA_LENGTH: number;
  NULLABLE: string;
  DATA_DEFAULT: string | null;
}

export interface ConstraintInfo {
  CONSTRAINT_NAME: string;
  CONSTRAINT_TYPE: string;
  STATUS: string;
  COLUMNS: string;
}

export interface IndexInfo {
  INDEX_NAME: string;
  INDEX_TYPE: string;
  UNIQUENESS: string;
  COLUMNS: string;
}

export interface TableSizeInfo {
  BYTES: number;
  BLOCKS: number;
  ROW_COUNT: number;
}

export type Row = Record<string, any>;

export type ViewMode = 'row' | 'transpose';

export interface FloatingWindow {
  id: string;
  title: string;
  content: any;
  type: 'row-json' | 'column-info' | 'table-info' | 'field-edit' | 'delete-confirm' | 'insert-form' | 'preset-query';
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface QueryParams {
  select: string;
  sort: string;
  sortDir: string;
  limit: number;
}
