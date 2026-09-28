import { useState } from "react";

import { IconPlus, IconUsers } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useSetAtom } from "jotai";
import { Button as AriaButton } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { EmptyState } from "#/components/ui/EmptyState";
import { Skeleton } from "#/components/ui/Skeleton";
import { focusRing } from "#/components/ui/styles";
import { PageHeader } from "#/features/layout/components/PageHeader";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";

import { useUserGroups } from "../hooks/useUserGroups";
import { UserGroupDialog } from "./UserGroupDialog";

// ユーザーグループの一覧。押すと右パネル（モバイルでは全画面）で詳細と編集を開く
export const UserGroupListPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const { data: groups, isLoading } = useUserGroups(workspaceId);
  const setRightPanel = useSetAtom(setRightSidePanelViewAtom);
  const [isCreating, setIsCreating] = useState(false);

  return (
    <>
      <PageHeader icon={<IconUsers />} title={t("userGroup.pageTitle")}>
        <Button
          size="sm"
          onPress={() => {
            setIsCreating(true);
          }}
        >
          <IconPlus aria-hidden />
          {t("userGroup.create")}
        </Button>
      </PageHeader>
      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        {isLoading ? (
          <Skeleton className="h-20 w-full" />
        ) : groups === undefined || groups.length === 0 ? (
          <EmptyState
            icon={<IconUsers />}
            title={t("userGroup.empty")}
            description={t("userGroup.emptyHint")}
          />
        ) : (
          <ul className="m-0 grid list-none grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-2.5 p-0">
            {groups.map((group) => (
              <li key={group.id}>
                <AriaButton
                  onPress={() => {
                    setRightPanel({ groupId: group.id, type: "user-group" });
                  }}
                  className={`flex h-full w-full cursor-pointer flex-col items-start gap-1 rounded-xl border border-border bg-surface p-3.5 text-left font-sans data-hovered:border-border-strong data-hovered:bg-hover ${focusRing}`}
                >
                  <span className="text-sm font-bold text-accent-text">@{group.name}</span>
                  {group.description !== undefined && group.description.length > 0 && (
                    <span className="line-clamp-2 text-[12.5px] text-muted">
                      {group.description}
                    </span>
                  )}
                </AriaButton>
              </li>
            ))}
          </ul>
        )}
      </div>
      {isCreating && (
        <UserGroupDialog
          workspaceId={workspaceId}
          group={null}
          onClose={(savedGroupId) => {
            setIsCreating(false);
            if (savedGroupId !== null) {
              setRightPanel({ groupId: savedGroupId, type: "user-group" });
            }
          }}
        />
      )}
    </>
  );
};
