// 名前空間ごとにファイルを分け、ここで束ねる。キーは「機能.文脈.項目」の camelCase。日本語辞書が正
import { attachment } from "./attachment";
import { bookmark } from "./bookmark";
import { codeBlock } from "./codeBlock";
import { common } from "./common";
import { link } from "./link";
import { message } from "./message";
import { notification } from "./notification";
import { pin } from "./pin";
import { preferences } from "./preferences";
import { reaction } from "./reaction";
import { ui } from "./ui";

export const ja = {
  attachment,
  bookmark,
  codeBlock,
  common,
  link,
  message,
  notification,
  pin,
  preferences,
  reaction,
  ui,
} as const;
