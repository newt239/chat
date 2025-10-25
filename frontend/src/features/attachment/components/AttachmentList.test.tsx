import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";

import { AttachmentList } from "./AttachmentList";

import type { PendingAttachment } from "../api/types";

describe("AttachmentList", () => {
  it("添付ファイルがない場合は何も表示しない", () => {
    const { container } = render(<AttachmentList attachments={[]} onRemove={vi.fn()} />);
    expect(container.firstChild).toBeNull();
  });

  it("添付ファイルを表示する", () => {
    const file = new File(["content"], "test.txt", { type: "text/plain" });
    const attachments: PendingAttachment[] = [
      {
        file,
        state: { status: "completed", attachmentId: "123" },
      },
    ];

    render(<AttachmentList attachments={attachments} onRemove={vi.fn()} />);
    expect(screen.getByText("test.txt")).toBeInTheDocument();
  });

  it("アップロード中の進捗を表示する", () => {
    const file = new File(["content"], "uploading.txt", { type: "text/plain" });
    const attachments: PendingAttachment[] = [
      {
        file,
        state: { status: "uploading", progress: 50 },
      },
    ];

    render(<AttachmentList attachments={attachments} onRemove={vi.fn()} />);
    expect(screen.getByText(/50%/)).toBeInTheDocument();
  });

  it("エラーメッセージを表示する", () => {
    const file = new File(["content"], "error.txt", { type: "text/plain" });
    const attachments: PendingAttachment[] = [
      {
        file,
        state: { status: "error", error: "アップロードに失敗しました" },
      },
    ];

    render(<AttachmentList attachments={attachments} onRemove={vi.fn()} />);
    expect(screen.getByText(/アップロードに失敗しました/)).toBeInTheDocument();
  });

  it("削除ボタンをクリックすると onRemove が呼ばれる", async () => {
    const user = userEvent.setup();
    const onRemove = vi.fn();
    const file = new File(["content"], "test.txt", { type: "text/plain" });
    const attachments: PendingAttachment[] = [
      {
        file,
        state: { status: "completed", attachmentId: "123" },
      },
    ];

    render(<AttachmentList attachments={attachments} onRemove={onRemove} />);
    const removeButton = screen.getByLabelText("削除");
    await user.click(removeButton);

    expect(onRemove).toHaveBeenCalledWith(0);
  });

  it("アップロード中は削除ボタンが無効化される", () => {
    const file = new File(["content"], "uploading.txt", { type: "text/plain" });
    const attachments: PendingAttachment[] = [
      {
        file,
        state: { status: "uploading", progress: 50 },
      },
    ];

    render(<AttachmentList attachments={attachments} onRemove={vi.fn()} />);
    const removeButton = screen.getByLabelText("削除");
    expect(removeButton).toBeDisabled();
  });
});
