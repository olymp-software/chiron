# Design

## Context

The archived `add-local-folder-library` change left `internal/library` (library.json, course copies, Add/Sync, Overview), `internal/manifest` (`course.yaml` with sections and activity refs including titles) and a frontend whose `App.vue` switches between `LibraryView` and `CourseOverview` with a `ref` and component events. Bound methods take the course source; `library.entry(source)` looks entries up by it. Nothing writes progress yet. See proposal.md for scope and specs/course-library for the navigation behavior.

## Goals / Non-Goals

**Goals:**
- A route table that later changes extend (activity content, Next) without reshaping it.
- One place that decides whether a course may be entered and which activity opens.
- Progress and settings files that later changes add fields to without migration.

**Non-Goals:**
- Activity locking, Next / Finish course and completed activities: `completed` is stored but nothing adds to it yet.
- Reading `activity.yaml`; the activity view shows only the title from `course.yaml`.
- The environment check: the place where it will gate entering is marked, but `check` is not evaluated.

## Decisions

### Routes with vue-router and hash history

```
/                                  startup redirect, no component
/library                           LibraryView
/library/:key                      CourseOverview
/courses/:key                      CourseScreen (header, tree, <RouterView>)
  ''                               redirect to the current activity
  activities/:slug                 ActivityView
```

- `createWebHashHistory`: the part after `#` never reaches the Wails asset server, so a webview reload on a deep route still gets `index.html`, and route paths never share a namespace with `/course-files/`.
- Parameters are storage keys and slugs. Both match `[a-z0-9-]`, so routes need no encoding, unlike course sources (`C:\...`).
- The overview stays under `/library`: it exists for courses that are not enrolled, while everything under `/courses` requires enrollment and can share one guard.
- Section slugs are not in the route: activity slugs are unique course-wide, so moving an activity between sections keeps its route.
- *Alternative: `createWebHistory`:* a reload on `/courses/...` reaches the asset server, which falls back to `library.FilesHandler()` and 404s; it would need an SPA fallback to `index.html` sharing the path space with `/course-files/`. Clean URLs give nothing in a desktop window.
- *Alternative: keep the `ref` switch, or a `useNavigation()` composable:* no dependency, but entering guards, nested routes and history would be hand-written.

### Push for the learner's moves, replace for the app's redirects
Every navigation the learner starts uses `router.push`: library ↔ overview ↔ course, the tree, the course dropdown, later Next. Redirects the app makes in place of the requested route use `replace` (in vue-router, a guard returning a location replaces the pending navigation): `/` on startup, `/courses/:key` to the current activity, the entering guard to the overview, an unknown slug to the current activity.
- *Alternative: replace for the dropdown:* replace swaps only the top entry, so Back after a switch lands on an older activity of the previous course instead of where the learner was.

### Back input
The app listens for `mouseup` with button 3 (back) and `keydown` Alt+ArrowLeft on `window`, calls `preventDefault` and `router.back()`. The overview's back button calls `router.back()` when `history.state.back` (set by vue-router) is not null, and pushes `/library` otherwise (the overview was opened by the startup redirect).
- *Alternative: rely on WebView2's own back handling:* whether Wails leaves browser accelerator keys enabled is not something the app controls or tests; handling it explicitly gives one behavior. See Risks for a double back.

### Entering a course in one Go call
A global `router.beforeEach` calls `EnterCourse(key)` whenever the target route is under `/courses` and its `:key` differs from the source route's `:key` (or the source is not a course route). A per-route `beforeEnter` is not enough: vue-router does not rerun it when only params change, which is exactly the dropdown switch from `/courses/a/...` to `/courses/b/...`.

`EnterCourse(key)`:
1. the course must be in the library and enrolled, and `course.yaml` must be valid; otherwise it returns the error and the guard redirects to `/library/:key`, where the overview shows the course error or Enroll;
2. (later: the environment check);
3. resolves the current activity against the activity order of `course.yaml` (next section) and saves the progress if the pointer moved;
4. stores the key as `lastCourse` in `settings.json`; a failure there is logged and does not block entering;
5. returns the current activity slug.

