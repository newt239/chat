import { useState } from "react";

import { useMutation } from "@connectrpc/connect-query";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { TextField } from "#/components/ui/TextField/TextField";
import { useCompleteLogin } from "#/features/auth/hooks/useCompleteLogin";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";

import { AuthCard } from "./AuthCard";
import { AuthMethods } from "./AuthMethods";

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
          <Form
            className="flex flex-col gap-4"
            onSubmit={(event) => {
              event.preventDefault();
              login.mutate({ email, password });
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
              label={t("auth.password")}
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={setPassword}
              minLength={8}
              isRequired
            />
            {login.isError && <p className="m-0 text-caption text-danger">{login.error.message}</p>}
            <Button type="submit" isPending={login.isPending}>
              {t("auth.login.submit")}
            </Button>
          </Form>
        }
      />
    </AuthCard>
  );
};
