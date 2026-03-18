# jeeves

A TUI client for your personal diary, powered by [pa2023](https://github.com/myAwesome/pa2023).

## Features

- 2-column TUI: months list on the left, posts on the right
- Browse diary history by month
- Write new entries with flexible date selection
- Search entries by content
- See entries written on this day in previous years
- Mouse wheel scrolling for post content

## Installation

```bash
go install jeeves@latest
```

Or build from source:

```bash
git clone <repo>
cd jeeves
go build -o jeeves .
```

## Configuration

On first run, jeeves looks for `~/.jeeves/config.json`:

```json
{
  "base_url": "https://your-pa2023-instance.com"
}
```

If the file doesn't exist, it defaults to `http://localhost:3030`.

Your session token is stored in `~/.jeeves/session.json` (mode 0600).

## Layout

```
 Jeeves  ●  history
┌────────────────┬──────────────────────────────────────┐
│ History        │ Posts                                │
│                │                                      │
│ ▶ March        │ ▶ Mon, 17 Mar 2026  #42             │
│   2026 · 3     │   Today I went for a walk…           │
│                │                                      │
│   February     │   Sun, 16 Mar 2026  #41             │
│   2026 · 5     │   Yesterday was productive…          │
│                │                                      │
└────────────────┴──────────────────────────────────────┘
 n new  · / search  · r recent  · t today  · tab switch  · q quit
```

## Key Bindings

### Global

| Key        | Action                                |
|------------|---------------------------------------|
| `q`        | Quit                                  |
| `ctrl+c`   | Quit (always)                         |
| `n`        | New entry (compose screen)            |
| `/`        | Search entries                        |
| `r`        | Recent posts (last 30)                |
| `t`        | On this day (entries from past years) |
| `h`        | History view                          |
| `L`        | Logout                                |
| `tab`      | Switch focus between panels           |

### Navigation

| Key       | Action                       |
|-----------|------------------------------|
| `↑` / `k` | Move up                      |
| `↓` / `j` | Move down                    |
| `↵`       | Select month / open post     |
| `esc`     | Back / close post viewer     |

### Compose screen

| Key      | Action                              |
|----------|-------------------------------------|
| `ctrl+s` | Save entry                          |
| `tab`    | Switch between date and body fields |
| `esc`    | Cancel                              |

### Search screen

| Key      | Action           |
|----------|------------------|
| `↵`      | Submit search    |
| `↑` / `↓`| Navigate results |
| `esc`    | Back to history  |

## Dev mode

Run with `--dev` to log all HTTP requests and responses:

```bash
jeeves --dev
```
