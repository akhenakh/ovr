# Tools to evaluate for ovr

Candidate transformations from the IT-tools-style catalog (token/hash/crypto,
converters, web helpers, network, text, data), filtered against what ovr's
action model can actually do well.

Effort: **S** ≈ under an hour, **M** ≈ half a day, **L** ≈ a day or more.

---

## What decides the fit

- **Actions are pure transforms.** Every action is `fmt → fmt` with typed
  parameters (`Definition[I, O]`), and **undo replays the whole stack from the
  raw input**. Actions that invent data (random, key generation, "now") must
  record their result — see below.
- **Single input, linear chain.** A tool that needs two documents (diff) has to
  take the second one as a parameter, or compare against the raw input.
- **Text, list, table, dict, time, geometry, bin** are the formats. A tool whose
  output is an image, a timer or an interactive widget does not fit the TUI, and
  only partially fits ovrui.
- **Action names are the UI**: they are typed, filtered and shown in the chain
  bubbles. Short, lowercase, one concept per action. Names must be unique per
  input format (duplicate registration panics).
- **Both frontends render `Data.String()`.** Text, dict (JSON) and table (CSV)
  all display well; anything else needs a rendering story first.
- **CGO is banned**, so pure-Go dependencies only.

## Deterministic replay (implemented)

Actions that invent data mark their definition `NonDeterministic` (or are
generators). The produced value is recorded on the stack entry, so undo replays
the same output instead of generating a new one. `repeat` runs the action
without its recording and records the whole list. This covers `uuid4`, `uuid7`,
`now`, `password`, `shuffle`, `token`, `newula`, `newrsa`, `bcrypt` and
`encrypt`, and it fixed the old undo quirk for the first four.

## Remaining prerequisites

1. **Two-input convention.** Decide that "compare against" and "replace with"
   take their second document as a parameter (long values via the ovrui edit
   modal). Needed for diff, JSON diff and regex replace.
2. **Encoding coverage.** URL-safe base64 and a `bin → text` path for
   `base64`/`tobase64` (Base64 file converter).
3. **Optional: a `color` format.** Nice for a color converter and for
   theme/style generators; text-only works fine to start.

## Implemented in this pass

| Action | What it does | Dependency |
|---|---|---|
| `intbase` | integer between bases 2–36, prefix auto-detect | stdlib `math/big` |
| `urlencode`, `urldecode` | query/path percent encoding | stdlib |
| `htmlescape`, `htmlunescape` | HTML entities | stdlib |
| `tobinary`, `frombinary` | text ↔ 8-bit binary groups | stdlib |
| `slug` | accent folding, punctuation, hyphens | `x/text/unicode/norm` |
| `replace` | regexp replace with `$1` groups | stdlib |
| `stats` | words, sentences, paragraphs, entropy, reading time | stdlib |
| `convert` | length, mass, data, time and temperature units | stdlib |
| `ipv4` | dotted ↔ decimal ↔ hex ↔ binary | `net/netip` |
| `subnet` | network, netmask, wildcard, broadcast, host range | `net/netip`, `math/big` |
| `ipv6` | compressed/expanded, prefix, reverse DNS name | `net/netip` |
| `newula` | random `fd00::/8` /48 prefix | `crypto/rand` |
| `toyaml` (dict) | YAML output for dicts, closing the YAML asymmetry | `yaml/v3` |
| `xmlpretty` | XML indentation and validation | `encoding/xml` |
| `markdown` | Markdown → HTML | `github.com/yuin/goldmark` |
| `uuid5` | deterministic name-based UUID | stdlib SHA-1 |
| `token` | random token, hex/base64/base64url | `crypto/rand` |
| `bcrypt`, `bcryptcheck` | password hash and verification | `golang.org/x/crypto` |
| `encrypt`, `decrypt` | AES-256-GCM envelope with a scrypt-derived key | `golang.org/x/crypto` |
| `newrsa` | RSA key pair as PEM, 2048–8192 bits | stdlib `crypto/rsa` |

`temp` was dropped: `convert` covers it with `c>f`, `f>c`, `k>c`, ….

## Already covered before this pass — do not re-implement

