import { Code, Modal, Stack, Table, Text } from "@mantine/core";

type MarkdownHelpModalProps = {
  opened: boolean;
  onClose: () => void;
};

const SYNTAX_ROWS = [
  { example: "**太字**", label: "太字" },
  { example: "*斜体*", label: "斜体" },
  { example: "~~打ち消し~~", label: "打ち消し線" },
  { example: "# 見出し", label: "見出し" },
  { example: "[リンク](https://example.com)", label: "リンク" },
  { example: "`コード`", label: "インラインコード" },
  { example: "> 引用", label: "引用" },
  { example: "- 項目", label: "箇条書き" },
  { example: "1. 項目", label: "番号付きリスト" },
  { example: "@ユーザー名", label: "メンション" },
  { example: "#チャンネル名", label: "チャンネルリンク" },
];

export const MarkdownHelpModal = ({ opened, onClose }: MarkdownHelpModalProps) => (
  <Modal opened={opened} onClose={onClose} title="書式のヘルプ" centered>
    <Stack gap="sm">
      <Text size="sm" c="dimmed">
        メッセージでは Markdown 記法が使えます。Enter で送信、Shift + Enter で改行します。
      </Text>
      <Table striped withTableBorder>
        <Table.Tbody>
          {SYNTAX_ROWS.map((row) => (
            <Table.Tr key={row.label}>
              <Table.Td>{row.label}</Table.Td>
              <Table.Td>
                <Code>{row.example}</Code>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </Stack>
  </Modal>
);
