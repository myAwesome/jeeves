# jeeves

A CLI client for your personal diary, powered by [pa2023](https://github.com/myAwesome/pa2023).

## Features

- Interactive REPL interface
- Write diary entries directly in the console (no editor required)
- Flexible date selection: today, yesterday, or any custom date
- Read recent posts
- Search entries by content
- Browse history by month
- See entries written on this day in previous years

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

## Usage

```
$ jeeves

     _
    | | ___  _____   _____  ___
    | |/ _ \/ _ \ \ / / _ \/ __|
 _  | |  __/  __/\ V /  __/\__ \
(_)_/ |\___|\___| \_/ \___||___/
    |__/

  Your personal diary. Type 'help' for commands.

> login
Email: you@example.com
Password: ****
Logged in.

> post
Date: 2026-03-16
Body (enter '.' on empty line to finish):
  Today was a good day.
  Went for a walk in the morning.
  .
Posted! (id: 42)

> post yesterday
Date: 2026-03-15
Body (enter '.' on empty line to finish):
  Forgot to write yesterday.
  .
Posted! (id: 43)

> post 2026-03-10
Date: 2026-03-10
Body (enter '.' on empty line to finish):
  A custom date entry.
  .
Posted! (id: 44)

> read
# shows last 10 entries

> read 25
# shows last 25 entries

> search morning coffee
# full-text search

> history
# lists months that have entries

> today
# shows entries written on this day in previous years

> logout
> exit
```

## Commands

| Command              | Description                                      |
|----------------------|--------------------------------------------------|
| `login`              | Authenticate with email + password              |
| `logout`             | Clear saved session                             |
| `post [date]`        | Write a new diary entry (see Date selection)    |
| `write [date]`       | Alias for `post`                                |
| `new [date]`         | Alias for `post`                                |
| `read [N]`           | Show last N entries (default: 10)               |
| `search <text>`      | Search entries by body content                  |
| `history`            | Browse months with entries                      |
| `today`              | Entries from this day in past years             |
| `help`               | Show command list                               |
| `exit`               | Quit                                            |

## Date selection

The `post` command accepts an optional date argument:

| Argument       | Result                        |
|----------------|-------------------------------|
| _(none)_       | Today (default)               |
| `yesterday`    | Yesterday                     |
| `y`            | Yesterday (short form)        |
| `YYYY-MM-DD`   | Any specific date             |

## Dev mode

Run with `--dev` to log all HTTP requests and responses:

```bash
jeeves --dev
```
