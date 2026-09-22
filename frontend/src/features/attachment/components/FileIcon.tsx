type FileIconProps = {
  mimeType: string;
};

export const FileIcon = ({ mimeType }: FileIconProps) => {
  const getIconColor = () => {
    if (mimeType.startsWith("image/")) {
      return "text-purple-500";
    }
    if (mimeType.startsWith("video/")) {
      return "text-red-500";
    }
    if (mimeType.startsWith("audio/")) {
      return "text-green-500";
    }
    if (mimeType.includes("pdf")) {
      return "text-red-600";
    }
    if (mimeType.includes("document") || mimeType.includes("word") || mimeType.includes("text")) {
      return "text-blue-500";
    }
    if (mimeType.includes("spreadsheet") || mimeType.includes("excel")) {
      return "text-green-600";
    }
    if (mimeType.includes("presentation") || mimeType.includes("powerpoint")) {
      return "text-orange-500";
    }
    return "text-gray-500";
  };

  return (
    <svg className={`w-8 h-8 ${getIconColor()}`} fill="currentColor" viewBox="0 0 24 24">
      <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8l-6-6z" />
      <path d="M14 2v6h6" />
    </svg>
  );
};
