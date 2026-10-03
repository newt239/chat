import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { useLogout } from "#/features/auth/hooks/useLogout";

type LogoutConfirmDialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
};

export const LogoutConfirmDialog = ({ isOpen, onOpenChange }: LogoutConfirmDialogProps) => {
  const { t } = useTranslation();
  const logout = useLogout();
  return (
    <AlertDialog
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      title={t("auth.logoutConfirm.title")}
      confirmLabel={t("shell.me.logout")}
      tone="danger"
      isPending={logout.isPending}
      onConfirm={() => {
        logout.mutate({});
      }}
    >
      <p className="m-0">{t("auth.logoutConfirm.body")}</p>
    </AlertDialog>
  );
};
