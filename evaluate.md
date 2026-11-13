# Tools to evaluate for ovr

Candidate transformations from the IT-tools-style catalog (token/hash/crypto,
converters, web helpers, network, text, data), filtered against what ovr's
action model can actually do well. Pick from here; nothing is implemented yet.

Effort: **S** ≈ under an hour, **M** ≈ half a day, **L** ≈ a day or more.

---

## What decides the fit

- **Actions are pure transforms.** Every action is `fmt → fmt` with typed
  parameters (`Definition[I, O]`), and **undo replays the whole stack from the
  raw input**. An action that invents data (random, key generation, "now") is
  re-run on undo and would produce something different. That is a bug today for
  `uuid4`, `uuid7`, `now` and `shuffle`.
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

## Prerequisites worth doing first

1. **Deterministic replay for generative actions.** Store the produced value as
   an internal/snapshotted parameter so undo re-applies the same output. This
   unlocks every random/keyed tool below (token, bcrypt, ULID, RSA, MAC, IPv6
   ULA, lorem, TOTP at a fixed time) and fixes the current `uuid4`/`shuffle`
   undo quirk.
2. **`dict → YAML`.** `toyaml` exists only for lists; dicts only have `tojson`.
   One action closes YAML↔JSON and YAML prettify.
3. **Two-input convention.** Decide that "compare against" and "replace with"
   take their second document as a parameter (long values via the ovrui edit
   modal). Needed for diff, JSON diff and regex replace.
4. **Encoding coverage.** `base64`/`tobase64` and `hex`/`tohex` already exist in
   both directions, but there is no URL-safe base64 and no `bin → text` path, so
   "Base64 file converter" needs a small format pass first.
5. **Optional: a `color` format.** Nice for a color converter and for
   theme/style generators; text-only works fine to start.

## Already covered — do not re-implement

| From the list | ovr today |
|---|---|
| Hash text (MD5/SHA1/SHA256/SHA512/CRC32) | `md5`, `sha1`, `sha256`, `sha512`, `crc32` |
| Hmac generator | `hmac` (SHA256) |
| UUID generator | `uuid4`, `uuid7` |
| Password generator | `password` (length, digits, symbols) |
| Base64 string encoder/decoder | `base64` (decode), `tobase64` (encode) |
| Hex encoder/decoder | `hex` (decode), `tohex` (encode) |
| JSON prettify / minify / formatter | `jsonpretty`, `jsoncompact` |
| JSON → YAML / YAML → JSON | `dict` then `tojson` (YAML output missing) |
| Timestamp / date-time converter | `date`, `epoch`, `iso`, `tojsondate`, `tz` (+ ~20 zones), `adddur`, `humanize` |
| List sorter / list converter | `sort`, `unique`, `reverse`, `index`, `first`, `last`, `join`, `tojson`, `toyaml` |
| Case converter | `upper`, `lower`, `title`, `camel`, `snake`, `kebab`, `pascal` |
| Math evaluator | `calc`, `newcalc` (`%` is modulo, not percentage) |
| JWT parser | `jwt` (decode only, no verification) |
| CSV parse | `csv`, `sortcol`, `tocsv` |
| String escaper | `quote`, `unquote`, `unescape` (`\n`, `\t` only) |
| Text statistics | `count`, `countlines` (no words/sentences/entropy) |

---

## Tier 1 — no new dependency

Most are an afternoon's work (**S**); the few that are not are marked.

