import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MessageEditor } from "./MessageEditor";

const render = (onSave: (body: string) => Promise<void>, onClose: () => void) =>
  renderWithProviders(
    <MessageEditor initialBody="before" onSave={onSave} onClose={onClose} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(WorkspaceService.method.listMembers, () => ({
        members: [
          { displayName: "Alice Johnson", userId: "u1" },
          { displayName: "Bob Smith", userId: "u2" },
        ],
      }));
      routes.rpc(UserGroupService.method.listUserGroups, () => ({ userGroups: [] }));
    },
  );

describe("MessageEditor", () => {
  test("Enter で前後の空白を除いた本文を保存して閉じる", async () => {
    const onSave = vi.fn<(body: string) => Promise<void>>().mockResolvedValue();
    const onClose = vi.fn<() => void>();
    await render(onSave, onClose);

    const textbox = await screen.findByRole("textbox", { name: "メッセージを編集" });
    await userEvent.clear(textbox);
    await userEvent.type(textbox, " after {Enter}");

    expect(onSave).toHaveBeenCalledWith("after");
    expect(onClose).toHaveBeenCalledOnce();
  });

  test("空にすると保存せずにエラーを出し、Esc で閉じる", async () => {
    const onSave = vi.fn<(body: string) => Promise<void>>();
    const onClose = vi.fn<() => void>();
    await render(onSave, onClose);

    const textbox = await screen.findByRole("textbox", { name: "メッセージを編集" });
    await userEvent.clear(textbox);
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    expect(screen.getByText("メッセージを入力してください")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();

    await userEvent.type(textbox, "{Escape}");
    expect(onClose).toHaveBeenCalledOnce();
  });

  test("@ で候補を出し、Enter で表示名を挿入する", async () => {
    const onSave = vi.fn<(body: string) => Promise<void>>().mockResolvedValue();
    const onClose = vi.fn<() => void>();
    await render(onSave, onClose);

    const textbox = await screen.findByRole("textbox", { name: "メッセージを編集" });
    await userEvent.type(textbox, " @bo");
    expect(await screen.findByRole("option", { name: /Bob Smith/ })).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: /Alice/ })).not.toBeInTheDocument();

    await userEvent.keyboard("{Enter}");
    expect(textbox).toHaveValue("before @Bob ");
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });
});
