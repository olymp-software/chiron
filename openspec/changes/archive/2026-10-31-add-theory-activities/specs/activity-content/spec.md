# Spec Delta

## ADDED Requirements

### Requirement: Activity error
An activity whose folder or `activity.yaml` is missing, or whose `activity.yaml` is not valid YAML, breaks its schema or fails a semantic check, SHALL show an activity error in place of its content. The error SHALL list every problem found with the YAML path of the offending value. The course tree and the other activities SHALL stay usable.

#### Scenario: Missing activity.yaml
- **WHEN** the manifest references `basics/hello` and `basics/hello/activity.yaml` does not exist
- **THEN** opening `hello` shows an activity error and the course tree is shown

#### Scenario: Several problems
- **WHEN** an `activity.yaml` has an unknown field `titel` and a theory block without `data`
- **THEN** the activity error names both problems with their YAML paths

#### Scenario: Fixed by Sync
- **WHEN** the author fixes a broken `activity.yaml` of a local-folder course and the Learner presses Sync and opens the activity again
- **THEN** the activity shows its theory blocks

## MODIFIED Requirements

### Requirement: Local images only
Images SHALL be referenced only by relative paths to files inside the activity folder. The course SHALL work offline.

#### Scenario: Relative image
- **WHEN** a `markdown` block's `content` contains `![diagram](images/diagram.png)` and the file exists in the activity folder
- **THEN** the image is shown

#### Scenario: Remote image
- **WHEN** a `markdown` block's `content` references an image by `https://` URL
- **THEN** the image is not loaded

#### Scenario: Image outside the activity folder
- **WHEN** a `markdown` block's `content` contains `![logo](../../images/logo.png)` and the file exists in the course
- **THEN** the image is not loaded
