import type { ReactNode } from "react";

type LinkComponentProps = {
  href?: string;
  children?: ReactNode;
};

// 本文中のリンク。見た目は markdownClassName で付ける
export const LinkComponent = ({ href, children }: LinkComponentProps) => {
  const isExternal = href?.startsWith("http://") || href?.startsWith("https://");

  return (
    <a
      href={href}
      target={isExternal ? "_blank" : undefined}
      rel={isExternal ? "noopener noreferrer" : undefined}
    >
      {children}
    </a>
  );
};
