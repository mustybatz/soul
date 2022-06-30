# Soul
---
Soul is a CLI tool written in Go that transforms csv files into Json.
## How to use 
``` bash
$ ./soul [options] <csvFilePath>
```
Example:
``` bash
$ ./soul data.csv
```

The command above will create a `data.json` file at exactly the same dir location of `data.csv`

## Options available
- **pretty**: If enabled, it will create a well-formatted JSON file instead of a compact one.
- **separator**: to indicate which character is used to separate cells. Only accepted options are `comma` (default) or `semicolon`.

Example using options:
``` bash
$ ./soul --pretty --separator=semicolon data.csv
```

The command above will create a formatted `data.json` file at exactly the same dir location of `data.csv` The row columns from this file are separated using semicolons instead of commas.

## Build
If you want to generate an executable for your platform, just run the following command, this tool is written using only standard lib packages, so no aditional installs are required.
``` bash
$ go build
```

## Tests
TDD was used to develop this utility, if you want to run the tests by yourself just run the following command:
``` bash
go run
```

