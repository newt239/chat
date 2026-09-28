// 名前空間ごとにファイルを分け、ここで束ねる。キーは「機能.文脈.項目」の camelCase。日本語辞書が正
import { admin } from "./admin";
import { attachment } from "./attachment";
import { auth } from "./auth";
import { bookmark } from "./bookmark";
import { channel } from "./channel";
import { codeBlock } from "./codeBlock";
import { common } from "./common";
import { dm } from "./dm";
import { insights } from "./insights";
import { link } from "./link";
import { member } from "./member";
import { message } from "./message";
import { notification } from "./notification";
import { pin } from "./pin";
import { preferences } from "./preferences";
import { reaction } from "./reaction";
import { search } from "./search";
import { settings } from "./settings";
import { shell } from "./shell";
import { ui } from "./ui";
import { userGroup } from "./userGroup";
import { workspace } from "./workspace";

export const ja = {
  admin,
  attachment,
  auth,
  bookmark,
  channel,
  codeBlock,
  common,
  dm,
  insights,
  link,
  member,
  message,
  notification,
  pin,
  preferences,
  reaction,
  search,
  settings,
  shell,
  ui,
  userGroup,
  workspace,
} as const;
