import { useQuery } from "@connectrpc/connect-query";
import { useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { CopyableUrl } from "#/components/block/CopyableUrl/CopyableUrl";
import { Switch } from "#/components/ui/Switch/Switch";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useAdminActions } from "#/features/admin/hooks/useAdminActions";
import { usePasswordAuthEnabled } from "#/features/auth/hooks/usePasswordAuthEnabled";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { toShareUrl } from "#/lib/shareUrl";

type AdminSignupSettingsProps = {
  workspaceId: string;
};

type SignupSettings = {
  signupEnabled?: boolean;
  emailSignupEnabled?: boolean;
};

// 参加リンクからの新規登録と、その中でメールアドレスでの登録を許可するかを切り替える
export const AdminSignupSettings = ({ workspaceId }: AdminSignupSettingsProps) => {
  const { t } = useTranslation();
  const router = useRouter();
  const passwordAuthEnabled = usePasswordAuthEnabled();
  const { updateWorkspace } = useAdminActions();
  const { data: workspace } = useQuery(
    WorkspaceService.method.getWorkspace,
    { workspaceId },
    { select: (res) => res.workspace },
  );
  if (workspace === undefined) {
    return null;
  }

  const update = (settings: SignupSettings) => {
    updateWorkspace.mutate(
      { workspaceId, ...settings },
      {
        onSuccess: () => {
          toast(t("admin.invitations.signup.updated"), { tone: "success" });
        },
      },
    );
  };
  const { href } = router.buildLocation({ params: { workspaceId }, to: "/join/$workspaceId" });

  return (
    <section className="flex flex-col gap-3">
      <h2 className="m-0 text-body-strong">{t("admin.invitations.signup.title")}</h2>
      <p className="m-0 text-caption text-muted">{t("admin.invitations.signup.note")}</p>
      <Switch
        isSelected={workspace.signupEnabled}
        isDisabled={updateWorkspace.isPending}
        onChange={(signupEnabled) => {
          update({ signupEnabled });
        }}
      >
        {t("admin.invitations.signup.enabled")}
      </Switch>
      <div className="flex flex-col gap-1 pl-6">
        <Switch
          isSelected={workspace.emailSignupEnabled}
          isDisabled={!workspace.signupEnabled || !passwordAuthEnabled || updateWorkspace.isPending}
          onChange={(emailSignupEnabled) => {
            update({ emailSignupEnabled });
          }}
        >
          {t("admin.invitations.signup.email")}
        </Switch>
        {!passwordAuthEnabled && (
          <p className="m-0 text-caption text-muted">
            {t("admin.invitations.signup.emailDisabled")}
          </p>
        )}
      </div>
      {workspace.signupEnabled && (
        <div className="flex flex-col gap-1">
          <span className="text-xs font-semibold text-muted">
            {t("admin.invitations.signup.link")}
          </span>
          <CopyableUrl url={toShareUrl(href)} />
        </div>
      )}
    </section>
  );
};
