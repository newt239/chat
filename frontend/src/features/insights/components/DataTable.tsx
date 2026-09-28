export type TableData = {
  columns: readonly string[];
  rows: readonly (readonly string[])[];
};

// グラフと同じ値を表で見せる。先頭の列は見出し、残りは数値として右寄せにする
export const DataTable = ({ columns, rows }: TableData) => (
  <div className="max-h-64 overflow-auto rounded-[10px] border border-border">
    <table className="w-full border-collapse text-[13px]">
      <thead>
        <tr>
          {columns.map((column, index) => (
            <th
              key={column}
              scope="col"
              className={`sticky top-0 border-b border-border bg-sunken px-3 py-2 text-[11.5px] font-semibold whitespace-nowrap text-muted ${index === 0 ? "text-left" : "text-right"}`}
            >
              {column}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((row) => (
          <tr key={row[0]} className="border-b border-border last:border-b-0">
            {row.map((cell, index) =>
              index === 0 ? (
                <th key={columns[index]} scope="row" className="px-3 py-1.5 text-left font-normal">
                  {cell}
                </th>
              ) : (
                <td
                  key={columns[index]}
                  className="px-3 py-1.5 text-right font-mono text-xs whitespace-nowrap tabular-nums"
                >
                  {cell}
                </td>
              ),
            )}
          </tr>
        ))}
      </tbody>
    </table>
  </div>
);
