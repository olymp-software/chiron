// @vitest-environment jsdom
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { describe, expect, it, onTestFinished, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { library } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import QuizActivity from './QuizActivity.vue'

// quiz has three single-choice questions Q1..Q3 with options Right (correct)
// and Wrong, in that written order; shuffle sets both shuffle flags.
function quiz(shuffle = false): library.Quiz {
  return library.Quiz.createFrom({
    shuffle,
    questions: [1, 2, 3].map((n) => ({
      kind: 'choice',
      prompt: `Q${n}`,
      explanation: '',
      multiple: false,
      shuffle,
      options: [
        { text: 'Right', correct: true, feedback: '' },
        { text: 'Wrong', correct: false, feedback: '' },
      ],
    })),
  })
}

async function mountQuiz(attachTo?: HTMLElement, shuffle = false) {
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  const w = mount(QuizActivity, {
    attachTo,
    props: {
      theoryBlocks: [library.TheoryBlock.createFrom({ kind: 'markdown', content: '# Intro' })],
      quiz: quiz(shuffle),
      resolveImage: () => null,
    },
    global: { plugins: [i18n] },
  })
  await flushPromises()
  return w
}

// answer selects option i of the open question and checks it.
async function answer(w: VueWrapper, i: number) {
  await w.findAll('.question input')[i].trigger('change')
  await w.get('.question button').trigger('click')
}

function mapLabels(w: VueWrapper): (string | undefined)[] {
  return w.findAll('nav button').map((b) => b.attributes('aria-label'))
}

function openQuestion(w: VueWrapper): string {
  return w.get('.position').text()
}

// shownOrder opens every question from the map and returns its prompt and
// option texts, in the order shown.
async function shownOrder(w: VueWrapper): Promise<string[]> {
  const out: string[] = []
  for (const b of w.findAll('nav button')) {
    await b.trigger('click')
    await flushPromises()
    const prompt = /Q\d/.exec(w.get('.question').text())?.[0]
    const options = w.findAll('.question .option .text').map((t) => t.text())
    out.push(`${prompt}: ${options.join(' ')}`)
  }
  return out
}

// randomness makes Math.random return value; 0 turns Fisher-Yates [a, b, c]
// into [b, c, a], 0.99 keeps the order.
function randomness(value: number) {
  const spy = vi.spyOn(Math, 'random').mockReturnValue(value)
  onTestFinished(() => spy.mockRestore())
  return spy
}

const shuffledOrder = ['Q2: Wrong Right', 'Q3: Wrong Right', 'Q1: Wrong Right']
const writtenOrder = ['Q1: Right Wrong', 'Q2: Right Wrong', 'Q3: Right Wrong']

describe('QuizActivity', () => {
  it('opens on the theory screen without questions', async () => {
    const w = await mountQuiz()

    expect(w.find('.theory').exists()).toBe(true)
    expect(w.get('.theory .start').text()).toBe('Start quiz')
    expect(w.find('.question').exists()).toBe(false)
  })

  it('Start quiz opens the first question', async () => {
    const w = await mountQuiz()

    await w.get('.theory .start').trigger('click')

    expect(w.find('.theory').exists()).toBe(false)
    expect(openQuestion(w)).toBe('Question 1 of 3')
    expect(w.get('.counter').text()).toBe('Correct: 0 of 3')
  })

  it('moves between questions without leaving their bounds', async () => {
    const w = await mountQuiz()
    await w.get('.theory .start').trigger('click')

    expect(w.get('.previous').attributes('disabled')).toBeDefined()
    await w.get('.next-question').trigger('click')
    await w.get('.next-question').trigger('click')

    expect(openQuestion(w)).toBe('Question 3 of 3')
    expect(w.get('.next-question').attributes('disabled')).toBeDefined()
    await w.get('.previous').trigger('click')
    expect(openQuestion(w)).toBe('Question 2 of 3')
  })

  it('stays on a question after checking it', async () => {
    const w = await mountQuiz()
    await w.get('.theory .start').trigger('click')

    await answer(w, 0)

    expect(openQuestion(w)).toBe('Question 1 of 3')
    expect(w.get('.result').text()).toBe('Correct')
    expect(w.get('.counter').text()).toBe('Correct: 1 of 3')
  })

  it('opens a question from the map', async () => {
    const w = await mountQuiz()
    await w.get('.theory .start').trigger('click')

    await w.findAll('nav button')[2].trigger('click')

    expect(openQuestion(w)).toBe('Question 3 of 3')
  })

  it('keeps the quiz across the theory screen', async () => {
    const w = await mountQuiz()
    await w.get('.theory .start').trigger('click')
    await answer(w, 1)
    await w.get('.next-question').trigger('click')
    await w.findAll('.question input')[0].trigger('change')

    await w.get('.to-theory').trigger('click')
    expect(w.find('.question').exists()).toBe(false)
    expect(w.get('.theory .start').text()).toBe('Back to quiz')
    await w.get('.theory .start').trigger('click')

    expect(openQuestion(w)).toBe('Question 2 of 3')
    expect((w.findAll('.question input')[0].element as HTMLInputElement).checked).toBe(true)
    expect(mapLabels(w)).toEqual([
      'Question 1: answered incorrectly',
      'Question 2: not answered',
      'Question 3: not answered',
    ])
    expect(w.get('.counter').text()).toBe('Correct: 0 of 3')
    await w.get('.previous').trigger('click')
    expect(w.get('.result').text()).toBe('Incorrect')
  })

  it('moves focus to the screen it switches to', async () => {
    // jsdom has no scrollIntoView.
    const scroll = vi.fn()
    Element.prototype.scrollIntoView = scroll
    onTestFinished(() => {
      delete (Element.prototype as Partial<Element>).scrollIntoView
    })
    const w = await mountQuiz(document.body)

    await w.get('.theory .start').trigger('click')
    await flushPromises()
    expect(document.activeElement).toBe(w.get('.quiz').element)
    await w.get('.to-theory').trigger('click')
    await flushPromises()

    expect(document.activeElement).toBe(w.get('.theory').element)
    expect(scroll).toHaveBeenCalledTimes(2)
    w.unmount()
  })

  it('Start over clears every answer and opens the first question', async () => {
    const w = await mountQuiz()
    await w.get('.theory .start').trigger('click')
    await answer(w, 0)
    await w.get('.next-question').trigger('click')
    await answer(w, 1)
    await w.get('.next-question').trigger('click')
    await answer(w, 0)

    await w.get('.start-over').trigger('click')

    expect(openQuestion(w)).toBe('Question 1 of 3')
    expect(w.get('.result').text()).toBe('')
    expect(w.findAll('.question input').some((i) => (i.element as HTMLInputElement).checked)).toBe(
      false,
    )
    expect(mapLabels(w).every((l) => l?.endsWith('not answered'))).toBe(true)
    expect(w.get('.counter').text()).toBe('Correct: 0 of 3')
  })
})

describe('QuizActivity shuffling', () => {
  it('shows the arranged order, numbered as shown', async () => {
    randomness(0)
    const w = await mountQuiz(undefined, true)
    await w.get('.theory .start').trigger('click')
    await flushPromises()

    expect(openQuestion(w)).toBe('Question 1 of 3')
    expect(w.get('.question').text()).toContain('Q2')
    expect(w.get('.previous').attributes('disabled')).toBeDefined()
    // Right is the second option shown; checking it is correct.
    await answer(w, 1)
    expect(w.get('.result').text()).toBe('Correct')
    expect(await shownOrder(w)).toEqual(shuffledOrder)
  })

  it('keeps the order across the theory screen', async () => {
    const random = randomness(0)
    const w = await mountQuiz(undefined, true)
    await w.get('.theory .start').trigger('click')
    random.mockReturnValue(0.99)

    await w.get('.to-theory').trigger('click')
    await w.get('.theory .start').trigger('click')

    expect(await shownOrder(w)).toEqual(shuffledOrder)
  })

  it('Start over draws a new order', async () => {
    const random = randomness(0)
    const w = await mountQuiz(undefined, true)
    await w.get('.theory .start').trigger('click')
    random.mockReturnValue(0.99)

    await w.get('.start-over').trigger('click')

    expect(openQuestion(w)).toBe('Question 1 of 3')
    expect(await shownOrder(w)).toEqual(writtenOrder)
  })

  it('draws a new order when mounted again', async () => {
    const random = randomness(0)
    const first = await mountQuiz(undefined, true)
    await first.get('.theory .start').trigger('click')
    expect(await shownOrder(first)).toEqual(shuffledOrder)
    first.unmount()
    random.mockReturnValue(0.99)

    const second = await mountQuiz(undefined, true)
    await second.get('.theory .start').trigger('click')

    expect(await shownOrder(second)).toEqual(writtenOrder)
  })

  it('keeps the written order without shuffle', async () => {
    randomness(0)
    const w = await mountQuiz()
    await w.get('.theory .start').trigger('click')

    expect(await shownOrder(w)).toEqual(writtenOrder)
  })
})
