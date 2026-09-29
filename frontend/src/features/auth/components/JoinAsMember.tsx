import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { LinkButton } from "#/components/ui/LinkButton/LinkButton";
import { useJoinPublicWorkspace } from "#/features/workspace/hooks/usePublicWorkspaces";
import { useWorkspaces } from "#/features/workspace/hooks/useWorkspace";
import { userAtom } from "#/providers/store/auth";

type JoinAsMemberProps = {
  workspaceId: string;
};

// ログイン済みのユーザーが参加リンクを開いたときは、今のアカウントのまま参加する
export const JoinAsMember = ({ workspaceId }: JoinAsMemberProps) => {
  const { t } = useTranslation();
  const user = useAtomValue(userAtom);
  const navigate = useNavigate();
  const workspaces = useWorkspaces();
  const join = useJoinPublicWorkspace();

  if (workspaces.isPending) {
    return null;
  }
  const joined = workspaces.data?.some((workspace) => workspace.id === workspaceId) === true;

  return (
    <>
      <p className="m-0 text-caption text-muted">
        {joined
          ? t("auth.join.alreadyJoined")
          : t("auth.join.loggedInLead", { name: user?.displayName ?? "" })}
      </p>
      {join.isError && <p className="m-0 text-caption text-danger">{join.error.message}</p>}
      {joined ? (
        <LinkButton to="/app/$workspaceId" params={{ workspaceId }}>
          {t("auth.join.open")}
        </LinkButton>
      ) : (
        <Button
          isPending={join.isPending}
          onPress={() => {
            join.mutate(
              { workspaceId },
              {
                onSuccess: () => {
                  void navigate({ params: { workspaceId }, to: "/app/$workspaceId" });
                },
              },
            );
          }}
        >
          {t("auth.join.submit")}
        </Button>
      )}
    </>
  );
};
