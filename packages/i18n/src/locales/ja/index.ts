// 名前空間ごとにファイルを分け、ここで束ねる。キーは「機能.文脈.項目」の camelCase。日本語辞書が正
import { channel } from "./channel";
import { codeBlock } from "./codeBlock";
import { common } from "./common";
import { dm } from "./dm";
import { member } from "./member";
import { preferences } from "./preferences";
import { search } from "./search";
import { ui } from "./ui";
import { userGroup } from "./userGroup";

export const ja = {
  channel,
  codeBlock,
  common,
  dm,
  member,
  preferences,
  search,
  ui,
  userGroup,
} as const;
