# Spec Delta

## MODIFIED Requirements

### Requirement: Unsupported courses
If a file's `version` is higher than the app supports, or `type` or `kind` is unknown to the app, the app SHALL report that the course needs a newer app version, link to the app's releases page, and MUST NOT try to parse that file further.

#### Scenario: Newer activity schema
- **WHEN** an `activity.yaml` has `version: 2` and the app supports only 1
- **THEN** the activity shows "requires a newer app version" with a link to the releases page

#### Scenario: Unknown activity type
- **WHEN** an `activity.yaml` has `type: terminal`
- **THEN** the activity shows "requires a newer app version"

#### Scenario: Unknown theory block kind
- **WHEN** an `activity.yaml` has a theory block with `kind: diagram`
- **THEN** the activity shows "requires a newer app version"
