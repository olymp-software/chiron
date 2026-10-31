# Design

## Context

`internal/manifest` parses `course.yaml`: it pre-checks `version`, validates against the embedded `schemas/course/v1.json` (santhosh-tekuri/jsonschema), maps schema errors to `apperr` problems with YAML paths (`schemaKeywordCodes`, with a test that every keyword the schemas use is mapped), then runs semantic checks (`checkPath`). `internal/progress` stores `{current, completed}` per course and has `Update` under a per-key lock and `Resolve`; nothing adds to `completed` yet. `library.Tree` returns sections, activity titles and the current slug; `ActivityView` shows the title and replaces an unknown slug with the current activity. `MarkdownView` and `courseImages(key, baseDir)` render the overview file. See proposal.md for scope.

## Goals / Non-Goals

**Goals:**
- One parser for the whole activity schema version 1, so quiz, controls and coding changes add UI and behavior, not schema.
- Go decides whether an activity is available and what Next does; the UI only reflects it.

**Non-Goals:**
- Showing controls, quiz questions or coding activity parts; they are parsed into typed structs but not returned to the UI.
- Caching parsed activities: every open reads `activity.yaml` again, which is what "content changes apply immediately" needs.

## Decisions

### `schemas/activity/v1.json`

One schema for activity schema version 1, embedded next to the course schema (`schemas.ActivityV1`) and compiled the same way.

```
version        const 1
type           enum [theory, quiz, coding]
theory_blocks  array, minItems 1, items: {kind: const markdown, data: {content: string, minLength 1}}
controls       array of {label (minLength 1), script (minLength 1), concurrency (minLength 1)}
data           per type, via allOf of if/then on type:
  theory -> data: false
  quiz   -> required data: {questions: minItems 1 of {kind: const choice, prompt, data: {multiple: bool,
            options: minItems 2 of {text (minLength 1), correct: bool, feedback}}, explanation}}
  coding -> required data: {workspace: minItems 1 of {path (minLength 1)}}
```

All objects have `additionalProperties: false`. No `pattern` (it is mapped to `invalidSlug`). `enum` is added to `schemaKeywordCodes` as `invalidValue`. A `false` subschema (`data` on a theory activity) is reported as `unknownField` for that field, which is what an author needs to read.
- *Alternative: `oneOf` over three activity shapes:* the validator reports every branch's failures, so one typo produces errors for all three types.
- *Alternative: validate only the theory parts now:* the activity schema version would change when quiz or coding land, and courses written against v1 would break.

### Pre-checks before the schema
As for `course.yaml`, `ParseActivity` decodes the YAML into `any` first and, before schema validation, reports "requires a newer app version" for `version` above 1, a string `type` that is not `theory`/`quiz`/`coding`, and a string theory block `kind` other than `markdown`. Values of the wrong type are left to the schema (`wrongType`).

### Semantic checks
After the schema, on the typed struct, with paths relative to the activity folder:
- control `script`: `checkPath` against the activity folder (relative, `/`, no `..`, an existing regular file);
- workspace `path`: the same rules, but the target may be a file or a folder;
- `choice` options: exactly one `correct` with `multiple: false`, at least one with `multiple: true`.

`checkPath` gets the base folder as a parameter; the course `check` keeps passing the course root.

### Error codes
Whole-file codes become file-agnostic with a `file` param holding the course-relative path (`course.yaml`, `basics/hello/activity.yaml`): `manifestInvalid`, `yamlSyntax`, `unsupportedSchemaVersion`; their `en`/`ru` messages interpolate `{file}`. New codes:

| code | params | when |
|---|---|---|
| `activityFileMissing` | `file` | the activity folder or its `activity.yaml` does not exist |
| `unsupportedActivityType` | `file`, `type` | pre-check |
| `unsupportedBlockKind` | `file`, `kind` | pre-check |
| `activityUnavailable` | `slug` | the slug is not in `course.yaml` or is after the current activity |
| `oneCorrectOptionRequired` | - | problem: `multiple: false` without exactly one correct option |
| `correctOptionRequired` | - | problem: `multiple: true` without a correct option |

`pathNotFound` changes from "file ... does not exist" to "“{path}” does not exist in the course", because a workspace path may be a folder. `ErrorView` shows the releases link for all three "newer app version" codes.
- *Alternative: separate `activityInvalid`, `activityYamlSyntax`, ... codes:* the same three messages twice, differing only in the file name.

### Availability and completion in `library`
Order is the activity slugs of `course.yaml` in course order (`order(c)`, already used by Enroll). An activity is available exactly when it is completed or is the current activity (after `Resolve`, see below); a slug not in the order is unavailable. Completed state comes from `progress.completed`; the course is completed when every slug of the order is in `completed`.

