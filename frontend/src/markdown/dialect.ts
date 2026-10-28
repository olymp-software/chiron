import type { MarkdownIt, StateCore, Token } from 'markdown-it'

export const calloutTypes = ['note', 'tip', 'warning', 'task'] as const
export type CalloutType = (typeof calloutTypes)[number]

// CalloutTitles are the shown titles of callouts, in the UI language.
export type CalloutTitles = Record<CalloutType, string>

// CalloutMeta marks both blockquote tokens of a callout.
type CalloutMeta = { callout: CalloutType }

const calloutMarker = /^\[!(\w+)\][ \t]*(?:\n|$)/

// callouts turns a blockquote whose first line is [!type] into a callout of
// that type, GitHub style. Unknown types stay plain blockquotes.
function callouts(state: StateCore) {
  const tokens = state.tokens
  for (let i = 0; i + 2 < tokens.length; i++) {
    if (tokens[i].type !== 'blockquote_open' || tokens[i + 1].type !== 'paragraph_open') {
      continue
    }
    const inline = tokens[i + 2]
    const m = calloutMarker.exec(inline.content)
    const type = m?.[1].toLowerCase()
    if (!m || !calloutTypes.includes(type as CalloutType)) {
      continue
    }
    inline.content = inline.content.slice(m[0].length)
    if (inline.content === '') {
      // The marker was the whole paragraph: drop the empty paragraph.
      tokens.splice(i + 1, 3)
    }
    const meta: CalloutMeta = { callout: type as CalloutType }
    tokens[i].meta = meta
    tokens[matchingClose(tokens, i)].meta = meta
  }
}

// matchingClose returns the index of the blockquote_close of the
// blockquote_open at open.
function matchingClose(tokens: Token[], open: number): number {
  let depth = 0
  for (let j = open + 1; j < tokens.length; j++) {
    if (tokens[j].type === 'blockquote_open') {
      depth++
    } else if (tokens[j].type === 'blockquote_close') {
      if (depth === 0) {
        return j
      }
      depth--
    }
  }
  throw new Error('unbalanced blockquote tokens')
}

// taskLists replaces a leading [ ] or [x] in a list item with a checkbox.
function taskLists(state: StateCore) {
  const tokens = state.tokens
  for (let i = 0; i + 2 < tokens.length; i++) {
    if (tokens[i].type !== 'list_item_open' || tokens[i + 1].type !== 'paragraph_open') {
      continue
    }
    const children = tokens[i + 2].children
    const first = children?.[0]
    const m = first?.type === 'text' ? /^\[([ xX])\][ \t]+/.exec(first.content) : null
    if (!children || !first || !m) {
      continue
    }
    first.content = first.content.slice(m[0].length)
    const box = new state.Token('task_checkbox', 'input', 0)
    box.meta = { checked: m[1] !== ' ' }
    children.unshift(box)
    tokens[i].attrJoin('class', 'task-list-item')
  }
}

// useDialect adds callouts and task lists to a markdown-it instance. The
// callout titles come from env.calloutTitles.
export function useDialect(md: MarkdownIt) {
  // Before inline parsing, so the marker is not part of the callout text.
  md.core.ruler.before('inline', 'callouts', callouts)
  md.core.ruler.push('task_lists', taskLists)

  md.renderer.rules.task_checkbox = (tokens: Token[], idx: number) => {
    const checked = (tokens[idx].meta as { checked: boolean }).checked
    return `<input type="checkbox" disabled${checked ? ' checked' : ''}> `
  }

  const renderQuoteOpen = md.renderer.rules.blockquote_open
  const renderQuoteClose = md.renderer.rules.blockquote_close
  md.renderer.rules.blockquote_open = (tokens, idx, options, env, self) => {
    const type = (tokens[idx].meta as CalloutMeta | null)?.callout
    if (!type) {
      return renderQuoteOpen
        ? renderQuoteOpen(tokens, idx, options, env, self)
        : self.renderToken(tokens, idx, options)
    }
    const titles: Partial<CalloutTitles> =
      (env as { calloutTitles?: CalloutTitles } | undefined)?.calloutTitles ?? {}
    const title = md.utils.escapeHtml(titles[type] ?? type)
    return `<div class="callout callout-${type}">\n<p class="callout-title">${title}</p>\n`
  }
  md.renderer.rules.blockquote_close = (tokens, idx, options, env, self) => {
    if ((tokens[idx].meta as CalloutMeta | null)?.callout) {
      return '</div>\n'
    }
    return renderQuoteClose
      ? renderQuoteClose(tokens, idx, options, env, self)
      : self.renderToken(tokens, idx, options)
  }
}
