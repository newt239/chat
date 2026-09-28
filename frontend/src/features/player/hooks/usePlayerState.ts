import { useSyncExternalStore } from "react";

import { mediaPlayer } from "../mediaPlayer";

export const usePlayerState = () =>
  useSyncExternalStore(mediaPlayer.subscribe, mediaPlayer.getState, mediaPlayer.getState);
