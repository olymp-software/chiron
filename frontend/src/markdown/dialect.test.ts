import { describe, expect, it } from 'vitest'
import en from '../i18n/en.json'
import ru from '../i18n/ru.json'
import { calloutTypes, type CalloutTitles } from './dialect'
import { externalLink } from './links'
import { renderMarkdown } from './markdown'

const calloutTitles: CalloutTitles = { note: 'N', tip: 'T', warning: 'W', task: 'K' }

describe('callouts', () => {
  it.each([
    ['NOTE', 'note', 'N'],
    ['tip', 'tip', 'T'],
    ['Warning', 'warning', 'W'],
    ['TASK', 'task', 'K'],
  ])('renders [!%s] as a %s callout', async (marker, type, title) => {
    const html = await renderMarkdown(`> [!${marker}]\n> Read *this*.`, { calloutTitles })
    expect(html).toBe(
      `<div class="callout callout-${type}">\n<p class="callout-title">${title}</p>\n` +
        '<p>Read <em>this</em>.</p>\n</div>\n',
    )
  })

  it('keeps an unknown type a plain blockquote', async () => {
    const html = await renderMarkdown('> [!CAUTION]\n> Text', { calloutTitles })
    expect(html).toBe('<blockquote>\n<p>[!CAUTION]\nText</p>\n</blockquote>\n')
  })

  it('keeps a marker followed by text on the same line a blockquote', async () => {
    const html = await renderMarkdown('> [!NOTE] Text', { calloutTitles })
    expect(html).toContain('<blockquote>')
  })

  it('keeps the following paragraphs and nested quotes inside', async () => {
    const html = await renderMarkdown('> [!TIP]\n>\n> First\n>\n> > Nested\n\nAfter', {
      calloutTitles,
    })
    expect(html).toBe(
      '<div class="callout callout-tip">\n<p class="callout-title">T</p>\n' +
        '<p>First</p>\n<blockquote>\n<p>Nested</p>\n</blockquote>\n</div>\n<p>After</p>\n',
    )
  })

  it('escapes the title', async () => {
    const html = await renderMarkdown('> [!note]\n> x', {
      calloutTitles: { ...calloutTitles, note: '<b>' },
    })
    expect(html).toContain('<p class="callout-title">&lt;b&gt;</p>')
  })

  it('has a translated title for every type', () => {
    for (const messages of [en, ru]) {
      for (const c of calloutTypes) {
        expect(messages.callout[c]).toBeTruthy()
      }
    }
  })
})

describe('GitHub extensions', () => {
  it('renders a task list', async () => {
    const html = await renderMarkdown('- [ ] todo\n- [x] done\n- plain')
    expect(html).toBe(
      '<ul>\n' +
        '<li class="task-list-item"><input type="checkbox" disabled> todo</li>\n' +
        '<li class="task-list-item"><input type="checkbox" disabled checked> done</li>\n' +
        '<li>plain</li>\n</ul>\n',
    )
  })

  it('links a bare URL', async () => {
    const html = await renderMarkdown('See https://example.com now')
    expect(html).toContain('<a href="https://example.com">https://example.com</a>')
  })

  it('renders a table and strikethrough', async () => {
    const html = await renderMarkdown('| a | b |\n|---|---|\n| 1 | ~~2~~ |')
    expect(html).toContain('<table>')
    expect(html).toContain('<td><s>2</s></td>')
  })
})

describe('externalLink', () => {
  it.each(['https://example.com', 'HTTP://example.com', 'mailto:a@example.com'])(
    'opens %s',
    (href) => {
      expect(externalLink(href)).toBe(href)
    },
  )

  it.each(['javascript:alert(1)', 'file:///C:/x', 'basics/hello', '#top', ''])(
    'does not open %s',
    (href) => {
      expect(externalLink(href)).toBeNull()
    },
  )
})
