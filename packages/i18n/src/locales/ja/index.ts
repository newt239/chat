// 名前空間ごとにファイルを分け、ここで束ねる。キーは「機能.文脈.項目」の camelCase。日本語辞書が正
import { codeBlock } from "./codeBlock";
import { common } from "./common";
import { preferences } from "./preferences";
import { ui } from "./ui";

export const ja = {
  codeBlock,
  common,
  preferences,
  ui,
} as const;
