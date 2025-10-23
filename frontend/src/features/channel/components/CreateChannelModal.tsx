import { useState } from "react";
import type { FormEvent } from "react";

import { Button, Modal, Switch, Text, TextInput, Textarea } from "@mantine/core";

import { useCreateChannel } from "../hooks/useChannel";

interface CreateChannelModalProps {
  workspaceId: string | null;
  opened: boolean;
  onClose: () => void;
}

export const CreateChannelModal = ({ workspaceId, opened, onClose }: CreateChannelModalProps) => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isPrivate, setIsPrivate] = useState(false);
  const createChannel = useCreateChannel(workspaceId);

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    createChannel.mutate(
      { name, description: description || undefined, isPrivate },
      {
        onSuccess: () => {
          setName("");
          setDescription("");
          setIsPrivate(false);
          onClose();
        },
      }
    );
  };

  const isDisabled = workspaceId === null || createChannel.isPending;

  return (
    <Modal opened={opened} onClose={onClose} title="新規チャンネル作成">
      <form onSubmit={handleSubmit}>
        <TextInput
          label="チャンネル名"
          placeholder="例: general"
          value={name}
          onChange={(event) => setName(event.currentTarget.value)}
          required
          className="mb-4"
          disabled={workspaceId === null}
        />

        <Textarea
          label="説明（任意）"
          placeholder="チャンネルの目的を記載"
          value={description}
          onChange={(event) => setDescription(event.currentTarget.value)}
          className="mb-4"
          disabled={workspaceId === null}
        />

        <Switch
          label="プライベートチャンネルにする"
          checked={isPrivate}
          onChange={(event) => setIsPrivate(event.currentTarget.checked)}
          className="mb-4"
          disabled={workspaceId === null}
        />

        {workspaceId === null && (
          <Text c="dimmed" size="sm" className="mb-2">
            先にワークスペースを選択してください
          </Text>
        )}

        {createChannel.isError && (
          <Text c="red" size="sm" className="mb-2">
            {createChannel.error?.message ?? "チャンネルの作成に失敗しました"}
          </Text>
        )}

        <Button type="submit" fullWidth disabled={isDisabled} loading={createChannel.isPending}>
          作成
        </Button>
      </form>
    </Modal>
  );
};