| From the list | ovr today |
|---|---|
| Hash text (MD5/SHA1/SHA256/SHA512/CRC32) | `md5`, `sha1`, `sha256`, `sha512`, `crc32` |
| Hmac generator | `hmac` (SHA256) |
| UUID generator | `uuid4`, `uuid7` |
| Password generator | `password` (length, digits, symbols) |
| Base64 string encoder/decoder | `base64` (decode), `tobase64` (encode) |
| Hex encoder/decoder | `hex` (decode), `tohex` (encode) |
| JSON prettify / minify / formatter | `jsonpretty`, `jsoncompact` |
| JSON → YAML / YAML → JSON | `dict` + `tojson` / `toyaml` |
| Timestamp / date-time converter | `date`, `epoch`, `iso`, `tojsondate`, `tz` (+ ~20 zones), `adddur`, `humanize` |
| List sorter / list converter | `sort`, `unique`, `reverse`, `index`, `first`, `last`, `join`, `tojson`, `toyaml` |
| Case converter | `upper`, `lower`, `title`, `camel`, `snake`, `kebab`, `pascal` |
| Math evaluator | `calc`, `newcalc` (`%` is modulo, not percentage) |
| JWT parser | `jwt` (decode only, no verification) |
| CSV parse | `csv`, `sortcol`, `tocsv` |
| String escaper | `quote`, `unquote`, `unescape` (`\n`, `\t` only) |
| Text statistics | `count`, `countlines` (now also `stats`) |

---

## Tier 1 candidates — no new dependency

Most are an afternoon's work (**S**); the few that are not are marked.

| Tool | Proposed action | Path | Notes |
|---|---|---|---|
| Roman numeral converter | `roman`, `fromroman` | t→t | 1–3999, both directions |
| Text to NATO alphabet | `nato` | t→t | static table; digits and punctuation spelled out too |
| Text to Unicode | `tounicode`, `fromunicode` | t→t | `\uXXXX`/`\UXXXXXXXX`; also bare code points |
| Numeronym | `numeronym` | t→t | `kubernetes` → `k8s`, `internationalization` → `i18n` |
| Email normalizer | `emailnorm` | t→t | lowercase domain, strip `+tag`, drop dots in gmail local part; fan out over a text list |
| Chmod calculator | `chmod` | t→t | octal ↔ `rwxr-xr-x` ↔ symbolic (`g+w`); explain each bit |
| HTTP status lookup | `status` | t→t | embedded `code → phrase + description` table (both directions) |
| MIME type lookup | `mime` | t→t | extension ↔ type (`mime.TypeByExtension`), plus content sniff reusing `GuessContentType` |
| Basic auth header | `basicauth` | t→t | `user:pass` → `Authorization: Basic …`; param for header vs bare token |
| Outlook Safelink decoder | `safelink` | t→t | base64 unwrap of `safelinks.protection.outlook.com` URLs |
| Color converter | `color` | t→t | hex ↔ `rgb()` ↔ `hsl()`; detect input, param picks output form |
| MAC normalization | `mac` | t→t | `aabb.ccdd.eeff`, `AA-BB-…`, `aabbccddeeff` ↔ each other; uppercase/lowercase |
| Case extensions | `constant`, `dot`, `sentence` | t→t | `CONSTANT_CASE`, `dot.case`, `Sentence case` |
| Unicode whitespace/zero-width cleanup | `clean` | t→t | strip ZWSP/BOM/soft hyphen, normalize spaces — common pasted-data fix |
| Regex extract (non-interactive) | `matches` | t→l | all RE2 matches as a list; the useful half of "regex tester" |
| Random port | `newport` | t→t | 1–65535 (or 1024+) |
| MAC generator | `newmac` | t→t | locally administered unicast bit set |
| Lorem ipsum | `lorem` | t→t | params words/paragraphs |
| TOTP/HOTP generator | `totp`, `hotp` | t→t | stdlib HMAC + `base32`; params secret, digits, period, timestamp (deterministic for a given timestamp) |
| SVG placeholder | `svgplaceholder` | t→t | width/height/label template |
| Docker run → compose | `tocompose` | t→t | flag parser → compose YAML; pure but fiddly (**M**) |
| SVG to JSX | `tojsx` | t→t | attribute renames (`stroke-width` → `strokeWidth`), `style` → object |
| JSON → TypeScript | `tots` | d→t | interfaces from a JSON sample; nested objects, arrays, optionals |

## Tier 2 candidates — one small pure-Go dependency (M)

