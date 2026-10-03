import { create } from "@bufbuild/protobuf";
import { Code, ConnectError } from "@connectrpc/connect";
import { cleanup, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { DraftService } from "#/gen/chat/v1/draft_service_pb";
import { MessageSchema } from "#/gen/chat/v1/message_pb";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { BaseMessageInput } from "./BaseMessageInput";

const setup = async (createMessage: (body: string) => Promise<void>, draftBody: string) => {
  const deleteDraft = vi.fn<() => void>();
  const saveDraft = vi.fn<(body: string) => void>();
  let storedDraft = draftBody;
  await renderWithProviders(
    <BaseMessageInput
      placeholder="メッセージを送信"
      channelId="c1"
      parentId={null}
      targetPicker={null}
      onSent={null}
    />,
    "/app/ws1",
    (routes) => {
      // 下書きはメンションの名前を引ける状態になってから入力欄に戻す
      routes.rpc(WorkspaceService.method.listMembers, () => ({ members: [] }));
      routes.rpc(UserGroupService.method.listUserGroups, () => ({ userGroups: [] }));
      routes.rpc(ChannelService.method.listChannels, () => ({ channels: [] }));
      routes.rpc(DraftService.method.getDraft, () =>
        storedDraft === "" ? {} : { draft: { body: storedDraft, channelId: "c1" } },
      );
      routes.rpc(DraftService.method.saveDraft, ({ body }) => {
        saveDraft(body);
        return {};
      });
      routes.rpc(DraftService.method.deleteDraft, () => {
        storedDraft = "";
        deleteDraft();
        return {};
      });
      routes.rpc(MessageService.method.createMessage, async (req) => {
        await createMessage(req.body);
        return { message: create(MessageSchema, { body: req.body, id: "m1" }) };
      });
    },
  );
  return { deleteDraft, saveDraft };
};

describe("BaseMessageInput", () => {
  test("送信に失敗したら本文と下書きを残す", async () => {
    const { saveDraft } = await setup(
      () => Promise.reject(new ConnectError("failed", Code.Unavailable)),
      "",
    );
    const input = await screen.findByRole("textbox", { name: "メッセージを送信" });
    await userEvent.type(input, "こんにちは{Enter}");

    expect(await screen.findByText("failed")).toBeInTheDocument();
    expect(input).toHaveValue("こんにちは");
    cleanup();
    await waitFor(() => {
      expect(saveDraft).toHaveBeenCalledWith("こんにちは");
    });
  });

  test("送信に成功したら本文と下書きを消す", async () => {
    const { deleteDraft } = await setup(() => Promise.resolve(), "");
    const input = await screen.findByRole("textbox", { name: "メッセージを送信" });
    await userEvent.type(input, "こんにちは{Enter}");

    await waitFor(() => {
      expect(input).toHaveValue("");
    });
    expect(deleteDraft).toHaveBeenCalled();
  });

  test("送信中に入力欄を離れても送った本文を下書きに残さない", async () => {
    const { deleteDraft, saveDraft } = await setup(() => new Promise(() => {}), "");
    const input = await screen.findByRole("textbox", { name: "メッセージを送信" });
    await userEvent.type(input, "こんにちは{Enter}");
    await waitFor(() => {
      expect(deleteDraft).toHaveBeenCalled();
    });

    cleanup();
    await new Promise((resolve) => {
      setTimeout(resolve, 50);
    });
    expect(saveDraft).not.toHaveBeenCalled();
  });

  test("復元した下書きをそのまま送っても、送信中は本文を入力欄に残す", async () => {
    const { deleteDraft } = await setup(() => new Promise(() => {}), "下書き");
    const input = await screen.findByRole("textbox", { name: "メッセージを送信" });
    await waitFor(() => {
      expect(input).toHaveValue("下書き");
    });
    await userEvent.type(input, "{Enter}");
    await waitFor(() => {
      expect(deleteDraft).toHaveBeenCalled();
    });

    await new Promise((resolve) => {
      setTimeout(resolve, 50);
    });
    expect(input).toHaveValue("下書き");
  });
});
