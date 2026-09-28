import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, test, vi } from "vite-plus/test";

import { MobileBackContext } from "../hooks/useMobileBack";
import { PageHeader } from "./PageHeader";

test("積み重ねた画面の外では「戻る」を出さない", () => {
  render(<PageHeader icon={null} title="スレッド" />);

  expect(screen.getByRole("heading", { name: "スレッド" })).toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "戻る" })).not.toBeInTheDocument();
});

test("積み重ねた画面の中では「戻る」で戻る", async () => {
  const back = vi.fn<() => void>();
  render(
    <MobileBackContext value={back}>
      <PageHeader icon={null} title="スレッド" />
    </MobileBackContext>,
  );

  await userEvent.click(screen.getByRole("button", { name: "戻る" }));
  expect(back).toHaveBeenCalledOnce();
});
