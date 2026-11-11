# Chiron Author Guide

This guide is for Authors: people who write courses and workshops for Chiron. It describes what the app supports **today**. Features that are specified but not yet in the app are listed separately in [Not yet available](#not-yet-available), so nothing in the main text is a promise.

The guide is kept in sync with the app: every change that affects what Authors write or what their scripts can rely on updates it. The normative requirements are in `openspec/specs/`; when this guide and the specs disagree, the specs win and the guide has a bug.

Contents:

1. [Quick start](#quick-start)
2. [Courses and workshops](#courses-and-workshops)
3. [Repository layout and file rules](#repository-layout-and-file-rules)
4. [The manifest: course.yaml and workshop.yaml](#the-manifest-courseyaml-and-workshopyaml)
5. [The overview file](#the-overview-file)
6. [Course activities](#course-activities)
7. [Theory blocks and Markdown](#theory-blocks-and-markdown)
8. [Quizzes](#quizzes)
9. [Workshops](#workshops)
10. [Scripts](#scripts)
11. [The environment check](#the-environment-check)
12. [Progress and changing a published course](#progress-and-changing-a-published-course)
13. [Editor support: JSON Schemas](#editor-support-json-schemas)
14. [Errors](#errors)
15. [Not yet available](#not-yet-available)

## Quick start

A minimal course is two files:

```
my-course/
  course.yaml
  basics/
    hello/
      activity.yaml
```

```yaml
# course.yaml
version: 1
title: My First Course
sections:
  - slug: basics
    title: Basics
    activities:
      - slug: hello
        title: Hello
```

```yaml
# basics/hello/activity.yaml
version: 1
type: theory
theory_blocks:
  - kind: markdown
    data:
      content: |
        # Hello
        This is the first activity.
```

To try it, open Chiron, press **Add from folder** in the library and pick `my-course/`. The course overview opens; press **Enroll**.

The authoring loop:

1. Edit files in your folder.
2. In the app, open the course overview (the **Overview** button on the course screen) and press **Sync**.
3. Reopen the activity.

The app never watches your folder: it works from its own copy, refreshed only by Sync. Sync includes uncommitted edits; the folder doesn't have to be a git repository.

## Courses and workshops

Chiron has two content types. The root file decides which one a repository is:

| | Course | Workshop |
|---|---|---|
| Root file | `course.yaml` | `workshop.yaml` |
| Activities | `theory`, `quiz`, `coding` | one form: theory blocks plus optional solution, Check, controls and files |
| Where the Learner works | (see [Not yet available](#not-yet-available)) | one **workshop folder** shared by all activities |
| Typical use | learn a topic step by step, test knowledge | build up one state across activities: files, a local cluster, a cloud account |

A repository with both files, or neither, is a course error. Everything below that's written for courses also applies to workshops unless [Workshops](#workshops) says otherwise. In the library, courses and workshops are listed in separate sections.

## Repository layout and file rules

```
course.yaml               # or workshop.yaml
OVERVIEW.md               # optional, shown in the course overview
README.md                 # optional, ignored by the app (for GitHub)
scripts/check.sh          # optional, environment check (any path)
<section-slug>/
  <activity-slug>/
    activity.yaml
    ...                   # images, scripts, files for the Learner
```

- There are exactly two levels: sections and activities. No deeper nesting.
- Folder names are slugs and carry no ordering numbers. Order comes only from the manifest.
- A folder that the manifest doesn't reference is ignored.

**Slugs** match `^[a-z0-9]+(-[a-z0-9]+)*$` (lowercase letters and digits, single hyphens) and equal their folder name. Section slugs are unique among sections; activity slugs are unique **across the whole course**, not just within a section.

**Common rules** for every file:

- Text files are UTF-8.
- YAML is parsed strictly: an unknown field (e.g. a typo like `titel`) or a value of the wrong type is an error.
- Every path in YAML is relative, uses `/`, never starts with `/` and never contains `..`. The target must exist.

**What the app copies.** When a course is added or synced, the app copies your folder **except** `.git` and files excluded by `.gitignore` files inside the folder (matched as git matches them; your global git settings and `.gitignore` files above the folder don't apply). Don't gitignore anything the course needs.

## The manifest: course.yaml and workshop.yaml

The manifest is the only description of sections, activities and their order. Activity titles come from here, not from activity files.

```yaml
# yaml-language-server: $schema=<schema-base>/course/v1.json
version: 1
title: Go Basics
description: Go fundamentals in 20 activities
check: scripts/check.sh
sections:
  - slug: basics
    title: First steps
    description: Variables, types and the toolchain
    activities:
      - slug: hello
        title: Hello, Go
      - slug: variables
        title: Variables
```

| Field | Required | Meaning |
|---|---|---|
| `version` | yes | Schema version of this file. Currently `1`. |
| `title` | yes | Course title, shown in the library and the course overview. |
| `description` | no | Plain text, shown in the course overview when there's no overview file. |
| `check` | no | Path from the course root to the [environment check](#the-environment-check) script. |
| `sections` | yes | At least one section. |
| `sections[].slug`, `.title` | yes | |
| `sections[].description` | no | Plain text. |
| `sections[].activities` | yes | At least one activity reference: `slug` and `title`, both required. |

`workshop.yaml` has the same fields plus optional `controls`, see [Workshop controls](#workshop-controls).

Progression runs through all sections in order: after the last activity of a section comes the first activity of the next.

The app reads only the manifest to build the course tree. An activity folder is read when the Learner opens that activity, so a broken `activity.yaml` breaks only that activity.

## The overview file

The course overview shows, in order of preference:

1. `OVERVIEW.md` in the course root (name matched case-insensitively, so `overview.md` works too), rendered as Markdown;
2. otherwise the manifest's `description` as plain text;
3. otherwise nothing.

Two matching files (say, `OVERVIEW.md` and `overview.md`) are a course error. `README.md` is never shown, so you can keep a GitHub-facing README separate from the in-app overview.

The overview file uses the same [Markdown dialect](#theory-blocks-and-markdown) as activities. Its images may come from anywhere inside the course, relative to the course root.

## Course activities

Every `activity.yaml` of a course has:

| Field | Required | Meaning |
|---|---|---|
| `version` | yes | Schema version of this file. Currently `1`. Independent of the manifest's version. |
| `type` | yes | `theory`, `quiz` or `coding`. |
| `theory_blocks` | yes | At least one [theory block](#theory-blocks-and-markdown). |
| `data` | depends on `type` | Type-specific fields: forbidden for `theory`, required for `quiz` and `coding`. |
| `controls` | no | Validated, but not yet run in courses; see [Not yet available](#not-yet-available). |

The type defines only what the Learner does in the window: read, answer questions or work with files. It doesn't define the topic, the language or how work is checked.

### theory

Theory blocks, nothing else. `data` must be absent.

### quiz

Theory blocks on a theory screen, and questions on a separate quiz screen. See [Quizzes](#quizzes).

### coding

```yaml
version: 1
type: coding
theory_blocks:
  - kind: markdown
    data:
      content: Implement `Add` in `main.go`.
data:
  workspace:
    - path: main.go
    - path: go.mod
```

`data.workspace` is a list of at least one item, each an object with `path` (a file or folder relative to the activity folder). Write `- path: main.go`, not `- main.go`.

The app validates coding activities today but shows only their theory blocks: copying workspace files to the Learner, controls, Reset and the external editor aren't implemented yet. Use a [workshop](#workshops) for hands-on work until then.

## Theory blocks and Markdown

`theory_blocks` is a list shown in order at the top of every activity:

```yaml
theory_blocks:
  - kind: markdown
    data:
      content: |
        # Pods
        A Pod is the smallest deployable unit.
  - kind: markdown
    data:
      content: |
        > [!TASK]
        > Create a Pod named `web`.
```

The only block kind is `markdown`; its `data.content` is required and non-empty. Several blocks are allowed but render the same as one longer block.

### Markdown dialect

All Markdown (theory blocks, the overview file, quiz prompts, options, feedback and explanations, workshop solutions) is CommonMark with GitHub extensions:

- tables, task lists, strikethrough;
- bare URLs become links;
- fenced code blocks with syntax highlighting by language tag (TextMate grammars as used by [Shiki](https://shiki.style/languages), light and dark themes). An unknown or missing tag is shown as plain text.

**Callouts**, GitHub style, case-insensitive:

```markdown
> [!NOTE]
> Useful background.

> [!TIP]
> A shortcut.

> [!WARNING]
> Something that can go wrong.

> [!TASK]
> What the Learner should do now.
```

`TASK` is a Chiron extension; GitHub shows it as a plain quote. Callout titles are shown in the Learner's UI language.

**HTML isn't rendered.** Tags are shown as literal text. The UI can run scripts, so HTML from a third-party course would amount to code execution.

**Links.** `http://`, `https://` and `mailto:` links open in the system browser. Other links don't navigate anywhere.

**Images** must be local files, so the course works offline:

- In an activity, the path is relative to the activity folder and must stay inside it: `![diagram](images/diagram.png)` works, `![logo](../../images/logo.png)` doesn't. Put shared images in each activity that uses them.
- In the overview file, the path is relative to the course root and must stay inside the course.
- Remote images (`https://…`) aren't loaded; their alt text is shown instead.

Course content isn't translated: write in the language of your audience. The app UI is in English or Russian.

## Quizzes

```yaml
version: 1
type: quiz
theory_blocks:
  - kind: markdown
    data:
      content: Kubernetes objects are either namespaced or cluster-scoped.
data:
  questions:
    - kind: choice
      prompt: Which Kubernetes objects are namespaced?
      data:
        multiple: true
        options:
          - text: "`Pod`"
            correct: true
          - text: "`Node`"
            feedback: Nodes are cluster-scoped.
          - text: "`Service`"
            correct: true
      explanation: Namespaced objects live inside a namespace.
```

**Question** fields:

| Field | Required | Meaning |
|---|---|---|
| `kind` | yes | Only `choice` exists today. |
| `prompt` | no | Markdown above the options. Without it only the options are shown. |
| `data` | yes | `multiple` and `options`. |
| `explanation` | no | Markdown shown after the Learner checks the question. |

**`choice` data:**

- `multiple` (default `false`): `false` shows radio buttons and needs exactly one correct option; `true` shows checkboxes and needs at least one. A checkbox question with a single correct option is valid on purpose, so that the input type doesn't hint at the answer.
- `options`: at least two, always shown in the order written (no shuffling: vary the position of correct options yourself).

**Option** fields: `text` (Markdown, required), `correct` (default `false`), `feedback` (Markdown, optional; shown only if the Learner selected that option).

**How the Learner takes a quiz.** The activity opens on the theory screen with a **Start quiz** button. The quiz screen shows one question at a time with **Check**, **Previous** / **Next question**, a question map and a correct-answer counter, and a **Theory** button to go back to the theory blocks. After **Check**, the question locks: every correct option is highlighted, selected options get a correct or incorrect mark, and the feedback of selected options and the explanation appear. An answer is correct only if the selected set equals the correct set: there is no partial credit. **Start over** clears everything.

**Quiz results don't count.** They don't affect progress, aren't saved, and the quiz starts fresh every time the activity opens. **Next** completes the activity regardless.

**Answers aren't protected.** `correct: true` is plain text in the course; a Learner can read it.

**YAML tips:**

- Quote option text that starts with a backtick, `*`, `&`, `!`, `[`, `{`, `>`, `|`, `@`, `%` or `-` followed by a space, or contains `: `: `text: "`Pod`"`.
- Quote `"True"`, `"False"`, `"Yes"`, `"No"`, `"On"`, `"Off"` and numbers you want as text; for a true/false question use a two-option `choice`.

## Workshops

A workshop is for activities that change one shared state that the app doesn't own: files in a folder, resources in a local cluster or a cloud account. Its root file is `workshop.yaml`; everything about the manifest, overview file, environment check, Sync and progress works as for a course.

### Workshop activity.yaml

```yaml
# yaml-language-server: $schema=<schema-base>/workshop-activity/v1.json
version: 1
theory_blocks:
  - kind: markdown
    data:
      content: |
        > [!TASK]
        > Start the server and make `/health` return `ok`.
files:
  - path: starter/main.go
  - path: manifests
check: check.sh
controls:
  - label: Start server
    script: server.sh
    concurrency: server
  - label: Show logs
    script: logs.sh
solution: |
  Run `go run ./starter` and add a handler:
  ```go
  http.HandleFunc("/health", ...)
  ```
```

| Field | Required | Meaning |
|---|---|---|
| `version` | yes | Schema version of this file. Currently `1`. |
| `theory_blocks` | yes | As in a course. |
| `files` | no | Files and folders copied to the workshop folder; items are `path:` objects relative to the activity folder. |
| `check` | no | Script for the built-in **Check** button, relative to the activity folder. |
| `controls` | no | Script buttons; see [Controls](#controls). |
| `solution` | no | Non-empty Markdown, hidden behind a **Solution** button. |

A workshop activity has no `type` and no `data`; both are errors.

### The workshop folder

Each workshop has one workshop folder, `~/Chiron/<storage key>`, shared by all its activities. The app creates it when an activity opens (also after the Learner deleted it). The workshop screen shows its absolute path with a copy button and an **Open in external editor** button (for now it opens the folder in the system file manager), so Learners can work there from a terminal or an editor.

Scripts get its path as `WORKSPACE_DIR`.

### files

Every time an activity opens, the app copies each file of its `files` items **that is missing** from the workshop folder, keeping relative paths. A listed folder is merged file by file. Existing files are never touched, so the Learner's edits survive reopening, and several activities can contribute files to one folder.

Files are copied byte for byte (no line-ending or encoding conversion) and keep the executable bit where the OS has one. A file appears only once it is fully copied. If a listed file's path is a folder in the workshop folder, the activity shows an error.

There's no Reset in a workshop. If a Learner needs to restore or clean up state, give them a control or an activity for it.

### Solution

`solution` is hidden until the Learner presses **Solution**, and hidden again when the activity is reopened. Showing it doesn't affect progress. Its images follow the activity image rules.

### Check

`check` adds a **Check** button with a fixed, translated label. It runs the script in a modal: output, **Cancel**, the exit code with a success or failure mark, then **OK**. Nothing about the run is kept after the modal closes, and the result gates nothing: **Next** always works, with or without a passing Check.

Check doesn't belong to any exclusion group and doesn't stop running controls, so it can test a server that a control started.

### Controls

```yaml
controls:
  - label: Start server
    script: server.sh
    concurrency: server   # runs alongside the others
  - label: Prepare
    script: setup.sh      # these two exclude each other
  - label: Check API
    script: check-api.sh
```

A control has `label` (button text), `script` (relative to the activity folder, must stay inside it) and optional `concurrency`, its exclusion group. A control always runs a file; there are no inline commands.

- The same control never runs twice at once.
- Controls with the same `concurrency` value exclude each other; while one runs, the others are disabled. Controls without `concurrency` share the activity's default group. Groups apply within one activity.
- Each running control has a **Stop** button.
- Each control has its own output, which the Learner can switch between. It's kept until that control starts again or the Learner leaves the activity.
- Leaving the activity (**Next**, the course tree, Back) while a control runs asks for confirmation and then stops it.

### Workshop controls

`workshop.yaml` may declare `controls` too, each with `label` and `script` (a path from the workshop root). They appear in the course overview once the Learner has enrolled, run in a modal like Check, and don't require the environment check to pass. Use them for things like "Check leftovers" before the Learner removes the workshop, or "Tear down the cluster".

```yaml
# workshop.yaml
version: 1
title: Kubernetes from Scratch
check: scripts/check.sh
controls:
  - label: Delete the cluster
    script: scripts/teardown.sh
sections:
  - …
```

A workshop control runs with the workshop root as working directory and gets `WORKSPACE_DIR`, `CONTENT_DIR` and `TEMP_DIR`, but not `ACTIVITY_DIR`.

## Scripts

Scripts are environment checks, workshop activity Checks, activity controls and workshop controls.

### Interpreter

Scripts run in a shell interpreter embedded in the app ([mvdan.cc/sh](https://github.com/mvdan/sh)) that behaves like non-interactive bash. No system bash is needed, including on Windows without Git Bash or WSL, and the same script behaves the same on Windows, Linux and macOS. Don't add a shebang expecting it to matter; the file is always run by the embedded interpreter.

Scripts run with the Learner's permissions and without a sandbox, like `make` in someone else's repository. The Learner accepts this once, when adding the course.

### Guaranteed commands

On every OS a script can rely on:

- bash builtins: `echo`, `printf`, `cd`, `test` / `[`, `export`, `exit`, `read`, `set`, …;
- `cat`, `cp`, `mkdir`, `mv`, `rm`, `touch`, provided by the app itself.

The app's utilities accept only these flags: `mkdir -p`, `cp -r`, `rm -r -f` (combinable, e.g. `rm -rf`) and `--`. Any other flag prints a message to stderr and exits with code 1 without doing anything.

Everything else (`ls`, `find`, `grep`, `sed`, `awk`, `head`, `diff`, `chmod`, `xargs`, `go`, `python3`, `docker`, `kubectl`, …) is looked up in the Learner's `PATH` and isn't guaranteed. On Windows most of these don't exist. Check for every tool you need in the [environment check](#the-environment-check).

### Environment

| | Environment check | Workshop activity: Check and controls | Workshop control |
|---|---|---|---|
| Working directory | course root (in the app's copy) | activity folder (in the app's copy) | workshop root (in the app's copy) |
| `HOST_OS` | `windows`, `linux` or `darwin` | yes | yes |
| `HOST_ARCH` | `amd64` or `arm64` | yes | yes |
| `ACTIVITY_DIR` | — | activity folder in the app's copy | — |
| `WORKSPACE_DIR` | — | the workshop folder | the workshop folder |
| `CONTENT_DIR` | — | root of the app's copy | root of the app's copy |
| `TEMP_DIR` | — | new empty folder for this run | new empty folder for this run |

Scripts also inherit the Learner's environment (`PATH`, `HOME`, …). They get no arguments; stdin is empty (reads get end of input immediately), and stdout and stderr are pipes, not a terminal.

- `CONTENT_DIR` lets activities share helpers: `. "$CONTENT_DIR/lib/server.sh"`.
- `TEMP_DIR` is outside the workshop folder, empty at the start of every run and deleted after it, also after Stop. Use it for build output and scratch files.

### Don't write to the course

The app's copy of the course is replaced on every Sync, so anything a script writes there is lost. Keep state in `$WORKSPACE_DIR`, scratch files in `$TEMP_DIR`. Watch out for implicit writes such as Python's `__pycache__` or `go build` output: `cd "$WORKSPACE_DIR"` (or `cd "$TEMP_DIR"`) first is the usual remedy.

### Output

stdout and stderr appear in the output panel as they arrive, decoded as UTF-8 (invalid bytes become `�`). ANSI colors and styles are rendered, `\r` returns to the start of the line (progress bars work), and other control sequences are stripped. The panel keeps the latest ~10,000 lines.

The app doesn't set `NO_COLOR` and doesn't force colors. Through a pipe most tools print without color; enable it yourself if you want it (`--color=always`, `FORCE_COLOR=1`).

Interpreters often buffer output when it isn't a terminal. A typical preamble for Python:

```sh
export PYTHONUNBUFFERED=1 PYTHONUTF8=1 PYTHONDONTWRITEBYTECODE=1
cd "$WORKSPACE_DIR"
python3 main.py
```

### Exit code

`0` is success (green mark); anything else is failure (red mark). Results never affect progress. There's no structured result yet: explain what passed and failed in the output.

### Stopping and cleanup

There's no timeout. **Stop**, **Cancel**, leaving the activity, Sync, removing the course and closing the app terminate the script **and every process it started** at once, with no signal the script can catch. When a script exits on its own, any background processes it left running are terminated too, so a server started in the background lives only as long as the script that started it.

Because a script gets no chance to clean up, clean up **at the start** of the next run:

```sh
docker rm -f chiron-web 2>/dev/null
docker run --name chiron-web -p 8080:80 nginx
```

Child processes don't open console windows on Windows.

### Portability tips

- Branch on `HOST_OS` / `HOST_ARCH` for platform-specific downloads or binary names (`app.exe` on Windows).
- Prefer the guaranteed utilities over `ls`, `grep` and friends, or check for those tools in the environment check.
- Quote paths: `"$WORKSPACE_DIR"` may contain spaces.

## The environment check

The manifest's `check` points to a script that verifies the Learner's machine has everything the course needs.

```sh
# scripts/check.sh
kubectl version --client || {
  echo "kubectl not found. Install it: https://kubernetes.io/docs/tasks/tools/"
  exit 1
}
```

- The course overview shows **Check environment** and the check status: not checked, passed, or check changed. **Enroll** and **Continue** are unavailable until the check passes.
- The check runs in a modal with live output, **Cancel**, the exit code and **OK**. Its output is the Learner's only guidance, so say what is missing and where to get it.
- Exit code `0` is remembered together with the SHA-256 of the `check` file. When the file changes (after a Sync), the check is required again; progress is kept.
- **Only the `check` file itself is hashed.** If it calls helpers that change, change the `check` file too, e.g. bump a `# rev 3` comment, to make Learners rerun it.
- A failed run forgets a pass for the same file; a cancelled run changes nothing.
- It runs with the course root as working directory and gets only `HOST_OS`, `HOST_ARCH` and the inherited environment.

Without `check`, the Learner can enroll right away.

## Progress and changing a published course

- Progression is strictly sequential: the current activity is the first one not completed; later ones are locked.
- Only **Next** (or **Finish course** / **Finish workshop** on the last activity) completes an activity, whatever the Learner did in it: quiz answers and script results don't matter. **Next** works even on an activity that shows an error.
- Progress is keyed by **activity slug**:
  - moving an activity to another section keeps its progress;
  - renaming a slug creates a new, not completed activity, and the old progress no longer applies;
  - an activity added before the current one becomes current, and a completed course with a new activity is no longer completed;
  - removing the current activity moves the Learner to the first activity not completed.
- Content changes (texts, questions, scripts) are visible right after Sync.
- The library keeps the last valid `title` from the manifest; a broken manifest after Sync doesn't rename the course.

## Editor support: JSON Schemas

The app validates every file against JSON Schemas that you can use in your editor for completion and validation. They live in the app repository under `schemas/`:

| File | Schema |
|---|---|
| `course.yaml` | `schemas/course/v1.json` |
| `activity.yaml` of a course | `schemas/course-activity/v1.json` |
| `workshop.yaml` | `schemas/workshop/v1.json` |
| `activity.yaml` of a workshop | `schemas/workshop-activity/v1.json` |

With the YAML extension for VS Code (or any editor using yaml-language-server), add a modeline as the first line:

```yaml
# yaml-language-server: $schema=<schema-base>/course-activity/v1.json
```

The published URL (`<schema-base>`) isn't fixed yet; until it is, point the modeline at a local copy of the schema file, e.g. `$schema=../../.schemas/course-activity/v1.json`.

**Schema versions.** Every manifest and `activity.yaml` has its own `version`, unrelated to the course version and independent of each other: one course may mix activities of different schema versions. Any schema change bumps the version. Current versions: course 1, course activity 1, workshop 1, workshop activity 1.

If a file has a `version` newer than the app supports, or an unknown activity `type` or theory block `kind`, the app shows "requires a newer app version" with a link to the releases page instead of trying to parse it.

## Errors

- **Course error**: the manifest is missing, invalid, or the root has both or neither manifest, or two overview files. The course can't be added; after a Sync, the course overview shows the error and keeps Sync available, and Enroll and Continue are unavailable.
- **Activity error**: the activity folder or `activity.yaml` is missing, isn't valid YAML, breaks the schema or a semantic check (e.g. wrong number of correct options, a path that leaves the activity folder). It shows in place of the activity; the rest of the course works.

Errors list every problem found with the YAML path of the offending value, e.g. `data.questions[2].data.options`. Fix the file, Sync, and reopen the activity.

Parse errors are also written to the app log (**Open log folder**). Script output never goes to the log.

## Not yet available

These are specified (see `openspec/specs/`) but not in the app yet. Don't rely on them; this list shrinks as they ship.

- **Repository courses and course versions.** Adding a course by git URL, course versions as `vX.Y.Z` tags, the version selector. Today every course is added from a local folder and updated with Sync.
- **Scripts in course activities.** `controls` in a course `activity.yaml` are validated but not shown or run. Per the spec they'll run with the activity folder as working directory and `ACTIVITY_DIR`, `WORKSPACE_DIR` (the per-activity learner folder), `HOST_OS`, `HOST_ARCH`.
- **The coding activity in full.** Copying `workspace` files to a per-activity learner folder (`~/Chiron/<storage key>/<activity slug>`), Reset and Full reset, the **Open in external editor** button. Files outside `workspace` (hidden tests, solutions) will stay in the course and remain readable by scripts.
- **Choosing an external editor** in settings; today folders open in the system file manager.
- **Changing the learner folder root** in settings; today it is `~/Chiron`.
- **App update notice** on startup.

Ideas that aren't specified yet, such as new question kinds, shuffling or an assessment activity type, are tracked in `openspec/feature-requests.md`.
