# e-ink-ui Specification

## Purpose
The e-ink mode adapts Sandwitches for kitchen e-ink displays, providing a high-contrast, large-text recipe viewer that's readable at counter distance and operable with messy hands.

## Requirements

### Requirement: E-ink mode toggle
The application SHALL provide a mechanism to enable e-ink mode via URL parameter (`?eink=1`) and optionally persist via cookie or user setting.

#### Scenario: Enable via URL parameter
- **WHEN** a user appends `?eink=1` to any page URL
- **THEN** the page SHALL render in e-ink mode for that session
- **AND** a cookie SHALL be set to persist the preference

#### Scenario: Enable via user settings
- **WHEN** a user toggles "E-ink mode" in their profile settings and saves
- **THEN** e-ink mode SHALL be enabled for that user across sessions

### Requirement: High contrast palette
All e-ink mode pages SHALL use strict black-on-white with no gradients, shadows, or semi-transparency.

#### Scenario: Colors
- **WHEN** e-ink mode is active
- **THEN** all text SHALL be `#000000` on `#ffffff` background
- **THEN** secondary/muted text SHALL use `#333333` on `#ffffff`
- **THEN** borders and dividers SHALL use `#cccccc` solid lines

### Requirement: No motion or CSS effects
- **WHEN** e-ink mode is active
- **THEN** all CSS transitions, animations, keyframes SHALL be disabled
- **THEN** no `box-shadow`, `text-shadow`, `backdrop-filter`, `gradient`, or `filter` SHALL render

### Requirement: Large touch targets
- **WHEN** e-ink mode is active
- **THEN** all buttons, links, and interactive elements SHALL be minimum 48×48px
- **THEN** recipe navigation (prev/next step) SHALL be minimum 56×56px
- **THEN** touch targets SHALL have 10px minimum spacing for greasy-finger operation

### Requirement: Cooking mode
A dedicated cooking mode SHALL provide a simplified, full-screen recipe view.

#### Scenario: Single recipe view
- **WHEN** e-ink mode is active and a user opens a recipe
- **THEN** the recipe title SHALL display in minimum 28px bold
- **THEN** ingredient list SHALL display in minimum 20px
- **THEN** each step SHALL display one at a time in minimum 24px with generous line-height (1.8)
- **THEN** step navigation buttons SHALL be positioned at the bottom, easily tappable

#### Scenario: Step-by-step mode
- **WHEN** a user enters step-by-step cooking mode in e-ink mode
- **THEN** only the current step SHALL be displayed (full screen, minimal UI)
- **THEN** the step number and total steps SHALL be shown as "Step 2 of 8"
- **THEN** large prev/next buttons SHALL be at the bottom (minimum 64px height)
- **THEN** a "Show All Steps" option SHALL be available to see the full recipe

### Requirement: Recipe grid/browse view
The recipe browsing view SHALL be simplified.

#### Scenario: Recipe list
- **WHEN** viewing the recipe grid in e-ink mode
- **THEN** recipe cards SHALL use solid black borders (no shadows, no rounded corners)
- **THEN** recipe names SHALL be minimum 20px bold
- **THEN** difficulty and time SHALL use text labels ("Easy · 30min"), not color badges
- **THEN** recipe images SHALL have a visible border and be optional (can be hidden)

#### Scenario: Search and filters
- **WHEN** using search in e-ink mode
- **THEN** search input SHALL be minimum 44px height
- **THEN** filter tags SHALL use text labels with visible borders and checkboxes, not colored pills
- **THEN** search results SHALL list recipe names only (no thumbnails) for density
