# onsave


`onsave` is a simple utility to allow quick script execution on file changes.

usage: `ls -l | onsave <command> [arguments]`

example: `find force-app/main/default/lwc/* -type f | onsave sfdx force:source:push`

It doesn't currently support directories but I/you could add it without too much effort.
