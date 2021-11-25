```
    ____  _            ____             __     
   / __ \(_)________  / __ \____ ______/ /____ 
  / /_/ / / ___/ __ \/ /_/ / __ `/ ___/ __/ _ \
 / ____/ / /__/ /_/ / ____/ /_/ (__  ) /_/  __/
/_/   /_/\___/\____/_/    \__,_/____/\__/\___/ 
                                               
```
A tiny self-contained pasting service with a built-in database.

## Features
- Zero dependencies
- Built in super fast [database](https://git.mills.io/prologic/bitcask)
- Monaco editor built in (the editor powering vscode)

## Installation
- You need to have go installed and `~/go/bin` added to your path.
- Installing is as easy as `go install github.com/swz-git/picopaste@latest`

## Getting started

All you need to do to get started is to run the `picopaste` command in your terminal. Configuration is done using environment variables:
```
PICOPASTE_DB_PATH - A path to a directory where picopaste will store its data.
PICOPASTE_PORT - The port picopaste will use.
```