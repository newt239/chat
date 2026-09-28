import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { HBarList } from "./HBarList";

describe("HBarList", () => {
  test("ラベルと値を並べる", () => {
    render(
      <HBarList
        emptyLabel="まだデータがありません"
        rows={[
          { icon: null, key: "a", label: "general", value: 120, valueLabel: "120" },
          { icon: null, key: "b", label: "random", value: 30, valueLabel: "30" },
        ]}
      />,
    );
    const items = screen.getAllByRole("listitem");
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent("general120");
  });

  test("行がなければ案内を出す", () => {
    render(<HBarList emptyLabel="まだデータがありません" rows={[]} />);
    expect(screen.getByText("まだデータがありません")).toBeInTheDocument();
  });
});
