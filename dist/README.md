# Legacy updater migration only

`puretokens-skill-install.zip` is not an installation package for current users or current hosts. New installation and update flows retrieve the official repository `main` source and run its source-only sync script.

This single archive remains only so Skills installed at version `0.13.25` can complete their already-published one-time updater path. That older updater is hard-coded to this exact repository path. The archive contains the current source-only sync implementation, so after that migration, later updates use the repository-source flow. Do not link this file in user installation instructions or create a second payload archive.
