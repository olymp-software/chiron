# app-platform Specification

## Purpose

Defines what the app is shipped as, how the learner learns about new app versions, which UI languages it offers, and what it logs and sends over the network.

## Requirements

### Requirement: Windows single-exe release
Version 1 SHALL ship for Windows only, as a single exe per architecture (amd64 and arm64), without an installer, published on GitHub Releases.

#### Scenario: Release contents
- **WHEN** a release is published
- **THEN** it contains a Windows amd64 exe and a Windows arm64 exe

### Requirement: App update notice
On startup the app SHALL check GitHub for the latest release and, if it is newer, show a message linking to its release page. Without network the check SHALL be skipped. There is no self-update.

#### Scenario: Newer release
- **WHEN** a newer release exists on startup
- **THEN** the app shows a message with a link to the release page

#### Scenario: Offline startup
- **WHEN** the app starts without network
- **THEN** it starts normally with no update message

### Requirement: UI languages
The UI SHALL be available in Russian and English, defaulting to the system language and changeable in settings. Course content is not translated.

#### Scenario: English system
- **WHEN** the system language is English on first start
- **THEN** the UI is in English

### Requirement: No telemetry
The app MUST NOT send telemetry or automatic error reports. It SHALL access the network only to reach git hosts for courses and GitHub for the release list.

#### Scenario: Error occurs
- **WHEN** the app hits an internal error
- **THEN** nothing is sent over the network about it

### Requirement: Local log
The app SHALL write a local log with course parse errors, git failures and crashes, and SHALL offer an "Open log folder" button. Script output MUST NOT be written to the log.

#### Scenario: Script output
- **WHEN** a control script prints output
- **THEN** it appears only in the output panel, not in the log
