import {
  IconFile,
  IconFileSpreadsheet,
  IconFileText,
  IconFileTypePdf,
  IconMusic,
  IconPhoto,
  IconPresentation,
  IconVideo,
} from "@tabler/icons-react";

type FileIconProps = {
  mimeType: string;
};

const iconOf = (mimeType: string) => {
  if (mimeType.startsWith("image/")) {
    return <IconPhoto aria-hidden />;
  }
  if (mimeType.startsWith("video/")) {
    return <IconVideo aria-hidden />;
  }
  if (mimeType.startsWith("audio/")) {
    return <IconMusic aria-hidden />;
  }
  if (mimeType.includes("pdf")) {
    return <IconFileTypePdf aria-hidden />;
  }
  if (mimeType.includes("spreadsheet") || mimeType.includes("excel")) {
    return <IconFileSpreadsheet aria-hidden />;
  }
  if (mimeType.includes("presentation") || mimeType.includes("powerpoint")) {
    return <IconPresentation aria-hidden />;
  }
  if (mimeType.includes("document") || mimeType.includes("word") || mimeType.includes("text")) {
    return <IconFileText aria-hidden />;
  }
  return <IconFile aria-hidden />;
};

export const FileIcon = ({ mimeType }: FileIconProps) => (
  <span className="grid size-8.5 shrink-0 place-items-center rounded-md bg-accent-soft text-accent-text [&_svg]:size-4.5">
    {iconOf(mimeType)}
  </span>
);
