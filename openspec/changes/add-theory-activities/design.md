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
- The "new" badge: it depends on version switching, which does not exist yet.

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
Order is the activity slugs of `course.yaml` in course order (`order(c)`, already used by Enroll). An activity is available exactly when its index is at most the index of `current`; a slug not in the order is unavailable. Completed state comes from `progress.completed`; the course is completed when the last slug of the order is in `completed`.

- `Activity(key, slug)` reads `course.yaml`, the progress and `activity.yaml` under the course copy lock, so a Sync cannot replace the copy halfway. It returns `activityUnavailable` as an error. A broken activity is not an error of the call: like `Overview`, the result carries the activity error next to what the UI still needs.

  ```
  Activity { slug, title, dir ("<section>/<slug>"), type, theoryBlocks [{kind, content}],
             last, completed, courseCompleted, error }
  ```

- `CompleteActivity(key, slug)` in one `progress.Update`: the slug must be available; it is added to `completed` if absent; if it is `current` and not last, `current` becomes the following slug. It returns the slug to open next (`""` on the last activity) and the updated `Tree`, so the UI needs no second call.
- `Tree` gains `completed` and `locked` per activity and `completed` for the course.
- Next never reads `activity.yaml`, so it works on a broken activity.
- *Alternative: lock check only in the UI:* the archived environment check set the rule that Go enforces gates; a stale tree or a history entry could otherwise open a locked activity.

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

None: `state/<key>.json` keeps its format and only starts receiving `completed` slugs; `library.json` and `settings.json` are unchanged.
