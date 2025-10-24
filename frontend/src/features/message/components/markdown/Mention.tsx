import type { ReactNode } from 'react';

import { Badge } from '@mantine/core';

type MentionProps = {
  'data-mention': string;
  children?: ReactNode;
}

export const Mention = ({ 'data-mention': username }: MentionProps) => {
  return (
    <Badge
      variant="light"
      color="blue"
      size="sm"
      className="cursor-pointer hover:bg-blue-100"
      component="span"
    >
      @{username}
    </Badge>
  );
}
