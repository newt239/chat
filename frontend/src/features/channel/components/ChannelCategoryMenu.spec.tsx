import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import {
  ChannelCategorySchema,
  ChannelCategoryService,
} from "#/gen/chat/v1/channel_category_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelCategoryMenu } from "./ChannelCategoryMenu";

import type {
  DeleteChannelCategoryRequest,
  ReorderChannelCategoriesRequest,
} from "#/gen/chat/v1/channel_category_service_pb";

const category = create(ChannelCategorySchema, { id: "a", name: "仕事" });

const setup = async () => {
  const reorder = vi.fn<(req: ReorderChannelCategoriesRequest) => void>();
  const remove = vi.fn<(req: DeleteChannelCategoryRequest) => void>();
  await renderWithProviders(
    <ChannelCategoryMenu workspaceId="ws1" category={category} categoryIds={["a", "b"]} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(ChannelCategoryService.method.listChannelCategories, () => ({}));
      routes.rpc(ChannelCategoryService.method.reorderChannelCategories, (req) => {
        reorder(req);
        return {};
      });
      routes.rpc(ChannelCategoryService.method.deleteChannelCategory, (req) => {
        remove(req);
        return {};
      });
    },
  );
  await userEvent.click(await screen.findByRole("button", { name: "仕事 の操作" }));
  return { remove, reorder };
};

describe("ChannelCategoryMenu", () => {
  test("下へ移動すると全カテゴリの新しい並びを送る。先頭では上へ移動できない", async () => {
    const { reorder } = await setup();
    expect(screen.getByRole("menuitem", { name: "上へ移動" })).toHaveAttribute(
      "aria-disabled",
      "true",
    );
    await userEvent.click(screen.getByRole("menuitem", { name: "下へ移動" }));
    await waitFor(() => {
      expect(reorder).toHaveBeenCalledWith(
        expect.objectContaining({ categoryIds: ["b", "a"], workspaceId: "ws1" }),
      );
    });
  });

  test("確認してから削除する", async () => {
    const { remove } = await setup();
    await userEvent.click(screen.getByRole("menuitem", { name: "カテゴリを削除" }));
    expect(await screen.findByRole("alertdialog")).toHaveTextContent("「仕事」を削除しますか？");
    expect(remove).not.toHaveBeenCalled();
    await userEvent.click(screen.getByRole("button", { name: "カテゴリを削除" }));
    await waitFor(() => {
      expect(remove).toHaveBeenCalledWith(expect.objectContaining({ categoryId: "a" }));
    });
  });
});
