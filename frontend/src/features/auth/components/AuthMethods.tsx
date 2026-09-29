import type { ReactNode } from "react";

import { useTranslation } from "react-i18next";

import { usePasswordAuthEnabled } from "#/features/auth/hooks/usePasswordAuthEnabled";

import { GoogleSignInButton } from "./GoogleSignInButton";

type AuthMethodsProps = {
  // パスワード認証が有効なときだけ表示する
  passwordForm: ReactNode;
};

// Google ログインを主に置き、パスワードのフォームは補助として下に並べる
export const AuthMethods = ({ passwordForm }: AuthMethodsProps) => {
  const { t } = useTranslation();
  const passwordAuthEnabled = usePasswordAuthEnabled();
  const googleClientId = import.meta.env.VITE_GOOGLE_OAUTH_CLIENT_ID;

  return (
    <>
      {googleClientId && <GoogleSignInButton clientId={googleClientId} />}
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
