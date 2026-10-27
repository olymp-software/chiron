# Spec Delta

## MODIFIED Requirements

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
