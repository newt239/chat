import { useState } from "react";

import {
  ActionIcon,
  Avatar,
  Box,
  Button,
  Group,
  Modal,
  Select,
  Stack,
  Text,
  Tooltip,
} from "@mantine/core";
import { IconCrown, IconTrash, IconUserPlus } from "@tabler/icons-react";

import {
  useChannelMembers,
  useInviteChannelMember,
  useRemoveChannelMember,
  useUpdateChannelMemberRole,
} from "@/features/channel/hooks/useChannelMembers";
import { useMembers } from "@/features/workspace/hooks/useMembers";


type ChannelMembersModalProps = {
  channelId: string | null;
  workspaceId: string | null;
  opened: boolean;
  onClose: () => void;
};

export const ChannelMembersModal = ({
  channelId,
  workspaceId,
  opened,
  onClose,
}: ChannelMembersModalProps) => {
  const { data: channelMembers } = useChannelMembers(channelId);
  const { data: workspaceMembers } = useMembers(workspaceId);
  const inviteMember = useInviteChannelMember(channelId ?? "");
  const removeMember = useRemoveChannelMember(channelId ?? "");
  const updateRole = useUpdateChannelMemberRole(channelId ?? "");

  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const [selectedRole, setSelectedRole] = useState<"member" | "admin">("member");

  const availableMembers =
    workspaceMembers?.filter(
      (wm) => !channelMembers?.some((cm) => cm.userId === wm.userId)
    ) ?? [];

  const handleInvite = () => {
    if (!selectedUserId) return;

    inviteMember.mutate(
      { userId: selectedUserId, role: selectedRole },
      {
        onSuccess: () => {
          setSelectedUserId(null);
          setSelectedRole("member");
        },
      }
    );
  };

  const handleRoleToggle = (userId: string, currentRole: "member" | "admin") => {
    const newRole = currentRole === "admin" ? "member" : "admin";
    updateRole.mutate({ userId, role: newRole });
  };

  const handleRemove = (userId: string) => {
    removeMember.mutate({ userId });
  };

  const adminCount = channelMembers?.filter((m) => m.role === "admin").length ?? 0;

  return (
    <Modal opened={opened} onClose={onClose} title="チャンネルメンバー管理" size="lg">
      <Stack gap="md">
        <Box>
          <Text fw={600} size="sm" mb="xs">
            メンバーを招待
          </Text>
          <Group gap="xs">
            <Select
              placeholder="メンバーを選択"
              data={availableMembers.map((m) => ({
                value: m.userId,
                label: m.displayName,
              }))}
              value={selectedUserId}
              onChange={(value) => setSelectedUserId(value)}
              flex={1}
              searchable
              nothingFoundMessage="招待可能なメンバーがいません"
            />
            <Select
              data={[
                { value: "member", label: "メンバー" },
                { value: "admin", label: "管理者" },
              ]}
              value={selectedRole}
              onChange={(value) => setSelectedRole(value as "member" | "admin")}
              w={120}
            />
            <Button
              onClick={handleInvite}
              disabled={!selectedUserId || inviteMember.isPending}
              loading={inviteMember.isPending}
              leftSection={<IconUserPlus size={16} />}
            >
              招待
            </Button>
          </Group>
          {inviteMember.isError && (
            <Text c="red" size="sm" mt="xs">
              {inviteMember.error?.message ?? "招待に失敗しました"}
            </Text>
          )}
        </Box>

        <Box>
          <Text fw={600} size="sm" mb="xs">
            現在のメンバー ({channelMembers?.length ?? 0})
          </Text>
          <Stack gap="xs">
            {channelMembers?.map((member) => (
              <Group key={member.userId} justify="space-between" p="xs" className="rounded">
                <Group gap="sm">
                  <Avatar src={member.avatarUrl} size="sm" />
                  <Box>
                    <Text size="sm" fw={500}>
                      {member.displayName}
                    </Text>
                    <Text size="xs" c="dimmed">
                      {member.email}
                    </Text>
                  </Box>
                </Group>
                <Group gap="xs">
                  <Tooltip label={member.role === "admin" ? "管理者権限を削除" : "管理者にする"}>
                    <ActionIcon
                      variant={member.role === "admin" ? "filled" : "subtle"}
                      color={member.role === "admin" ? "yellow" : "gray"}
                      onClick={() => handleRoleToggle(member.userId, member.role)}
                      disabled={
                        updateRole.isPending || (member.role === "admin" && adminCount <= 1)
                      }
                    >
                      <IconCrown size={16} />
                    </ActionIcon>
                  </Tooltip>
                  <Tooltip
                    label={
                      member.role === "admin" && adminCount <= 1
                        ? "最後の管理者は削除できません"
                        : "メンバーを削除"
                    }
                  >
                    <ActionIcon
                      variant="subtle"
                      color="red"
                      onClick={() => handleRemove(member.userId)}
                      disabled={
                        removeMember.isPending || (member.role === "admin" && adminCount <= 1)
                      }
                    >
                      <IconTrash size={16} />
                    </ActionIcon>
                  </Tooltip>
                </Group>
              </Group>
            ))}
          </Stack>
          {(removeMember.isError || updateRole.isError) && (
            <Text c="red" size="sm" mt="xs">
              {removeMember.error?.message ?? updateRole.error?.message ?? "操作に失敗しました"}
            </Text>
          )}
        </Box>
      </Stack>
    </Modal>
  );
};
