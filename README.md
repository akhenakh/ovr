# ovr

A CLI tool to pipe anything into and apply transformations with an advanced UI.

## Build
```sh
go build -o ovr ./cmd/ovr
```

Enable geo features.
```sh
go build -tags geo -o ovr ./cmd/ovr
```
## Features
- Fuzzy search for block names
- Apply actions, cancel actions using backspace
- Parse text, chain & transform
- Known formats (multiline, csv, json ..) filtering, transforming
- Plot 
- Highlight known code
- Create scripts using TUI, replay scripts with simple CLI options

## Inputs Outputs
- from/to clipboard
- stdin
- editor https://github.com/charmbracelet/bubbletea/tree/master/examples/textarea
- file?


## Format

- Text
- Lines
- CSV
- JSON
- YAML
- TOML
- Images
- Geometry

## Values Types

- numbers
- durations
- time, epoch, parse
- bin



## Transformations

- [X] to upper/lower
- [X] Title
- [ ] CamelCase
- [X] encoding from/to (b64, hex ...)
- [X] hashes
- [ ] count inputs
- [X] time parse transform, epoch 
- [ ] Time timezone, not completed
- [ ] duration add substract
- [X] escape unescape
- [ ] reformat input, prettifie
- [X] JWT decode
- [ ] JWT Validate
- [ ] known payloads (AWS...), logs severity, golang stack, java stack...
- [X] Minify 
- [ ] sort by a column/property
- [ ] Add/Set value
- [ ] conversion (json, csv, yaml, toml)
- [ ] output to a configurable filename, xxx-%Y%m%d.txt
- [ ] execute a shell command
- [ ] Colors, RGBtoHex, js names to colors
- [X] WKB/WKT/GeoJSON (geometry)
- [X] Geometry: area, centroid, timezone, 
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
