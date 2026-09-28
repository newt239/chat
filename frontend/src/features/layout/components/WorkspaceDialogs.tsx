import { useNavigate, useParams } from "@tanstack/react-router";

import { ChannelLinkDialogLoader } from "#/features/channel/components/ChannelLinkDialogLoader";
import { CreateChannelModal } from "#/features/channel/components/CreateChannelModal";
import { CreateDMModal } from "#/features/dm/components/CreateDMModal";
import { MarkdownHelpModal } from "#/features/message/components/MarkdownHelpModal";
import { SettingsDialog } from "#/features/settings/components/SettingsDialog";
import { UserGroupDialog } from "#/features/userGroup/components/UserGroupDialog";
import { useCanManageUserGroups } from "#/features/userGroup/hooks/useCanManageUserGroups";
import { useUserGroups } from "#/features/userGroup/hooks/useUserGroups";
import { CreateWorkspaceModal } from "#/features/workspace/components/CreateWorkspaceModal";
import { WorkspaceSettingsModal } from "#/features/workspace/components/WorkspaceSettingsModal";
import { useWorkspaces } from "#/features/workspace/hooks/useWorkspace";

import { closeDialog, openPanel } from "../utils/overlaySearch";
import { workspaceRoute } from "../utils/workspaceRoute";

type WorkspaceDialogsProps = {
  workspaceId: string;
};

// ?dialog= で開くダイアログ。開くボタンが複数の画面にあっても、ここで一度だけ描く
export const WorkspaceDialogs = ({ workspaceId }: WorkspaceDialogsProps) => {
  const navigate = useNavigate();
  const { dialog, group, link } = workspaceRoute.useSearch();
  const channelId = useParams({ select: (params) => params.channelId, strict: false });
  const { data: workspaces } = useWorkspaces();
  const { data: groups } = useUserGroups(workspaceId);
  const canManageGroups = useCanManageUserGroups(workspaceId);
  const workspace = workspaces?.find((candidate) => candidate.id === workspaceId);
  const editingGroup = groups?.find((candidate) => candidate.id === group);

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
      <CreateDMModal workspaceId={workspaceId} opened={dialog === "create-dm"} onClose={close} />
      <CreateWorkspaceModal isOpen={dialog === "create-workspace"} onOpenChange={onOpenChange} />
      {workspace && (
        <WorkspaceSettingsModal
          isOpen={dialog === "workspace-settings"}
          onOpenChange={onOpenChange}
          workspace={workspace}
        />
      )}
      <MarkdownHelpModal isOpen={dialog === "markdown-help"} onOpenChange={onOpenChange} />
      <SettingsDialog />
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
    </>
  );
};
