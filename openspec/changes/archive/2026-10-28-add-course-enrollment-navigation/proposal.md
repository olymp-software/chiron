# Proposal

## Why

The library and the course overview exist, but the learner cannot take a course: there is no Enroll, no course screen and no way to move between the library, the course overview and a course. Enrollment and navigation are the frame every later feature (activity content, controls, quizzes, progress) is rendered in, so they come next, before any activity content.

## What Changes

- Enroll and Continue in the course overview. Enroll creates the course's progress with the first activity as the current activity; a course is enrolled exactly when it has progress. The library marks enrolled courses.
- The course screen: a header with a Library button, a dropdown of enrolled courses and an Overview button; a course tree listing sections and their activities from `course.yaml`; and the open activity, shown in this change only by its title. Clicking an activity in the tree opens it.
- Entering a course opens its current activity; if the current activity no longer exists after a Sync, the pointer moves as the `progress` spec defines. Entering a course whose `course.yaml` is invalid opens its course overview instead.
- On startup the app enters the course entered last, or opens the library.
- Back navigation: the app keeps a history of the screens and activities the learner opened, and Back returns to the previous one.
- Screens are addressed by routes with the storage key: the library, a course overview, a course and an activity of a course. Bound Go methods take the storage key instead of the course source.
- App-level state is stored in a new `settings.json` in the data folder; in this change it holds only the course entered last.
- Not in this change: the environment check (a declared `check` is ignored, so Enroll and Continue are available whenever `course.yaml` is valid), locked activities (every activity in the tree can be opened), Next and Finish course, activity content (`activity.yaml`, theory blocks), and settings UI.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `course-library`: the course screen gets Library and Overview buttons; Back returns to the previously opened screen; entering a course with an invalid `course.yaml` opens its course overview; startup falls back to the library when the course entered last is no longer enrolled.

## Impact

- Code: new `internal/progress` (`state/<key>.json`: current activity and completed activities) and `internal/settings` (`settings.json`); `internal/library` looks courses up by storage key, reports enrollment, enrolls and computes the course tree; `app.go` binds `EnrollCourse`, `EnterCourse`, `CourseTree`, `StartCourse` (the startup target) and switches `CourseOverview` and `SyncCourse` to the storage key.
- Frontend: `vue-router` with hash history replaces the view switch in `App.vue`; new course screen components; `LibraryView` and `CourseOverview` navigate through the router; new i18n strings in `en.json` and `ru.json`.
- New dependency: `vue-router`.
- Data folder: new `settings.json` and `state/<key>.json` files; `library.json` is unchanged.
