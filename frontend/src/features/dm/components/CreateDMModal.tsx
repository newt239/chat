import { useState } from "react";

import { Modal, Button, Select, Text, Group } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router";

import { api } from "#/lib/api/client";
import { paths } from "#/lib/paths";

import { useCreateDM } from "../hooks/useDM";

type CreateDMModalProps = {
  workspaceId: string;
  opened: boolean;
  onClose: () => void;
};

export const CreateDMModal = ({ workspaceId, opened, onClose }: CreateDMModalProps) => {
  const navigate = useNavigate();
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);

  const { data: members } = useQuery({
    enabled: Boolean(workspaceId) && opened,
    queryFn: async () => {
      const response = await api.GET("/api/workspaces/{id}/members", {
        params: { path: { id: workspaceId } },
      });
      if (response.error) {
        throw new Error("ワークスペースメンバーの取得に失敗しました");
      }
      return response.data;
    },
    queryKey: ["workspace-members", workspaceId],
  });

  const createDMMutation = useCreateDM(workspaceId);

  const handleSubmit = async () => {
    if (!selectedUserId) {
      return;
    }

    try {
      const dm = await createDMMutation.mutateAsync({ userId: selectedUserId });
      onClose();
      setSelectedUserId(null);

      void navigate(paths.channel(workspaceId, dm.id));
    } catch (error) {
      console.error("DMの作成に失敗しました:", error);
    }
  };

  const handleClose = () => {
    setSelectedUserId(null);
    onClose();
  };

  const memberOptions =
    members?.members.map((member: { userId: string; displayName: string }) => ({
      label: member.displayName,
      value: member.userId,
    })) ?? [];

  return (
    <Modal opened={opened} onClose={handleClose} title="DM を作成">
      <div className="space-y-4">
        <div>
          <Text size="sm" fw={500} mb={4}>
            ユーザーを選択
          </Text>
          <Select
            placeholder="ユーザーを選択してください"
            data={memberOptions}
            value={selectedUserId}
            onChange={setSelectedUserId}
            searchable
            required
          />
        </div>

        <Group justify="flex-end">
          <Button variant="subtle" onClick={handleClose}>
            キャンセル
          </Button>
          <Button
            onClick={() => {
              void handleSubmit();
            }}
            disabled={!selectedUserId}
            loading={createDMMutation.isPending}
          >
            作成
          </Button>
        </Group>
      </div>
    </Modal>
  );
};
