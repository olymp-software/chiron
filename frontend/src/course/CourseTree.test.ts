// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import { library } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import CourseTree from './CourseTree.vue'

// The first two activities are completed, the third is current.
const tree = library.Tree.createFrom({
  title: 'Go',
  current: 'a3',
  completed: false,
  sections: [
    {
      slug: 'basics',
      title: 'Basics',
      activities: [
        { slug: 'a1', title: 'A1', completed: true, locked: false },
        { slug: 'a2', title: 'A2', completed: true, locked: false },
      ],
    },
    {
      slug: 'more',
      title: 'More',
      activities: [
        { slug: 'a3', title: 'A3', completed: false, locked: false },
        { slug: 'a4', title: 'A4', completed: false, locked: true },
        { slug: 'a5', title: 'A5', completed: false, locked: true },
      ],
    },
  ],
})

async function mountTree() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/courses/:key/activities/:slug', component: { render: () => null } }],
  })
  await router.push('/courses/go-1/activities/a3')
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  const w = mount(CourseTree, {
    props: { courseKey: 'go-1', tree },
    global: { plugins: [i18n, router] },
  })
  return { w, router }
}

// item returns the tree item of the activity titled title.
function item(w: Awaited<ReturnType<typeof mountTree>>['w'], title: string) {
  const li = w.findAll('li').find((l) => l.text().startsWith(title))
  if (!li) {
    throw new Error(`no item ${title}`)
  }
  return li
}

function marks(w: Awaited<ReturnType<typeof mountTree>>['w'], title: string): string[] {
  return item(w, title)
    .findAll('.mark')
    .map((m) => m.attributes('aria-label') ?? '')
}

it('marks completed, current and locked activities', async () => {
  const { w } = await mountTree()

  expect(marks(w, 'A1')).toEqual(['Completed'])
  expect(marks(w, 'A2')).toEqual(['Completed'])
  expect(marks(w, 'A3')).toEqual(['Current activity'])
  for (const title of ['A1', 'A2', 'A3']) {
    expect(item(w, title).find('a').exists()).toBe(true)
  }
  for (const title of ['A4', 'A5']) {
    expect(item(w, title).find('a').exists()).toBe(false)
    expect(item(w, title).get('[aria-disabled=true]').text()).toBe(title)
  }
})

it('does not open a locked activity on click', async () => {
  const { w, router } = await mountTree()

  await item(w, 'A4').get('[aria-disabled=true]').trigger('click')
  await flushPromises()

  expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/a3')
})
