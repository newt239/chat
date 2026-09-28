import { timestampDate } from "@bufbuild/protobuf/wkt";

import type { Timestamp } from "@bufbuild/protobuf/wkt";

/** 未設定の日時は UNIX エポックとして扱う */
export const toDate = (timestamp: Timestamp | undefined) =>
  timestamp === undefined ? new Date(0) : timestampDate(timestamp);
