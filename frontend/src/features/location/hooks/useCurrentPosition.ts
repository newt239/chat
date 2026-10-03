import { useEffect, useState } from "react";

import { create } from "@bufbuild/protobuf";

import { MessageLocationSchema } from "#/gen/chat/v1/message_pb";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

type PositionState =
  | { status: "locating" }
  | { status: "failed"; reason: "denied" | "unavailable" }
  | { status: "ready"; location: MessageLocation };

// マウント時に Geolocation API で現在地を 1 回取得する。locate で取り直せる
export const useCurrentPosition = () => {
  const [state, setState] = useState<PositionState>({ status: "locating" });

  const locate = () => {
    if (!("geolocation" in navigator)) {
      setState({ reason: "unavailable", status: "failed" });
      return;
    }
    setState({ status: "locating" });
    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        setState({
          location: create(MessageLocationSchema, {
            accuracyMeters: coords.accuracy,
            latitude: coords.latitude,
            longitude: coords.longitude,
          }),
          status: "ready",
        });
      },
      (error) => {
        setState({
          reason: error.code === error.PERMISSION_DENIED ? "denied" : "unavailable",
          status: "failed",
        });
      },
      { enableHighAccuracy: true, maximumAge: 30_000, timeout: 15_000 },
    );
  };

  useEffect(locate, [locate]);

  return { locate, state };
};
