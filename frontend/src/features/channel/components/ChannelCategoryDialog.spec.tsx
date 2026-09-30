import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import {
  ChannelCategorySchema,
  ChannelCategoryService,
} from "#/gen/chat/v1/channel_category_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelCategoryDialog } from "./ChannelCategoryDialog";

import type {
  CreateChannelCategoryRequest,
  SetChannelCategoryRequest,
  UpdateChannelCategoryRequest,
} from "#/gen/chat/v1/channel_category_service_pb";

describe("ChannelCategoryDialog", () => {
  test("作成したカテゴリへ指定のチャンネルを移して閉じる", async () => {
    const created = vi.fn<(req: CreateChannelCategoryRequest) => void>();
    const assigned = vi.fn<(req: SetChannelCategoryRequest) => void>();
    const onClose = vi.fn<() => void>();
    await renderWithProviders(
      <ChannelCategoryDialog
        workspaceId="ws1"
        category={null}
        assignChannelId="c1"
        onClose={onClose}
      />,
      "/app/ws1",
      (routes) => {
        routes.rpc(ChannelCategoryService.method.createChannelCategory, (req) => {
          created(req);
          return { category: { id: "cat1", name: req.name } };
        });
        routes.rpc(ChannelCategoryService.method.setChannelCategory, (req) => {
          assigned(req);
          return {};
        });
        routes.rpc(ChannelCategoryService.method.listChannelCategories, () => ({}));
      },
    );

    await userEvent.type(await screen.findByRole("textbox", { name: "カテゴリ名" }), " 仕事 ");
    await userEvent.click(screen.getByRole("button", { name: "作成" }));

    await waitFor(() => {
      expect(assigned).toHaveBeenCalledWith(
        expect.objectContaining({ categoryId: "cat1", channelId: "c1" }),
      );
    });
    expect(created).toHaveBeenCalledWith(
      expect.objectContaining({ name: "仕事", workspaceId: "ws1" }),
    );
    expect(onClose).toHaveBeenCalled();
  });

  test("既存のカテゴリは名前を変更する", async () => {
    const updated = vi.fn<(req: UpdateChannelCategoryRequest) => void>();
    await renderWithProviders(
      <ChannelCategoryDialog
        workspaceId="ws1"
        category={create(ChannelCategorySchema, { id: "cat1", name: "仕事" })}
        assignChannelId={null}
        onClose={() => {}}
      />,
      "/app/ws1",
      (routes) => {
        routes.rpc(ChannelCategoryService.method.updateChannelCategory, (req) => {
          updated(req);
          return {};
        });
      },
    );

    const name = await screen.findByRole("textbox", { name: "カテゴリ名" });
    expect(name).toHaveValue("仕事");
    await userEvent.clear(name);
    await userEvent.type(name, "趣味");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    await waitFor(() => {
      expect(updated).toHaveBeenCalledWith(
        expect.objectContaining({ categoryId: "cat1", name: "趣味" }),
      );
    });
  });
});
