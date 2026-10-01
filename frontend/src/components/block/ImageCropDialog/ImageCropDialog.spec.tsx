import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ImageCropDialog } from "./ImageCropDialog";

describe("ImageCropDialog", () => {
  test("画像がなければ開かない", () => {
    render(
      <ImageCropDialog
        src={null}
        isPending={false}
        onCancel={vi.fn<() => void>()}
        onCrop={vi.fn<(image: Blob) => void>()}
      />,
    );
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  test("範囲が決まるまで適用できず、キャンセルを伝える", async () => {
    const onCancel = vi.fn<() => void>();
    render(
      <ImageCropDialog
        src="blob:icon"
        isPending={false}
        onCancel={onCancel}
        onCrop={vi.fn<(image: Blob) => void>()}
      />,
    );
    expect(await screen.findByRole("dialog", { name: "画像を切り抜く" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "適用" })).toBeDisabled();

    await userEvent.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(onCancel).toHaveBeenCalled();
  });
});
