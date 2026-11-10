# Spec Delta

## Purpose

Defines the workshop: a content type whose activities change one shared state, such as files in a folder, resources in a cloud account or a local cluster, which the app does not own.

## ADDED Requirements

### Requirement: Content type by root file
A repository with `workshop.yaml` in its root SHALL be a workshop, and one with `course.yaml` SHALL be a course. A repository with both files or neither SHALL be a course error.

#### Scenario: Workshop manifest
- **WHEN** the learner adds a folder with a valid `workshop.yaml` and no `course.yaml`
- **THEN** it is added to the library as a workshop

#### Scenario: Both manifests
- **WHEN** a folder has both `course.yaml` and `workshop.yaml`
- **THEN** the app shows a course error and the folder is not added

#### Scenario: Neither manifest
- **WHEN** a folder has neither `course.yaml` nor `workshop.yaml`
- **THEN** the app shows a course error and the folder is not added

### Requirement: workshop.yaml fields
`workshop.yaml` SHALL accept the same fields with the same meaning as `course.yaml`: `version`, `title`, `description`, `check` and `sections`, plus optional `controls`. Its schema version SHALL be independent of the `course.yaml` schema version.

#### Scenario: Schema versions are independent
- **WHEN** a workshop has `version: 1` in `workshop.yaml`
- **THEN** the app treats it as workshop schema 1, whatever course schema versions it supports

#### Scenario: Newer workshop schema
- **WHEN** `workshop.yaml` has a `version` higher than the app supports
- **THEN** the app reports that a newer app version is required

### Requirement: Workshops reuse course behavior
Except where this specification says otherwise, a workshop SHALL behave as a course in course source, course versions, Sync, the library, the course overview, Enroll, Continue, the environment check, the course tree, progression and removal. Every rule for `course.yaml` SHALL apply to `workshop.yaml`.

#### Scenario: Environment check
- **WHEN** `workshop.yaml` declares `check` and the check has not succeeded for the current file
- **THEN** Enroll and Continue are unavailable

#### Scenario: Version switch
- **WHEN** the learner switches a workshop to another course version
- **THEN** the workshop folder is not touched and progress is kept by activity slug

### Requirement: Library separates courses and workshops
The library SHALL list courses and workshops in separate sections, titled "Courses" and "Workshops". A section with no entries SHALL NOT be shown. The course overview of a workshop SHALL show that it is a workshop. On the last activity "Finish course" SHALL read "Finish workshop".

#### Scenario: Both kinds in the library
- **WHEN** the library holds a course and a workshop
- **THEN** the course is listed under "Courses" and the workshop under "Workshops"

#### Scenario: Only workshops
- **WHEN** the library holds only workshops
- **THEN** only the "Workshops" section is shown

### Requirement: Adding names neither content type
The library button that adds from a folder, the folder dialog, the empty library text and the script warning SHALL apply to both courses and workshops: the content type is known only after the folder is read.

#### Scenario: Add button
- **WHEN** the learner opens the library
- **THEN** the button reads "Add from folder"

### Requirement: Next is unconditional
"Next" SHALL complete an activity of a workshop whatever the outcome of its Check, as in a course. The app SHALL NOT warn about a failed or skipped Check.

#### Scenario: Next without Check
- **WHEN** the learner presses "Next" without pressing Check
- **THEN** the activity is completed, the next one opens and no warning is shown

#### Scenario: Next after a failed Check
- **WHEN** Check exited with code 1 and the learner presses "Next"
- **THEN** the activity is completed and the next one opens

### Requirement: Workshop activity fields
`activity.yaml` of a workshop SHALL accept `version` (integer, required), `theory_blocks` (required, as in a course), `solution`, `check`, `controls` and `files` (all optional). It SHALL have no activity type and no questions. Its schema version SHALL be independent of the course activity schema version.

#### Scenario: Activity with only theory blocks
- **WHEN** an `activity.yaml` has only `version` and `theory_blocks`
- **THEN** the activity shows its theory blocks

#### Scenario: Activity type in a workshop
- **WHEN** an `activity.yaml` of a workshop has a `type` field
- **THEN** opening the activity shows an activity error naming the field

#### Scenario: Questions in a workshop
- **WHEN** an `activity.yaml` of a workshop has a `data` field
- **THEN** opening the activity shows an activity error naming the field

#### Scenario: Newer schema
- **WHEN** an `activity.yaml` of a workshop has a `version` higher than the app supports
- **THEN** the activity shows "requires a newer app version"

### Requirement: Solution
`solution` SHALL be non-empty Markdown, rendered as theory block content. It SHALL be hidden until the learner presses a Solution button and SHALL be hidden again when the activity is reopened. Showing it SHALL not affect progress. An activity without `solution` SHALL show no Solution button.

#### Scenario: Show the solution
- **WHEN** the learner presses Solution
- **THEN** the rendered `solution` is shown in the activity

