import { render, screen } from "@testing-library/react";
import { expect, test } from "vite-plus/test";

import { EmptyState } from "./EmptyState";

test("アイコン・見出し・説明を表示する", () => {
  render(<EmptyState icon={<svg data-testid="icon" />} title="空です" description="説明文" />);

  expect(screen.getByText("空です")).toBeInTheDocument();
  expect(screen.getByText("説明文")).toBeInTheDocument();
  expect(screen.getByTestId("icon")).toBeInTheDocument();
});
