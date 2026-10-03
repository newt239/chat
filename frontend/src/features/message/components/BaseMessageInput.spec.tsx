import { create } from "@bufbuild/protobuf";
import { Code, ConnectError } from "@connectrpc/connect";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { DraftService } from "#/gen/chat/v1/draft_service_pb";
import { MessageSchema } from "#/gen/chat/v1/message_pb";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { BaseMessageInput } from "./BaseMessageInput";

const setup = async (fails: boolean) => {
  const deleteDraft = vi.fn<() => void>();
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
      routes.rpc(DraftService.method.saveDraft, () => ({}));
      routes.rpc(DraftService.method.deleteDraft, () => {
        deleteDraft();
        return {};
      });
      routes.rpc(MessageService.method.createMessage, (req) => {
        if (fails) {
          throw new ConnectError("failed", Code.Unavailable);
        }
        return { message: create(MessageSchema, { body: req.body, id: "m1" }) };
      });
    },
  );
  return { deleteDraft };
};

describe("BaseMessageInput", () => {
  test("送信に失敗したら本文と下書きを残す", async () => {
    const { deleteDraft } = await setup(true);
    const input = await screen.findByRole("textbox", { name: "メッセージを送信" });
    await userEvent.type(input, "こんにちは{Enter}");

    expect(await screen.findByText("failed")).toBeInTheDocument();
    expect(input).toHaveValue("こんにちは");
    expect(deleteDraft).not.toHaveBeenCalled();
  });

  test("送信に成功したら本文と下書きを消す", async () => {
    const { deleteDraft } = await setup(false);
    const input = await screen.findByRole("textbox", { name: "メッセージを送信" });
    await userEvent.type(input, "こんにちは{Enter}");

    await waitFor(() => {
      expect(input).toHaveValue("");
    });
    expect(deleteDraft).toHaveBeenCalled();
  });
});
