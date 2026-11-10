// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'
import { Activity, CompleteActivity } from '../../wailsjs/go/main/App'
import { apperr, library, main } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import ActivityView from './ActivityView.vue'

vi.mock('../../wailsjs/go/main/App', () => ({
  Activity: vi.fn(),
  CompleteActivity: vi.fn(),
}))

const loadActivity = vi.mocked(Activity)
const completeActivity = vi.mocked(CompleteActivity)

function tree(fields: Partial<library.Tree> = {}): library.Tree {
  return library.Tree.createFrom({
    title: 'Go',
    current: 'hello',
    completed: false,
    sections: [],
    ...fields,
  })
}

function activity(fields: Partial<library.Activity> = {}): library.Activity {
  return library.Activity.createFrom({
    slug: 'hello',
    title: 'Hello',
    dir: 'basics/hello',
    type: 'theory',
    theoryBlocks: [{ kind: 'markdown', content: '# Hi' }],
    last: false,
    courseCompleted: false,
    ...fields,
  })
}

function loads(a: library.Activity) {
  loadActivity.mockResolvedValue(main.ActivityResult.createFrom({ activity: a }))
}

let router: Router

beforeEach(async () => {
  vi.resetAllMocks()
  router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/courses/:key/activities/:slug', component: { render: () => null } }],
  })
  await router.push('/courses/go-1/activities/hello')
})

function mountView(slug = 'hello', t = tree()) {
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  return mount(ActivityView, {
    props: { courseKey: 'go-1', slug, tree: t },
    global: { plugins: [i18n, router] },
  })
}

