# Spec Delta

## MODIFIED Requirements

### Requirement: Check script declaration
A course MAY declare an environment check via `check` in `course.yaml`, a path from the course root. If `check` is absent, Enroll and Continue in the course overview SHALL be available without a check.

```sh
# scripts/check.sh
kubectl version --client || {
  echo "kubectl not found. Install it: https://kubernetes.io/docs/tasks/tools/"
  exit 1
}
```

#### Scenario: No check declared
- **WHEN** `course.yaml` has no `check`
- **THEN** the learner can enroll right after adding the course

### Requirement: Check gates the course
If `check` is declared, Enroll and Continue MUST NOT be available until the check exits with code 0. The course overview SHALL show a built-in "Check environment" button, the check status and an output panel. The check's output SHALL explain to the learner what is missing and where to get it.

#### Scenario: Failed check
- **WHEN** the check exits with code 1
- **THEN** the output panel in the course overview shows the script output and Enroll or Continue stays unavailable

#### Scenario: Successful check
- **WHEN** the check exits with code 0
- **THEN** Enroll or Continue becomes available

#### Scenario: No check button on the course screen
- **WHEN** the learner is on the course screen
- **THEN** the "Check environment" button is not shown there

### Requirement: Remembered by file hash
A successful check SHALL be remembered together with the SHA-256 of the `check` file. On opening the course overview and on entering a course the app SHALL recompute the hash; if it differs, the check SHALL be required again. The learner SHALL be able to rerun the check manually at any time from the course overview.

#### Scenario: Version with a changed check
- **WHEN** the learner switches to a course version whose `check` file differs
- **THEN** the course overview requires the check again before Enroll or Continue is available

#### Scenario: Unchanged check
- **WHEN** the learner reopens the course and the `check` file is unchanged
- **THEN** the check is not required again

#### Scenario: Required check survives a restart
- **WHEN** the check became required after a version switch and the learner restarts the app without running it
- **THEN** the check is still required
