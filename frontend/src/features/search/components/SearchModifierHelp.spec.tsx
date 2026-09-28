import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { SearchModifierHelp } from "./SearchModifierHelp";

describe("SearchModifierHelp", () => {
  test("押した修飾子を渡す", async () => {
    const onInsert = vi.fn<(modifier: string) => void>();
    render(<SearchModifierHelp onInsert={onInsert} />);
    await userEvent.click(screen.getByRole("button", { name: "from:@ を挿入" }));
    await userEvent.click(screen.getByRole("button", { name: "during:week を挿入" }));
    expect(onInsert.mock.calls).toEqual([["from:@"], ["during:week"]]);
  });
});
