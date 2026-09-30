import { useNavigate, useParams } from "@tanstack/react-router";

import { ChannelCategoryDialog } from "#/features/channel/components/ChannelCategoryDialog";
import { ChannelLinkDialogLoader } from "#/features/channel/components/ChannelLinkDialogLoader";
import { CreateChannelModal } from "#/features/channel/components/CreateChannelModal";
import { useChannelCategories } from "#/features/channel/hooks/useChannelCategories";
import { CreateDMModal } from "#/features/dm/components/CreateDMModal";
import { MarkdownHelpModal } from "#/features/message/components/MarkdownHelpModal";
import { UserGroupDialog } from "#/features/userGroup/components/UserGroupDialog";
import { useCanManageUserGroups } from "#/features/userGroup/hooks/useCanManageUserGroups";
import { useUserGroups } from "#/features/userGroup/hooks/useUserGroups";
import { WebhookDialogLoader } from "#/features/webhook/components/WebhookDialogLoader";
import { CreateWorkspaceModal } from "#/features/workspace/components/CreateWorkspaceModal";

import { closeDialog, openPanel } from "../utils/overlaySearch";
import { workspaceRoute } from "../utils/workspaceRoute";

type WorkspaceDialogsProps = {
  workspaceId: string;
};

// ?dialog= で開くダイアログ。開くボタンが複数の画面にあっても、ここで一度だけ描く
export const WorkspaceDialogs = ({ workspaceId }: WorkspaceDialogsProps) => {
  const navigate = useNavigate();
  const { assign, category, dialog, group, link, webhook } = workspaceRoute.useSearch();
  const channelId = useParams({ select: (params) => params.channelId, strict: false });
  const { data: groups } = useUserGroups(workspaceId);
  const canManageGroups = useCanManageUserGroups(workspaceId);
  const editingGroup = groups?.find((candidate) => candidate.id === group);
  const { data: categories } = useChannelCategories(workspaceId);
  const editingCategory = categories?.find((candidate) => candidate.id === category);

  const close = () => {
    void navigate({ search: closeDialog, to: "." });
  };
  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      close();
    }
  };

  return (
    <>
      <CreateChannelModal
        workspaceId={workspaceId}
        opened={dialog === "create-channel"}
        onClose={close}
      />
      {(dialog === "create-category" || (dialog === "edit-category" && editingCategory)) && (
        <ChannelCategoryDialog
          key={dialog}
          workspaceId={workspaceId}
          category={dialog === "edit-category" ? (editingCategory ?? null) : null}
          assignChannelId={dialog === "create-category" ? (assign ?? null) : null}
          onClose={close}
        />
      )}
      <CreateDMModal workspaceId={workspaceId} opened={dialog === "create-dm"} onClose={close} />
      <CreateWorkspaceModal isOpen={dialog === "create-workspace"} onOpenChange={onOpenChange} />
      <MarkdownHelpModal isOpen={dialog === "markdown-help"} onOpenChange={onOpenChange} />
      {canManageGroups &&
        (dialog === "create-group" || (dialog === "edit-group" && editingGroup)) && (
          <UserGroupDialog
            key={dialog}
            workspaceId={workspaceId}
            group={dialog === "edit-group" ? (editingGroup ?? null) : null}
            onClose={(savedGroupId) => {
              // 作成したグループはそのまま右パネルで開く
              if (dialog === "create-group" && savedGroupId !== null) {
                void navigate({ search: openPanel({ group: savedGroupId }), to: "." });
                return;
              }
              close();
            }}
          />
        )}
      {channelId !== undefined &&
        (dialog === "add-link" || (dialog === "edit-link" && link !== undefined)) && (
          <ChannelLinkDialogLoader
            channelId={channelId}
            linkId={dialog === "edit-link" ? (link ?? null) : null}
            onClose={close}
          />
        )}
      {channelId !== undefined &&
        (dialog === "add-webhook" || (dialog === "edit-webhook" && webhook !== undefined)) && (
          <WebhookDialogLoader
            channelId={channelId}
            webhookId={dialog === "edit-webhook" ? (webhook ?? null) : null}
            onClose={close}
          />
        )}
    </>
  );
};
