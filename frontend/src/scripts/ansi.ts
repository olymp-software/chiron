// Renders script output the way the script-runtime spec defines: ANSI SGR
// sequences become styles, \r returns to the start of the line, every other
// control sequence is stripped. Output arrives in chunks that may split a
// sequence, so the parser keeps state between writes.

// Span is a run of text in one style: cls are CSS classes for the 16 basic
// colors and text attributes, css is an inline style for other colors.
export interface Span {
  text: string
  cls: string
  css: string
}

// Line is one output line. A Line array is never changed once it has been
// returned by Terminal.lines, so views can skip unchanged lines by identity.
export type Line = readonly Span[]

// Color is a basic color index (0-15) or a CSS color.
type Color = number | string

interface Style {
  fg?: Color
  bg?: Color
  bold?: boolean
  dim?: boolean
  italic?: boolean
  underline?: boolean
}

const ESC = '\x1b'
const BEL = '\x07'

// An escape sequence that has not ended after this many characters is
// dropped, so a broken script cannot make the parser buffer forever.
const maxPending = 4096

export class Terminal {
  private done: Line[] = []
  private current: Line = []
  private droppedLines = 0
  // returned is set by \r: the next text replaces the current line, while a
  // following \n (a CRLF line ending) keeps it.
  private returned = false
  private style: Style = {}
  private pending = ''

  constructor(private readonly maxLines: number) {}

  // dropped is the number of lines removed from the top so far; it keeps
  // line keys stable while old lines are dropped.
  get dropped(): number {
    return this.droppedLines
  }

  // lines returns the kept lines; an empty last line is left out.
  lines(): Line[] {
    return this.current.length > 0 ? [...this.done, this.current] : this.done.slice()
  }

  write(chunk: string) {
    const s = this.pending + chunk
    this.pending = ''
    let cur = this.current.slice()
    let text = ''
    const flush = () => {
      if (text === '') return
      if (this.returned) {
        cur = []
        this.returned = false
      }
      cur = append(cur, text, this.style)
      text = ''
    }
    let i = 0
    while (i < s.length) {
      const c = s[i]
      if (c === '\n') {
        flush()
        this.done.push(cur)
        cur = []
        this.returned = false
        i++
      } else if (c === '\r') {
        flush()
        this.returned = true
        i++
      } else if (c === ESC) {
        flush()
        const end = sequenceEnd(s, i)
        if (end < 0) {
          if (s.length - i <= maxPending) this.pending = s.slice(i)
          break
        }
        const seq = s.slice(i, end)
        if (seq.startsWith(ESC + '[') && seq.endsWith('m')) {
          this.style = applySGR(this.style, seq.slice(2, -1))
        }
        i = end
      } else if (c === '\t' || c >= ' ') {
        text += c
        i++
      } else {
        // Other C0 control characters (BEL, backspace, ...) are stripped.
        i++
      }
    }
    flush()
    this.current = cur
    this.trim()
  }

  private trim() {
    const limit = this.current.length > 0 ? this.maxLines - 1 : this.maxLines
    const extra = this.done.length - limit
    if (extra > 0) {
      this.done.splice(0, extra)
      this.droppedLines += extra
    }
  }
}

// append returns line with text added in style, merged into the last span
// when the style is the same. Spans are replaced, never changed.
function append(line: Span[], text: string, style: Style): Span[] {
  const { cls, css } = render(style)
  const last = line[line.length - 1]
  if (last && last.cls === cls && last.css === css) {
    line[line.length - 1] = { text: last.text + text, cls, css }
  } else {
    line.push({ text, cls, css })
  }
  return line
}

// sequenceEnd returns the index after the escape sequence starting at s[i],
// or -1 if it is incomplete.
function sequenceEnd(s: string, i: number): number {
  if (i + 1 >= s.length) return -1
  const kind = s[i + 1]
  if (kind === '[') {
    // CSI: parameter and intermediate bytes, then a final byte @ to ~.
    for (let j = i + 2; j < s.length; j++) {
      const code = s.charCodeAt(j)
      if (code >= 0x40 && code <= 0x7e) return j + 1
      if (code < 0x20 || code > 0x3f) return j // malformed: drop what came before
    }
    return -1
  }
  if (kind === ']' || kind === 'P' || kind === '_' || kind === '^' || kind === 'X') {
    // OSC and other strings end with BEL or ST (ESC \).
    for (let j = i + 2; j < s.length; j++) {
      if (s[j] === BEL) return j + 1
      if (s[j] === ESC) {
        if (j + 1 >= s.length) return -1
        if (s[j + 1] === '\\') return j + 2
      }
    }
    return -1
  }
  // Any other escape is two characters long, e.g. ESC 7.
  return i + 2
}

