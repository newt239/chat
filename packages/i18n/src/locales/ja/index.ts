// 名前空間ごとにファイルを分け、ここで束ねる。キーは「機能.文脈.項目」の camelCase。日本語辞書が正
import { auth } from "./auth";
import { codeBlock } from "./codeBlock";
import { common } from "./common";
import { preferences } from "./preferences";
import { settings } from "./settings";
import { shell } from "./shell";
import { ui } from "./ui";
import { workspace } from "./workspace";

export const ja = {
  auth,
  codeBlock,
  common,
  preferences,
  settings,
  shell,
  ui,
  workspace,
} as const;
