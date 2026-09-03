# 06 · First steps

Do not treat this as an implementation task list. Nothing here is authorized to be built until Taḋg passes Gate 1 on the corresponding issue (see `05-process.md`). This file lists **the sequence of actions you will need to take, and the choices to escalate first**, so that when Taḋg does authorise you, you are not starting cold.

## Zero: confirm you are in the right cwd

The new repo's location and name are not yet recorded in this handover. Do not assume. Ask Taḋg for the path, `cd` there, and then confirm:

```
git rev-parse --is-inside-work-tree     # must print 'true'
git remote -v                            # note the origin URL
ls -la                                   # is it empty? already scaffolded?
```

If the repo already contains anything, read what is there before writing anything new. Do not overwrite Taḋg's edits (`AGENTS.md §1.4`).

## One: verify the environment

Before drafting even the first issue:

```
which folio && folio --version           # expect ~/.local/bin/folio, ≥ 0.4.10
which typst && typst --version           # required for PDF output
which pandoc && pandoc --version         # required for manuscript md/org
xcodebuild -version                      # Xcode ≥ 26.x (confirmed available 2026-07-20)
swift --version                          # comes with Xcode
```

Record anything missing on a first-run issue in the new repo before drafting the app itself.

## Two: draft the master issue

Recommended path (per `05-process.md`): `/draft-design-issue` for the MVP master.

Suggested master issue shape (adjust to `ISSUES.md`):

- **Title:** `feat: MVP macOS companion app`
- **Body:** the confirmed goal, definition of done, scope, exclusions, human-only checkpoints. Copy from `02-decisions.md` §Decided and any Recommended items Taḋg has confirmed by the time you draft this.
- **Child issues (draft alongside):**
  1. `feat: FolioCore — FolioRunner and ArgBuilder` — plumbing layer, no UI. Testable purely in `FolioCoreTests`.
  2. `feat: FolioCore — ConfigReader` — reads `script.yaml`, prefill struct.
  3. `feat: FolioCore — OutputParser` — parses Letter's `Generated:` lines, exposes emitted-path events.
  4. `feat: App shell — sidebar and empty views` — `NavigationSplitView`, three view stubs, no wiring.
  5. `feat: Convert screen` — full behaviour per `docs/ux_proposal.md` §Screen: Convert.
  6. `feat: Letter screen` — full behaviour per §Screen: Letter. **Blocked on CLI issue `folio letter --list`** if R4 is confirmed (see `07-open-questions.md`).
  7. `feat: Manuscript screen` — full behaviour per §Screen: Manuscript.
  8. `chore: Makefile with build/test/lint/run` — matches `~/code/agents/SDLC.md` §Makefile.
  9. `docs: README, ARCHITECTURE, and this project's CLAUDE.md` — mirror the first-folio repo's structure.

**Do not implement any child until its ACs have been drafted and Taḋg has PROCEED'd (or the master is under confirmed MODE DELIVER with per-child audits passing).**

## Three: bootstrap the Xcode project (only after PROCEED on issue 4)

Suggested shape — adjust to what actually works cleanly in Xcode 26:

1. `Package.swift` at the repo root defining a `FolioCore` library target and its test target. Library only; no app target here.
2. Xcode project `FirstFolio.xcodeproj` in the repo root, with an app target that depends on `FolioCore` via the local package.
3. Info.plist keys:
   - `LSUIElement`: false — this is a normal Dock app.
   - `NSSupportsAutomaticGraphicsSwitching`: true.
   - `LSMinimumSystemVersion`: whatever Taḋg confirms for O4 in `02-decisions.md`.
4. Entitlements: **App Sandbox disabled** (per R2). Do not enable "Outgoing Connections" or any other network entitlement — the app makes no network calls.
5. `.gitignore` covering `.build/`, `xcuserdata/`, `DerivedData/`, `*.xcodeproj/project.xcworkspace/xcuserdata/`, `.DS_Store`.

## Four: PATH capture

When `Process` spawns `folio`, it inherits the app's environment, and a GUI-launched app on macOS has a minimal `PATH` (often `/usr/bin:/bin:/usr/sbin:/sbin`). `folio` at `~/.local/bin/folio` is therefore not automatically visible.

Strategy:

1. On first launch, run `bash -lc 'echo $PATH'` once via `Process` to capture the shell login PATH, and persist it in `UserDefaults` (`FolioResolvedShellPATH`).
2. On every subsequent `folio` spawn, set `Process.environment["PATH"]` to the persisted PATH (falling back to a hard-coded reasonable default: `~/.local/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin`).
3. Discover `folio` by searching the persisted PATH (or the user-overridden explicit path) with `FileManager.isExecutableFile(atPath:)`.

This is the same trick Alfred, Raycast, and every other GUI-native macOS tool uses. Document it in the repo's `ARCHITECTURE.md` when you write that.

## Five: hello-world end-to-end

Before wiring any real UI, prove the plumbing:

1. `FolioRunner` unit test that spawns `folio --version` and asserts the stdout contains "folio ".
2. `ArgBuilder+Convert` unit tests for every combination that matters (blank fields omitted, style flag mapping, preview-as-text vs save-as-file).
3. Manual smoke: an empty `RootView` that has a "Test folio" button; tapping it runs `folio --version` via `FolioRunner` and shows the output. Verify by running the app once.

Only after this passes do you start on real views.

## Six: fixtures

Copy (do not symlink) representative test fixtures from `/Users/tigger/code/tigoss/first-folio/examples/` into the new repo under `Tests/Fixtures/`. Keep a NOTE.md in that directory recording the source repo, the source commit hash, and the copy date, so the next agent knows how to refresh them.

Do not vendor `about-time.org` if it contains real personal correspondence — check first with Taḋg. `examples/dummy-manuscript.org` and similar `dummy-*` files are safe.

## Seven: documentation to write as you go

Mirror first-folio's shape:

- `README.md` — quickstart, install, screenshots.
- `ARCHITECTURE.md` — the runtime shape once implemented (source: `03-architecture.md`).
- `docs/ACs.org` — post-approval AC migration target, per SDLC §Phase 5.
- `docs/ux.md` — a stable copy of the app's own UX contract (initially: a link back to `first-folio/docs/ux_proposal.md` and a note of any deltas; over time, becomes canonical for the app).
- `CLAUDE.md` at the repo root — repo-specific agent instructions if any diverge from `~/code/agents/*`.

Do not create documentation files that duplicate this handover. Once the code exists, this handover directory can be deleted (or moved to an `archive/` directory in the app repo with a note pointing to what replaced it).
