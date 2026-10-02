import { useEffect, useRef, useState } from "react";

import { useTranslation } from "react-i18next";

import { useLoginWithGoogle } from "#/features/auth/hooks/useLoginWithGoogle";
import { loadGoogleIdentity } from "#/features/auth/utils/googleIdentity";
import { usePreferences } from "#/hooks/usePreferences";

type GoogleSignInButtonProps = {
  clientId: string;
  // 参加リンクから来たときのワークスペース
  workspaceId: string | null;
};

// Google Identity Services が描画するボタン。受け取った ID トークンをサーバーで検証してログインする
export const GoogleSignInButton = ({ clientId, workspaceId }: GoogleSignInButtonProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const loginWithGoogle = useLoginWithGoogle(workspaceId);
  const { mutate } = loginWithGoogle;
  const containerRef = useRef<HTMLDivElement>(null);
  const [loadFailed, setLoadFailed] = useState(false);

  useEffect(() => {
    let active = true;
    loadGoogleIdentity().then(
      () => {
        const container = containerRef.current;
        if (!active || container === null) {
          return;
        }
        globalThis.google.accounts.id.initialize({
          callback: ({ credential }) => {
            mutate({ idToken: credential, workspaceId: workspaceId ?? undefined });
          },
          client_id: clientId,
        });
        globalThis.google.accounts.id.renderButton(container, {
          locale,
          size: "large",
          text: "continue_with",
          theme: "outline",
          type: "standard",
          width: 320,
        });
      },
      () => {
        if (active) {
          setLoadFailed(true);
        }
      },
    );
    return () => {
      active = false;
    };
  }, [clientId, locale, mutate, workspaceId]);

  return (
    <div className="flex flex-col items-center gap-2">
      <div ref={containerRef} className="flex min-h-10 justify-center" />
      {loadFailed && <p className="m-0 text-caption text-danger">{t("auth.google.loadFailed")}</p>}
      {loginWithGoogle.isError && (
        <p className="m-0 text-caption text-danger">{loginWithGoogle.error.message}</p>
      )}
    </div>
  );
};
