import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Button } from "react-aria-components";
import { describe, expect, test, vi } from "vite-plus/test";

import { Menu } from "#/components/ui/Menu/Menu";
import {
  ChannelCategorySchema,
  ChannelCategoryService,
} from "#/gen/chat/v1/channel_category_service_pb";
import { ChannelSchema } from "#/gen/chat/v1/channel_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MoveToCategoryMenu } from "./MoveToCategoryMenu";

import type { SetChannelCategoryRequest } from "#/gen/chat/v1/channel_category_service_pb";

const dev = create(ChannelSchema, { id: "dev", name: "dev" });
const frontend = create(ChannelSchema, { id: "fe", name: "dev/frontend", parentId: "dev" });

describe("MoveToCategoryMenu", () => {
  test("親の所属カテゴリに印を付け、選んだカテゴリへ移す", async () => {
    const assigned = vi.fn<(req: SetChannelCategoryRequest) => void>();
    await renderWithProviders(
      <Menu trigger={<Button>開く</Button>}>
        <MoveToCategoryMenu workspaceId="ws1" channel={frontend} channels={[dev, frontend]} />
      </Menu>,
      "/app/ws1",
      (routes) => {
        routes.rpc(ChannelCategoryService.method.listChannelCategories, () => ({
          categories: [
            create(ChannelCategorySchema, { channelIds: ["dev"], id: "work", name: "仕事" }),
            create(ChannelCategorySchema, { id: "fun", name: "趣味" }),
          ],
        }));
        routes.rpc(ChannelCategoryService.method.setChannelCategory, (req) => {
          assigned(req);
          return {};
        });
      },
    );

    await userEvent.click(await screen.findByRole("button", { name: "開く" }));
    await userEvent.click(screen.getByRole("menuitem", { name: "カテゴリに移動" }));
    const current = await screen.findByRole("menuitem", { name: "仕事" });
    expect(current.querySelector("svg")).not.toHaveClass("invisible");
    await userEvent.click(screen.getByRole("menuitem", { name: "趣味" }));

    await waitFor(() => {
      expect(assigned).toHaveBeenCalledWith(
        expect.objectContaining({ categoryId: "fun", channelId: "fe" }),
      );
    });
  });
});
