import { useState } from "react";

import { useMutation } from "@connectrpc/connect-query";
import { useTranslation } from "react-i18next";

import { TextField } from "#/components/ui/TextField/TextField";
import { useCompleteLogin } from "#/features/auth/hooks/useCompleteLogin";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";

import { AuthCard } from "./AuthCard";
import { AuthMethods } from "./AuthMethods";
import { PasswordAuthForm } from "./PasswordAuthForm";

export const LoginPage = () => {
  const { t } = useTranslation();
  const login = useMutation(AuthService.method.login, { onSuccess: useCompleteLogin(null) });
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  return (
    <AuthCard title={t("auth.login.title")} footer={t("auth.login.invitationOnly")}>
      <AuthMethods
        workspaceId={null}
        passwordForm={
          <PasswordAuthForm
            onSubmit={() => {
              login.mutate({ email, password });
            }}
            error={login.error}
            isPending={login.isPending}
            submitLabel={t("auth.login.submit")}
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
              label={t("auth.password")}
              type="password"
              autoComplete="current-password"
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
