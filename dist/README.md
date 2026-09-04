# Legacy updater migration only

Neither archive in this directory is an installation package for current users or current hosts. New installation and update flows retrieve the official repository `main` source and run its source-only sync script.

- `puretokens-skill-install.zip` is the source-only migration archive. It contains the current sync implementation and six Skills; it must not contain a Node runtime, a direct API executor, or retired monolithic model-selection files.
- `puretokens-skill-install-payload.zip` exists only at the exact path hard-coded into published `0.13.x` updaters. Those old updaters refuse an archive unless it has their legacy runtime marker. The marker performs no API request and is not used by the current Skills; it lets the updater copy the current source-only Skills, after which later updates use the repository-source flow.

Do not link either file in user installation instructions.
