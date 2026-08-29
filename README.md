# ovr

A CLI tool to pipe anything into and apply transformations with an advanced UI.

## Install

With [Homebrew](https://brew.sh) (Linux and macOS, formulas from [akhenakh/homebrew-tap](https://github.com/akhenakh/homebrew-tap)):
```sh
brew install akhenakh/tap/ovr
```

Or from source (requires Go 1.27):
```sh
go build -o ovr ./cmd/ovr
```

Enable geo features.
```sh
go build -tags geo -o ovr ./cmd/ovr
```

## Usage

```sh
ovr [flags]
```

By default ovr reads the clipboard. If the clipboard is unavailable
(eg. running over SSH), use an alternative input:

- `-s` read from stdin, eg. `cat data.txt | ovr -s`
- `-f <filename>` read from a file

Flags:

- `-s` use stdin as input data
- `-f <filename>` read input from a file
- `-r` raw output, only print the final result
- `-o <filename>` write the output to a file
- `-debug` log to debug.log

On exit, ovr prints the applied action chain, eg. `Split(" "),Index(2),Upper`,
followed by the result.

## Features
- [X] Fuzzy search for block names
- [X] Apply actions, cancel actions using backspace
- [X] Parse text, chain & transform
- [ ] Known formats (multiline, csv, json ..) filtering, transforming
- [ ] Plot 
- [ ] Highlight known code
- [ ] Create scripts using TUI, replay scripts with simple CLI options

## Inputs Outputs
- [X] from/to clipboard (with a fallback message and `-s`/`-f` hint when unavailable, eg. over SSH)
- [X] stdin (`-s`)
- [X] file (`-f <filename>`)
- [X] output to file (`-o <filename>`)
- editor https://github.com/charmbracelet/bubbletea/tree/master/examples/textarea

## Format

- [X] Text
- [X] Lines (text lists)
- [X] CSV (table of rows)
- JSON
- YAML
- TOML
- Images
- [X] Time
- [X] Geometry (WKT input like `POINT(-0.4539761 48.0930043)` is auto-detected
  and parsed, geo actions are offered immediately)

## Values Types

- numbers
- durations
- time, epoch, parse
- bin

## Config file
- [ ] Font name & Size
- [ ] add commands by invoking shell, txt to txt
      - name: count_lines
        pipe: "wc -l"

## Transformations

- [X] to upper/lower
- [X] Title
- [ ] CamelCase
- [X] encoding from/to (b64, hex ...)
- [X] hashes (md5, sha1, sha256, sha512, crc32, hmac)
- [X] count inputs
- [X] split text into a list with any separator (`split`), by comma, space, pipe
- [X] list actions: sort, reverse, first, last, index, count, join
- [X] time parse transform, epoch
- [X] date parsing: ISO 8601/JSON dates, Go time strings (`date`)
- [X] time to ISO, epoch, JSON date string
- [X] timezones: est, et, utc, pt, mst, cst, brt, gmt, cet, eet, msk, ist, sgt, hkt, jst, kst, aest, nzst, hst
- [X] duration add/substract, Go durations like `1s`, `2h30m`, days `2d`, weeks `3w`, negative to substract (`adddur`)
- [X] escape unescape
- [X] reformat input, prettifie
- [X] JSON prettify
- [X] JWT decode
- [ ] JWT Validate
- [ ] known payloads (AWS...), logs severity, golang stack, java stack...
- [X] JSON Minify 
- [X] CSV: parse into a table (`csv`), sort by a column (`sortcol`), output as csv (`tocsv`)
- [X] strip all whitespace (`strip`)
- [ ] sort yaml
- [ ] Add/Set value
- [ ] conversion (json, csv, yaml, toml)
- [ ] output to a configurable filename, xxx-%Y%m%d.txt
- [X] execute a shell command, input piped to stdin (`exec`)
- [ ] Colors, RGBtoHex, js names to colors
- [X] WKB/WKT/GeoJSON (geometry)
- [X] Geometry: area, centroid, timezone, 
- [X] s2/h3 cell covers
- [X] Interactive map view of geometries (tiletea, Kitty graphics protocol)
- [ ] Skip entries
- [ ] to qrcode
- [ ] ip address
- [ ] URL
- [ ] Source Code, format, colorize
- [ ] Save to file

## Filter 
- [ ] dedup from a list
- [ ] Filter fields, select values
- [ ] JMESPath
- [ ] Regexp
- [ ] https://github.com/tidwall/gjson

## Create Data (not from stdin or pasteboard)
- [ ] Time Now
- [ ] UUID
- [ ] From an HTTP Request
- [ ] Multiple Create (will create as many above)
- [ ] From Editor

## Real workflows

- from clipboard, unescape json, parse json, prettyfier, colorize
- from pipe, recognize CSV, apply sort by 3rd column, display output

## Libraries to consider

### Code Highlight color

- https://github.com/alecthomas/chroma

### UI

- https://github.com/charmbracelet/bubbletea 
- https://github.com/rivo/tview
- https://github.com/gdamore/tcell

### Screen recording

- https://asciinema.org/


## Content type guess

- https://github.com/h2non/filetype

### Encode

- json https://github.com/multiprocessio/go-json

### Logs

- https://lnav.org/

## Search

- https://github.com/Vivino/go-autocomplete-trie

## Markdown

- https://github.com/yuin/goldmark

### Transform

- https://github.com/TomWright/dasel
- JMESPATH https://jmespath.org/
- https://github.com/tidwall/gjson
- https://github.com/tidwall/sjson

## Inspirations

- https://github.com/IvanMathy/Boop
- https://open-vsx.org/extension/qcz/text-power-tools
- https://github.com/d-akara/vscode-extension-transformer

## Crazy Ideas

### Geography

- display GeoJSON as a map
- Find Centroid
- to/from s2
- to/from h3

### Online

- to pastebin
- to imgur render image
- to geojson.io
- to s2 map https://s2.inair.space/

### Unicode

- code to unicode
- name to unicode
## Name

The name over is based on [Over and Over from Hot Chip](https://www.youtube.com/watch?v=pDJKgi2e-Aw)

## TODO

- add a time action, then filter by time action when the type is time
- hex dump
