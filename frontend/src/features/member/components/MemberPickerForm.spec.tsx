import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MemberPickerForm } from "./MemberPickerForm";

describe("MemberPickerForm", () => {
  test("未参加のメンバーを表示名で選び、送信する", async () => {
    const onSubmit = vi.fn<(userId: string) => void>();
    await renderWithProviders(
      <MemberPickerForm
        workspaceId="ws1"
        memberIds={new Set(["u-bob"])}
        label="メンバーを追加"
        placeholder="名前で検索"
        submitLabel="追加"
        isPending={false}
        onSubmit={onSubmit}
      />,
      "/app/ws1",
      (routes) => {
        routes.rpc(WorkspaceService.method.listMembers, () => ({
          members: [
            create(WorkspaceMemberSchema, { displayName: "Bob", userId: "u-bob" }),
            create(WorkspaceMemberSchema, {
              displayName: "Carol",
              nickname: "かろる",
              userId: "u-carol",
            }),
          ],
        }));
      },
    );

    const submit = screen.getByRole("button", { name: "追加" });
    expect(submit).toBeDisabled();
    await userEvent.click(screen.getByRole("combobox", { name: "メンバーを追加" }));
    expect(screen.queryByRole("option", { name: "Bob" })).not.toBeInTheDocument();
    await userEvent.click(await screen.findByRole("option", { name: "かろる" }));
    await userEvent.click(submit);

    expect(onSubmit).toHaveBeenCalledWith("u-carol");
    expect(submit).toBeDisabled();
  });
});