The route `/courses/:key` without a slug redirects to `activities/<slug returned by EnterCourse>`; the guard keeps the slug of the last `EnterCourse` call for that. `ActivityView` with a slug that is not in the tree replaces it with the current activity.

### Progress: `internal/progress`

```
<data>/state/<key>.json
{ "version": 1, "current": "hello-world", "completed": [] }
```

- Enrollment is the existence of this file, as decided in the archived enrollment change; there is no enrolled flag elsewhere.
- `Enroll` writes it with `current` set to the first activity of `course.yaml` (the schema guarantees at least one); on an existing file it does nothing, so a second Enroll cannot reset progress.
- `Resolve(order []string, p)` is a pure function: if `current` is in `order` it stays; otherwise it becomes the activity after the last completed one in the new order, the first activity if nothing completed is left, and the last activity if the last completed one is the last. Callers save only when it changed.
- Writes use `platform.WriteFileAtomic`; a newer `version` is read-only, as for library.json.
- `library` owns the progress store (created in `library.Open` for `<data>/state`) and uses its per-key lock for enroll and resolve, because enrollment, the course copy and the manifest are needed together.
- *Alternative: an `enrolled` flag in library.json:* rejected earlier; it can disagree with progress after removing a course with "Keep progress" and adding it again.

### Settings: `internal/settings`

```
<data>/settings.json
{ "version": 1, "lastCourse": "go-basics-3f2a9c1e" }
```

- App-level state and, in later changes, learner settings: UI language, external editor, learner folder root. `lastCourse` is written by the app, the rest will be written by the learner, so every write is a read-modify-write under one mutex.
- Errors reading or writing it are logged and never shown: without it the app opens the library, which is always a valid start.
- `StartCourse()` returns `lastCourse` if that course is in the library and enrolled, else `""`; the route `/` redirects to `/courses/<key>` (which runs the entering guard) or `/library`.
- *Alternative: `lastCourse` in library.json:* app state mixed into the course list, rewritten on every course entry. *`localStorage`:* lives in the WebView2 profile, not the data folder.

### Bound API by storage key
`library.entry(source)` becomes a lookup by key; `CourseOverview(key)`, `SyncCourse(key)` take the key. New bound methods: `EnrollCourse(key)`, `EnterCourse(key)`, `CourseTree(key)` (course title, current activity, sections with their activities' slugs and titles from `course.yaml`), `StartCourse()`. `ListCourses` and `CourseOverview` return an `enrolled` flag; it is computed from the progress store and is not a field of the persisted `Entry`, so library.json is unchanged. All results keep the `*Result` with `*apperr.Error` convention.

New error codes: `notEnrolled`, `progressReadFailed`, `progressWriteFailed`, with `en` and `ru` messages.

### Frontend layout

```
frontend/src/
  router.ts                  routes, entering guard, back input
  App.vue                    app header + <RouterView>
  library/LibraryView.vue    /library; Add pushes /library/<key>
  overview/CourseOverview.vue /library/:key; Enroll / Continue push /courses/<key>
  course/CourseScreen.vue    header (Library, enrolled dropdown, Overview), CourseTree, <RouterView>
  course/CourseTree.vue      sections and activity links (RouterLink, push)
  course/ActivityView.vue    activity title placeholder
```

Components get `key` and `slug` as props (`props: true`), not from `useRoute()`, so they stay testable without a router. `CourseScreen` is keyed by `:key` so switching courses reloads its tree.

## Risks / Trade-offs

- [WebView2 also handles mouse back or Alt+Left, so Back runs twice] → `preventDefault` on both events; verified in the built exe as part of the tasks.
- [`EnterCourse` runs on every course switch and on startup, reading `course.yaml` and the progress file] → both are small local files; acceptable.
- [Activities are not locked, so the learner can open any activity] → temporary, listed in the proposal; locking only needs a check in `ActivityView` and the tree.
- [`check` is ignored, so a course that needs a tool can be entered without it] → temporary until the environment check change; the guard has a marked step for it.
- [A `lastCourse` that fails to save makes the next start open the library] → harmless and logged.

## Migration Plan

No migration: `state/` and `settings.json` are new, and library.json keeps its format.
