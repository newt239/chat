import { useState } from "react";

import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog";
import { Button } from "#/components/ui/Button";
import { toast } from "#/components/ui/toast";
import { useAdminActions } from "#/features/admin/hooks/useAdminActions";

import type { AdminMember } from "#/gen/chat/v1/admin_service_pb";

type MemberSuspendButtonProps = {
  workspaceId: string;
  member: AdminMember;
};

// 停止は確認を挟み、再開はすぐに行う
export const MemberSuspendButton = ({ workspaceId, member }: MemberSuspendButtonProps) => {
  const { t } = useTranslation();
  const { suspend, resume } = useAdminActions();
  const [isConfirming, setIsConfirming] = useState(false);
  const name = member.displayName;
  const input = { userId: member.userId, workspaceId };

  if (member.suspendedAt !== undefined) {
    return (
      <Button
        variant="secondary"
        size="sm"
        isPending={resume.isPending}
        onPress={() => {
          resume.mutate(input, {
            onSuccess: () => {
              toast(t("admin.members.resumed", { name }), { tone: "success" });
            },
          });
        }}
      >
        {t("admin.members.resume")}
      </Button>
    );
  }
  return (
    <>
      <Button
        variant="secondary"
        size="sm"
        onPress={() => {
          setIsConfirming(true);
        }}
      >
        {t("admin.members.suspend")}
      </Button>
      <AlertDialog
        isOpen={isConfirming}
        onOpenChange={setIsConfirming}
        title={t("admin.members.suspendTitle", { name })}
        confirmLabel={t("admin.members.suspendConfirm")}
        tone="danger"
        isPending={suspend.isPending}
        onConfirm={() => {
          suspend.mutate(input, {
            onSettled: () => {
              setIsConfirming(false);
            },
            onSuccess: () => {
              toast(t("admin.members.suspended", { name }), { tone: "success" });
            },
          });
        }}
      >
        <p className="m-0">{t("admin.members.suspendBody")}</p>
      </AlertDialog>
    </>
  );
};
