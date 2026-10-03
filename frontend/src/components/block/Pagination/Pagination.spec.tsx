import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, test, vi } from "vite-plus/test";

import { Pagination } from "./Pagination";

test("1 ページしかなければ何も出さない", () => {
  const { container } = render(<Pagination page={1} totalPages={1} onChange={() => {}} />);

  expect(container).toBeEmptyDOMElement();
});

test("前後のページへ移動でき、端では押せない", async () => {
  const onChange = vi.fn<(page: number) => void>();
  render(<Pagination page={1} totalPages={3} onChange={onChange} />);

  expect(screen.getByText("1 / 3 ページ")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "前のページ" })).toBeDisabled();
  await userEvent.click(screen.getByRole("button", { name: "次のページ" }));
  expect(onChange).toHaveBeenCalledWith(2);
});
