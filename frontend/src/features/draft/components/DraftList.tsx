import { IconPencil } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/ui/EmptyState/EmptyState";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useConversationLabel } from "#/features/channel/hooks/useConversationLabel";
import { toastError } from "#/lib/toastError";

import { useDeleteDraft, useDrafts } from "../hooks/useDrafts";
import { DraftListItem } from "./DraftListItem";

type DraftListProps = {
  workspaceId: string;
};

export const DraftList = ({ workspaceId }: DraftListProps) => {
  const { t } = useTranslation();
  const { data: drafts, isLoading } = useDrafts(workspaceId);
  const labelOf = useConversationLabel(workspaceId);
  const deleteDraft = useDeleteDraft();

  if (isLoading) {
    return (
      <div className="flex flex-col gap-2 p-3">
        <Skeleton className="h-14 w-full" />
        <Skeleton className="h-14 w-full" />
      </div>
    );
  }
  if (!drafts || drafts.length === 0) {
    return (
      <EmptyState
        icon={<IconPencil />}
        title={t("draft.list.empty")}
        description={t("draft.list.emptyHint")}
      />
    );
  }

  return (
    <div className="flex flex-col gap-2 p-3">
      {drafts.map((draft) => (
        <DraftListItem
          key={draft.id}
          workspaceId={workspaceId}
          draft={draft}
          label={labelOf(draft.channelId)}
          onDelete={() => {
            deleteDraft.mutate(
              { channelId: draft.channelId, parentId: draft.parentId },
              {
                onError: toastError,
                onSuccess: () => {
                  toast(t("draft.list.deleted"), { tone: "success" });
                },
              },
            );
          }}
        />
      ))}
    </div>
  );
};
