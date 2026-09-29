import type { MessageLocation } from "#/gen/chat/v1/message_pb";

// 端末の地図アプリ（モバイルでは Google マップのアプリ）で開ける URL
export const externalMapUrl = ({ latitude, longitude }: MessageLocation) =>
  `https://www.google.com/maps/search/?api=1&query=${latitude},${longitude}`;

export const formatCoordinates = ({ latitude, longitude }: MessageLocation) =>
  `${latitude.toFixed(5)}, ${longitude.toFixed(5)}`;
