import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { AppSchema, AppService } from "#/gen/chat/v1/app_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AdminAppsTab } from "./AdminAppsTab";

describe("AdminAppsTab", () => {
  test("公式アプリには印を付けて編集させず、ほかのアプリは編集ダイアログで開く", async () => {
    const { router } = await renderWithProviders(
      <AdminAppsTab workspaceId="ws1" />,
      "/app/ws1",
      (routes) => {
        routes.rpc(AppService.method.listApps, () => ({
          apps: [
            create(AppSchema, {
              description: "リマインダーなど",
              id: "a0",
              isOfficial: true,
              name: "Chat",
            }),
            create(AppSchema, { canManage: true, id: "a1", name: "監視" }),
          ],
        }));
      },
    );

    expect(await screen.findByText("公式")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Chat を編集" })).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "監視 を編集" }));
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ app: "a1", dialog: "edit-app" });
    });
  });
});