// applySGR returns style changed by the parameters of an SGR sequence.
function applySGR(style: Style, params: string): Style {
  // Parameters are separated by ";"; a parameter may have ":" subparameters.
  const groups = params.split(';').map((g) => g.split(':'))
  const codes = groups.map((g) => param(g[0]))
  let next: Style = { ...style }
  for (let i = 0; i < codes.length; i++) {
    const code = codes[i]
    if (code === 0) next = {}
    else if (code === 1) next.bold = true
    else if (code === 2) next.dim = true
    else if (code === 3) next.italic = true
    else if (code === 4) next.underline = true
    else if (code === 22) next.bold = next.dim = false
    else if (code === 23) next.italic = false
    else if (code === 24) next.underline = false
    else if (code >= 30 && code <= 37) next.fg = code - 30
    else if (code >= 90 && code <= 97) next.fg = code - 90 + 8
    else if (code === 39) next.fg = undefined
    else if (code >= 40 && code <= 47) next.bg = code - 40
    else if (code >= 100 && code <= 107) next.bg = code - 100 + 8
    else if (code === 49) next.bg = undefined
    else if (code === 38 || code === 48) {
      let color: Color | undefined
      if (groups[i].length > 1) {
        color = colonColor(groups[i].slice(1))
      } else {
        let used: number
        ;[color, used] = extendedColor(codes, i + 1)
        i += used
      }
      if (code === 38) next.fg = color
      else next.bg = color
    }
  }
  return next
}

// param returns the value of an SGR parameter; an empty one is 0.
function param(p: string): number {
  return p === '' ? 0 : Number(p)
}

// colonColor reads the subparameters of a color in colon form: 5:n, 2:r:g:b
// or ITU T.416 2:cs:r:g:b with a color space id that may be empty.
function colonColor(sub: string[]): Color | undefined {
  const codes = sub.map(param)
  if (codes[0] === 2 && codes.length >= 5) {
    return extendedColor([2, ...codes.slice(2)], 0)[0]
  }
  return extendedColor(codes, 0)[0]
}

// extendedColor reads a 256-color (5;n) or RGB (2;r;g;b) color from codes
// at i and returns it with the number of codes used. A value outside 0-255
// gives no color.
function extendedColor(codes: number[], i: number): [Color | undefined, number] {
  if (codes[i] === 5 && i + 1 < codes.length) {
    const n = codes[i + 1]
    if (!isByte(n)) return [undefined, 2]
    return [n < 16 ? n : xterm256(n), 2]
  }
  if (codes[i] === 2 && i + 3 < codes.length) {
    const [r, g, b] = codes.slice(i + 1, i + 4)
    if (![r, g, b].every(isByte)) return [undefined, 4]
    return [`rgb(${r}, ${g}, ${b})`, 4]
  }
  return [undefined, codes.length - i]
}

function isByte(n: number): boolean {
  return Number.isInteger(n) && n >= 0 && n <= 255
}

// xterm256 returns the CSS color of an xterm 256-color index from 16 on.
function xterm256(n: number): string {
  if (n >= 232) {
    const v = 8 + (n - 232) * 10
    return `rgb(${v}, ${v}, ${v})`
  }
  const i = n - 16
  const level = (c: number) => (c === 0 ? 0 : 55 + c * 40)
  return `rgb(${level(Math.floor(i / 36))}, ${level(Math.floor(i / 6) % 6)}, ${level(i % 6)})`
}

function render(style: Style): { cls: string; css: string } {
  const cls: string[] = []
  const css: string[] = []
  if (typeof style.fg === 'number') cls.push(`fg-${style.fg}`)
  else if (style.fg) css.push(`color: ${style.fg}`)
  if (typeof style.bg === 'number') cls.push(`bg-${style.bg}`)
  else if (style.bg) css.push(`background-color: ${style.bg}`)
  if (style.bold) cls.push('bold')
  if (style.dim) cls.push('dim')
  if (style.italic) cls.push('italic')
  if (style.underline) cls.push('underline')
  return { cls: cls.join(' '), css: css.join('; ') }
}
