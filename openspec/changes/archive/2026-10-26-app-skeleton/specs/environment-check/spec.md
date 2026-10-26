# Spec Delta

## MODIFIED Requirements

### Requirement: Check run environment
The check SHALL run under the same contract as control scripts (interpreter, guaranteed commands, output, exit code, Stop), except that its working directory SHALL be the course root and `ACTIVITY_DIR` and `WORKSPACE_DIR` SHALL NOT be set. `HOST_OS`, `HOST_ARCH` and the inherited environment SHALL be present.

#### Scenario: Working directory
- **WHEN** the check runs `test -f course.yaml`
- **THEN** it exits with code 0

#### Scenario: No activity variables
- **WHEN** the check prints `$WORKSPACE_DIR`
- **THEN** the value is empty
