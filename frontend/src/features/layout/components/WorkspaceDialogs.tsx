import { useNavigate, useParams } from "@tanstack/react-router";

import { AppDialog } from "#/features/app/components/AppDialog";
import { useApps } from "#/features/app/hooks/useApps";
import { ChannelCategoryDialog } from "#/features/channel/components/ChannelCategoryDialog";
import { ChannelLinkDialog } from "#/features/channel/components/ChannelLinkDialog";
import { CreateChannelModal } from "#/features/channel/components/CreateChannelModal";
import { CreateDMModal } from "#/features/channel/components/CreateDMModal";
import { useChannelCategories } from "#/features/channel/hooks/useChannelCategories";
import { useChannelLinks } from "#/features/channel/hooks/useChannelLinks";
import { MarkdownHelpModal } from "#/features/message/components/MarkdownHelpModal";
import { UserGroupDialog } from "#/features/userGroup/components/UserGroupDialog";
import { useUserGroups } from "#/features/userGroup/hooks/useUserGroups";
import { CreateWorkspaceModal } from "#/features/workspace/components/CreateWorkspaceModal";
import { useIsWorkspaceAdmin } from "#/hooks/useIsWorkspaceAdmin";
import { closeDialog, openPanel, workspaceRoute } from "#/lib/overlaySearch";

type WorkspaceDialogsProps = {
  workspaceId: string;
};

// ?dialog= で開くダイアログ。開くボタンが複数の画面にあっても、ここで一度だけ描く
export const WorkspaceDialogs = ({ workspaceId }: WorkspaceDialogsProps) => {
  const navigate = useNavigate();
  const { app, assign, category, dialog, group, link, parent } = workspaceRoute.useSearch();
  const channelId = useParams({ select: (params) => params.channelId, strict: false });
  const { data: groups } = useUserGroups(workspaceId);
  const canManageGroups = useIsWorkspaceAdmin(workspaceId);
  const editingGroup = groups?.find((candidate) => candidate.id === group);
  const { data: categories } = useChannelCategories(workspaceId);
  const editingCategory = categories?.find((candidate) => candidate.id === category);
  const isLinkDialog = dialog === "add-link" || dialog === "edit-link";
  const { data: channelLinks } = useChannelLinks(
    isLinkDialog && channelId !== undefined ? channelId : null,
  );
  const editingLink = channelLinks?.links.find((candidate) => candidate.id === link);
  const { data: apps } = useApps(dialog === "edit-app" ? workspaceId : null);
  const editingApp = apps?.apps.find((candidate) => candidate.id === app && candidate.canManage);

  const close = () => {
    void navigate({ search: closeDialog, to: "." });
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
      {dialog === "create-workspace" && <CreateWorkspaceModal onClose={close} />}
      <MarkdownHelpModal isOpen={dialog === "markdown-help"} onOpenChange={close} />
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
      {(dialog === "add-app" || (dialog === "edit-app" && editingApp)) && (
        <AppDialog
          key={dialog}
          workspaceId={workspaceId}
          app={dialog === "edit-app" ? (editingApp ?? null) : null}
          initialChannelId={channelId ?? null}
          onClose={close}
        />
      )}
    </>
  );
};
