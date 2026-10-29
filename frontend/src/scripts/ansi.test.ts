import { describe, expect, it } from 'vitest'
import { type Line, Terminal } from './ansi'

// text returns the plain text of the lines.
function text(lines: Line[]): string[] {
  return lines.map((l) => l.map((s) => s.text).join(''))
}

function render(...chunks: string[]): Line[] {
  const t = new Terminal(100)
  for (const c of chunks) t.write(c)
  return t.lines()
}

describe('Terminal', () => {
  it('splits lines and leaves out an empty last line', () => {
    expect(text(render('a\nb\n'))).toEqual(['a', 'b'])
    expect(text(render('a\nb'))).toEqual(['a', 'b'])
    expect(text(render('a\n\nb\n'))).toEqual(['a', '', 'b'])
  })

  it('shows a progress bar as one line', () => {
    expect(text(render('10%\r20%\r30%'))).toEqual(['30%'])
    expect(text(render('10%\r', '20%\r', '30%\ndone\n'))).toEqual(['30%', 'done'])
  })

  it('keeps lines ending in CRLF', () => {
    expect(text(render('a\r\nb\r', '\nc'))).toEqual(['a', 'b', 'c'])
  })

  it('renders colored text without escape characters', () => {
    const [line] = render('\x1b[31mred\x1b[0m plain \x1b[1;92mbold green\x1b[m')
    expect(line).toEqual([
      { text: 'red', cls: 'fg-1', css: '' },
      { text: ' plain ', cls: '', css: '' },
      { text: 'bold green', cls: 'fg-10 bold', css: '' },
    ])
  })

  it('renders 256 and RGB colors inline', () => {
    const [line] = render('\x1b[38;5;196ma\x1b[48;2;1;2;3mb\x1b[39;49mc')
    expect(line).toEqual([
      { text: 'a', cls: '', css: 'color: rgb(255, 0, 0)' },
      { text: 'b', cls: '', css: 'color: rgb(255, 0, 0); background-color: rgb(1, 2, 3)' },
      { text: 'c', cls: '', css: '' },
    ])
  })

  it('keeps the style across writes and split sequences', () => {
    const [line] = render('\x1b[3', '2mgr', 'een\x1b', '[0m!')
    expect(line).toEqual([
      { text: 'green', cls: 'fg-2', css: '' },
      { text: '!', cls: '', css: '' },
    ])
  })

  it('strips cursor and other control sequences', () => {
    const lines = render(
      '\x1b[2K\x1b[1Ga\x1b[?25lb\x1b]0;title\x07c\x1b]8;;http://x\x1b\\d\x1b7e\x07\bf\tg\n',
    )
    expect(text(lines)).toEqual(['abcdef\tg'])
  })

  it('does not change lines it has returned', () => {
    const t = new Terminal(100)
    t.write('a\nb')
    const before = t.lines()
    const snapshot = text(before)

    t.write('c\nd')

    expect(text(before)).toEqual(snapshot)
    const after = t.lines()
    expect(after[0]).toBe(before[0])
    expect(text(after)).toEqual(['a', 'bc', 'd'])
  })

  it('keeps the latest lines', () => {
    const t = new Terminal(10_000)
    for (let i = 0; i < 20_000; i += 100) {
      let chunk = ''
      for (let j = i; j < i + 100; j++) chunk += `line ${j}\n`
      t.write(chunk)
    }

    const lines = t.lines()
    expect(lines).toHaveLength(10_000)
    expect(text(lines)[0]).toBe('line 10000')
    expect(text(lines)[9_999]).toBe('line 19999')
    expect(t.dropped).toBe(10_000)
  })
})
