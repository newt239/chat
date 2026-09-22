import { useState } from "react";

import { Modal, Button, MultiSelect, Text, Group, TextInput } from "@mantine/core";
import { useNavigate } from "react-router";

import { useMembers } from "#/features/member/hooks/useMembers";
import { paths } from "#/lib/paths";

import { useCreateDM, useCreateGroupDM } from "../hooks/useDM";

type CreateDMModalProps = {
  workspaceId: string;
  opened: boolean;
  onClose: () => void;
};

export const CreateDMModal = ({ workspaceId, opened, onClose }: CreateDMModalProps) => {
  const navigate = useNavigate();
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const [groupName, setGroupName] = useState("");

  const { data: members } = useMembers(opened ? workspaceId : null);
  const createDM = useCreateDM(workspaceId);
  const createGroupDM = useCreateGroupDM(workspaceId);

  const isGroup = selectedUserIds.length > 1;
  const isPending = createDM.isPending || createGroupDM.isPending;

  const handleClose = () => {
    setSelectedUserIds([]);
    setGroupName("");
    onClose();
  };

  const handleSubmit = async () => {
    const [firstUserId] = selectedUserIds;
    if (firstUserId === undefined) {
      return;
    }

    const dm = isGroup
      ? await createGroupDM.mutateAsync({
          name: groupName.trim() === "" ? undefined : groupName.trim(),
          userIds: selectedUserIds,
        })
      : await createDM.mutateAsync({ userId: firstUserId });

    handleClose();
    void navigate(paths.channel(workspaceId, dm.id));
  };

  const memberOptions = (members ?? []).map((member) => ({
    label: member.displayName,
    value: member.userId,
  }));

  return (
    <Modal opened={opened} onClose={handleClose} title="DM を作成">
      <div className="space-y-4">
        <div>
          <Text size="sm" fw={500} mb={4}>
            ユーザーを選択
          </Text>
          <MultiSelect
            placeholder="ユーザーを選択してください（複数選択でグループ DM）"
            data={memberOptions}
            value={selectedUserIds}
            onChange={setSelectedUserIds}
            searchable
            maxValues={8}
          />
        </div>

        {isGroup && (
          <TextInput
            label="グループ名（任意）"
            placeholder="未入力の場合はメンバー名から自動生成されます"
            value={groupName}
            onChange={(event) => {
              setGroupName(event.currentTarget.value);
            }}
          />
        )}

        {(createDM.isError || createGroupDM.isError) && (
          <Text c="red" size="sm">
            {(createDM.error ?? createGroupDM.error)?.message}
          </Text>
        )}

        <Group justify="flex-end">
          <Button variant="subtle" onClick={handleClose}>
            キャンセル
          </Button>
          <Button
            onClick={() => {
              void handleSubmit();
            }}
            disabled={selectedUserIds.length === 0}
            loading={isPending}
          >
            作成
          </Button>
        </Group>
      </div>
    </Modal>
  );
};
