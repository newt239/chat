import { useSyncExternalStore } from "react";

import { mediaPlayer } from "../mediaPlayer";

import type { PlayerState } from "../mediaPlayer";

/** プレイヤーの状態のうち selector で選んだ部分だけを購読する。選んだ値が変わらなければ描き直さない */
export const usePlayerState = <T>(selector: (state: PlayerState) => T) => {
  const getSnapshot = () => selector(mediaPlayer.getState());
  return useSyncExternalStore(mediaPlayer.subscribe, getSnapshot, getSnapshot);
};
