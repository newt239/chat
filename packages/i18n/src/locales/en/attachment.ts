import type { Messages } from "../../messages";

export const attachment: Messages["attachment"] = {
  completed: "Done",
  crop: {
    tall: "Tall · View full",
    wide: "Wide · View full",
  },
  download: "Download",
  downloadFailed: "Couldn't download the file",
  errors: {
    aborted: "Upload canceled",
    empty: "The file is empty",
    http: "Upload failed (HTTP {{status}})",
    network: "A network error occurred",
    tooLarge: "The file exceeds the 1 GB limit: {{size}}",
    unknown: "Upload failed",
  },
  expand: "Enlarge {{name}}",
  failed: "Error: {{error}}",
  lightbox: {
    label: "Image viewer",
    next: "Next image",
    page: "Image {{index}}",
    position: "{{index}} / {{total}}",
    previous: "Previous image",
    tallHint: "Scroll to see the whole image",
  },
  loadFailed: "Couldn't load",
  player: {
    close: "Stop playback",
    jump: "Go to the message",
    label: "Now playing",
    pause: "Pause",
    play: "Play",
    playFailed: "Couldn't play the media",
    playFile: "Play {{name}}",
    seek: "Playback position",
    speed: "Playback speed",
  },
  remove: "Remove attachment",
};
