import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { PollMode } from "#/gen/chat/v1/message_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { PollComposerDialog } from "./PollComposerDialog";

import type { PollInput } from "#/gen/chat/v1/message_pb";

const setup = async () => {
  const onConfirm = vi.fn<(poll: PollInput) => void>();
  await renderWithProviders(
    <PollComposerDialog isOpen onOpenChange={() => {}} onConfirm={onConfirm} />,
    "/app/ws1",
    () => {},
  );
  await screen.findByRole("dialog", { name: "投票を作成" });
  return { onConfirm };
};

describe("PollComposerDialog", () => {
  test("質問と 2 つ以上の選択肢がなければ投稿しない", async () => {
    const { onConfirm } = await setup();
    await userEvent.click(screen.getByRole("button", { name: "投稿" }));

    expect(screen.getByText("質問を入力してください")).toBeInTheDocument();
    expect(screen.getByText("選択肢を 2 つ以上入力してください")).toBeInTheDocument();
    expect(onConfirm).not.toHaveBeenCalled();
  });

  test("空の選択肢を除いて投票を組み立てる", async () => {
    const { onConfirm } = await setup();
    await userEvent.type(screen.getByRole("textbox", { name: /質問/ }), "昼ご飯");
    await userEvent.type(screen.getByRole("textbox", { name: "選択肢 1" }), "そば");
    await userEvent.click(screen.getByRole("button", { name: "選択肢を追加" }));
    await userEvent.type(screen.getByRole("textbox", { name: "選択肢 3" }), "うどん");
    await userEvent.click(screen.getByRole("switch", { name: "複数選択できる" }));
    await userEvent.click(screen.getByRole("button", { name: "投稿" }));

    const poll = onConfirm.mock.calls[0]?.[0];
    expect(poll?.question).toBe("昼ご飯");
    expect(poll?.mode).toBe(PollMode.TEXT);
    expect(poll?.allowMultiple).toBe(true);
    expect(poll?.options.map((option) => option.label)).toEqual(["そば", "うどん"]);
  });

  test("日程調整では日時の候補を送る", async () => {
    const { onConfirm } = await setup();
    await userEvent.click(screen.getByRole("radio", { name: "日程調整" }));
    await userEvent.type(screen.getByRole("textbox", { name: /質問/ }), "打ち上げ");
    // 1 つ目の候補だけを終日にする
    for (const checkbox of screen.getAllByRole("checkbox", { name: "終日" }).slice(0, 1)) {
      await userEvent.click(checkbox);
    }
    await userEvent.click(screen.getByRole("button", { name: "投稿" }));

    const poll = onConfirm.mock.calls[0]?.[0];
    expect(poll?.mode).toBe(PollMode.DATE);
    expect(poll?.options).toHaveLength(2);
    expect(poll?.options[0]?.allDay).toBe(true);
    expect(poll?.options[0]?.startsAt).toBeDefined();
  });
});