| Tool | Proposed action | Path | Notes |
|---|---|---|---|
| Integer base converter | `intbase` | t→t | `strconv.ParseInt(s, from, 64)` → format base `to`; params `from`, `to` (2–36), 0 = auto-detect `0x`/`0b`/`0o` |
| Roman numeral converter | `roman`, `fromroman` | t→t | 1–3999, both directions |
| URL encode/decode | `urlencode`, `urldecode` | t→t | `url.QueryEscape`/`PathEscape` + `Unescape`; param for the escape mode |
| Escape HTML entities | `htmlescape`, `htmlunescape` | t→t | `html.EscapeString`/`UnescapeString` |
| Text to NATO alphabet | `nato` | t→t | static table; digits and punctuation spelled out too |
| Text to ASCII binary | `tobinary`, `frombinary` | t→t | 8-bit groups; decode accepts spaced or continuous input |
| Text to Unicode | `tounicode`, `fromunicode` | t→t | `\uXXXX`/`\UXXXXXXXX`; also bare code points |
| Slugify | `slug` | t→t | accent folding via `golang.org/x/text/unicode/norm` (already a dep) |
| Numeronym | `numeronym` | t→t | `kubernetes` → `k8s`, `internationalization` → `i18n` |
| Email normalizer | `emailnorm` | t→t | lowercase domain, strip `+tag`, drop dots in gmail local part; fan out over a text list |
| Chmod calculator | `chmod` | t→t | octal ↔ `rwxr-xr-x` ↔ symbolic (`g+w`); explain each bit |
| HTTP status lookup | `status` | t→t | embedded `code → phrase + description` table (both directions) |
| MIME type lookup | `mime` | t→t | extension ↔ type (`mime.TypeByExtension`), plus content sniff reusing `GuessContentType` |
| Basic auth header | `basicauth` | t→t | `user:pass` → `Authorization: Basic …`; param for header vs bare token |
| Outlook Safelink decoder | `safelink` | t→t | base64 unwrap of `safelinks.protection.outlook.com` URLs |
| Color converter | `color` | t→t | hex ↔ `rgb()` ↔ `hsl()`; detect input, param picks output form |
| Temperature converter | `temp` | t→t | C/F/K; input `12C`, output requested unit |
| MAC normalization | `mac` | t→t | `aabb.ccdd.eeff`, `AA-BB-…`, `aabbccddeeff` ↔ each other; uppercase/lowercase |
| IPv4 address converter | `ipv4` | t→t | dotted ↔ uint32 ↔ hex ↔ binary (`net/netip`) |
| IPv4 subnet calculator | `subnet` | t→d | CIDR → network, broadcast, mask, wildcard, host range, host count as a dict |
| IPv6 prefix info | `ipv6` | t→d | `netip.Prefix` math; works for both families (ULA *generation* is in the generator batch) |
| JSON → YAML (dict) | `toyaml` | d→t | fills the asymmetry with lists |
| XML formatter | `xmlpretty` | t→t | `encoding/xml` encoder with indentation; validate as a side effect |
| Text stats extension | `stats` | t→t | words, sentences, paragraphs, unique words, reading time, Shannon entropy |
| Case extensions | `constant`, `dot`, `sentence` | t→t | `CONSTANT_CASE`, `dot.case`, `Sentence case` |
| Unicode whitespace/zero-width cleanup | `clean` | t→t | strip ZWSP/BOM/soft hyphen, normalize spaces — common pasted-data fix |
| Regex extract (non-interactive) | `matches` | t→l | all RE2 matches as a list; the useful half of "regex tester" |
| Regex replace | `replace` | t→t | `re.ReplaceAllString` with `$1` groups; params pattern, replacement |
| UUID v5 | `uuid5` | t→t | deterministic name-based UUID (namespace + input); no dependency |
| Token generator | `token` | t→t | 16/32 bytes, hex or base64url; needs prerequisite 1 |
| Random port | `newport` | t→t | 1–65535 (or 1024+); needs prerequisite 1 |
| MAC generator | `newmac` | t→t | locally administered unicast bit set; needs prerequisite 1 |
| IPv6 ULA generator | `newula` | t→t | random `fd00::/8` prefix from `crypto/rand`; needs prerequisite 1 |
| Lorem ipsum | `lorem` | t→t | params words/paragraphs; needs prerequisite 1 (or a fixed seed) |
| TOTP/HOTP generator | `totp`, `hotp` | t→t | stdlib HMAC + `base32`; params secret, digits, period, timestamp (deterministic for a given timestamp) |
| SVG placeholder | `svgplaceholder` | t→t | width/height/label template |
| Docker run → compose | `tocompose` | t→t | flag parser → compose YAML; pure but fiddly (**M**) |
| SVG to JSX | `tojsx` | t→t | attribute renames (`stroke-width` → `strokeWidth`), `style` → object |
| JSON → TypeScript | `tots` | d→t | interfaces from a JSON sample; nested objects, arrays, optionals |

