import { MantineProvider } from '@mantine/core';
import { render } from '@testing-library/react';
import { describe, it, expect } from 'vitest';

import { renderMarkdown } from '../../utils/markdown/renderer';


const renderWithMantine = (element: React.ReactElement) => {
  return render(<MantineProvider>{element}</MantineProvider>);
};

describe('renderMarkdown', () => {
  it('基本的なMarkdownをレンダリングできる', () => {
    const content = '# Hello\n\nThis is **bold** text.';
    const result = renderMarkdown(content);
    const { container } = renderWithMantine(<>{result}</>);

    expect(container.querySelector('h1')).toHaveTextContent('Hello');
    expect(container.querySelector('strong')).toHaveTextContent('bold');
  });

  it('リンクをレンダリングできる', () => {
    const content = '[Google](https://google.com)';
    const result = renderMarkdown(content);
    const { container } = renderWithMantine(<>{result}</>);

    const link = container.querySelector('a');
    expect(link).toHaveAttribute('href', 'https://google.com');
    expect(link).toHaveTextContent('Google');
  });

  it('URLを自動リンクできる', () => {
    const content = 'Visit https://example.com for more info.';
    const result = renderMarkdown(content);
    const { container } = renderWithMantine(<>{result}</>);

    const link = container.querySelector('a');
    expect(link).toHaveAttribute('href', 'https://example.com');
  });
});
