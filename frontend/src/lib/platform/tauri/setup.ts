import { platform } from "@tauri-apps/plugin-os";

import { interceptExternalLinks } from "./externalLinks";

export const setupTauri = () => {
  interceptExternalLinks();
  return { isMobileApp: platform() === "android" || platform() === "ios" };
};
