// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import { AddRepoCourse, CancelAdd, RemoteVersions } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import AddRepoDialog from './AddRepoDialog.vue'

vi.mock('../../wailsjs/go/main/App', () => ({
  AddRepoCourse: vi.fn(),
  CancelAdd: vi.fn(),
  RemoteVersions: vi.fn(),
}))

const source = 'https://github.com/owner/go'

beforeEach(() => {
  vi.resetAllMocks()
  HTMLDialogElement.prototype.showModal = vi.fn()
  vi.mocked(CancelAdd).mockResolvedValue(main.ErrorResult.createFrom({}))
})

let router: ReturnType<typeof createRouter>

function mountDialog() {
  router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/:p(.*)*', component: { render: () => null } }],
  })
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  return mount(AddRepoDialog, { global: { plugins: [i18n, router] } })
}

function versions(list: string[]) {
  return main.RemoteVersionsResult.createFrom({ remote: { source, versions: list } })
}

async function enterAddress(w: ReturnType<typeof mountDialog>, address = 'github.com/owner/go') {
  await w.get('input[name=address]').setValue(address)
  await w.get('form').trigger('submit')
  await flushPromises()
}

it('opens the overview of a course already in the library', async () => {
  vi.mocked(RemoteVersions).mockResolvedValue(
    main.RemoteVersionsResult.createFrom({
      remote: {
        source,
        versions: [],
        existing: { origin: 'repo', source, key: 'go-1', title: 'Go' },
      },
    }),
  )
  const w = mountDialog()
  const push = vi.spyOn(router, 'push')

  await enterAddress(w)

  expect(RemoteVersions).toHaveBeenCalledExactlyOnceWith('github.com/owner/go')
  expect(push).toHaveBeenCalledExactlyOnceWith('/library/go-1')
  expect(AddRepoCourse).not.toHaveBeenCalled()
  expect(w.emitted('close')).toHaveLength(1)
})

it('preselects the highest version and shows the script warning', async () => {
  vi.mocked(RemoteVersions).mockResolvedValue(versions(['v1.10.0', 'v1.2.0', 'v1.0.0']))
  const w = mountDialog()

  await enterAddress(w)

  const select = w.get<HTMLSelectElement>('select[name=version]')
  expect(select.element.value).toBe('v1.10.0')
  expect(select.findAll('option').map((o) => o.text())).toEqual(['v1.10.0', 'v1.2.0', 'v1.0.0'])
  expect(w.text()).toContain(en.scriptWarning.text)
  expect(w.text()).toContain(source)
})

it('adds the chosen version and opens its overview', async () => {
  vi.mocked(RemoteVersions).mockResolvedValue(versions(['v1.1.0', 'v1.0.0']))
  vi.mocked(AddRepoCourse).mockResolvedValue(
    main.AddRepoResult.createFrom({ course: { origin: 'repo', source, key: 'go-1', title: 'Go' } }),
  )
  const w = mountDialog()
  const push = vi.spyOn(router, 'push')
  await enterAddress(w)

  await w.get('select[name=version]').setValue('v1.0.0')
  await w.get('form').trigger('submit')
  await flushPromises()

  expect(AddRepoCourse).toHaveBeenCalledExactlyOnceWith(source, 'v1.0.0')
  expect(push).toHaveBeenCalledExactlyOnceWith('/library/go-1')
  expect(w.emitted('close')).toHaveLength(1)
})

it('cancels the clone and returns to the version choice', async () => {
  vi.mocked(RemoteVersions).mockResolvedValue(versions(['v1.0.0']))
  let finish: (r: main.AddRepoResult) => void = () => {}
  vi.mocked(AddRepoCourse).mockReturnValue(new Promise((resolve) => (finish = resolve)))
  const w = mountDialog()
  await enterAddress(w)
  await w.get('form').trigger('submit')
  await flushPromises()

  expect(w.get('[role=status]').text()).toBe(en.addRepo.cloning)

  await w.get('.cancel').trigger('click')
  await flushPromises()

  expect(CancelAdd).toHaveBeenCalledOnce()
  expect(w.emitted('close')).toBeUndefined()

  finish(main.AddRepoResult.createFrom({ cancelled: true }))
  await flushPromises()

  expect(w.find('select[name=version]').exists()).toBe(true)
  expect(w.find('.error').exists()).toBe(false)
})

it('shows a listing error at the address', async () => {
  vi.mocked(RemoteVersions).mockResolvedValue(
    main.RemoteVersionsResult.createFrom({
      remote: { versions: [] },
      error: { code: 'repoNotFound', params: {} },
    }),
  )
  const w = mountDialog()

  await enterAddress(w)

  expect(w.find('input[name=address]').exists()).toBe(true)
  expect(w.get('.error').text()).toContain(en.errors.repoNotFound)
})

it('shows a clone error at the version choice', async () => {
  vi.mocked(RemoteVersions).mockResolvedValue(versions(['v1.0.0']))
  vi.mocked(AddRepoCourse).mockResolvedValue(
    main.AddRepoResult.createFrom({
      error: { code: 'manifestMissing', params: { file: 'course.yaml' } },
    }),
  )
  const w = mountDialog()
  await enterAddress(w)

  await w.get('form').trigger('submit')
  await flushPromises()

  expect(w.find('select[name=version]').exists()).toBe(true)
  expect(w.find('.error').exists()).toBe(true)
  expect(w.emitted('close')).toBeUndefined()
})

it('shows the output of an unknown git failure', async () => {
  vi.mocked(RemoteVersions).mockResolvedValue(
    main.RemoteVersionsResult.createFrom({
      remote: { versions: [] },
      error: { code: 'gitFailed', params: { output: 'fatal: something odd' } },
    }),
  )
  const w = mountDialog()

  await enterAddress(w)

  expect(w.get('.error .output').text()).toBe('fatal: something odd')
})
