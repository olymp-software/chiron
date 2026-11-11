# Spec Delta

## MODIFIED Requirements

### Requirement: Theory blocks
`theory_blocks` SHALL be a list of at least one theory block, shown in order at the top of the activity in every activity type; a `quiz` activity shows them on its theory screen instead. A theory block SHALL have `kind` (required) and `data` (object whose shape depends on `kind`, required).

```yaml
theory_blocks:
  - kind: markdown
    data:
      content: |
        # Pods
        A Pod is the smallest deployable unit.
```

#### Scenario: Empty theory blocks
- **WHEN** `theory_blocks` is an empty list
- **THEN** opening the activity shows an activity error

#### Scenario: Missing theory blocks
- **WHEN** a `quiz` activity has no `theory_blocks`
- **THEN** opening the activity shows an activity error

#### Scenario: Block without data
- **WHEN** a theory block has `kind: markdown` and no `data`
- **THEN** opening the activity shows an activity error

#### Scenario: Blocks keep their order
- **WHEN** an activity has two `markdown` blocks
- **THEN** they are shown in the order written

### Requirement: quiz type
A `quiz` activity SHALL show its theory blocks on a theory screen and its questions on a separate quiz screen opened from it. `data` SHALL be required and SHALL hold the questions.

#### Scenario: Quiz without data
- **WHEN** a `quiz` activity has no `data`
- **THEN** opening the activity shows an activity error

#### Scenario: Theory and questions are separate
- **WHEN** the learner opens a valid `quiz` activity
- **THEN** its theory blocks are shown without its questions
