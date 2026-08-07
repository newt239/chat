import { useState } from "react";

import { Modal, TextInput, Textarea, Button, Text } from "@mantine/core";

import { useCreateWorkspace } from "../hooks/useWorkspace";

type CreateWorkspaceModalProps = {
  opened: boolean;
  onClose: () => void;
};

// OpenAPI の CreateWorkspaceRequest.id と同じ制約
const WORKSPACE_ID_PATTERN = /^[a-z0-9][a-z0-9-]*[a-z0-9]$/;
const WORKSPACE_ID_MIN_LENGTH = 3;
const WORKSPACE_ID_MAX_LENGTH = 12;

const validateWorkspaceId = (value: string) => {
  if (value.length < WORKSPACE_ID_MIN_LENGTH || value.length > WORKSPACE_ID_MAX_LENGTH) {
    return `${WORKSPACE_ID_MIN_LENGTH}文字以上${WORKSPACE_ID_MAX_LENGTH}文字以内で入力してください`;
  }
  if (!WORKSPACE_ID_PATTERN.test(value)) {
    return "小文字英数字とハイフンのみ使用できます（先頭と末尾にハイフンは使えません）";
  }
  return null;
};

export const CreateWorkspaceModal = ({ opened, onClose }: CreateWorkspaceModalProps) => {
  const [id, setId] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [idError, setIdError] = useState<string | null>(null);
  const createWorkspace = useCreateWorkspace();

  const resetForm = () => {
    setId("");
    setName("");
    setDescription("");
    setIdError(null);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const validationError = validateWorkspaceId(id);
    if (validationError !== null) {
      setIdError(validationError);
      return;
    }
    setIdError(null);

    createWorkspace.mutate(
      { id, name, description: description || undefined },
      {
        onSuccess: () => {
          resetForm();
          onClose();
        },
      },
    );
  };

  return (
    <Modal opened={opened} onClose={onClose} title="新規ワークスペース作成">
      <form onSubmit={handleSubmit}>
        <TextInput
          label="ワークスペースID"
          description="URL に使われます。小文字英数字とハイフンで3〜12文字"
          placeholder="例: team-dev"
          value={id}
          onChange={(e) => {
            setId(e.currentTarget.value);
          }}
          error={idError}
          required
          className="mb-4"
        />

        <TextInput
          label="ワークスペース名"
          placeholder="例: チーム開発"
          value={name}
          onChange={(e) => {
            setName(e.currentTarget.value);
          }}
          required
          className="mb-4"
        />

        <Textarea
          label="説明（任意）"
          placeholder="ワークスペースの説明を入力"
          value={description}
          onChange={(e) => {
            setDescription(e.currentTarget.value);
          }}
          className="mb-4"
        />

        {createWorkspace.isError && (
          <Text c="red" size="sm" className="mb-4">
            {createWorkspace.error?.message || "作成に失敗しました"}
          </Text>
        )}

        <Button type="submit" fullWidth loading={createWorkspace.isPending}>
          作成
        </Button>
      </form>
    </Modal>
  );
};
