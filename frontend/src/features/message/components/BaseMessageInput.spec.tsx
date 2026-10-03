import { create } from "@bufbuild/protobuf";
import { Code, ConnectError } from "@connectrpc/connect";
import { cleanup, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { DraftService } from "#/gen/chat/v1/draft_service_pb";
import { MessageSchema } from "#/gen/chat/v1/message_pb";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { BaseMessageInput } from "./BaseMessageInput";

const setup = async (createMessage: (body: string) => Promise<void>) => {
  const deleteDraft = vi.fn<() => void>();
  const saveDraft = vi.fn<(body: string) => void>();
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
      routes.rpc(DraftService.method.getDraft, () => ({}));
      routes.rpc(DraftService.method.saveDraft, ({ body }) => {
        saveDraft(body);
        return {};
      });
      routes.rpc(DraftService.method.deleteDraft, () => {
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
    const { saveDraft } = await setup(() =>
      Promise.reject(new ConnectError("failed", Code.Unavailable)),
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
    const { deleteDraft } = await setup(() => Promise.resolve());
    const input = await screen.findByRole("textbox", { name: "メッセージを送信" });
    await userEvent.type(input, "こんにちは{Enter}");

    await waitFor(() => {
      expect(input).toHaveValue("");
    });
    expect(deleteDraft).toHaveBeenCalled();
  });

  test("送信中に入力欄を離れても送った本文を下書きに残さない", async () => {
    const { deleteDraft, saveDraft } = await setup(() => new Promise(() => {}));
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
});
