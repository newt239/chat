import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { renderWithProviders } from "#/test/renderWithProviders";

import { MemberRow } from "./MemberRow";

describe("MemberRow", () => {
  test("名前と補足を表示し、押すとプロフィールを開く", async () => {
    const { router } = await renderWithProviders(
      <MemberRow userId="u-bob" name="Bob" avatarUrl={undefined} detail="bob@example.com" />,
      "/app/ws1",
      () => {},
    );
    const row = screen.getByRole("button", { name: /Bob/ });
    expect(row).toHaveTextContent("bob@example.com");
    await userEvent.click(row);
    expect(router.state.location.search).toEqual({ profile: "u-bob" });
  });
});
