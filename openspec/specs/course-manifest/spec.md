# course-manifest Specification

## Purpose

Defines the layout of a course repository, the `course.yaml` course manifest, slugs, rules shared by all course files, and how file schemas are versioned.

## Requirements

### Requirement: Repository layout
A course repository SHALL contain `course.yaml` at its root. Each section SHALL be a folder `<section-slug>/` and each activity a folder `<section-slug>/<activity-slug>/` containing `activity.yaml`. There SHALL be no other nesting levels. The course root MAY contain the overview file `OVERVIEW.md`, matched case-insensitively, shown in the course overview. `README.md` files are allowed for GitHub and ignored by the app.

```
course.yaml
OVERVIEW.md               # optional, shown in the course overview
README.md                 # optional, ignored by the app
scripts/check.sh          # optional, environment check
<section-slug>/
  <activity-slug>/
    activity.yaml
    ...                   # workspace files, tests, scripts, images
```

#### Scenario: Valid layout
- **WHEN** the course has `course.yaml` and `basics/hello/activity.yaml` referenced from the manifest
- **THEN** the app builds the course tree with section `basics` and activity `hello`

#### Scenario: Missing course manifest
- **WHEN** the repository root has no `course.yaml`
- **THEN** the app shows a course error

#### Scenario: README is not the overview
- **WHEN** the course root has `README.md` and no overview file
- **THEN** the course overview does not show the `README.md` content

### Requirement: Manifest is the only source of course structure
The set of sections and activities and their order SHALL come only from `course.yaml`. Folder names MUST NOT carry ordering numbers. Progression order SHALL run through all sections: after the last activity of a section comes the first activity of the next.

#### Scenario: Folder not in the manifest
- **WHEN** the repository has an activity folder that the manifest does not reference
- **THEN** the activity does not appear in the course

#### Scenario: Cross-section order
- **WHEN** the learner completes the last activity of section `basics`
- **THEN** the next activity is the first activity of the following section

### Requirement: Lazy reading of activities
The app SHALL build the course tree from `course.yaml` alone and read an activity folder only when the learner opens that activity. A mismatch between manifest and disk SHALL be an error of that activity only.

#### Scenario: Activity folder missing
- **WHEN** the manifest references `basics/hello` but the folder or its `activity.yaml` does not exist
- **THEN** the course tree is shown, and opening `hello` shows an activity error

### Requirement: course.yaml fields
`course.yaml` SHALL accept: `version` (integer, required), `title` (string, required), `description` (plain-text string, optional; shown in the course overview when there is no overview file), `check` (path from the course root, optional), `sections` (list, required, at least one). A section SHALL have `slug`, `title` (required), `description` (optional) and `activities` (required, at least one). An activity reference SHALL have `slug` and `title`, both required.

```yaml
# yaml-language-server: $schema=<schema-base>/course/v1.json
version: 1
title: Go Basics
description: Go fundamentals in 20 activities
check: scripts/check.sh
sections:
  - slug: basics
    title: First steps
    activities:
      - slug: hello
        title: Hello, Go
```

#### Scenario: Empty section
- **WHEN** a section has an empty `activities` list
- **THEN** the app shows a course error

#### Scenario: Activity title comes from the manifest
- **WHEN** the app shows the course tree
- **THEN** activity titles come from `course.yaml` without reading activity folders

### Requirement: Slug format and uniqueness
A slug SHALL match `^[a-z0-9]+(-[a-z0-9]+)*$` and equal its folder name. Section slugs SHALL be unique among sections. Activity slugs SHALL be unique across the whole course, not only within a section.

#### Scenario: Invalid slug
- **WHEN** a slug is `Hello_World`
- **THEN** the app shows a course error

#### Scenario: Duplicate activity slug in different sections
- **WHEN** sections `basics` and `advanced` both contain activity slug `intro`
- **THEN** the app shows a course error

### Requirement: Activity slug is the activity identity
Progress and the learner folder SHALL be keyed by activity slug. Moving an activity to another section SHALL keep its identity. Renaming a slug SHALL create a new activity.

#### Scenario: Activity moved to another section
- **WHEN** a new course version moves activity `loops` from `basics` to `control-flow`
- **THEN** its progress and learner folder are kept

#### Scenario: Slug renamed
- **WHEN** a new course version renames activity `loops` to `for-loops`
- **THEN** `for-loops` is a new activity and the progress for `loops` no longer applies

### Requirement: Common file rules
All text files SHALL be UTF-8. YAML SHALL be parsed strictly: an unknown field or a value of the wrong type is an error. All paths in YAML SHALL be relative with `/` separators; absolute paths and `..` are forbidden, and the target SHALL exist.

#### Scenario: Unknown field
- **WHEN** `course.yaml` contains a field `autor`
- **THEN** the app shows a course error naming the field

#### Scenario: Path escapes the course
- **WHEN** `check` is `../outside.sh`
- **THEN** the app shows a course error

### Requirement: Independent schema versions
`course.yaml` and every `activity.yaml` SHALL have an integer `version`, the schema version of that file. Course and activity schema versions SHALL be independent; one course MAY mix activities of different schema versions. Any schema change, including a new optional field, SHALL bump the version. Current versions: course 1, activity 1.

#### Scenario: Schema version is not course version
- **WHEN** a course at tag `v2.3.0` has `version: 1` in `course.yaml`
- **THEN** the app treats it as course schema 1

### Requirement: Published JSON Schemas
For every schema version a JSON Schema SHALL be published at `<schema-base>/course/v<N>.json` and `<schema-base>/activity/v<N>.json`. The app SHALL use the same schemas to validate files that authors use for editor completion.

#### Scenario: Editor completion
- **WHEN** an author adds `# yaml-language-server: $schema=<schema-base>/activity/v1.json` as the first line of `activity.yaml`
- **THEN** a YAML-aware editor validates and autocompletes against the published schema

### Requirement: Unsupported courses
If a file's `version` is higher than the app supports, or `type` or `kind` is unknown to the app, the app SHALL report that the course needs a newer app version, link to the app's releases page, and MUST NOT try to parse that file further.

#### Scenario: Newer activity schema
- **WHEN** an `activity.yaml` has `version: 2` and the app supports only 1
- **THEN** the activity shows "requires a newer app version" with a link to the releases page

#### Scenario: Unknown activity type
- **WHEN** an `activity.yaml` has `type: terminal`
- **THEN** the activity shows "requires a newer app version"
