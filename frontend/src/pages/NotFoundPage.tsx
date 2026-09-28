import { Button, Stack, Text } from "@mantine/core";
import { Link } from "@tanstack/react-router";

export const NotFoundPage = () => (
  <Stack align="center" justify="center" className="h-screen" gap="md">
    <Text size="xl" fw={700}>
      404
    </Text>
    <Text size="sm" c="dimmed">
      お探しのページは見つかりませんでした。
    </Text>
    <Button renderRoot={(props) => <Link {...props} to="/app" />}>トップへ戻る</Button>
  </Stack>
);