describe('ActivityView', () => {
  it('shows the theory blocks in order', async () => {
    loads(
      activity({
        theoryBlocks: [
          { kind: 'markdown', content: '# One' },
          { kind: 'markdown', content: '# Two' },
        ],
      }),
    )

    const w = mountView()

    await vi.waitFor(() =>
      expect(w.findAll('.markdown h1').map((h) => h.text())).toEqual(['One', 'Two']),
    )
    expect(loadActivity).toHaveBeenCalledExactlyOnceWith('go-1', 'hello')
  })

  it('shows the activity error and Next', async () => {
    loads(
      activity({
        theoryBlocks: [],
        error: apperr.Error.createFrom({
          code: 'activityFileMissing',
          params: { file: 'basics/hello/activity.yaml' },
        }),
      }),
    )

    const w = mountView()
    await flushPromises()

    expect(w.get('[role=alert]').text()).toContain(
      'basics/hello/activity.yaml does not exist in the course.',
    )
    expect(w.get('footer button').text()).toBe('Next')
  })

  it('shows Finish course on the last activity', async () => {
    loads(activity({ last: true }))

    const w = mountView()
    await flushPromises()

    expect(w.get('footer button').text()).toBe('Finish course')
  })

  it('shows the completed course instead of Finish course', async () => {
    loads(activity({ last: true, courseCompleted: true }))

    const w = mountView()
    await flushPromises()

    expect(w.get('footer [role=status]').text()).toBe('Course completed')
    expect(w.find('footer button').exists()).toBe(false)
  })

  it('Next opens the next activity and updates the tree', async () => {
    loads(activity())
    const updated = tree({ current: 'loops' })
    completeActivity.mockResolvedValue(
      main.CompleteResult.createFrom({ next: 'loops', tree: updated }),
    )
    const w = mountView()
    await flushPromises()

    await w.get('footer button').trigger('click')
    await flushPromises()

    expect(completeActivity).toHaveBeenCalledExactlyOnceWith('go-1', 'hello')
    expect(w.emitted('tree')).toEqual([[updated]])
    expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/loops')
  })

  it('Finish course shows the completed course', async () => {
    loads(activity({ last: true }))
    completeActivity.mockResolvedValue(
      main.CompleteResult.createFrom({ next: '', tree: tree({ completed: true }) }),
    )
    const w = mountView()
    await flushPromises()

    await w.get('footer button').trigger('click')
    await flushPromises()

    expect(w.get('footer [role=status]').text()).toBe('Course completed')
    expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/hello')
  })

  it('Finish course with an earlier activity left opens it', async () => {
    loads(activity({ last: true }))
    completeActivity.mockResolvedValue(
      main.CompleteResult.createFrom({ next: '', tree: tree({ current: 'middle' }) }),
    )
    const w = mountView()
    await flushPromises()

    await w.get('footer button').trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/middle')
  })

  it('stays where the learner went while Next was running', async () => {
    loads(activity())
    let answer: (res: main.CompleteResult) => void = () => {}
    completeActivity.mockReturnValue(new Promise((resolve) => (answer = resolve)))
    const updated = tree({ current: 'loops' })
    const w = mountView()
    await flushPromises()

    await w.get('footer button').trigger('click')
    await router.push('/courses/go-1/activities/maps')
    await w.setProps({ slug: 'maps' })
    await flushPromises()
    answer(main.CompleteResult.createFrom({ next: 'loops', tree: updated }))
    await flushPromises()

    expect(w.emitted('tree')).toEqual([[updated]])
    expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/maps')
  })

  it('does not redirect after it is unmounted', async () => {
    await router.push('/courses/go-1/activities/maps')
    let answer: (res: main.ActivityResult) => void = () => {}
    loadActivity.mockReturnValue(new Promise((resolve) => (answer = resolve)))
    const replace = vi.spyOn(router, 'replace')
    const w = mountView('maps')

    w.unmount()
    answer(
      main.ActivityResult.createFrom({
        error: { code: 'activityUnavailable', params: { slug: 'maps' } },
      }),
    )
    await flushPromises()

    expect(replace).not.toHaveBeenCalled()
  })

  function quizActivity(slug = 'hello'): library.Activity {
    return activity({
      slug,
      type: 'quiz',
      quiz: library.Quiz.createFrom({
        questions: [
          {
            kind: 'choice',
            prompt: 'Pick',
            explanation: '',
            multiple: false,
            options: [
              { text: 'Right', correct: true, feedback: '' },
              { text: 'Wrong', correct: false, feedback: '' },
            ],
          },
        ],
      }),
    })
  }

  it('opens a quiz on its theory screen', async () => {
    loads(quizActivity())

    const w = mountView()
    await flushPromises()

    await vi.waitFor(() => expect(w.get('.theory .markdown h1').text()).toBe('Hi'))
    expect(w.get('.theory .start').text()).toBe('Start quiz')
    expect(w.find('.question').exists()).toBe(false)
  })

  it('offers Next on both quiz screens', async () => {
    loads(quizActivity())
    completeActivity.mockResolvedValue(
      main.CompleteResult.createFrom({ next: 'loops', tree: tree({ current: 'loops' }) }),
    )
    const w = mountView()
    await flushPromises()

    expect(w.get('footer button').text()).toBe('Next')
    await w.get('.theory .start').trigger('click')
    expect(w.get('footer button').text()).toBe('Next')
    await w.get('footer button').trigger('click')
    await flushPromises()

    expect(completeActivity).toHaveBeenCalledExactlyOnceWith('go-1', 'hello')
  })

  it('starts a reopened quiz fresh', async () => {
    loads(quizActivity())
    const w = mountView()
    await flushPromises()
    await w.get('.theory .start').trigger('click')
    await w.findAll('.question input')[0].trigger('change')
    await w.get('.question button').trigger('click')
    expect(w.get('.result').text()).toBe('Correct')

    loads(activity({ slug: 'loops' }))
    await w.setProps({ slug: 'loops' })
    await flushPromises()
    loads(quizActivity())
    await w.setProps({ slug: 'hello' })
    await flushPromises()

    expect(w.get('.theory .start').text()).toBe('Start quiz')
    await w.get('.theory .start').trigger('click')
    expect(w.get('.result').text()).toBe('')
    expect(w.get('.counter').text()).toBe('Correct: 0 of 1')
  })

  it('opens the current activity instead of an unavailable one', async () => {
    await router.push('/courses/go-1/activities/maps')
    loadActivity.mockResolvedValue(
      main.ActivityResult.createFrom({
        error: { code: 'activityUnavailable', params: { slug: 'maps' } },
      }),
    )

    mountView('maps')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/hello')
  })
})
