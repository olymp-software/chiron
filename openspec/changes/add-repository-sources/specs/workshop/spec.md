## MODIFIED Requirements

### Requirement: Adding names neither content type
The library buttons that add from a folder and from a repository, the folder dialog, the repository add dialog, the empty library text and the script warning SHALL apply to both courses and workshops: the content type is known only after the folder or the chosen version of the repository is read.

#### Scenario: Add button
- **WHEN** the learner opens the library
- **THEN** the buttons read "Add from folder" and "Add from repository"

#### Scenario: Workshop from a repository
- **WHEN** the learner adds a repository whose chosen version has `workshop.yaml` in its root
- **THEN** it is added as a workshop and listed under "Workshops"
