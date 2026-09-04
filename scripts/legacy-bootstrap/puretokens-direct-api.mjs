#!/usr/bin/env node

process.stderr.write("This compatibility marker is only for migrating a published legacy updater. Current Pure Tokens Skills do not run a Node media runtime. Start a new host conversation after the update.\n");
process.exitCode = 1;
