import { useState } from "react";

import { useRouter } from "@tanstack/react-router";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { CopyableUrl } from "#/components/block/CopyableUrl/CopyableUrl";
import { Button } from "#/components/ui/Button/Button";
import { Select } from "#/components/ui/Select/Select";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useInvitationActions } from "#/features/workspace/hooks/useInvitationActions";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";
import { toShareUrl } from "#/lib/shareUrl";

const roles = {
  admin: WorkspaceRole.ADMIN,
  guest: WorkspaceRole.GUEST,
  member: WorkspaceRole.MEMBER,
} as const;
type RoleKey = keyof typeof roles;
const roleKeys: readonly RoleKey[] = ["member", "guest", "admin"];

type IssuedInvitation = {
  email: string;
  url: string;
};

type InviteMemberFormProps = {
  workspaceId: string;
};

// 登録済みのメールアドレスは直ちに追加し、未登録なら一度だけ表示する招待リンクを発行する
export const InviteMemberForm = ({ workspaceId }: InviteMemberFormProps) => {
  const { t } = useTranslation();
  const router = useRouter();
  const { create } = useInvitationActions();
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<RoleKey>("member");
  const [issued, setIssued] = useState<IssuedInvitation | null>(null);

  const submit = () => {
    const target = email.trim();
    create.mutate(
      { email: target, role: roles[role], workspaceId },
      {
        onError: (error) => {
          toast(t("workspace.invite.failed"), { description: error.message, tone: "danger" });
        },
        onSuccess: ({ addedDirectly, token }) => {
          setEmail("");
          if (addedDirectly) {
            setIssued(null);
            toast(t("workspace.invite.addedDirectly", { email: target }), { tone: "success" });
            return;
          }
          const { href } = router.buildLocation({ params: { token }, to: "/invite/$token" });
          setIssued({ email: target, url: toShareUrl(href) });
        },
      },
    );
  };

  return (
    <div className="flex flex-col gap-3">
      <Form
        className="flex flex-wrap items-end gap-2"
        onSubmit={(event) => {
          event.preventDefault();
          submit();
        }}
      >
        <TextField
          className="min-w-48 flex-1"
          type="email"
          label={t("workspace.invite.email")}
          placeholder="email@example.com"
          value={email}
          onChange={setEmail}
          isRequired
        />
        <Select
          label={t("workspace.invite.role")}
          className="w-32"
          value={role}
          onChange={setRole}
          options={roleKeys.map((value) => ({ label: t(`member.role.${value}`), value }))}
        />
        <Button type="submit" isDisabled={email.trim().length === 0} isPending={create.isPending}>
          {t("workspace.invite.submit")}
        </Button>
      </Form>
      {issued !== null && (
        <div className="flex flex-col gap-2">
          <p className="m-0 rounded-md bg-accent-soft px-3 py-2 text-caption text-accent-text">
            {t("workspace.invite.linkOnce")}
          </p>
          <span className="text-xs font-semibold text-muted">
            {t("workspace.invite.link", { email: issued.email })}
          </span>
          <CopyableUrl url={issued.url} />
        </div>
      )}
    </div>
  );
};
