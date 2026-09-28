import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { AdminMemberSchema, AdminService } from "#/gen/chat/v1/admin_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MemberSuspendButton } from "./MemberSuspendButton";

import type { ResumeMemberRequest, SuspendMemberRequest } from "#/gen/chat/v1/admin_service_pb";

const bob = create(AdminMemberSchema, { displayName: "Bob", userId: "u-bob" });

describe("MemberSuspendButton", () => {
  test("確認してから停止する", async () => {
    const suspend = vi.fn<(req: SuspendMemberRequest) => void>();
    await renderWithProviders(
      <MemberSuspendButton workspaceId="ws1" member={bob} />,
      "/app/ws1/admin",
      (routes) => {
        routes.rpc(AdminService.method.suspendMember, (req) => {
          suspend(req);
          return {};
        });
      },
    );

    await userEvent.click(screen.getByRole("button", { name: "停止" }));
    const dialog = screen.getByRole("alertdialog", { name: "Bob を停止しますか？" });
    expect(dialog).toHaveTextContent("投稿はそのまま残ります");
    expect(suspend).not.toHaveBeenCalled();

    await userEvent.click(screen.getByRole("button", { name: "停止する" }));
    await waitFor(() => {
      expect(suspend).toHaveBeenCalledWith(
        expect.objectContaining({ userId: "u-bob", workspaceId: "ws1" }),
      );
    });
    await waitFor(() => {
      expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    });
  });

  test("停止中なら確認なしで再開する", async () => {
    const resume = vi.fn<(req: ResumeMemberRequest) => void>();
    await renderWithProviders(
      <MemberSuspendButton
        workspaceId="ws1"
        member={create(AdminMemberSchema, {
          displayName: "Bob",
          suspendedAt: timestampFromDate(new Date()),
          userId: "u-bob",
        })}
      />,
      "/app/ws1/admin",
      (routes) => {
        routes.rpc(AdminService.method.resumeMember, (req) => {
          resume(req);
          return {};
        });
      },
    );

    await userEvent.click(screen.getByRole("button", { name: "再開" }));
    await waitFor(() => {
      expect(resume).toHaveBeenCalledWith(expect.objectContaining({ userId: "u-bob" }));
    });
  });
});
