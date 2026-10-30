# Tasks

## 1. Error codes

- [ ] 1.1 Add `file` param to `manifestInvalid`, `yamlSyntax`, `unsupportedSchemaVersion` everywhere they are created (course.yaml passes `course.yaml`) and update their `en`/`ru` messages to `{file}`; change the `pathNotFound` message to "“{path}” does not exist in the course"; verify with `task go:test -- -run TestParseCourse` asserting the `file` param
- [ ] 1.2 Add codes `activityFileMissing`, `unsupportedActivityType`, `unsupportedBlockKind`, `activityUnavailable`, `oneCorrectOptionRequired`, `correctOptionRequired` to `internal/apperr` (constants and `All`), run `go generate ./internal/apperr`, add `en`/`ru` messages; verify with `task go:test -- -run 'TestAllListsEveryCode|TestTypeScriptUpToDate'` and `task frontend:test -- src/errors/messages.test.ts`
- [ ] 1.3 Show the releases link in `ErrorView` for `unsupportedActivityType` and `unsupportedBlockKind` too; verify with a component test or `task frontend:check`

## 2. Activity schema and parser

- [ ] 2.1 Write `schemas/activity/v1.json` as in design.md (`$id` under the same base as the course schema) and embed it as `schemas.ActivityV1`; add `enum` to `schemaKeywordCodes` and report a `false` subschema as `unknownField`; verify `TestEverySchemaKeywordHasCode` covers both schemas and compile/`$id` tests exist for the activity schema like `TestCourseSchemaCompiles`/`TestCourseSchemaID`
- [ ] 2.2 Add `manifest.ParseActivity(courseRoot, section, slug)` with typed structs for theory blocks, controls, quiz and coding data: missing folder or file → `activityFileMissing`, YAML syntax, pre-checks for `version`/`type`/`kind`, schema validation with problems; verify with table tests for each activity-content and quiz error scenario (content instead of theory_blocks, empty/missing theory_blocks, block without data, empty markdown content, theory with data, quiz without data, workspace item as string, empty workspace, no questions, single option, version 2, `type: terminal`, `kind: diagram`) and a valid file of each type
- [ ] 2.3 Add semantic checks: `checkPath` takes the base folder; control `script` and workspace `path` checked against the activity folder (workspace may be a folder); correct option count for `multiple: false`/`true`; verify with tests for `../shared/test.sh`, a missing workspace folder, two correct radio options, one correct checkbox option (valid), and the unchanged course `check` tests

## 3. Availability and completion in library

- [ ] 3.1 Extend `Tree` with per-activity `completed`/`locked` and course `completed`; verify with `library` tests for a fresh enrollment (only the first available) and the "Marks in the tree" scenario
- [ ] 3.2 Add `Library.Activity(key, slug)` reading course.yaml, progress and activity.yaml under the course copy lock, returning `activityUnavailable` for a locked or unknown slug and carrying a broken activity's error in the result with `last`/`completed`/`courseCompleted`; verify with tests for a valid theory activity, a missing `activity.yaml` (tree still loads), a locked slug and a fixed-by-Sync activity
- [ ] 3.3 Add `Library.CompleteActivity(key, slug)` in one `progress.Update` returning the next slug and the updated tree; verify with tests for Next on the current activity, Next on an earlier activity (current unchanged), Next on a broken activity, Finish course (pointer stays, course completed) and a locked slug (`activityUnavailable`, progress unchanged)
- [ ] 3.4 Bind `App.Activity` and `App.CompleteActivity` with `ActivityResult` / `CompleteResult`; verify `task go:lint` and `task go:lint:internal` pass and `wails build` regenerates the bindings

## 4. Frontend

- [ ] 4.1 Add `activityImages(key, dir)` to `markdown/images.ts`; verify with unit tests: relative image inside the activity folder resolves, `../../images/logo.png` and `https://` give null
- [ ] 4.2 Rewrite `ActivityView`: load the activity on slug change, replace the route with the current activity on `activityUnavailable`, render markdown theory blocks in order or the activity error, and the footer with Next / Finish course / "Course completed"; add i18n strings in `en.json` and `ru.json`; verify with component tests for blocks in order, the error with Next, and the completed notice
- [ ] 4.3 Wire Next / Finish: call `CompleteActivity`, pass the returned tree up to `CourseScreen`, `router.push` the next activity; verify with a component test that Next pushes the next route and updates the tree
- [ ] 4.4 Update `CourseTree`: links for available activities, plain `aria-disabled` text for locked ones, check mark for completed, marker for current; verify with a component test of the "Marks in the tree" scenario and that a locked item is not a link
- [ ] 4.5 Run `task check` and verify it passes

## 5. Integration

- [ ] 5.1 Build and run the app (`task run`) with a local-folder test course of two sections with theory activities (one with an image, one with a broken `activity.yaml`): verify locked activities in the tree, Next across sections, Next on the broken activity, Finish course, the completed notice after restart, and the image outside the activity folder not loading
