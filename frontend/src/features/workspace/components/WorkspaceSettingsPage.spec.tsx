import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { WorkspaceRole, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { WorkspaceSettingsPage } from "./WorkspaceSettingsPage";

const render = (role: WorkspaceRole) =>
  renderWithProviders(
    <WorkspaceSettingsPage />,
    "/app/ws1/workspace-settings/general",
    (routes) => {
      routes.rpc(WorkspaceService.method.listWorkspaces, () => ({
        workspaces: [{ id: "ws1", isPublic: false, name: "Acme", role }],
      }));
    },
  );

describe("WorkspaceSettingsPage", () => {
  test("オーナーは編集でき、削除の操作も出る", async () => {
    await render(WorkspaceRole.OWNER);
    expect(await screen.findByRole("textbox", { name: /名前/ })).toBeEnabled();
    expect(screen.getByText("ワークスペースを削除")).toBeInTheDocument();
  });

  test("一般メンバーは閲覧だけで、削除の操作は出ない", async () => {
    await render(WorkspaceRole.MEMBER);
    expect(await screen.findByRole("textbox", { name: /名前/ })).toBeDisabled();
    expect(screen.queryByText("ワークスペースを削除")).toBeNull();
  });
});
