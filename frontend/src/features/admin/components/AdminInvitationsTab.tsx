import { useState } from "react";

import { useQuery } from "@connectrpc/connect-query";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { Button } from "#/components/ui/Button/Button";
import { cn } from "#/components/ui/styles/styles";
import { toast } from "#/components/ui/ToastRegion/toast";
import { AdminSignupSettings } from "#/features/admin/components/AdminSignupSettings";
import { tableClassNames } from "#/features/admin/utils/tableClassNames";
import { workspaceRoleKey } from "#/features/member/utils/workspaceRoleKeys";
import { InvitationService } from "#/gen/chat/v1/invitation_service_pb";
import { useDateFormat } from "#/hooks/useDateFormat";
import { toDate } from "#/lib/timestamp";

import { useAdminActions } from "../hooks/useAdminActions";
import { InviteMemberForm } from "./InviteMemberForm";

import type { Invitation } from "#/gen/chat/v1/invitation_service_pb";

const columns = ["email", "role", "invitedBy", "expiresAt"] as const;

type AdminInvitationsTabProps = {
  workspaceId: string;
};

export const AdminInvitationsTab = ({ workspaceId }: AdminInvitationsTabProps) => {
  const { t } = useTranslation();
  const { formatDateTime } = useDateFormat();
  const { data: invitations = [] } = useQuery(
    InvitationService.method.listInvitations,
    { workspaceId },
    { select: (res) => res.invitations },
  );
  const { revokeInvitation } = useAdminActions();
  const [revoking, setRevoking] = useState<Invitation | null>(null);

  return (
    <div className="flex flex-col gap-5">
      <AdminSignupSettings workspaceId={workspaceId} />
      <section className="flex flex-col gap-2">
        <p className="m-0 text-caption text-muted">{t("admin.invitations.note")}</p>
        <InviteMemberForm workspaceId={workspaceId} />
      </section>
      <section className="flex flex-col gap-2">
        <h2 className="m-0 text-body-strong">{t("admin.invitations.pending")}</h2>
        {invitations.length === 0 ? (
          <p className="m-0 text-caption text-muted">{t("admin.invitations.empty")}</p>
        ) : (
          <div className={tableClassNames.wrapper}>
            <table className={tableClassNames.table}>
              <thead>
                <tr>
                  {columns.map((key) => (
                    <th key={key} scope="col" className={tableClassNames.header}>
                      {t(`admin.invitations.columns.${key}`)}
                    </th>
                  ))}
                  <th scope="col" className={tableClassNames.header}>
                    <span className="sr-only">{t("admin.members.columns.actions")}</span>
                  </th>
                </tr>
              </thead>
              <tbody>
                {invitations.map((invitation) => (
                  <tr key={invitation.id} className={tableClassNames.row}>
                    <td className={tableClassNames.cell}>{invitation.email}</td>
                    <td className={tableClassNames.cell}>
                      {t(`member.role.${workspaceRoleKey(invitation.role)}`)}
                    </td>
                    <td className={tableClassNames.cell}>{invitation.invitedByName}</td>
                    <td className={cn(tableClassNames.cell, tableClassNames.numeric)}>
                      {formatDateTime(toDate(invitation.expiresAt))}
                    </td>
                    <td className={cn(tableClassNames.cell, "text-right")}>
                      <Button
                        size="sm"
                        variant="ghost"
                        aria-label={t("admin.invitations.revokeLabel", { email: invitation.email })}
                        onPress={() => {
                          setRevoking(invitation);
                        }}
                      >
                        {t("admin.invitations.revoke")}
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
      <AlertDialog
        isOpen={revoking !== null}
        onOpenChange={() => {
          setRevoking(null);
        }}
        title={t("admin.invitations.revokeTitle", { email: revoking?.email ?? "" })}
        confirmLabel={t("admin.invitations.revoke")}
        tone="danger"
        isPending={revokeInvitation.isPending}
        onConfirm={() => {
          if (revoking === null) {
            return;
          }
          revokeInvitation.mutate(
            { invitationId: revoking.id, workspaceId },
            {
              onSettled: () => {
                setRevoking(null);
              },
              onSuccess: () => {
                toast(t("admin.invitations.revoked"), { tone: "success" });
              },
            },
          );
        }}
      >
        <p className="m-0">{t("admin.invitations.revokeBody")}</p>
      </AlertDialog>
    </div>
  );
};
