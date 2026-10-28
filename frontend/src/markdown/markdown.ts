import githubDark from '@shikijs/themes/github-dark'
import githubLight from '@shikijs/themes/github-light'
import markdownIt, { type MarkdownIt } from 'markdown-it'
import { createHighlighterCore, type HighlighterCore } from 'shiki/core'
import { createOnigurumaEngine } from 'shiki/engine/oniguruma'
import { bundledLanguages } from 'shiki/langs'
import { useDialect, type CalloutTitles } from './dialect'
import type { ImageResolver } from './images'

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

// RenderOptions adapt rendering to where the Markdown comes from.
export type RenderOptions = {
  // resolveImage gives image URLs; without it images show their alt text.
  resolveImage?: ImageResolver
  // calloutTitles are the callout titles in the UI language.
  calloutTitles?: CalloutTitles
}

function createRenderer(hl: HighlighterCore): MarkdownIt {
  const md = markdownIt({
    html: false,
    linkify: true,
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
  const renderImage = md.renderer.rules.image!
  // Images load only through the resolver; without one, or when it refuses a
  // src, the alt text is shown.
  md.renderer.rules.image = (tokens, idx, options, env, self) => {
    const token = tokens[idx]
    const resolve = (env as RenderOptions | undefined)?.resolveImage
    const src = resolve?.(String(token.attrGet('src') ?? '')) ?? null
    if (src === null) {
      return md.utils.escapeHtml(self.renderInlineAsText(token.children ?? [], options, env))
    }
    token.attrSet('src', src)
    return renderImage(tokens, idx, options, env, self)
  }
  useDialect(md)
  return md
}

let renderer: MarkdownIt | undefined

// renderMarkdown renders Markdown to HTML with raw HTML escaped. Grammars of
// the fence languages are loaded before rendering; unknown languages render
// as plain text.
export async function renderMarkdown(source: string, opts: RenderOptions = {}): Promise<string> {
  const hl = await getHighlighter()
  renderer ??= createRenderer(hl)
  const env = { ...opts }
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
