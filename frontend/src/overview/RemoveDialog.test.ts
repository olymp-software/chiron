// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { RemoveCourse } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import RemoveDialog from './RemoveDialog.vue'

vi.mock('../../wailsjs/go/main/App', () => ({ RemoveCourse: vi.fn() }))

beforeEach(() => {
  vi.resetAllMocks()
  HTMLDialogElement.prototype.showModal = vi.fn()
  HTMLDialogElement.prototype.close = vi.fn()
})

function mountDialog(workshop = false) {
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  return mount(RemoveDialog, {
    props: { courseKey: 'go-1', title: 'Go', workshop },
    global: { plugins: [i18n] },
  })
}

function checkbox(w: ReturnType<typeof mountDialog>, name: string) {
  return w.get<HTMLInputElement>(`input[name=${name}]`)
}

it('asks with both boxes unchecked', () => {
  const w = mountDialog()

  expect(w.get('h2').text()).toBe('Remove the course “Go”?')
  expect(checkbox(w, 'keepProgress').element.checked).toBe(false)
  expect(checkbox(w, 'keepFiles').element.checked).toBe(false)
  expect(w.findAll('.actions button').map((b) => b.text())).toEqual(['Cancel', 'Remove'])
})

it('names a workshop and its folder', () => {
  const w = mountDialog(true)

  expect(w.get('h2').text()).toBe('Remove the workshop “Go”?')
  expect(w.text()).toContain('Keep my files (the workshop folder)')
})

it('cancels without removing', async () => {
  const w = mountDialog()

  await w.get('.cancel').trigger('click')

  expect(RemoveCourse).not.toHaveBeenCalled()
  expect(w.emitted('cancel')).toHaveLength(1)
})

it('removes with the checked boxes and disables the buttons meanwhile', async () => {
  let resolve: (r: main.RemoveResult) => void = () => {}
  vi.mocked(RemoveCourse).mockReturnValue(new Promise((r) => (resolve = r)))
  const w = mountDialog()
  await checkbox(w, 'keepFiles').setValue(true)

  await w.get('.confirm').trigger('click')

  expect(RemoveCourse).toHaveBeenCalledExactlyOnceWith('go-1', false, true)
  expect(w.findAll('.actions button').every((b) => b.attributes('disabled') !== undefined)).toBe(
    true,
  )

  resolve(main.RemoveResult.createFrom({ left: [] }))
  await flushPromises()

  expect(w.emitted('removed')).toHaveLength(1)
})

it('lists the paths left before reporting the removal', async () => {
  vi.mocked(RemoveCourse).mockResolvedValue(
    main.RemoveResult.createFrom({ left: ['C:\\learner\\go-1'] }),
  )
  const w = mountDialog()

  await w.get('.confirm').trigger('click')
  await flushPromises()

  expect(w.get('h2').text()).toBe('Some files are left on disk')
  expect(w.findAll('.left li').map((li) => li.text())).toEqual(['C:\\learner\\go-1'])
  expect(w.emitted('removed')).toBeUndefined()

  await w.get('.actions button').trigger('click')

  expect(w.emitted('removed')).toHaveLength(1)
})

it('reports an error when nothing was removed', async () => {
  const error = { code: 'courseNotFound', params: { key: 'go-1' } }
  vi.mocked(RemoveCourse).mockResolvedValue(main.RemoveResult.createFrom({ left: [], error }))
  const w = mountDialog()

  await w.get('.confirm').trigger('click')
  await flushPromises()

  expect(w.emitted('failed')?.[0]).toEqual([error])
  expect(w.emitted('removed')).toBeUndefined()
})
