import { useState } from "react";

import { useMutation, useQuery } from "@connectrpc/connect-query";
import { getRouteApi } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Link } from "#/components/ui/Link/Link";
import { TextField } from "#/components/ui/TextField/TextField";
import { useCompleteLogin } from "#/features/auth/hooks/useCompleteLogin";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { sessionAtom } from "#/providers/store/auth";

import { AuthCard } from "./AuthCard";
import { AuthMethods } from "./AuthMethods";
import { JoinAsMember } from "./JoinAsMember";

const joinRoute = getRouteApi("/join/$workspaceId");

// 新規登録を許可したワークスペースの参加リンクの受け口。未ログインならアカウントを作って参加する
export const JoinWorkspace = () => {
  const { t } = useTranslation();
  const { workspaceId } = joinRoute.useParams();
  const info = useQuery(WorkspaceService.method.getWorkspaceSignupInfo, { workspaceId });
  const isAuthenticated = useAtomValue(sessionAtom) !== null;
  const signUp = useMutation(AuthService.method.signUp, {
    onSuccess: useCompleteLogin(workspaceId),
  });
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");

  const footer = isAuthenticated ? null : (
    <>
      {t("auth.invite.hasAccount")} <Link to="/login">{t("auth.login.title")}</Link>
    </>
  );

  if (info.isPending) {
    return null;
  }
  if (info.isError) {
    return (
      <AuthCard title={t("auth.invite.linkTitle")} footer={footer}>
        <p role="alert" className="m-0 text-center text-caption text-danger">
          {t("auth.join.notFound")}
        </p>
      </AuthCard>
    );
  }
  const { name, emailSignupEnabled } = info.data;

  return (
    <AuthCard title={t("auth.join.title", { workspace: name })} footer={footer}>
      {isAuthenticated ? (
        <JoinAsMember workspaceId={workspaceId} />
      ) : (
        <>
          <p className="m-0 text-caption text-muted">
            {t(emailSignupEnabled ? "auth.join.lead" : "auth.join.leadGoogleOnly")}
          </p>
          <AuthMethods
            workspaceId={workspaceId}
            passwordForm={
              emailSignupEnabled ? (
                <Form
                  className="flex flex-col gap-4"
                  onSubmit={(event) => {
                    event.preventDefault();
                    signUp.mutate({ displayName, email, password, workspaceId });
                  }}
                >
                  <TextField
                    label={t("auth.email")}
                    type="email"
                    autoComplete="email"
                    placeholder="email@example.com"
                    value={email}
                    onChange={setEmail}
                    isRequired
                  />
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
                    {t("auth.join.signUp")}
                  </Button>
                </Form>
              ) : null
            }
          />
        </>
      )}
    </AuthCard>
  );
};
