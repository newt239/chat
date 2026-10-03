import { useState } from "react";

import { useMutation } from "@connectrpc/connect-query";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { Button } from "#/components/ui/Button/Button";
import { LinkButton } from "#/components/ui/LinkButton/LinkButton";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";
import { LogoutConfirmDialog } from "#/features/auth/components/LogoutConfirmDialog";
import { useDisablePushBeforeSignOut } from "#/features/auth/hooks/useLogout";
import { UserService } from "#/gen/chat/v1/user_service_pb";
import { useMe } from "#/hooks/useMe";
import { openPanel } from "#/lib/overlaySearch";
import { signOut } from "#/lib/session";
import { toastError } from "#/lib/toastError";

import { SettingRow } from "./SettingRow";

export const AccountSettings = () => {
  const { t } = useTranslation();
  const { data: user } = useMe();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [isDeleteConfirming, setIsDeleteConfirming] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  // 変更後はサーバー側の全セッションが失効する
  const updatePassword = useMutation(UserService.method.updatePassword);
  const disablePush = useDisablePushBeforeSignOut();
  const deleteAccount = useMutation(UserService.method.deleteMe, {
    onError: (error) => {
      setIsDeleteConfirming(false);
      toastError(error);
    },
    onSuccess: async () => {
      await disablePush();
      signOut();
    },
  });

  return (
    <div className="flex flex-col gap-5">
      <section className="flex flex-col">
        <SettingRow title={t("settings.account.profile")} description={null}>
          <LinkButton variant="secondary" to="." search={openPanel({ profile: user?.id })}>
            {t("settings.account.openProfile")}
          </LinkButton>
        </SettingRow>
        <SettingRow title={t("auth.email")} description={user?.email ?? null}>
          <Button
            variant="secondary"
            onPress={() => {
              setIsLoggingOut(true);
            }}
          >
            {t("shell.me.logout")}
          </Button>
        </SettingRow>
        <LogoutConfirmDialog isOpen={isLoggingOut} onOpenChange={setIsLoggingOut} />
      </section>

      <Form
        className="flex max-w-sm flex-col gap-3"
        onSubmit={(event) => {
          event.preventDefault();
          updatePassword.mutate(
            { currentPassword, newPassword },
            {
              onSuccess: () => {
                setCurrentPassword("");
                setNewPassword("");
                toast(t("settings.account.passwordChanged"), { tone: "success" });
              },
            },
          );
        }}
      >
        <h3 className="m-0 text-body-strong">{t("settings.account.password")}</h3>
        <TextField
          label={t("settings.account.currentPassword")}
          type="password"
          autoComplete="current-password"
          value={currentPassword}
          onChange={setCurrentPassword}
          isRequired
        />
        <TextField
          label={t("settings.account.newPassword")}
          description={t("auth.passwordRule")}
          type="password"
          autoComplete="new-password"
          value={newPassword}
          onChange={setNewPassword}
          minLength={8}
          isRequired
          errorMessage={updatePassword.isError ? updatePassword.error.message : undefined}
        />
        <Button type="submit" className="self-start" isPending={updatePassword.isPending}>
          {t("settings.account.changePassword")}
        </Button>
      </Form>

      <section className="flex flex-wrap items-center gap-3 rounded-lg border border-danger p-3">
        <div className="flex min-w-40 flex-1 flex-col [word-break:auto-phrase]">
          <b className="text-body-strong text-danger">{t("settings.account.delete")}</b>
          <span className="text-caption text-muted">{t("settings.account.deleteDescription")}</span>
        </div>
        <Button
          variant="danger"
          onPress={() => {
            setIsDeleteConfirming(true);
          }}
        >
          {t("common.delete")}
        </Button>
      </section>
      <AlertDialog
        isOpen={isDeleteConfirming}
        onOpenChange={setIsDeleteConfirming}
        title={t("settings.account.deleteConfirm")}
        confirmLabel={t("common.delete")}
        tone="danger"
        isPending={deleteAccount.isPending}
        onConfirm={() => {
          deleteAccount.mutate({});
        }}
      >
        {t("settings.account.deleteDescription")}
      </AlertDialog>
    </div>
  );
};
