import { describe, expect, it } from 'vitest'
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
