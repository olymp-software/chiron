# Feature requests

Ideas not yet worked out into OpenSpec changes. When one is taken up, create a change in `openspec/changes/` and remove the entry here.

## Quiz activities

Feedback from writing the Terraform Associate exam-prep course (`tmp/courses/15-terraform-associate`, 32 quiz activities, 394 choice questions).

### `true_false` question kind

A statement the Learner marks as true or false. Today Authors emulate it with a two-option `choice` question, quoting `"True"` / `"False"` so YAML does not read them as booleans, and repeating the same markup in every such question.

### `matching` question kind

The Learner matches items of two lists (e.g. plan symbol ↔ action, flag ↔ command, CLI command ↔ purpose). Today Authors split one matching task into several choice questions.

### Shuffling

Option order (and optionally question order) is shuffled on every attempt. Today Authors vary the position of the correct option by hand, and a Learner retaking a quiz remembers answers by position.

### `select` in multiple-choice questions

`select: N` on a `multiple: true` question: the UI tells the Learner how many options to select ("Select two"), and the manifest check verifies that exactly N options are correct. Today "Select two." is only text in the prompt.

### `references` on questions

A list of links (e.g. official docs) shown after the Learner answers, separate from `explanation`. Today Authors put links inside `explanation`.

## `assessment` activity type

A new Activity type for exam simulation, unlike a quiz where every answer is checked and explained right away:

- the Learner answers all questions, results and explanations are shown only at the end;
- final score in percent and a pass threshold set by the Author;
- optional time limit.
