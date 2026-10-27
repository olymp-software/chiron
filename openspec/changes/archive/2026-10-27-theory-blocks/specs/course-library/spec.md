# Spec Delta

## MODIFIED Requirements

### Requirement: Overview file rendering
The overview file SHALL be rendered with the same Markdown dialect, code highlighting, callouts and HTML restriction as activity Markdown. Images SHALL be referenced only by relative paths to files inside the course; remote images SHALL NOT be loaded.

#### Scenario: Script tag in the overview file
- **WHEN** `OVERVIEW.md` contains `<script>alert(1)</script>`
- **THEN** the text is shown literally and nothing executes

#### Scenario: Relative image
- **WHEN** `OVERVIEW.md` contains `![setup](images/setup.png)` and the file exists in the course
- **THEN** the image is shown
