# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

@SadFavBot is a Telegram bot (Go) for saving "favs" (stickers, pictures, GIFs, voices, texts, etc.) under textual aliases and retrieving them via inline mode. User-facing docs live in the GitHub wiki; `README.md` lists the commands.

The bot is built on the author's framework [`goSadTgBot`](https://github.com/kozalosev/goSadTgBot) (sibling checkout at `../goSadTgBot`), which provides `base` (handler interfaces, `ApplicationEnv`/`RequestEnv`), `wizard` (multi-step forms backed by Redis), `app`, `server`, `storage`, `metrics` and `logconst`. Telegram API is `github.com/OvyFlash/telegram-bot-api`. Many "how does X work" answers are in that framework, not here.

## Commands

```bash
go build ./...
go test -v ./...                         # needs Docker: tests spin up Postgres via testcontainers
go test ./db/repo -run TestName -v       # single test / package
make start | stop                        # docker-compose (+ override) up --build -d / down
make dev-env                             # patch .env for local run: *_HOST=localhost, DEBUG=true
```

CI (`.github/workflows/ci-build.yml`) runs `go build` + `go test` with `TESTCONTAINERS_RYUK_DISABLED=true`, plus a Docker image build.

Configuration is entirely via env vars in `.env` (documented inline there). A new variable must also be passed in the `environment` of `docker-compose.yaml` here and of `SadFavBot/docker-compose.yml` in the server-configs repo, otherwise it silently doesn't reach the container. `DEBUG=true` → long polling; otherwise a webhook server is started (`main.go`). `.env` is tracked in git, so avoid committing local edits (e.g. from `make dev-env`).

## Architecture

- `main.go` — wiring: connects Redis (wizard state) + Postgres (runs migrations via `storage.RunMigrations`), builds the handler lists in `initHandlers`, registers bot commands and Prometheus counters. **A new command handler must be added to `messageHandlers` there**; the order matters: `NewIgnoreForeignCommandsHandler` (commands like `/help@OtherBot`) must stay first and `NewIgnoreUnknownInGroupsHandler` last (fallback).
- `strings.go` (package main) — all `en`/`ru` l10n strings registered into `locpool` (go-l10n) in `init()`. New user-facing text needs both languages. Long help/privacy texts are Markdown files in `handlers/help/*.md` and `handlers/privacy/` (`<name>.md` + `<name>.ru.md`).
- `handlers/` — one file per command (`save.go`, `link.go`, `list.go`, ...), plus `inline.go` (inline query → favs), `guest.go` + `guest_commands.go` + `guest_save.go` (Bot API 10.0 guest mode: `base.GuestHandler` returns one result and the framework sends the single `answerGuestQuery`; `GuestCommandsHandler` routes keywords like save/delete/ref/list/link/rmlink/help in `guestKeywords` and must precede the search handler; the search result gets "◀ n/N ▶" buttons handled by `GuestShowCallbackHandler` in `guest_show.go`, which edits the message by `inline_message_id` and shares `ShowCache` with `/show`) and callback handlers (`list` pagination, help). Most commands are **wizards**: a handler embeds `base.CommandHandlerTrait`, implements `GetWizardEnv`/`GetWizardDescriptor`, declares fields with validators, and supplies a form-action callback invoked when all fields are collected. Translation keys follow `commands.<cmd>.fields.<field>` / `commands.<cmd>.status.<status>` (prefix constants in `handlers/consts.go` and each handler file). `handlers/common` holds shared helpers (group-chat behaviour, repliers).
- Self-destroying messages (replies in groups, questions of the wizards, guest placeholders) are scheduled by goSadTgBot's `selfdestruct` in a Redis queue started in `main.go`, so they're deleted after a restart too. The framework only gives the mechanism; the rules of the bot are in `handlers/common/cleanup.go` (the policy of the wizards, set by `wizard.SetGroupCleanupPolicy` in `main.go`) and `repliers.go`: in group chats, the replies and the whole dialogue of a wizard are deleted together with the messages of the user only if the bot may delete messages of other users there, after the bot's messages are replaced with a warning for `SELF_DELETION_WARNING_SECS`; otherwise nothing is deleted, so that no lone messages are left.
- Metrics: the framework counts handler usages (`used_*_handler_*`), updates, languages, panics, completed wizards and Telegram API requests (via `metrics.NewTelegramHTTPClient` in `main.go`). Bot-specific counters (search results, save/install outcomes, carousel, `ShowCache`) live in `handlers/metrics.go`.
- `db/repo/` — service layer over `pgxpool` (`FavService`, aliases, links, packages, users); constructed from `*base.ApplicationEnv`. `db/dto` has the entities. `db/migrations/` — sequential golang-migrate SQL pairs (`NNNNNN_name.up.sql`/`.down.sql`); add the next number for schema changes.
- Domain model: a **fav** = (user, alias, object) where object is a file or a text (`texts` table, with entities and hash); **links** map extra phrases to an existing alias; **packages** are sets of aliases owned by a user that others can `/install` (copying objects at install time); alias **visibility** hides aliases from listings; user **search mode** toggles substring (pg_trgm) search.

## Testing

- Tests are integration-style: `db/repo/common_test.go` and `handlers/common_test.go` start a Postgres container in `TestMain`, run migrations, and share a `*pgxpool.Pool`. Shared fixtures and constants (`UID`, `Alias`, `FileID`, ...) plus `InsertTestData`/`InsertTestPackages` live in `test/` — they wipe all tables first, so call them at the start of each test.
- Handler tests build fake updates (see `handlers/helper_test.go`) and use `testify`.
