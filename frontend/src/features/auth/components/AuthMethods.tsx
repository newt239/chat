import type { ReactNode } from "react";

import { useTranslation } from "react-i18next";

import { usePasswordAuthEnabled } from "#/features/auth/hooks/usePasswordAuthEnabled";
import { isTauri } from "#/lib/platform/platform";

import { GoogleSignInButton } from "./GoogleSignInButton";
import { GoogleSignInButtonNative } from "./GoogleSignInButtonNative";

type AuthMethodsProps = {
  passwordForm: ReactNode;
  workspaceId: string | null;
};

// Google ログインを主に置き、パスワードのフォームは補助として下に並べる
export const AuthMethods = ({ passwordForm, workspaceId }: AuthMethodsProps) => {
  const { t } = useTranslation();
  const passwordAuthEnabled = usePasswordAuthEnabled() && passwordForm !== null;
  const googleClientId = import.meta.env.VITE_GOOGLE_OAUTH_CLIENT_ID;

  return (
    <>
      {googleClientId &&
        (isTauri ? (
          <GoogleSignInButtonNative workspaceId={workspaceId} />
        ) : (
          <GoogleSignInButton clientId={googleClientId} workspaceId={workspaceId} />
        ))}
      {googleClientId && passwordAuthEnabled && (
        <div className="flex items-center gap-3 text-caption text-muted">
          <hr className="m-0 flex-1 border-border" />
          {t("auth.login.or")}
          <hr className="m-0 flex-1 border-border" />
        </div>
      )}
      {passwordAuthEnabled && passwordForm}
    </>
  );
};