## Tier 2 — one small pure-Go dependency (M)

| Tool | Proposed action | Path | Dependency / notes |
|---|---|---|---|
| TOML ↔ JSON/YAML | `toml`, `totoml` | t→d, d→t | `github.com/pelletier/go-toml/v2` |
| Markdown → HTML | `markdown` | t→t | `github.com/yuin/goldmark` (already in the module graph) |
| Base64 URL-safe + file | `b64url`, bin support for `base64`/`tobase64` | t↔b | no dep, but touches `Format` plumbing (**M**) |
| Minify CSS/JS/HTML | `minify` | t→t | `github.com/tdewolff/minify/v2` (already in the module graph) |
| Text diff | `diff` | t→t | `github.com/pmezard/go-difflib` (already in the module graph); second doc as parameter |
| JSON diff | `jsondiff` | d→t | no dep; recursive walk, output as a readable report or unified diff |
| XML ↔ JSON | `xmltojson`, `jsontoxml` | t→d, d→t | `github.com/clbanning/mxj/v2` — attribute/array ambiguity needs a documented mapping |
| JSON → CSV | `jsontocsv` | d→r | no dep; arrays of objects → table, then existing `tocsv` |
| Unit converter | `convert` | t→t | no dep; param pair (`km>mi`, `kg>lb`, `B>MiB`…) — supersedes `temp` |
| Cron describe / build | `crondesc` | t→t | own 5-field parser (~100 LOC) or `github.com/robfig/cron/v3` |
| ULID generator | `ulid` | t→t | `github.com/oklog/ulid/v2`; prerequisite 1 |
| Bcrypt | `bcrypt`, `bcryptcheck` | t→t | `golang.org/x/crypto/bcrypt`; hash is non-deterministic (prerequisite 1), verify is pure |
| AES encrypt/decrypt | `encrypt`, `decrypt` | t→t | stdlib AES-GCM + `x/crypto/scrypt` or argon2; params passphrase, KDF; base64 envelope with salt+nonce |
| RSA key pair | `newrsa` | t→t | stdlib `crypto/rsa` + `x509`; PEM output, param bits (2048/3072/4096); prerequisite 1 |
| SHA-3 / BLAKE2b hashes | `sha3`, `blake2b` | t→t | `golang.org/x/crypto` (add alongside the SHA family) |
| QR code (block art) | `qrcode` | t→t | `github.com/skip2/go-qrcode`, render `▀▄█ ` half-blocks so it works in the TUI |
| WiFi QR payload | `wifiqr` | t→t | same renderer, `WIFI:T:WPA;S:ssid;P:pass;;` |
| BIP39 passphrase | `bip39` | t→t | `github.com/tyler-smith/go-bip39` (English wordlist only, to keep the binary small); prerequisite 1 |
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

## Suggested first batch

Highest value per line of code, all pure and dependency-free:

1. `urlencode`/`urldecode`, `htmlescape`/`htmlunescape` — the most common
   pasted-data fixes, and the current `unescape` only knows `\n`/`\t`.
2. `tobinary`/`frombinary`, `tounicode`/`fromunicode`, `nato` — classic ovr
   material, a handful of lines each.
3. `slug`, `emailnorm`, `numeronym`, `clean` — text cleanup that chains well.
4. `intbase`, `roman`, `chmod`, `status`, `mime`, `basicauth`, `safelink` —
   lookups and conversions with embedded data, no parameters to explain.
5. `toyaml` for dicts — removes a real asymmetry in the current YAML support.
6. `subnet`, `ipv4`, `mac` — `net/netip` does the work.
7. `uuid5`, `totp` — deterministic cryptography with no dependency.
8. `matches`, `replace` — the scriptable core of the regex tool.

Then do prerequisite 1 (recorded generator outputs) before the random/keyed
batch: `token`, `newport`, `newmac`, `lorem`, `newula`, `ulid`, `bcrypt`,
`encrypt`, `newrsa`, `bip39`.
