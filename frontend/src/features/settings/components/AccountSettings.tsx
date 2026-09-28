import { useState } from "react";

import { useAtomValue, useSetAtom } from "jotai";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog";
import { Button } from "#/components/ui/Button";
import { TextField } from "#/components/ui/TextField";
import { toast } from "#/components/ui/toast";
import { useLogout } from "#/features/auth/hooks/useLogout";
import { userAtom } from "#/providers/store/auth";
import { setRightSidePanelViewAtom, settingsSectionAtom } from "#/providers/store/ui";

import { useDeleteAccount, useUpdatePassword } from "../hooks/useAccount";
import { SettingRow } from "./SettingRow";

export const AccountSettings = () => {
  const { t } = useTranslation();
  const user = useAtomValue(userAtom);
  const setRightPanel = useSetAtom(setRightSidePanelViewAtom);
  const setSettingsSection = useSetAtom(settingsSectionAtom);
  const updatePassword = useUpdatePassword();
  const deleteAccount = useDeleteAccount();
  const logout = useLogout();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [isDeleteConfirming, setIsDeleteConfirming] = useState(false);

  return (
    <div className="flex flex-col gap-5">
      <section className="flex flex-col">
        <SettingRow
          title={t("settings.account.profile")}
          description={t("settings.account.profileDescription")}
        >
          <Button
            variant="secondary"
            onPress={() => {
              if (user) {
                setSettingsSection(null);
                setRightPanel({ type: "user-profile", userId: user.id });
              }
            }}
          >
            {t("settings.account.openProfile")}
          </Button>
        </SettingRow>
        <SettingRow title={t("auth.email")} description={user?.email ?? null}>
          <Button
            variant="secondary"
            isPending={logout.isPending}
            onPress={() => {
              logout.mutate({});
            }}
          >
            {t("shell.me.logout")}
          </Button>
        </SettingRow>
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
