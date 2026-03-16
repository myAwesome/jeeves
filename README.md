# jeeves

A CLI client for your personal diary, powered by [pa2023](https://github.com/myAwesome/pa2023).

## Features

- Interactive REPL interface
- Write diary entries in your `$EDITOR`
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
# opens $EDITOR (or nano as fallback)
# save and close to publish

Posted! (id: 42)

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

| Command          | Description                          |
|------------------|--------------------------------------|
| `login`          | Authenticate with email + password   |
| `logout`         | Clear saved session                  |
| `post`           | Write a new diary entry              |
| `read [N]`       | Show last N entries (default: 10)    |
| `search <text>`  | Search entries by body content       |
| `history`        | Browse months with entries           |
| `today`          | Entries from this day in past years  |
| `help`           | Show command list                    |
| `exit`           | Quit                                 |

## Editor

Set `$EDITOR` or `$VISUAL` to your preferred editor. Falls back to `nano`.

```bash
export EDITOR=vim   # or nvim, hx, code --wait, etc.
```
