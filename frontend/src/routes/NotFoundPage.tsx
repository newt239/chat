import { Button, Stack, Text } from "@mantine/core";
import { Link } from "react-router";

import { paths } from "#/lib/paths";

export const NotFoundPage = () => (
  <Stack align="center" justify="center" className="h-screen" gap="md">
    <Text size="xl" fw={700}>
      404
    </Text>
    <Text size="sm" c="dimmed">
      お探しのページは見つかりませんでした。
    </Text>
    <Button component={Link} to={paths.app()}>
      トップへ戻る
    </Button>
  </Stack>
);
