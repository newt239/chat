import type { ReactNode } from "react";

import { useQuery } from "@connectrpc/connect-query";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";
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
  const config = useQuery(
    AuthService.method.getAuthConfig,
    {},
    { select: (res) => res.passwordAuthEnabled },
  );
  const passwordAuthEnabled = config.data === true && passwordForm !== null;
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
      {/* 設定が取れるまでフォームを出せないので、空白にせず読み込み中と失敗を示す */}
      {config.isPending && passwordForm !== null && (
        <div aria-busy className="flex flex-col gap-3">
          <Skeleton className="h-9 w-full" />
          <Skeleton className="h-9 w-full" />
          <Skeleton className="h-9 w-full" />
        </div>
      )}
      {config.isError && (
        <div role="alert" className="flex flex-col items-center gap-2 text-caption text-danger">
          {t("auth.login.configFailed")}
          <Button
            variant="secondary"
            size="sm"
            isPending={config.isFetching}
            onPress={() => {
              void config.refetch();
            }}
          >
            {t("common.retry")}
          </Button>
        </div>
      )}
    </>
  );
};
