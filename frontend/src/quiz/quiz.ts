import type { library } from '../../wailsjs/go/models'

// Result is what the question map shows for a question.
export type Result = 'unchecked' | 'correct' | 'incorrect'

// Answer is the learner's work on one question: the selected option indexes
// and, once checked, whether the answer is correct. A checked answer is locked.
export interface Answer {
  selected: number[]
  checked: boolean
  correct: boolean
}

// arrange returns a copy of quiz in the order to show: the questions permuted
// when quiz.shuffle is set and the options of each question permuted when its
// own shuffle is set. Options are moved whole, so correctness and feedback
// stay with them; the input is left untouched. random returns numbers in
// [0, 1), like Math.random.
export function arrange(quiz: library.Quiz, random: () => number = Math.random): library.Quiz {
  const questions = quiz.questions.map((q) =>
    copyWith(q, { options: q.shuffle ? shuffled(q.options, random) : [...q.options] }),
  )
  return copyWith(quiz, { questions: quiz.shuffle ? shuffled(questions, random) : questions })
}

// copyWith returns a shallow copy of obj with fields replaced. The copy keeps
// the class of obj, so it stays a generated model without rebuilding it.
function copyWith<T extends object>(obj: T, fields: Partial<T>): T {
  const copy = Object.create(Object.getPrototypeOf(obj) as object | null) as T
  return Object.assign(copy, obj, fields)
}

// shuffled returns a Fisher-Yates permutation of items.
function shuffled<T>(items: readonly T[], random: () => number): T[] {
  const out = [...items]
  for (let i = out.length - 1; i > 0; i--) {
    const j = Math.floor(random() * (i + 1))
    const t = out[i]
    out[i] = out[j]
    out[j] = t
  }
  return out
}

// freshAnswers returns an unchecked answer with no selection per question.
export function freshAnswers(quiz: library.Quiz): Answer[] {
  return quiz.questions.map(() => ({ selected: [], checked: false, correct: false }))
}

// isCorrect reports whether selected, a set of option indexes, is exactly the
// set of correct options of q; there is no partial credit.
export function isCorrect(q: library.Question, selected: readonly number[]): boolean {
  const chosen = new Set(selected)
  return q.options.every((o, i) => o.correct === chosen.has(i))
}

// result returns the map state of an answer.
export function result(a: Answer): Result {
  if (!a.checked) {
    return 'unchecked'
  }
  return a.correct ? 'correct' : 'incorrect'
}
