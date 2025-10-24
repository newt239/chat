import { Card, Image, Text, Anchor } from "@mantine/core";

import type { LinkInfo } from "../types";

type LinkPreviewEmbedProps = {
  link: LinkInfo;
}

export const LinkPreviewEmbed = ({ link }: LinkPreviewEmbedProps) => {
  const { url, title, description, imageUrl, siteName, cardType } = link;

  const isLargeImage = cardType === "summary_large_image";

  return (
    <Card withBorder radius="md" p="md" className="max-w-md my-2">
      {imageUrl && (
        <Image
          src={imageUrl}
          alt={title || "Preview image"}
          height={isLargeImage ? 200 : 120}
          radius="md"
          mb="sm"
          className="object-cover"
        />
      )}

      <div className="space-y-1">
        {title && (
          <Text fw={600} size="sm" lineClamp={2}>
            {title}
          </Text>
        )}

        {description && (
          <Text size="xs" c="dimmed" lineClamp={2}>
            {description}
          </Text>
        )}

        <div className="flex items-center space-x-1">
          {siteName && (
            <Text size="xs" c="dimmed">
              {siteName}
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
      </div>
    </Card>
  );
};
