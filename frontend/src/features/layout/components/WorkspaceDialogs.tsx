import { skipToken, useQuery } from "@connectrpc/connect-query";
import { useNavigate, useParams } from "@tanstack/react-router";

import { AppDialogLoader } from "#/features/app/components/AppDialogLoader";
import { ChannelCategoryDialog } from "#/features/channel/components/ChannelCategoryDialog";
import { ChannelLinkDialog } from "#/features/channel/components/ChannelLinkDialog";
import { CreateChannelModal } from "#/features/channel/components/CreateChannelModal";
import { useChannelCategories } from "#/features/channel/hooks/useChannelCategories";
import { CreateDMModal } from "#/features/dm/components/CreateDMModal";
import { MarkdownHelpModal } from "#/features/message/components/MarkdownHelpModal";
import { UserGroupDialog } from "#/features/userGroup/components/UserGroupDialog";
import { useUserGroups } from "#/features/userGroup/hooks/useUserGroups";
import { CreateWorkspaceModal } from "#/features/workspace/components/CreateWorkspaceModal";
import { ChannelLinkService } from "#/gen/chat/v1/channel_link_service_pb";
import { useMyWorkspaceRole } from "#/hooks/useMyWorkspaceRole";
import { isAdminRole } from "#/lib/isAdminRole";

import { closeDialog, openPanel } from "../utils/overlaySearch";
import { workspaceRoute } from "../utils/workspaceRoute";

type WorkspaceDialogsProps = {
  workspaceId: string;
};

// ?dialog= で開くダイアログ。開くボタンが複数の画面にあっても、ここで一度だけ描く
export const WorkspaceDialogs = ({ workspaceId }: WorkspaceDialogsProps) => {
  const navigate = useNavigate();
  const { app, assign, category, dialog, group, link, parent } = workspaceRoute.useSearch();
  const channelId = useParams({ select: (params) => params.channelId, strict: false });
  const { data: groups } = useUserGroups(workspaceId);
  const canManageGroups = isAdminRole(useMyWorkspaceRole(workspaceId).data);
  const editingGroup = groups?.find((candidate) => candidate.id === group);
  const { data: categories } = useChannelCategories(workspaceId);
  const editingCategory = categories?.find((candidate) => candidate.id === category);
  const isLinkDialog = dialog === "add-link" || dialog === "edit-link";
  const { data: channelLinks } = useQuery(
    ChannelLinkService.method.listChannelLinks,
    isLinkDialog && channelId !== undefined ? { channelId } : skipToken,
  );
  const editingLink = channelLinks?.links.find((candidate) => candidate.id === link);

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
      {dialog === "create-channel" && (
        <CreateChannelModal
          key={parent ?? ""}
          workspaceId={workspaceId}
          parentId={parent ?? null}
          onClose={close}
        />
      )}
      {(dialog === "create-category" || (dialog === "edit-category" && editingCategory)) && (
        <ChannelCategoryDialog
          key={dialog}
          workspaceId={workspaceId}
          category={dialog === "edit-category" ? (editingCategory ?? null) : null}
          assignChannelId={dialog === "create-category" ? (assign ?? null) : null}
          onClose={close}
        />
      )}
      {dialog === "create-dm" && <CreateDMModal workspaceId={workspaceId} onClose={close} />}
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
        channelLinks?.canEdit &&
        (dialog === "add-link" || (dialog === "edit-link" && editingLink)) && (
          <ChannelLinkDialog
            key={dialog}
            channelId={channelId}
            link={dialog === "edit-link" ? (editingLink ?? null) : null}
            onClose={close}
          />
        )}
      {(dialog === "add-app" || (dialog === "edit-app" && app !== undefined)) && (
        <AppDialogLoader
          workspaceId={workspaceId}
          appId={dialog === "edit-app" ? (app ?? null) : null}
          initialChannelId={channelId ?? null}
          onClose={close}
        />
      )}
    </>
  );
};
