import { MantineProvider } from '@mantine/core';
import { render } from '@testing-library/react';
import { describe, it, expect } from 'vitest';

import { renderMarkdown } from "@/features/message/utils/markdown/renderer";


const renderWithMantine = (element: React.ReactElement) => {
  return render(<MantineProvider>{element}</MantineProvider>);
};

describe('Mention rendering', () => {
  it('メンションをレンダリングできる', () => {
    const content = 'Hello @john, how are you?';
    const result = renderMarkdown(content);
    const { container } = renderWithMantine(<>{result}</>);

    const mention = container.querySelector('.mention');
    expect(mention).toBeDefined();
    expect(mention).toHaveAttribute('data-mention', 'john');
    expect(mention).toHaveTextContent('@john');
  });

  it('複数のメンションをレンダリングできる', () => {
    const content = '@alice and @bob are here.';
    const result = renderMarkdown(content);
    const { container } = renderWithMantine(<>{result}</>);

    const mentions = container.querySelectorAll('.mention');
    expect(mentions).toHaveLength(2);
    expect(mentions[0]).toHaveAttribute('data-mention', 'alice');
    expect(mentions[1]).toHaveAttribute('data-mention', 'bob');
  });
});
