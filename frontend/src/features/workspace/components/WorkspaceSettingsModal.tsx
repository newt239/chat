import { useState } from "react";

import {
  Button,
  Checkbox,
  Divider,
  Group,
  Modal,
  Stack,
  Text,
  TextInput,
  Textarea,
} from "@mantine/core";
import { useNavigate } from "@tanstack/react-router";

import { UserGroupManager } from "#/features/userGroup/components/UserGroupManager";
import { WorkspaceMemberManager } from "#/features/workspace/components/WorkspaceMemberManager";
import { useWorkspaceActions } from "#/features/workspace/hooks/useWorkspaceActions";

import type { WorkspaceSummary } from "#/features/workspace/types";

type WorkspaceSettingsModalProps = {
  opened: boolean;
  onClose: () => void;
  workspace: WorkspaceSummary;
};

export const WorkspaceSettingsModal = ({
  opened,
  onClose,
  workspace,
}: WorkspaceSettingsModalProps) => {
  const { update, remove } = useWorkspaceActions();
  const navigate = useNavigate();

  const [name, setName] = useState(workspace.name);
  const [description, setDescription] = useState(workspace.description ?? "");
  const [isPublic, setIsPublic] = useState(workspace.isPublic);
  const [isDeleteConfirming, setIsDeleteConfirming] = useState(false);

  const handleSave = () => {
    update.mutate({ description, isPublic, name, workspaceId: workspace.id });
  };

  const handleDelete = () => {
    remove.mutate(
      { workspaceId: workspace.id },
      {
        onSuccess: () => {
          onClose();
          void navigate({ to: "/app" });
        },
      },
    );
  };

  return (
    <Modal opened={opened} onClose={onClose} title="ワークスペース設定" centered size="lg">
      <Stack gap="md">
        <TextInput
          label="名前"
          value={name}
          onChange={(event) => {
            setName(event.currentTarget.value);
          }}
          required
        />
        <Textarea
          label="説明"
          value={description}
          onChange={(event) => {
            setDescription(event.currentTarget.value);
          }}
          autosize
          minRows={2}
        />
        <Checkbox
          label="公開ワークスペースにする（誰でも参加できます）"
          checked={isPublic}
          onChange={(event) => {
            setIsPublic(event.currentTarget.checked);
          }}
        />
        <Group justify="flex-end">
          <Button loading={update.isPending} onClick={handleSave}>
            保存
          </Button>
        </Group>
        {update.isError && (
          <Text c="red" size="sm">
            {update.error.message}
          </Text>
        )}

        <Divider />

        <WorkspaceMemberManager workspaceId={workspace.id} />

        <Divider />

        <UserGroupManager workspaceId={workspace.id} />

        <Divider />

        <Stack gap="xs">
          <Text fw={600} c="red">
            ワークスペースの削除
          </Text>
          <Text size="xs" c="dimmed">
            チャンネルとメッセージもすべて削除されます。この操作は取り消せません。
          </Text>
          {isDeleteConfirming ? (
            <Group justify="flex-end">
              <Button
                variant="subtle"
                onClick={() => {
                  setIsDeleteConfirming(false);
                }}
              >
                キャンセル
              </Button>
              <Button color="red" loading={remove.isPending} onClick={handleDelete}>
                削除する
              </Button>
            </Group>
          ) : (
            <Group justify="flex-end">
              <Button
                variant="outline"
                color="red"
                onClick={() => {
                  setIsDeleteConfirming(true);
                }}
              >
                削除
              </Button>
            </Group>
          )}
          {remove.isError && (
            <Text c="red" size="sm">
              {remove.error.message}
            </Text>
          )}
        </Stack>
      </Stack>
    </Modal>
  );
};
