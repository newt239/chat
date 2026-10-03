import { IconBell, IconKey, IconKeyboard, IconLanguage, IconPalette } from "@tabler/icons-react";

import type { SettingsSection } from "../schemas";

export const settingsSectionIcons: Record<SettingsSection, typeof IconKey> = {
  account: IconKey,
  display: IconLanguage,
  notifications: IconBell,
  shortcuts: IconKeyboard,
  theme: IconPalette,
};
