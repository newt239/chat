import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { rightSidePanelViewAtom } from "#/providers/store/ui";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MemberRow } from "./MemberRow";

describe("MemberRow", () => {
  test("名前と補足を表示し、押すとプロフィールを開く", async () => {
    const { store } = await renderWithProviders(
      <MemberRow userId="u-bob" name="Bob" avatarUrl={undefined} detail="bob@example.com" />,
      "/app/ws1",
      () => {},
    );
    const row = screen.getByRole("button", { name: /Bob/ });
    expect(row).toHaveTextContent("bob@example.com");
    await userEvent.click(row);
    expect(store.get(rightSidePanelViewAtom)).toEqual({ type: "user-profile", userId: "u-bob" });
  });
});
