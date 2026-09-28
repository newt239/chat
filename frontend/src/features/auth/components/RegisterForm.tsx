import { useState } from "react";

import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { Link } from "#/components/ui/Link";
import { TextField } from "#/components/ui/TextField";
import { useRegister } from "#/features/auth/hooks/useRegister";

import { AuthCard } from "./AuthCard";

export const RegisterForm = () => {
  const { t } = useTranslation();
  const register = useRegister();
  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  return (
    <AuthCard
      title={t("auth.register.title")}
      footer={
        <>
          {t("auth.register.hasAccount")} <Link to="/login">{t("auth.login.title")}</Link>
        </>
      }
    >
      <Form
        className="flex flex-col gap-4"
        onSubmit={(event) => {
          event.preventDefault();
          register.mutate({ displayName, email, password });
        }}
      >
        <TextField
          label={t("auth.displayName")}
          autoComplete="nickname"
          value={displayName}
          onChange={setDisplayName}
          isRequired
        />
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
          description={t("auth.passwordRule")}
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={setPassword}
          minLength={8}
          isRequired
        />
        {register.isError && (
          <p className="m-0 text-caption text-danger">{register.error.message}</p>
        )}
        <Button type="submit" isPending={register.isPending}>
          {t("auth.register.submit")}
        </Button>
      </Form>
    </AuthCard>
  );
};
