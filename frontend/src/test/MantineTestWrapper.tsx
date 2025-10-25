import type { ReactNode } from "react";

import { MantineProvider } from "@mantine/core";

type MantineTestWrapperProps = {
  children: ReactNode;
};

export const MantineTestWrapper = ({ children }: MantineTestWrapperProps) => {
  return <MantineProvider>{children}</MantineProvider>;
};
