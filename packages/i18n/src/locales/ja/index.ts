// 名前空間ごとにファイルを分け、ここで束ねる。キーは「機能.文脈.項目」の camelCase。日本語辞書が正
import { attachment } from "./attachment";
import { bookmark } from "./bookmark";
import { channel } from "./channel";
import { codeBlock } from "./codeBlock";
import { common } from "./common";
import { dm } from "./dm";
import { link } from "./link";
import { member } from "./member";
import { message } from "./message";
import { notification } from "./notification";
import { pin } from "./pin";
import { preferences } from "./preferences";
import { reaction } from "./reaction";
import { search } from "./search";
import { ui } from "./ui";
import { userGroup } from "./userGroup";

export const ja = {
  attachment,
  bookmark,
  channel,
  codeBlock,
  common,
  dm,
  link,
  member,
  message,
  notification,
  pin,
  preferences,
  reaction,
  search,
  ui,
  userGroup,
} as const;
