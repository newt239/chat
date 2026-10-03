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
    empty: "The file is empty",
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
  remove: "Remove attachment",
};
