import { useState } from "react";

import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { IconImageField } from "#/components/block/IconImageField/IconImageField";
import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { Button } from "#/components/ui/Button/Button";
import { Checkbox } from "#/components/ui/Checkbox/Checkbox";
import { TextArea } from "#/components/ui/TextArea/TextArea";
import { TextField } from "#/components/ui/TextField/TextField";
import { useWorkspaceActions } from "#/features/workspace/hooks/useWorkspaceActions";
import { ImagePurpose } from "#/gen/chat/v1/image_service_pb";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

import type { Workspace as WorkspaceSummary } from "#/gen/chat/v1/workspace_service_pb";

type WorkspaceGeneralSettingsProps = {
  workspace: WorkspaceSummary;
};

// 名前・説明・公開設定。編集は管理者以上、削除はオーナーだけ（API 側でも同じ制限）
export const WorkspaceGeneralSettings = ({ workspace }: WorkspaceGeneralSettingsProps) => {
  const { t } = useTranslation();
  const { update, remove } = useWorkspaceActions();
  const navigate = useNavigate();
  const isOwner = workspace.role === WorkspaceRole.OWNER;
  const canEdit = isOwner || workspace.role === WorkspaceRole.ADMIN;

  const [name, setName] = useState(workspace.name);
  const [description, setDescription] = useState(workspace.description ?? "");
  const [isPublic, setIsPublic] = useState(workspace.isPublic);
  const [iconUrl, setIconUrl] = useState(workspace.iconUrl ?? "");
  const [isDeleteConfirming, setIsDeleteConfirming] = useState(false);

  return (
    <div className="flex flex-col gap-6">
      <section className="flex flex-col gap-3">
        {canEdit && (
          <IconImageField
            label={t("workspace.settings.icon")}
            name={name || workspace.name}
            value={iconUrl}
            onChange={setIconUrl}
            purpose={ImagePurpose.WORKSPACE_ICON}
            workspaceId={workspace.id}
          />
        )}
        <TextField
          label={t("workspace.settings.name")}
          value={name}
          onChange={setName}
          isRequired
          isDisabled={!canEdit}
        />
        <TextArea
          label={t("workspace.settings.description")}
          value={description}
          onChange={setDescription}
          rows={2}
          isDisabled={!canEdit}
        />
        <Checkbox isSelected={isPublic} onChange={setIsPublic} isDisabled={!canEdit}>
          {t("workspace.settings.isPublic")}
        </Checkbox>
        {canEdit ? (
          <div className="flex items-center justify-end gap-3">
            {update.isError && (
              <p className="m-0 flex-1 text-caption text-danger">{update.error.message}</p>
            )}
            <Button
              isPending={update.isPending}
              onPress={() => {
                update.mutate({ description, iconUrl, isPublic, name, workspaceId: workspace.id });
              }}
            >
              {t("common.save")}
            </Button>
          </div>
        ) : (
          <p className="m-0 text-caption text-muted">{t("workspace.settings.adminOnly")}</p>
        )}
      </section>

      {isOwner && (
        <section className="flex flex-wrap items-center gap-3 rounded-lg border border-danger p-3">
          <div className="flex min-w-0 flex-1 flex-col">
            <b className="text-body-strong text-danger">{t("workspace.settings.delete")}</b>
            <span className="text-caption text-muted">
              {t("workspace.settings.deleteDescription")}
            </span>
          </div>
          <Button
            variant="danger"
            onPress={() => {
              setIsDeleteConfirming(true);
            }}
          >
            {t("common.delete")}
          </Button>
        </section>
      )}

      <AlertDialog
        isOpen={isDeleteConfirming}
        onOpenChange={setIsDeleteConfirming}
        title={t("workspace.settings.deleteConfirm", { name: workspace.name })}
        confirmLabel={t("common.delete")}
        tone="danger"
        isPending={remove.isPending}
        onConfirm={() => {
          remove.mutate(
            { workspaceId: workspace.id },
            {
              onSuccess: () => {
                setIsDeleteConfirming(false);
                void navigate({ to: "/app" });
              },
            },
          );
        }}
      >
        {t("workspace.settings.deleteDescription")}
      </AlertDialog>
    </div>
  );
};
