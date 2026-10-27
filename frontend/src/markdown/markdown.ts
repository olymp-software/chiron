import githubDark from '@shikijs/themes/github-dark'
import githubLight from '@shikijs/themes/github-light'
import markdownIt, { type MarkdownIt } from 'markdown-it'
import { createHighlighterCore, type HighlighterCore } from 'shiki/core'
import { createOnigurumaEngine } from 'shiki/engine/oniguruma'
import { bundledLanguages } from 'shiki/langs'

let highlighter: Promise<HighlighterCore> | undefined

// getHighlighter creates the highlighter once; grammars are loaded per render.
function getHighlighter(): Promise<HighlighterCore> {
  highlighter ??= createHighlighterCore({
    themes: [githubLight, githubDark],
    langs: [],
    engine: createOnigurumaEngine(import('shiki/wasm')),
  })
  return highlighter
}

// fenceLanguage returns the language of a fence info string ("go title=x" → "go").
function fenceLanguage(info: string): string {
  return info.trim().split(/\s+/)[0].toLowerCase()
}

function isBundled(lang: string): lang is keyof typeof bundledLanguages {
  return Object.hasOwn(bundledLanguages, lang)
}

function createRenderer(hl: HighlighterCore): MarkdownIt {
  return markdownIt({
    html: false,
    // An empty result makes markdown-it render an escaped plain <pre><code>.
    highlight(code, info) {
      const lang = fenceLanguage(info)
      if (!hl.getLoadedLanguages().includes(lang)) {
        return ''
      }
      return hl.codeToHtml(code, {
        lang,
        themes: { light: 'github-light', dark: 'github-dark' },
      })
    },
  })
}

let renderer: MarkdownIt | undefined

// renderMarkdown renders Markdown to HTML with raw HTML escaped. Grammars of
// the fence languages are loaded before rendering; unknown languages render
// as plain text.
export async function renderMarkdown(source: string): Promise<string> {
  const hl = await getHighlighter()
  renderer ??= createRenderer(hl)
  const env = {}
  const tokens = renderer.parse(source, env)
  const langs = new Set(
    tokens
      .filter((t) => t.type === 'fence')
      .map((t) => fenceLanguage(t.info))
      .filter(isBundled),
  )
  await Promise.all([...langs].map((lang) => hl.loadLanguage(bundledLanguages[lang])))
  return renderer.renderer.render(tokens, renderer.options, env)
}
