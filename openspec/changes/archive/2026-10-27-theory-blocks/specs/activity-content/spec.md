# Spec Delta

## ADDED Requirements

### Requirement: Theory blocks
`theory_blocks` SHALL be a list of at least one theory block, shown in order at the top of the activity in every activity type. A theory block SHALL have `kind` (required) and `data` (object whose shape depends on `kind`, required).

```yaml
theory_blocks:
  - kind: markdown
    data:
      content: |
        # Pods
        A Pod is the smallest deployable unit.
```

#### Scenario: Empty theory blocks
- **WHEN** `theory_blocks` is an empty list
- **THEN** opening the activity shows an activity error

#### Scenario: Missing theory blocks
- **WHEN** a `quiz` activity has no `theory_blocks`
- **THEN** opening the activity shows an activity error

#### Scenario: Block without data
- **WHEN** a theory block has `kind: markdown` and no `data`
- **THEN** opening the activity shows an activity error

#### Scenario: Blocks keep their order
- **WHEN** an activity has two `markdown` blocks
- **THEN** they are shown in the order written

### Requirement: markdown block kind
The only theory block kind SHALL be `markdown`. Its `data` SHALL have `content` (Markdown, required, non-empty), rendered as Markdown.

#### Scenario: Empty markdown content
- **WHEN** a `markdown` block's `content` is an empty string
- **THEN** opening the activity shows an activity error

## MODIFIED Requirements

### Requirement: activity.yaml fields
`activity.yaml` SHALL accept: `version` (integer, required), `type` (`theory`, `quiz` or `coding`, required), `theory_blocks` (list of theory blocks, required), `controls` (list, optional; absent means no controls), `data` (object, presence depends on `type`). Fields common to all types SHALL live at the top level and type-specific fields under `data`.

#### Scenario: Empty content
- **WHEN** `activity.yaml` has a top-level `content` field instead of `theory_blocks`
- **THEN** opening the activity shows an activity error

#### Scenario: No controls
- **WHEN** `controls` is absent
- **THEN** the activity shows no control buttons

### Requirement: theory type
A `theory` activity SHALL show its theory blocks and its controls, if any. `data` MUST NOT be present.

#### Scenario: Theory with a control
- **WHEN** a `theory` activity has a control "Check Docker installation"
- **THEN** the activity shows the theory blocks and the control button

#### Scenario: Theory with data
- **WHEN** a `theory` activity has a `data` field
- **THEN** opening the activity shows an activity error

### Requirement: quiz type
A `quiz` activity SHALL show its theory blocks followed by its questions. `data` SHALL be required and SHALL hold the questions.

#### Scenario: Quiz without data
- **WHEN** a `quiz` activity has no `data`
- **THEN** opening the activity shows an activity error

### Requirement: coding type
A `coding` activity SHALL show its theory blocks, an "Open in external editor" button, the output panel and its controls. `data` SHALL be required and SHALL contain `workspace`: a list of at least one item, each an object with a required `path` to a file or folder relative to the activity folder.

```yaml
type: coding
controls:
  - label: Run
    script: run.sh
data:
  workspace:
    - path: main.go
    - path: go.mod
```

#### Scenario: Workspace item as a plain string
- **WHEN** a `workspace` item is written as `- main.go` instead of `- path: main.go`
- **THEN** opening the activity shows an activity error

#### Scenario: Empty workspace
- **WHEN** `workspace` is an empty list
- **THEN** opening the activity shows an activity error

### Requirement: Markdown dialect
All Markdown fields (`markdown` block `content` and quiz text fields) SHALL be rendered as CommonMark with GitHub extensions: tables, task lists, strikethrough, autolinks, fenced code blocks with a language tag.

#### Scenario: Table
- **WHEN** a `markdown` block's `content` contains a GitHub-style table
- **THEN** it is rendered as a table

### Requirement: Callouts
GitHub-style callouts `> [!type]` SHALL be rendered for types `note`, `tip`, `warning` and `task`, case-insensitive. `task` is an app extension; on GitHub it appears as a plain blockquote.

#### Scenario: Task callout
- **WHEN** a `markdown` block's `content` contains `> [!TASK]` followed by text
- **THEN** it is rendered as a task callout

### Requirement: No HTML rendering
HTML in Markdown MUST NOT be rendered; it SHALL be shown as text. The UI runs in a webview with access to app functions, including running scripts, so HTML from a third-party course would amount to code execution.

#### Scenario: Script tag in content
- **WHEN** a `markdown` block's `content` contains `<script>alert(1)</script>`
- **THEN** the text is shown literally and nothing executes

### Requirement: Local images only
Images SHALL be referenced only by relative paths to files inside the activity folder. The course SHALL work offline.

#### Scenario: Relative image
- **WHEN** a `markdown` block's `content` contains `![diagram](images/diagram.png)` and the file exists in the activity folder
- **THEN** the image is shown

#### Scenario: Remote image
- **WHEN** a `markdown` block's `content` references an image by `https://` URL
- **THEN** the image is not loaded
