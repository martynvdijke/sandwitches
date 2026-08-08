## Why

Sandwitches is a recipe management app — the perfect companion for a kitchen e-ink display. E-ink is ideal in the kitchen: no glare from overhead lights, readable while cooking (hands may be messy for swiping), low power, and always-on for quick reference. An e-ink mode transforms Sandwitches into the ultimate kitchen recipe display.

## What Changes

- Add `?eink=1` URL parameter and/or cookie-activated e-ink mode
- Apply high-contrast black-on-white palette to all pages
- Remove all animations, transitions, gradients, shadows, backdrop filters
- Enforce 48px minimum touch targets — especially for recipe navigation while cooking
- Replace color-coded tags/difficulty ratings with icon + text labels
- Add a "cooking mode" — single-recipe view with large step text, auto-advance option
- Remove hover-dependent tooltips; show all info inline
- Simplify card layouts — use solid borders instead of shadows
- Increase font sizes for readability at kitchen counter distance

## Capabilities

### New
- `eink-mode-stylesheet`: Alternative high-contrast CSS
- `eink-mode-toggle`: URL param and settings toggle mechanism
- `eink-cooking-mode`: Simplified single-recipe view with large step-by-step instructions

### Modified
- *(none)*

## Impact

- **Frontend**: New eink.css stylesheet; toggle mechanism; cooking mode template/view
- **Backend**: (minimal) Persist e-ink preference in user settings
- **Database**: Optional
- **Dependencies**: None
