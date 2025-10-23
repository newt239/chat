import { useState } from "react";

import { Modal, TextInput, Textarea, Button, Text } from "@mantine/core";

import { useCreateWorkspace } from "../hooks/useWorkspace";

interface CreateWorkspaceModalProps {
  opened: boolean;
  onClose: () => void;
}

export const CreateWorkspaceModal = ({ opened, onClose }: CreateWorkspaceModalProps) => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const createWorkspace = useCreateWorkspace();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createWorkspace.mutate(
      { name, description: description || undefined },
      {
        onSuccess: () => {
          setName("");
          setDescription("");
          onClose();
        },
      }
    );
  };

  return (
    <Modal opened={opened} onClose={onClose} title="新規ワークスペース作成">
      <form onSubmit={handleSubmit}>
        <TextInput
          label="ワークスペース名"
          placeholder="例: チーム開発"
          value={name}
          onChange={(e) => setName(e.currentTarget.value)}
          required
          className="mb-4"
        />

        <Textarea
          label="説明（任意）"
          placeholder="ワークスペースの説明を入力"
          value={description}
          onChange={(e) => setDescription(e.currentTarget.value)}
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
}
