import { useState } from "react";

import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog";
import { Button } from "#/components/ui/Button";
import { Checkbox } from "#/components/ui/Checkbox";
import { Dialog } from "#/components/ui/Dialog";
import { TextArea } from "#/components/ui/TextArea";
import { TextField } from "#/components/ui/TextField";
import { WorkspaceMemberManager } from "#/features/workspace/components/WorkspaceMemberManager";
import { useWorkspaceActions } from "#/features/workspace/hooks/useWorkspaceActions";

import type { WorkspaceSummary } from "#/features/workspace/types";

type WorkspaceSettingsModalProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  workspace: WorkspaceSummary;
};

export const WorkspaceSettingsModal = ({
  isOpen,
  onOpenChange,
  workspace,
}: WorkspaceSettingsModalProps) => {
  const { t } = useTranslation();
  const { update, remove } = useWorkspaceActions();
  const navigate = useNavigate();

  const [name, setName] = useState(workspace.name);
  const [description, setDescription] = useState(workspace.description ?? "");
  const [isPublic, setIsPublic] = useState(workspace.isPublic);
  const [isDeleteConfirming, setIsDeleteConfirming] = useState(false);

  return (
    <Dialog
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      title={t("workspace.settings.title")}
      size="lg"
    >
      <section className="flex flex-col gap-3">
        <TextField
          label={t("workspace.settings.name")}
          value={name}
          onChange={setName}
          isRequired
        />
        <TextArea
          label={t("workspace.settings.description")}
          value={description}
          onChange={setDescription}
          rows={2}
        />
        <Checkbox isSelected={isPublic} onChange={setIsPublic}>
          {t("workspace.settings.isPublic")}
        </Checkbox>
        <div className="flex items-center justify-end gap-3">
          {update.isError && (
            <p className="m-0 flex-1 text-caption text-danger">{update.error.message}</p>
          )}
          <Button
            isPending={update.isPending}
            onPress={() => {
              update.mutate({ description, isPublic, name, workspaceId: workspace.id });
            }}
          >
            {t("common.save")}
          </Button>
        </div>
      </section>

      <hr className="my-1 h-px border-0 bg-border" />
      <WorkspaceMemberManager workspaceId={workspace.id} />
      <hr className="my-1 h-px border-0 bg-border" />

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
                onOpenChange(false);
                void navigate({ to: "/app" });
              },
            },
          );
        }}
      >
        {t("workspace.settings.deleteDescription")}
      </AlertDialog>
    </Dialog>
  );
};
