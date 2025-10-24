import { Card, Image, Skeleton, Text, Anchor } from "@mantine/core";

import type { LinkPreview } from "../types";

type LinkPreviewCardProps = {
  preview: LinkPreview;
  onRemove?: () => void;
}

export const LinkPreviewCard = ({ preview, onRemove }: LinkPreviewCardProps) => {
  const { url, ogpData, isLoading, error } = preview;

  if (isLoading) {
    return (
      <Card withBorder radius="md" p="md" className="max-w-md">
        <Skeleton height={120} radius="md" mb="sm" />
        <Skeleton height={16} width="80%" mb="xs" />
        <Skeleton height={14} width="60%" />
      </Card>
    );
  }

  if (error) {
    return (
      <Card withBorder radius="md" p="md" className="max-w-md bg-red-50">
        <Text size="sm" c="red">
          Failed to load preview
        </Text>
        <Anchor href={url} target="_blank" rel="noopener noreferrer" size="sm">
          {url}
        </Anchor>
        {onRemove && (
          <button onClick={onRemove} className="ml-2 text-red-500 hover:text-red-700" type="button">
            Remove
          </button>
        )}
      </Card>
    );
  }

  return (
    <Card withBorder radius="md" p="md" className="max-w-md hover:shadow-md transition-shadow">
      {ogpData.imageUrl && (
        <Image
          src={ogpData.imageUrl}
          alt={ogpData.title || "Preview image"}
          height={120}
          radius="md"
          mb="sm"
          className="object-cover"
        />
      )}

      <div className="space-y-1">
        {ogpData.title && (
          <Text fw={600} size="sm" lineClamp={2}>
            {ogpData.title}
          </Text>
        )}

        {ogpData.description && (
          <Text size="xs" c="dimmed" lineClamp={2}>
            {ogpData.description}
          </Text>
        )}

        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-1">
            {ogpData.siteName && (
              <Text size="xs" c="dimmed">
                {ogpData.siteName}
              </Text>
            )}
            <Text size="xs" c="dimmed">
              •
            </Text>
            <Anchor
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              size="xs"
              className="truncate max-w-32"
            >
              {new URL(url).hostname}
            </Anchor>
          </div>

          {onRemove && (
            <button onClick={onRemove} className="text-gray-400 hover:text-gray-600" type="button">
              ×
            </button>
          )}
        </div>
      </div>
    </Card>
  );
};
