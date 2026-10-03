import { useState } from "react";

import { useMutation, useQuery } from "@connectrpc/connect-query";
import { getRouteApi } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";
import { TextField } from "#/components/ui/TextField/TextField";
import { useCompleteLogin } from "#/features/auth/hooks/useCompleteLogin";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { InvitationService } from "#/gen/chat/v1/invitation_service_pb";

import { AuthCard } from "./AuthCard";
import { AuthMethods } from "./AuthMethods";
import { PasswordAuthForm } from "./PasswordAuthForm";

const inviteRoute = getRouteApi("/invite/$token");

// 招待リンクの受け口。Google でログインするか、パスワードを設定してアカウントを作る
export const InvitationAccept = () => {
  const { t } = useTranslation();
  const { token } = inviteRoute.useParams();
  const invitation = useQuery(InvitationService.method.getInvitation, { token });
  const signUp = useMutation(AuthService.method.signUpWithInvitation, {
    onSuccess: useCompleteLogin(null),
  });
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");

  const footer = (
    <>
      {t("auth.invite.hasAccount")} <Link to="/login">{t("auth.login.title")}</Link>
    </>
  );

  if (invitation.isPending) {
    return null;
  }
  if (invitation.isError) {
    return (
      <AuthCard title={t("auth.invite.linkTitle")} footer={footer}>
        <p role="alert" className="m-0 text-center text-caption text-danger">
          {t("auth.invite.notFound")}
        </p>
      </AuthCard>
    );
  }
  const { workspaceName, email } = invitation.data;

  return (
    <AuthCard title={t("auth.invite.title", { workspace: workspaceName })} footer={footer}>
      <p className="m-0 text-caption text-muted">{t("auth.invite.lead", { email })}</p>
      <AuthMethods
        workspaceId={null}
        passwordForm={
          <PasswordAuthForm
            onSubmit={() => {
              signUp.mutate({ displayName, password, token });
            }}
            error={signUp.error}
            isPending={signUp.isPending}
            submitLabel={t("auth.invite.submit")}
          >
            <p className="m-0 text-caption text-muted">{t("auth.invite.passwordLead")}</p>
            <TextField
              label={t("auth.displayName")}
              autoComplete="nickname"
              value={displayName}
              onChange={setDisplayName}
              isRequired
            />
            <TextField
              label={t("auth.password")}
              description={t("auth.passwordRule")}
              type="password"
              autoComplete="new-password"
              value={password}
              onChange={setPassword}
              minLength={8}
              isRequired
            />
          </PasswordAuthForm>
        }
      />
    </AuthCard>
  );
};
