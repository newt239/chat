import type { ComponentProps } from "react";

import { createLink } from "@tanstack/react-router";

import { MenuItem } from "#/components/ui/MenuItem/MenuItem";

// 遷移先を to / params で指定するメニュー項目。target="_blank" で新しいタブに開く
export const MenuItemLink = createLink((props: ComponentProps<typeof MenuItem>) => (
  <MenuItem {...props} />
));
