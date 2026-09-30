import type { Literal } from "mdast";

// remark プラグインで追加するカスタムノードを mdast に登録する
declare module "mdast" {
  interface MentionNode extends Literal {
    type: "mention";
  }

  interface ChannelLinkNode extends Literal {
    type: "channelLink";
  }

  interface CustomEmojiNode extends Literal {
    type: "customEmoji";
  }

  interface RootContentMap {
    channelLink: ChannelLinkNode;
    customEmoji: CustomEmojiNode;
    mention: MentionNode;
  }

  // mdast-util-to-hast が参照する変換ヒント
  interface Data {
    hName?: string;
    hProperties?: Record<string, unknown>;
  }
}
