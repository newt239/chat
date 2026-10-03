import type { MessageLocation } from "#/gen/chat/v1/message_pb";

import type { TFunction } from "i18next";

// 端末の地図アプリ（モバイルでは Google マップのアプリ）で開ける URL
export const externalMapUrl = ({ latitude, longitude }: MessageLocation) =>
  `https://www.google.com/maps/search/?api=1&query=${latitude},${longitude}`;

export const formatCoordinates = (
  { latitude, longitude, accuracyMeters }: MessageLocation,
  t: TFunction,
) => {
  const coordinates = `${latitude.toFixed(5)}, ${longitude.toFixed(5)}`;
  return accuracyMeters === undefined
    ? coordinates
    : `${coordinates} · ${t("location.accuracy", { meters: Math.round(accuracyMeters) })}`;
};
