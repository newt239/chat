import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

import { RoleSelect } from "./RoleSelect";

describe("RoleSelect", () => {
  test("選んだロールを通知し、同じロールでは通知しない", async () => {
    const onChange = vi.fn<(role: WorkspaceRole) => void>();
    render(
      <RoleSelect
        ariaLabel="Bob のロール"
        value={WorkspaceRole.MEMBER}
        onChange={onChange}
        isDisabled={false}
      />,
    );
    const trigger = screen.getByRole("button", { name: /Bob のロール/ });
    expect(trigger).toHaveTextContent("メンバー");

    await userEvent.click(trigger);
    await userEvent.click(screen.getByRole("option", { name: "メンバー" }));
    expect(onChange).not.toHaveBeenCalled();

    await userEvent.click(trigger);
    await userEvent.click(screen.getByRole("option", { name: "ゲスト" }));
    expect(onChange).toHaveBeenCalledWith(WorkspaceRole.GUEST);
  });
});
