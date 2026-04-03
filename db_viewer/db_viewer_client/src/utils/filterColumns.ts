export interface FilteredColumn {
  type: 'column' | 'space' | 'commentary' | 'the-rest';
  name?: string;
  text?: string;
}

export function parseFilterColumns(
  filterCols: string[],
  allColumns: string[]
): FilteredColumn[] {
  const result: FilteredColumn[] = [];
  const usedCols = new Set<string>();

  for (const entry of filterCols) {
    if (entry === '<SPACE>') {
      result.push({ type: 'space' });
    } else if (entry.startsWith('<COMMENTARY>')) {
      result.push({ type: 'commentary', text: entry.replace('<COMMENTARY>', '').trim() });
    } else if (entry === '<THE REST>') {
      result.push({ type: 'space' });
      for (const col of allColumns) {
        if (!usedCols.has(col)) {
          result.push({ type: 'column', name: col });
        }
      }
    } else {
      usedCols.add(entry);
      result.push({ type: 'column', name: entry });
    }
  }

  return result;
}
