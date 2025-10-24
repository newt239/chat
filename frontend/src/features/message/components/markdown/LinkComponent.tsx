import type { ReactNode } from 'react';

import { Anchor } from '@mantine/core';

interface LinkComponentProps {
  href?: string;
  children?: ReactNode;
}

export const LinkComponent = ({ href, children }: LinkComponentProps) => {
  // 外部リンクかどうかを判定
  const isExternal = href?.startsWith('http://') || href?.startsWith('https://');

  return (
    <Anchor
      href={href}
      target={isExternal ? '_blank' : undefined}
      rel={isExternal ? 'noopener noreferrer' : undefined}
    >
      {children}
    </Anchor>
  );
}
