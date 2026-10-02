import { useState } from "react";

import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { renderMarkdown } from "../../utils/markdown/renderer";
import { MarkdownSpan } from "./MarkdownSpan";

const Rerenderable = () => {
  const [count, setCount] = useState(0);
  return (
    <>
      <button
        type="button"
        onClick={() => {
          setCount(count + 1);
        }}
      >
        再描画 {count}
      </button>
      <p>{renderMarkdown("<@11111111-1111-1111-1111-111111111111> さん", [])}</p>
    </>
  );
};

describe("MarkdownSpan", () => {
  test("メンションでも絵文字でもない span はそのまま出す", async () => {
    await renderWithProviders(
      <MarkdownSpan className="note">本文</MarkdownSpan>,
      "/app/ws1",
      () => {},
    );
    expect(screen.getByText("本文")).toHaveClass("note");
  });

  test("親が再描画されてもメンションのボタンを作り直さない", async () => {
    await renderWithProviders(<Rerenderable />, "/app/ws1", (routes) => {
      routes.rpc(WorkspaceService.method.listMembers, () => ({
        members: [
          create(WorkspaceMemberSchema, {
            displayName: "Bob Smith",
            userId: "11111111-1111-1111-1111-111111111111",
          }),
        ],
      }));
    });
    const mention = await screen.findByRole("button", { name: "@Bob Smith" });

    await userEvent.click(screen.getByRole("button", { name: /再描画/ }));

    expect(screen.getByRole("button", { name: "@Bob Smith" })).toBe(mention);
  });
});
