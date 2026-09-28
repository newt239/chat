import { create } from "@bufbuild/protobuf";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { AttachmentService } from "#/gen/chat/v1/attachment_service_pb";
import { MessageAttachmentSchema, MessageSchema } from "#/gen/chat/v1/message_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ImageGallery } from "./ImageGallery";

const image = (id: string, width: number, height: number) =>
  create(MessageAttachmentSchema, {
    fileName: `${id}.png`,
    id,
    media: { height, width },
    mimeType: "image/png",
  });

const render = (images: ReturnType<typeof image>[]) =>
  renderWithProviders(
    <ImageGallery
      images={images}
      message={create(MessageSchema, { id: "m1", user: { displayName: "Alice", id: "u1" } })}
    />,
    "/app/ws1/ch1",
    (routes) => {
      routes.rpc(AttachmentService.method.getDownloadUrl, ({ attachmentId }) => ({
        url: `https://storage.example.com/${attachmentId}`,
      }));
    },
  );

describe("ImageGallery", () => {
  test("1 枚の画像は寸法で枠を予約し、極端な比率なら切り取りの印を出す", async () => {
    await render([image("tall", 400, 4000)]);

    const tile = screen.getByRole("button", { name: "tall.png を拡大" });
    expect(tile).toHaveStyle({ aspectRatio: "140 / 300", width: "140px" });
    expect(within(tile).getByText("縦長 · 全体を表示")).toBeInTheDocument();
    expect(await within(tile).findByRole("img", { name: "tall.png" })).toHaveAttribute(
      "src",
      "https://storage.example.com/tall",
    );
  });

  test("4 枚を超えた分は +N で示し、ライトボックスで前後の画像へ移れる", async () => {
    await render(["a", "b", "c", "d", "e", "f"].map((id) => image(id, 800, 600)));

    expect(screen.getAllByRole("button", { name: /を拡大/ })).toHaveLength(4);
    expect(screen.getByText("+2")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "d.png を拡大" }));
    const dialog = await screen.findByRole("dialog", { name: "画像ビューア" });
    expect(within(dialog).getByText("4 / 6")).toBeInTheDocument();

    await userEvent.keyboard("{ArrowRight}");
    expect(within(dialog).getByText("5 / 6")).toBeInTheDocument();
    await userEvent.click(within(dialog).getByRole("button", { name: "前の画像" }));
    await userEvent.click(within(dialog).getByRole("button", { name: "前の画像" }));
    expect(within(dialog).getByText("3 / 6")).toBeInTheDocument();
    expect(within(dialog).getByText("c.png")).toBeInTheDocument();

    await userEvent.keyboard("{Escape}");
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).toBeNull();
    });
  });
});
