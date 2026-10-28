# Spec Delta

## MODIFIED Requirements

### Requirement: A learner folder per activity
On first open of any activity, regardless of type, the app SHALL create a learner folder for it, outside the course clone. The default root is `~/<App>/`; inside it the course folder is named by the course's storage key, with one subfolder per activity slug.

#### Scenario: Repository course
- **WHEN** the learner first opens activity `hello` of course `github.com/owner/repo` with storage key `repo-1a2b3c4d`
- **THEN** the folder `~/<App>/repo-1a2b3c4d/hello` exists

#### Scenario: Local-folder course
- **WHEN** the learner first opens activity `hello` of the local-folder course `C:\Courses\go-basics` with storage key `go-basics-5e6f7a8b`
- **THEN** the folder `~/<App>/go-basics-5e6f7a8b/hello` exists

#### Scenario: Theory activity
- **WHEN** the learner opens a `theory` activity
- **THEN** it also has a learner folder, available to its controls as `WORKSPACE_DIR`
