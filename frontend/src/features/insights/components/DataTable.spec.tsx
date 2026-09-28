import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { DataTable } from "./DataTable";

describe("DataTable", () => {
  test("先頭の列を行見出しにして値を並べる", () => {
    render(
      <DataTable
        columns={["チャンネル", "件数"]}
        rows={[
          ["general", "120"],
          ["random", "8"],
        ]}
      />,
    );
    expect(screen.getAllByRole("columnheader").map((cell) => cell.textContent)).toStrictEqual([
      "チャンネル",
      "件数",
    ]);
    expect(screen.getByRole("rowheader", { name: "random" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "120" })).toBeInTheDocument();
  });
});
