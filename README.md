# ovr
A CLI tool to pipe anything into and apply transformations with an advanced UI.

## Features
- Parse text, chain & transform
- Known formats (multiline, csv, json ..) filtering, transforming
- Plot 
- Highlight known code
- Create scripts using TUI, replay scripts with simple CLI options


## Format
- CSV
- JSON
- TOML
- Images

## Transformations

- [ ] to upper/lower
- [ ] Title
- [ ] CamelCase
- [ ] encoding from/to (b64, hex ...)
- [ ] hashes
- [ ] from/to clipboard
- [ ] count inputs
- [ ] time parse transform, epoch 
- [ ] escape unescape
- [ ] reformat input, prettifie
- [ ] JWT decode, known payloads (GeoJSON, AWS...), logs severity, golang stack, java stack...
- [ ] Minify 
- [ ] sort by a column/property
- [ ] dedup
- [ ] conversion (json, csv, yaml)
- [ ] Filter fields, select values
- [ ] output to a configurable filename, xxx-%Y%m%d.txt
- [ ] execute a shell command

## Real workflows
- from clipboard, unescape json, parse json, prettryfier, colorize
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

### Transform 
- https://github.com/TomWright/dasel
- JMESPATH https://jmespath.org/
- https://github.com/tidwall/gjson

## Inspirations
- https://github.com/IvanMathy/Boop

## Crazy Ideas

### Geography 
- display GeoJSON as a map
- Find Centroid