| Tool | Proposed action | Path | Dependency / notes |
|---|---|---|---|
| TOML ↔ JSON/YAML | `toml`, `totoml` | t→d, d→t | `github.com/pelletier/go-toml/v2` |
| Base64 URL-safe + file | `b64url`, bin support for `base64`/`tobase64` | t↔b | no dep, but touches `Format` plumbing (**M**) |
| Minify CSS/JS/HTML | `minify` | t→t | `github.com/tdewolff/minify/v2` (already in the module graph) |
| Text diff | `diff` | t→t | `github.com/pmezard/go-difflib` (already in the module graph); second doc as parameter |
| JSON diff | `jsondiff` | d→t | no dep; recursive walk, output as a readable report or unified diff |
| XML ↔ JSON | `xmltojson`, `jsontoxml` | t→d, d→t | `github.com/clbanning/mxj/v2` — attribute/array ambiguity needs a documented mapping |
| JSON → CSV | `jsontocsv` | d→r | no dep; arrays of objects → table, then existing `tocsv` |
| Cron describe / build | `crondesc` | t→t | own 5-field parser (~100 LOC) or `github.com/robfig/cron/v3` |
| ULID generator | `ulid` | t→t | `github.com/oklog/ulid/v2` |
| SHA-3 / BLAKE2b hashes | `sha3`, `blake2b` | t→t | `golang.org/x/crypto` (add alongside the SHA family) |
| QR code (block art) | `qrcode` | t→t | `github.com/skip2/go-qrcode`, render `▀▄█ ` half-blocks so it works in the TUI |
| WiFi QR payload | `wifiqr` | t→t | same renderer, `WIFI:T:WPA;S:ssid;P:pass;;` |
| BIP39 passphrase | `bip39` | t→t | `github.com/tyler-smith/go-bip39` (English wordlist only, to keep the binary small) |
| IBAN validate/parse | `iban` | t→d | no dep: checksum + embedded country-length table; returns country, bank, account, validity |
| Phone parse/format | `phone` | t→d | `github.com/nyaruka/phonenumbers` (large metadata; consider a build tag) |
| User-agent parser | `ua` | t→d | `github.com/mssola/user_agent` |
| Password strength | `strength` | t→t | no dep: entropy + character-class heuristic; zxcvbn is tempting but a heavy wordlist |
| SQL format | `sqlpretty` | t→t | risky: no maintained cgo-free formatter — either a conservative keyword/indent formatter (S) or skip |
| ASCII art text | `figlet` | t→t | needs embedded font assets (small subset of standard.flf) |

## Tier 3 — needs a new format or UI work (L)

| Tool | Why it is expensive | Recommendation |
|---|---|---|
| Color converter with real semantics | a `color` format would enable palettes, CSS gradient/shadow/etc. generators | do the text-level `color` first, add the format only if generators follow |
| CSS gradient / shadow / clip-path / glassmorphism / timing / grid | parameter-heavy generators that want sliders and previews, not text params | ovrui-only panels later; the pure string builders are S once the panel exists |
| Favicon / robots / meta / Open Graph generators | template output, low value without a form | low priority |
| Image tools (resize/convert), camera recorder | needs an image format plus TUI rendering | out of scope for the action model |
| PDF signature checker | PDF parsing + PKCS#7 chain validation, heavy pure-Go surface | park; revisit if a maintained CGO-free lib appears |
| HTML WYSIWYG, Regex tester UI, diff viewer UI, chronometer, device info | interactive/stateful UI, not a transform | keep only the transform half (extract, diff, markdown) |
| Crontab builder with a form | the parser (`crondesc`) is the useful part | ship `crondesc`, skip the form |
| Docker compose / Bash / benchmark script builders | template scaffolds, each needs its own UX | low priority after the converters |

---

## Suggested next batch

Highest value per line of code, still dependency-free:

1. `matches` (regex extract), `clean`, `emailnorm`, `numeronym` — text cleanup
   that chains well with `replace`.
2. `status`, `mime`, `chmod`, `basicauth`, `safelink`, `nato`, `roman` —
   lookups and conversions with embedded data.
3. `uuid5` is done; next deterministic crypto is `totp`.
4. `toyaml` is done; then TOML (`toml`, `totoml`) closes the config-format
   triangle, and `jsontocsv` makes dict output usable as a table.