#### Scenario: Reopening the activity
- **WHEN** the learner showed the solution, left and reopened the activity
- **THEN** the solution is hidden

#### Scenario: Empty solution
- **WHEN** `solution` is an empty string
- **THEN** opening the activity shows an activity error

### Requirement: Check button
`check` SHALL be a path relative to the activity folder that does not leave it. An activity with `check` SHALL show a built-in Check button with a fixed label translated by the app. An activity without `check` SHALL show no Check button.

#### Scenario: Activity with a check
- **WHEN** an activity declares `check: check.sh`
- **THEN** the activity shows a Check button

#### Scenario: Check outside the activity folder
- **WHEN** `check` is `../shared/check.sh`
- **THEN** opening the activity shows an activity error

### Requirement: Check runs in a modal
Pressing Check SHALL open the modal of the environment check, with the same blocking, Cancel, exit code display and OK. The output SHALL not be kept after the modal is closed, and the result SHALL not be remembered and SHALL gate nothing.

#### Scenario: Cancel a hanging check
- **WHEN** Check does not finish and the learner presses Cancel
- **THEN** the script and its child processes are terminated and the modal closes

#### Scenario: Closing a finished check
- **WHEN** Check exited with code 1 and the learner presses OK
- **THEN** the modal closes and nothing about the run stays on the activity

#### Scenario: Reopening the activity
- **WHEN** Check succeeded and the learner reopens the activity
- **THEN** the activity shows no result of that run

### Requirement: Check and controls run together
Check SHALL NOT belong to an exclusion group and SHALL NOT stop running controls. While the Check modal is open no control can be started.

#### Scenario: Checking a running server
- **WHEN** a control "Start server" is running and the learner presses Check
- **THEN** Check runs and the server keeps running

### Requirement: Controls in a workshop
Controls of a workshop activity SHALL be declared and behave as in a course, including exclusion groups and Stop.

#### Scenario: Stop a control
- **WHEN** a control is running and the learner presses Stop
- **THEN** the script and its child processes are terminated

### Requirement: Output of each control
Each control of a workshop activity SHALL have its own output, which the learner can select while several controls run. It SHALL be kept until that control starts again or the learner leaves the activity.

#### Scenario: Two controls running
- **WHEN** two controls in different exclusion groups are running
- **THEN** the learner can switch between their outputs and neither is mixed into the other

#### Scenario: Finished control
- **WHEN** a control has finished and the learner stays on the activity
- **THEN** its output stays available

#### Scenario: Restarting a control
- **WHEN** a finished control is started again
- **THEN** its output starts empty

### Requirement: Leaving with a running control
When the learner presses "Next" or otherwise leaves an activity while a control of that activity is running, the app SHALL ask for confirmation. If confirmed, it SHALL terminate the running scripts and leave. If declined, the learner SHALL stay and the scripts SHALL keep running.

#### Scenario: Confirmed
- **WHEN** a control is running, the learner presses "Next" and confirms
- **THEN** the script is terminated, the activity is completed and the next one opens

#### Scenario: Declined
- **WHEN** a control is running, the learner presses "Next" and declines
- **THEN** the activity is not completed and the script keeps running

#### Scenario: Nothing running
- **WHEN** no control is running and the learner presses "Next"
- **THEN** no confirmation is shown

#### Scenario: Opening another activity
- **WHEN** a control is running and the learner picks another activity in the course tree
- **THEN** the app asks for confirmation first

### Requirement: Workshop controls
`workshop.yaml` MAY declare `controls`, each with `label` and `script`, a path from the workshop root. They SHALL be shown in the course overview of an enrolled workshop, and run in the modal of the environment check. They SHALL NOT depend on the environment check being satisfied.

#### Scenario: Looking for leftovers before removing
- **WHEN** a workshop declares a control "Check leftovers" and the learner presses it in the overview
- **THEN** the modal shows its output, and it is not kept after the modal closes

#### Scenario: Check not satisfied
- **WHEN** the environment check is not satisfied
- **THEN** workshop controls can still be run

#### Scenario: Not enrolled
- **WHEN** the learner is not enrolled in a workshop
- **THEN** its controls are not shown

#### Scenario: Script outside the workshop
- **WHEN** a workshop control has `script: ../leftovers.sh`
- **THEN** the app shows a course error

### Requirement: Workshop control run environment
A workshop control SHALL run with the workshop root in the course copy as working directory, with `WORKSPACE_DIR`, `CONTENT_DIR` and `TEMP_DIR` set and `ACTIVITY_DIR` not set.

#### Scenario: No activity variable
- **WHEN** a workshop control prints `$ACTIVITY_DIR`
- **THEN** the value is empty

### Requirement: One workshop folder
A workshop SHALL have one workshop folder at `<learner folder root>/<storage key>`, shared by all its activities, created when an activity opens if it does not exist and left empty if nothing is put there. It SHALL have no per-activity subfolders and no path chosen by the learner.

