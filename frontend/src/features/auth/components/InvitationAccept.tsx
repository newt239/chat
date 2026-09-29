import { useState } from "react";

import { useQuery } from "@connectrpc/connect-query";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Link } from "#/components/ui/Link/Link";
import { TextField } from "#/components/ui/TextField/TextField";
import { useSignUpWithInvitation } from "#/features/auth/hooks/useSignUpWithInvitation";
import { InvitationService } from "#/gen/chat/v1/invitation_service_pb";

import { AuthCard } from "./AuthCard";
import { AuthMethods } from "./AuthMethods";

type InvitationAcceptProps = {
  token: string;
};

// 招待リンクの受け口。Google でログインするか、パスワードを設定してアカウントを作る
export const InvitationAccept = ({ token }: InvitationAcceptProps) => {
  const { t } = useTranslation();
  const invitation = useQuery(InvitationService.method.getInvitation, { token });
  const signUp = useSignUpWithInvitation();
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
          <Form
            className="flex flex-col gap-4"
            onSubmit={(event) => {
              event.preventDefault();
              signUp.mutate({ displayName, password, token });
            }}
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
            {signUp.isError && (
              <p className="m-0 text-caption text-danger">{signUp.error.message}</p>
            )}
            <Button type="submit" isPending={signUp.isPending}>
              {t("auth.invite.submit")}
            </Button>
          </Form>
        }
      />
    </AuthCard>
  );
};
