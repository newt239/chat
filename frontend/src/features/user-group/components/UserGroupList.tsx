import { useState, useEffect } from "react";

import { Button, Card, Group, Text, Badge, ActionIcon, Menu, Loader } from "@mantine/core";
import { IconPlus, IconDots, IconEdit, IconTrash, IconUsers } from "@tabler/icons-react";

import { useUserGroups } from "../hooks/useUserGroups";

import { CreateUserGroupModal } from "./CreateUserGroupModal";

interface UserGroupListProps {
  workspaceId: string;
}

export const UserGroupList = ({ workspaceId }: UserGroupListProps) => {
  const { groups, isLoading, error, fetchGroups, deleteGroup } = useUserGroups(workspaceId);
  const [createModalOpen, setCreateModalOpen] = useState(false);

  useEffect(() => {
    fetchGroups();
  }, [fetchGroups]);

  const handleDeleteGroup = async (groupId: string) => {
    if (window.confirm("Are you sure you want to delete this group?")) {
      try {
        await deleteGroup(groupId);
      } catch (error) {
        console.error("Failed to delete group:", error);
      }
    }
  };

  if (isLoading) {
    return (
      <div className="flex justify-center p-4">
        <Loader size="sm" />
      </div>
    );
  }

  if (error) {
    return <div className="p-4 text-red-600">Error: {error}</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <Text size="lg" fw={600}>
          User Groups
        </Text>
        <Button leftSection={<IconPlus size={16} />} onClick={() => setCreateModalOpen(true)}>
          Create Group
        </Button>
      </div>

      {groups.length === 0 ? (
        <Card withBorder p="xl" className="text-center">
          <Text c="dimmed">No groups found. Create your first group to get started.</Text>
        </Card>
      ) : (
        <div className="grid gap-4">
          {groups.map((group) => (
            <Card key={group.id} withBorder p="md">
              <div className="flex justify-between items-start">
                <div className="flex-1">
                  <Group gap="xs" mb="xs">
                    <Text fw={600}>{group.name}</Text>
                    <Badge size="sm" variant="light">
                      @{group.name}
                    </Badge>
                  </Group>

                  {group.description && (
                    <Text size="sm" c="dimmed" mb="sm">
                      {group.description}
                    </Text>
                  )}

                  <Group gap="xs">
                    <ActionIcon variant="subtle" size="sm">
                      <IconUsers size={16} />
                    </ActionIcon>
                    <Text size="xs" c="dimmed">
                      Created by you
                    </Text>
                  </Group>
                </div>

                <Menu shadow="md" width={200}>
                  <Menu.Target>
                    <ActionIcon variant="subtle">
                      <IconDots size={16} />
                    </ActionIcon>
                  </Menu.Target>

                  <Menu.Dropdown>
                    <Menu.Item leftSection={<IconEdit size={14} />}>Edit Group</Menu.Item>
                    <Menu.Item leftSection={<IconUsers size={14} />}>Manage Members</Menu.Item>
                    <Menu.Divider />
                    <Menu.Item
                      leftSection={<IconTrash size={14} />}
                      color="red"
                      onClick={() => handleDeleteGroup(group.id)}
                    >
                      Delete Group
                    </Menu.Item>
                  </Menu.Dropdown>
                </Menu>
              </div>
            </Card>
          ))}
        </div>
      )}

      <CreateUserGroupModal
        opened={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        workspaceId={workspaceId}
        onSuccess={() => {
          setCreateModalOpen(false);
          fetchGroups();
        }}
      />
    </div>
  );
};
