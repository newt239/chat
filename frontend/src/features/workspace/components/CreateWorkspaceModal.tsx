import { useState } from "react";

import { useNavigate } from "@tanstack/react-router";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { Dialog } from "#/components/ui/Dialog";
import { TextArea } from "#/components/ui/TextArea";
import { TextField } from "#/components/ui/TextField";

import { useCreateWorkspace } from "../hooks/useWorkspace";

type CreateWorkspaceModalProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
};

// CreateWorkspaceRequest.id と同じ制約
const WORKSPACE_ID_PATTERN = /^[a-z0-9][a-z0-9-]{1,10}[a-z0-9]$/;
const FORM_ID = "create-workspace";

export const CreateWorkspaceModal = ({ isOpen, onOpenChange }: CreateWorkspaceModalProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [id, setId] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isSubmitted, setIsSubmitted] = useState(false);
  const createWorkspace = useCreateWorkspace();
  const idError =
    isSubmitted && !WORKSPACE_ID_PATTERN.test(id) ? t("workspace.create.idInvalid") : undefined;

  return (
    <Dialog
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      title={t("workspace.create.title")}
      footer={
        <>
          <Button
            variant="secondary"
            onPress={() => {
              onOpenChange(false);
            }}
          >
            {t("common.cancel")}
          </Button>
          <Button type="submit" form={FORM_ID} isPending={createWorkspace.isPending}>
            {t("workspace.create.submit")}
          </Button>
        </>
      }
    >
      <Form
        id={FORM_ID}
        className="flex flex-col gap-3"
        onSubmit={(event) => {
          event.preventDefault();
          setIsSubmitted(true);
          if (!WORKSPACE_ID_PATTERN.test(id)) {
            return;
          }
          createWorkspace.mutate(
            { description: description || undefined, id, name },
            {
              onSuccess: () => {
                onOpenChange(false);
                void navigate({ params: { workspaceId: id }, to: "/app/$workspaceId" });
              },
            },
          );
        }}
      >
        <TextField
          label={t("workspace.create.id")}
          description={t("workspace.create.idDescription")}
          placeholder="team-dev"
          value={id}
          onChange={setId}
          errorMessage={idError}
          isRequired
        />
        <TextField label={t("workspace.create.name")} value={name} onChange={setName} isRequired />
        <TextArea
          label={t("workspace.create.description")}
          value={description}
          onChange={setDescription}
        />
        {createWorkspace.isError && (
          <p className="m-0 text-caption text-danger">{createWorkspace.error.message}</p>
        )}
      </Form>
    </Dialog>
  );
};