#### Scenario: Shared between activities
- **WHEN** the learner creates a file in the workshop folder in one activity and opens the next
- **THEN** the file is there and `WORKSPACE_DIR` of both activities is the same folder

#### Scenario: Workshop without files
- **WHEN** a workshop that declares no `files` opens its first activity
- **THEN** the workshop folder exists and is empty

#### Scenario: Folder deleted by the learner
- **WHEN** the learner deleted the workshop folder and opens an activity
- **THEN** the folder exists again, with the `files` of that activity

### Requirement: Workshop folder path is visible
The workshop screen SHALL show the absolute path of the workshop folder on every activity, with a button that copies it.

#### Scenario: Working in a terminal
- **WHEN** the learner presses the copy button
- **THEN** the clipboard holds the absolute path of the workshop folder

### Requirement: Files are copied without overwriting
Every time an activity opens, the app SHALL copy each file of its `files` items that is missing from the workshop folder, preserving relative paths. A listed folder SHALL be merged file by file. Existing files MUST NOT be touched. Each item SHALL be an object with a required `path` relative to the activity folder.

#### Scenario: First activity provides starter files
- **WHEN** the first activity lists `starter/main.go`
- **THEN** the workshop folder has `starter/main.go` after it opens

#### Scenario: Learner edited a file
- **WHEN** the learner edited a listed file and reopens the activity
- **THEN** the edits are kept

#### Scenario: Two activities fill one folder
- **WHEN** two activities list a folder `manifests` with different files
- **THEN** after both open the workshop folder has the files of both

#### Scenario: Path outside the activity folder
- **WHEN** a `files` item has `path: ../other/data.csv`
- **THEN** opening the activity shows an activity error

### Requirement: Files are copied exactly and atomically
Each file SHALL be copied byte for byte, without line-ending or encoding conversion, keeping the executable bit where the OS has one. A file SHALL appear in the workshop folder only when complete. A failed copy SHALL show an activity error naming the file and leave existing files untouched.

#### Scenario: Binary file
- **WHEN** an activity lists a binary file
- **THEN** the copy in the workshop folder is identical to the file in the course copy

#### Scenario: Interrupted copy
- **WHEN** the app is closed while a large file is being copied
- **THEN** the file is not in the workshop folder and is copied in full on the next open

#### Scenario: Folder where a file is expected
- **WHEN** a listed file's path is a folder in the workshop folder
- **THEN** the activity shows an error naming the path and the folder is untouched

### Requirement: No built-in reset
A workshop SHALL have no Reset and no Full reset. Restoring or cleaning up state SHALL be the author's control or activity.

#### Scenario: Reset button
- **WHEN** the learner opens an activity of a workshop
- **THEN** no Reset button is shown

### Requirement: External editor on the workshop screen
The workshop screen SHALL have an "Open in external editor" button outside the activity content, available on every activity, that opens the workshop folder the way a `coding` activity opens its learner folder.

#### Scenario: Open from an activity without files
- **WHEN** the learner presses the button on an activity that declares no `files`
- **THEN** the workshop folder opens

### Requirement: Workshop script environment
Controls and the activity Check of a workshop SHALL run under the control script contract, with `WORKSPACE_DIR` set to the workshop folder. They SHALL also receive `CONTENT_DIR`, the root of the course copy, and `TEMP_DIR`, an empty folder outside the workshop folder created for that run. The app SHALL delete it after the run and SHALL delete leftovers on start. The environment check SHALL NOT receive these variables.

#### Scenario: Helper shared between activities
- **WHEN** a Check runs `. "$CONTENT_DIR/lib/server.sh"` and the file exists in the workshop root
- **THEN** the helper is loaded, whichever section the activity is in

#### Scenario: Build output outside the workshop folder
- **WHEN** Check builds a binary into `$TEMP_DIR`
- **THEN** the workshop folder is unchanged

#### Scenario: Nothing kept between runs
- **WHEN** a script writes a file to `$TEMP_DIR` and is run again
- **THEN** the second run starts with an empty `$TEMP_DIR`

#### Scenario: Stopped run
- **WHEN** a script writes to `$TEMP_DIR` and the learner cancels it
- **THEN** the next run starts with an empty `$TEMP_DIR`

#### Scenario: Environment check
- **WHEN** the environment check of a workshop prints `$TEMP_DIR`
- **THEN** the value is empty

### Requirement: Processes end with their script
When a script of a workshop exits on its own, the app SHALL terminate every process it started that is still running, as Stop does, before showing the exit code and deleting `TEMP_DIR`.

#### Scenario: Check leaves a server running
- **WHEN** Check starts a server in the background and exits
- **THEN** the server is terminated and its port is free before the modal shows the exit code

#### Scenario: Binary in TEMP_DIR
- **WHEN** a script exits while a process started from `$TEMP_DIR` is running
- **THEN** the process is terminated and `TEMP_DIR` is deleted
