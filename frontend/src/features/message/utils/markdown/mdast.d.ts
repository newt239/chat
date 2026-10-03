import type { Literal } from "mdast";

// renderMarkdown が処理ごとに remarkHideLinks へ渡す値
declare module "vfile" {
  interface DataMap {
    hiddenUrls: readonly string[];
  }
}

// remark プラグインで追加するカスタムノードを mdast に登録する
declare module "mdast" {
  interface MentionNode extends Literal {
    type: "mention";
  }

  interface CustomEmojiNode extends Literal {
    type: "customEmoji";
  }

  interface RootContentMap {
    customEmoji: CustomEmojiNode;
    mention: MentionNode;
  }

  // mdast-util-to-hast が参照する変換ヒント
  interface Data {
    hName?: string;
    hProperties?: Record<string, unknown>;
  }
}
