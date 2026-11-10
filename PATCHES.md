# Changes from upstream

This is the Asciiverse packaging of [El Poblador](https://sr.ht/~vicho/el_poblador/)
by Vicente Reyes (EUPL-1.2), built from upstream's `main` at `265c83d`
(5 November 2026, after its last tag `v0.3.0`). Upstream is the `upstream`
remote (git.sr.ht/~vicho/el_poblador); the `asciiverse` branch carries the
changes below.

## Ours

- **A start screen.** Run with no command, as av starts it, `el_poblador`
  opens a start screen instead of printing its usage: a new game for three
  or four players at one keyboard, with their names (empty keeps "Player
  1" and so on, and names must differ), or a saved game to continue, listed
  newest first with its players. `setup.go`, and a few lines in `main.go`;
  the `new`, `load` and `serve` commands are unchanged, and `-help` prints
  the usage.
- **Saves in the player's state folder.** Save & Quit wrote
  `game_save_<time>.gob` to the current folder, which under a launcher is
  wherever it happened to start. It now writes to `$XDG_STATE_HOME/el-poblador`,
  else `~/.local/state/el-poblador` (Linux), `~/Library/Application Support/el-poblador`
  (macOS) or `%LocalAppData%\el-poblador` (Windows). `game/savedir.go`.
  `el_poblador load <file>` still opens a save from anywhere.
- **How to play it, at the top of README.md.** A short note on installing
  and starting it with av; the original text follows unchanged.

## Packaging

- `.github/workflows/release.yml`: builds av archives for Linux (static,
  x86-64 and ARM64), macOS (x86-64 and ARM64) and Windows (x86-64 and
  ARM64), plus SHA256SUMS, and .deb, .rpm, .apk and Arch packages.
  Upstream builds on builds.sr.ht and has no GitHub workflow.
- `asciiverse/nfpm.yaml`: the system packages.
- `asciiverse/el-poblador.cast`: the catalog preview.
