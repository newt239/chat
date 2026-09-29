import { useState } from "react";

import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { TextField } from "#/components/ui/TextField";
import { useLogin } from "#/features/auth/hooks/useLogin";

import { AuthCard } from "./AuthCard";
import { AuthMethods } from "./AuthMethods";

export const LoginForm = () => {
  const { t } = useTranslation();
  const login = useLogin();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  return (
    <AuthCard title={t("auth.login.title")} footer={t("auth.login.invitationOnly")}>
      <AuthMethods
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
