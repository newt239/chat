import { useEffect, useState } from "react";

import { IconBrandGoogle } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { useLoginWithGoogleCode } from "#/features/auth/hooks/useLoginWithGoogleCode";
import { startGoogleOAuth, takeGoogleOAuthResult } from "#/features/auth/utils/googleOAuthPkce";
import { openExternal } from "#/lib/platform/openExternal";

type GoogleSignInButtonNativeProps = {
  // 参加リンクから来たときのワークスペース
  workspaceId: string | null;
};

// React Compiler はコンポーネント内の動的 import を扱えないため外に出す
const loadDeepLink = () => import("#/lib/platform/tauri/deepLink");

// ネイティブアプリでは WebView で Google のボタンを使えないため、システムのブラウザでログインしてディープリンクで戻る
export const GoogleSignInButtonNative = ({ workspaceId }: GoogleSignInButtonNativeProps) => {
  const { t } = useTranslation();
  const { error, isError, isPending, mutate } = useLoginWithGoogleCode(workspaceId);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let stop: (() => void) | null = null;
    let active = true;
    void loadDeepLink().then(({ listenDeepLinks }) => {
      if (!active) {
        return;
      }
      stop = listenDeepLinks((url) => {
        const result = takeGoogleOAuthResult(url);
        if (result === null) {
          return;
        }
        if (result.type === "error") {
          setFailed(true);
          return;
        }
        setFailed(false);
        mutate({
          code: result.code,
          codeVerifier: result.codeVerifier,
          nonce: result.nonce,
          workspaceId: result.workspaceId ?? undefined,
        });
      });
    });
    return () => {
      active = false;
      stop?.();
    };
  }, [mutate]);

  return (
    <div className="flex flex-col items-center gap-2">
      <Button
        variant="secondary"
        className="h-10 w-80 max-w-full"
        isPending={isPending}
        onPress={() => {
          setFailed(false);
          void startGoogleOAuth(workspaceId).then(openExternal);
        }}
      >
        <IconBrandGoogle aria-hidden />
        {t("auth.google.continueInBrowser")}
      </Button>
      {failed && <p className="m-0 text-caption text-danger">{t("auth.google.browserFailed")}</p>}
      {isError && <p className="m-0 text-caption text-danger">{error.message}</p>}
    </div>
  );
};
