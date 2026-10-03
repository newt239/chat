import { useRef } from "react";

import { decodeMentions, encodeMentions } from "../utils/mentionCodec";
import { useMentionDirectory } from "./useMentionDirectory";

import type { MentionLabels } from "../utils/mentionCodec";

/** 入力欄では名前で見せ、送信・下書きの保存では ID 記法にする */
export const useMentionCodec = () => {
  const { isReady, labelOf } = useMentionDirectory();
  // 選んだ候補と読み込んだ本文から覚えた、名前と ID 記法の対応
  const labelsRef = useRef<MentionLabels>(new Map());

  return {
    decode: (body: string) => decodeMentions(body, labelsRef.current, labelOf),
    encode: (text: string) => encodeMentions(text, labelsRef.current),
    isReady,
    register: (label: string, token: string) => {
      labelsRef.current.set(label, token);
    },
  };
};
