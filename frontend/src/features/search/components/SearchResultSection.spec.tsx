import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { SearchResultSection } from "./SearchResultSection";

describe("SearchResultSection", () => {
  test("見出しと中身を表示する", () => {
    render(<SearchResultSection title="チャンネル">中身</SearchResultSection>);
    expect(screen.getByRole("heading", { name: "チャンネル" })).toBeInTheDocument();
    expect(screen.getByText("中身")).toBeInTheDocument();
  });
});
