import { Button, Card, Group, Loader, Stack, Text } from "@mantine/core";

import {
  useJoinPublicWorkspace,
  usePublicWorkspaces,
} from "#/features/workspace/hooks/usePublicWorkspaces";

export const PublicWorkspaceList = () => {
  const { data: workspaces, isLoading } = usePublicWorkspaces();
  const join = useJoinPublicWorkspace();

  if (isLoading) {
    return (
      <div className="flex justify-center py-4">
        <Loader size="sm" />
      </div>
    );
  }

  const joinable = workspaces?.filter((workspace) => !workspace.isJoined) ?? [];

  if (joinable.length === 0) {
    return null;
  }

  return (
    <Stack gap="md">
      <Text size="lg" fw={500}>
        参加できる公開ワークスペース
      </Text>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {joinable.map((workspace) => (
          <Card key={workspace.id} shadow="sm" padding="lg" radius="md" withBorder>
            <Text fw={500} size="lg" className="mb-2">
              {workspace.name}
            </Text>
            {workspace.description && (
              <Text size="sm" c="dimmed" className="mb-4">
                {workspace.description}
              </Text>
            )}
            <Group justify="space-between" align="center">
              <Text size="xs" c="dimmed">
                {workspace.memberCount}人
              </Text>
              <Button
                size="xs"
                variant="light"
                loading={join.isPending && join.variables.workspaceId === workspace.id}
                onClick={() => {
                  join.mutate({ workspaceId: workspace.id });
                }}
              >
                参加する
              </Button>
            </Group>
          </Card>
        ))}
      </div>
    </Stack>
  );
};