- `Activity(key, slug)` reads `course.yaml`, the progress and `activity.yaml` under the course copy lock, so a Sync cannot replace the copy halfway. It returns `activityUnavailable` as an error. A broken activity is not an error of the call: like `Overview`, the result carries the activity error next to what the UI still needs.

  ```
  Activity { slug, title, dir ("<section>/<slug>"), type, theoryBlocks [{kind, content}],
             last, completed, courseCompleted, error }
  ```

- `CompleteActivity(key, slug)` in one `progress.Update`: the slug must be available; it is added to `completed` if absent, and `Resolve` moves `current` to the first activity that is not completed. It returns the slug to open next (`""` on the last activity) and the updated `Tree`, so the UI needs no second call.
- `Tree` gains `completed` and `locked` per activity and `completed` for the course.
- Next never reads `activity.yaml`, so it works on a broken activity.
- *Alternative: lock check only in the UI:* the archived environment check set the rule that Go enforces gates; a stale tree or a history entry could otherwise open a locked activity.

### Completed course in the library
The library list reads no course copies (it shows stored titles), but completion depends on the last activity of `course.yaml`. So the progress stores it: `Progress` gains `courseCompleted` (`"courseCompleted": true` in `state/<key>.json`), and `progress.Resolve(order, p)`, which already reconciles the current activity with the order, also sets `courseCompleted`. Every place that has `course.yaml` and the progress reconciles them: `Enter`, `CompleteActivity`, and `Sync` when the course is enrolled and the new `course.yaml` is valid (an invalid one keeps the stored progress). `Courses()` loads the progress instead of only checking it exists and returns `completed` next to `enrolled`; `LibraryView` shows Completed instead of Enrolled. `Tree` and `Activity` keep computing completion from `course.yaml`, which they read anyway.
- *Alternative: parse `course.yaml` of every enrolled course in the list:* always current, but the list starts reading course copies and a broken `course.yaml` hides completion. A stored flag also keeps working if progression stops being linear.
- The format version stays 1: the field is additive and a file without it reads as not completed until the next reconcile.
- `Resolve` derives the pointer from `completed`: `current` becomes the first slug of the order that is not completed, or the last slug when all are, and `courseCompleted` becomes whether every slug is completed. In linear progress `completed` is a prefix of the order, so this is the activity Next reached; it also covers a removed current activity and activities a Sync or version switch adds anywhere in the course: the first added activity becomes current and the course stops being completed. The stored `current` is kept in the file for the format, but never decides anything.
- `Overview` gains `completed` from the stored flag. The overview shows a "Course completed" status and Review instead of Continue; Review does what Continue does. After Sync the overview reloads, so both follow the reconciled progress.

### Bound API
`App.Activity(key, slug) ActivityResult` and `App.CompleteActivity(key, slug) CompleteResult{next, tree, error}`, following the `*Result` convention. Neither checks the environment check again: like `CourseTree`, they are reached only after `EnterCourse`, which enforces it.

### Frontend
- `ActivityView` loads the activity whenever `slug` changes. `activityUnavailable` replaces the route with the current activity, as an unknown slug does today. It renders each `markdown` block with `MarkdownView` in order, or `ErrorView` for an activity error, then a footer: Next, Finish course, or a "Course completed" notice when the course is completed and this is the last activity.
- Next / Finish call `CompleteActivity`, emit the returned tree to `CourseScreen` (it owns the tree and passes it to the tree and the `RouterView`), and `router.push` the next activity, since Next is a move the Learner starts.
- `CourseTree` renders available activities as `RouterLink` and locked ones as plain text with `aria-disabled`; completed ones get a check mark, the current one a marker.
- Images: `activityImages(key, dir)` in `markdown/images.ts`, built on the same resolution as `courseImages` but returning `null` when `..` would leave `dir`. The overview keeps `courseImages`.

## Risks / Trade-offs

- [The quiz and coding parts of the schema are written before their UI exists; using them may show the schema is wrong] → they follow the baseline `quiz` and `activity-content` specs field by field, and the app has no release yet, so v1 can still be corrected before the first one.
- [Every activity open parses `course.yaml`, the progress file and `activity.yaml`] → small local files, as with `EnterCourse`; caching would break immediate content changes.
- [Changing `manifestInvalid`/`yamlSyntax`/`unsupportedSchemaVersion` messages to take `{file}`] → every place that creates them passes `file`, and the manifest tests assert the `file` param, since no test checks message params against error params.

## Migration Plan

None: `state/<key>.json` keeps format version 1, starts receiving `completed` slugs and gains the additive `courseCompleted` flag; `library.json` and `settings.json` are unchanged.
