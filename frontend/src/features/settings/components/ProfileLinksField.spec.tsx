import { useState } from "react";

import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { MAX_PROFILE_LINKS } from "../utils/profileLinks";
import { ProfileLinksField } from "./ProfileLinksField";

const Harness = ({ initial }: { initial: string[] }) => {
  const [links, setLinks] = useState(initial);
  return (
    <>
      <ProfileLinksField value={links} onChange={setLinks} />
      <output>{JSON.stringify(links)}</output>
    </>
  );
};

describe("ProfileLinksField", () => {
  test("リンクを追加して入力し、削除できる", async () => {
    render(<Harness initial={[]} />);
    await userEvent.click(screen.getByRole("button", { name: "リンクを追加" }));
    await userEvent.type(screen.getByRole("textbox", { name: "リンク 1" }), "https://github.com");
    expect(screen.getByRole("status")).toHaveTextContent(JSON.stringify(["https://github.com"]));

    await userEvent.click(screen.getByRole("button", { name: "リンク 1 を削除" }));
    expect(screen.getByRole("status")).toHaveTextContent("[]");
  });

  test("上限まで入れたら追加ボタンを出さない", () => {
    render(
      <Harness initial={Array.from({ length: MAX_PROFILE_LINKS }, () => "https://a.example")} />,
    );
    expect(screen.queryByRole("button", { name: "リンクを追加" })).not.toBeInTheDocument();
  });
});
