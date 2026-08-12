# CLAUDE.md

Project-specific instructions for Claude Code working in this repo.

## Rolling to a new Playwright version

Always use the [`roll-playwright`](.claude/skills/roll-playwright/SKILL.md) skill for this — do not bump `playwrightCliVersion` or edit `patches/main.patch` by hand outside of it. It encodes the submodule/patch/codegen workflow and the cross-binding parity checks against python/java/dotnet.

## Cutting a release

Always use the [`release-playwright`](.claude/skills/release-playwright/SKILL.md) skill for this — do not hand-tag a release. It encodes the version-numbering convention (`v0.<MM><PP>.<G>`, tracking the upstream driver's own minor+patch) and the release-notes format; guessing the version number by hand has produced at least one inconsistent tag (`v0.6100.0`) in the past.
