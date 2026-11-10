# Spec Delta

## MODIFIED Requirements

### Requirement: Repository layout
A course repository SHALL contain `course.yaml` at its root; a workshop repository SHALL contain `workshop.yaml` there instead. Each section SHALL be a folder `<section-slug>/` and each activity a folder `<section-slug>/<activity-slug>/` containing `activity.yaml`. There SHALL be no other nesting levels. The course root MAY contain the overview file `OVERVIEW.md`, matched case-insensitively, shown in the course overview. `README.md` files are allowed for GitHub and ignored by the app.

```
course.yaml               # or workshop.yaml
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
- **WHEN** the repository root has neither `course.yaml` nor `workshop.yaml`
- **THEN** the app shows a course error

#### Scenario: README is not the overview
- **WHEN** the course root has `README.md` and no overview file
- **THEN** the course overview does not show the `README.md` content

### Requirement: Published JSON Schemas
For every schema version a JSON Schema SHALL be published at `<schema-base>/course/v<N>.json` and `<schema-base>/course-activity/v<N>.json`, and for workshops at `<schema-base>/workshop/v<N>.json` and `<schema-base>/workshop-activity/v<N>.json`. The app SHALL use the same schemas to validate files that authors use for editor completion.

#### Scenario: Editor completion
- **WHEN** an author adds `# yaml-language-server: $schema=<schema-base>/course-activity/v1.json` as the first line of `activity.yaml`
- **THEN** a YAML-aware editor validates and autocompletes against the published schema

#### Scenario: Workshop activity completion
- **WHEN** an author adds `# yaml-language-server: $schema=<schema-base>/workshop-activity/v1.json` as the first line of `activity.yaml` of a workshop
- **THEN** a YAML-aware editor validates and autocompletes against the published schema
