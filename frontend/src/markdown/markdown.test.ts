import { describe, expect, it } from 'vitest'
import { courseImages } from './images'
import { renderMarkdown } from './markdown'

describe('renderMarkdown', () => {
  it('escapes raw HTML', async () => {
    const html = await renderMarkdown('<script>alert(1)</script>\n\n<b>x</b>')
    expect(html).not.toContain('<script>')
    expect(html).not.toContain('<b>')
    expect(html).toContain('&lt;script&gt;')
  })

  it('highlights a go fence', async () => {
    const html = await renderMarkdown('```go\nfunc main() {}\n```')
    expect(html).toContain('<pre class="shiki shiki-themes github-light github-dark"')
    expect(html).toMatch(/<span style="color:#[0-9A-F]+;--shiki-dark:#[0-9A-F]+">func<\/span>/)
  })

  it('highlights an alias with extra info', async () => {
    const html = await renderMarkdown('```JS title=x\nconst a = 1\n```')
    expect(html).toContain('class="shiki')
  })

  it('renders an unknown fence as plain text', async () => {
    const html = await renderMarkdown('```nosuchlang\na < b\n```')
    expect(html).toBe('<pre><code class="language-nosuchlang">a &lt; b\n</code></pre>\n')
  })

  it('renders a fence without language as plain text', async () => {
    const html = await renderMarkdown('```\nx\n```')
    expect(html).toBe('<pre><code>x\n</code></pre>\n')
  })
})

describe('renderMarkdown images', () => {
  const resolve = { resolveImage: courseImages('go-basics-4fb70bd4') }

  it('loads a relative image from the course', async () => {
    const html = await renderMarkdown('![setup](images/setup.png)', resolve)
    expect(html).toContain(
      '<img src="/course-files/go-basics-4fb70bd4/images/setup.png" alt="setup">',
    )
  })

  it('resolves a nested relative path', async () => {
    const html = await renderMarkdown('![d](./img/../images/a/b.png)', resolve)
    expect(html).toContain('src="/course-files/go-basics-4fb70bd4/images/a/b.png"')
  })

  it.each([
    'https://example.com/x.png',
    '//example.com/x.png',
    '/etc/x.png',
    '../outside.png',
    'data:image/png;base64,AAAA',
  ])('shows alt text instead of %s', async (src) => {
    const html = await renderMarkdown(`![a <b>](${src})`, resolve)
    expect(html).not.toContain('<img')
    expect(html).toContain('a &lt;b&gt;')
  })

  it('shows alt text without a resolver', async () => {
    const html = await renderMarkdown('![setup](images/setup.png)')
    expect(html).toBe('<p>setup</p>\n')
  })

  it('keeps a script tag literal next to images', async () => {
    const html = await renderMarkdown('<script>alert(1)</script>\n\n![x](x.png)', resolve)
    expect(html).not.toContain('<script>')
    expect(html).toContain('&lt;script&gt;')
  })
})

describe('courseImages', () => {
  it('resolves against a base folder', () => {
    expect(courseImages('k', 'basics/hello')('../shared/a.png')).toBe(
      '/course-files/k/basics/shared/a.png',
    )
  })

  it('keeps encoded characters and drops query and hash', () => {
    expect(courseImages('k')('my%20pic.png?v=1#x')).toBe('/course-files/k/my%20pic.png')
  })
})
