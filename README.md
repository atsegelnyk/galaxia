# 🌌 Galaxia — Telegram Bot Constructor

**Galaxia** is a flexible **Go** library for building Telegram bots with **declarative conversational flows**: register entities (commands, stages, actions, callbacks) in an **Entity Registry** and let the **Processor** run the show.

> General workflow: define abstractions → register them in the Entity Registry → start the `Processor`.  
> You can override behaviors per user to provide private functionality.

---

## Table of Contents
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Core Concepts](#-core-concepts)
    - [Processor](#processor)
    - [Entity Registry](#entity-registry)
    - [Abstractions](#abstractions)
- [Usage Examples](#usage-examples)
    - [Action](#action)
    - [Command](#command)
    - [Stage](#stage)
    - [Callback Handler](#callback-handler)
- [Features](#-features)
- [Authentication](#-authentication)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Roadmap](#-roadmap)
- [License](#license)

---

## 📦 Installation

```bash
go get github.com/atsegelnyk/galaxia
```

**Requirements**
- Go ≥ 1.24
- Telegram Bot API client:  
  `github.com/go-telegram-bot-api/telegram-bot-api`

---

## ⚡ Quick Start

```go
package main

import (
  "context"
  "github.com/atsegelnyk/galaxia/entityregistry"
  tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

  "github.com/atsegelnyk/galaxia"
  "github.com/atsegelnyk/galaxia/model"
  "github.com/atsegelnyk/galaxia/session"
)

func main() {
  // 1) Create the registry
  entityReg := entityregistry.New()

  // 2) Define an action
  startAction := model.NewAction("start", func(ctx *model.UserContext, update *tgbotapi.Update) model.Updater {
    msg := model.NewMessage(model.WithText("Hello, World!"))
    return model.NewUserUpdate(
      update.Message.Chat.ID,
      model.WithMessages(msg),
      model.WithTransit("main", false), // go to "main" stage; don't clean old messages
    )
  })

  // 3) Register action + command
  _ = entityReg.RegisterAction(startAction)
  startCmd := model.NewCommand("start", startAction.SelfRef())
  _ = entityReg.RegisterCommand(startCmd)

  // 4) Minimal stage
  stage := model.NewStage("main",
    model.WithInitializer(model.NewStaticStageInitializer(
      model.NewMessage(model.WithText("Choose an option from the keyboard.")),
    )),
  )
  _ = entityReg.RegisterStage(stage)

  // 5) Start processor
  gp := galaxia.NewProcessor(
    galaxia.WithEntityRegistry(entityReg),
    galaxia.WithBotToken("<TOKEN>"),
    galaxia.WithSessionRepository(session.NewInMemorySessionRepository()),
  )
  gp.Start(context.Background())
}
```

> ⚠️ **Required:** `/start` must be registered or the processor will refuse to start.

---

## 🧠 Core Concepts

### Processor
The main engine: consumes **updates**, manages **sessions**, and orchestrates **entity workflows**.

```go
gp := processor.NewGalaxiaProcessor(
    processor.WithEntityRegistry(entityReg),
    processor.WithBotToken("<TOKEN>"),
    processor.WithSessionRepository(session.NewInMemorySessionRepository()),
)
gp.Start(context.Background())
```

### Entity Registry
Central place to register:
- **Commands** — e.g. `start`
- **Stages** — multi-step conversations
- **Actions** — message-driven handlers
- **Callbacks** — inline button responses

### Abstractions
- **Actions** — user-triggered handlers that **return** a `model.Updater`.
- **Commands** — named entrypoints (usually Telegram slash commands).
- **Stages** — structured steps; each stage has an **initializer**.
- **Callbacks** — actions invoked by inline button presses.

---

## Usage Examples

### Action

`Action` is a function with signature:
```go
func(ctx *model.UserContext, update *tgbotapi.Update) model.Updater
```

`UserContext` includes fields like `UserID`, `Username`, `Name`, `LastName`, `Lang`, and `Misc`.  
`Misc` may contain arbitrary data but **must be JSON/Proto-marshalable** if you plan to persist or sync across clusters.

```go
startAction := model.NewAction("start", func(ctx *model.UserContext, update *tgbotapi.Update) model.Updater {
	msg := model.NewMessage(
		model.WithText("Hello World!"),
	)
	return model.NewUserUpdate(
		update.Message.Chat.ID,
		model.WithMessages(msg),
		model.WithTransit("main", false),
	)
})
// Register it
_ = entityReg.RegisterAction(startAction)
```

> You can return your own `Updater` implementation if you need custom in-flight behavior.

---

### Command

Bind a command name to an action:

```go
startHandler := model.NewAction("start", func(ctx *model.UserContext, update *tgbotapi.Update) model.Updater {
	msg := model.NewMessage(model.WithText("Hello World!"))
	return model.NewUserUpdate(
		update.Message.Chat.ID,
		model.WithMessages(msg),
		model.WithTransit("main", false),
	)
})
_ = entityReg.RegisterAction(startHandler)

startCmd := model.NewCommand("start", startHandler.SelfRef())
_ = entityReg.RegisterCommand(startCmd)
```

- `WithTransit(stageName, clean bool)` moves the user to a stage.  
  When `clean == true`, previously sent stage messages will be deleted after the transition.

---

### Stage

A stage typically shows a keyboard and waits for user input.

```go
usernameAction := model.NewAction("get_username", func(ctx *model.UserContext, update *tgbotapi.Update) model.Updater {
	msg := model.NewMessage(model.WithText(ctx.Username))
	return model.NewUserUpdate(update.Message.Chat.ID, model.WithMessages(msg))
})
_ = entityReg.RegisterAction(usernameAction)

userLangAction := model.NewAction("get_user_lang", func(ctx *model.UserContext, update *tgbotapi.Update) model.Updater {
	msg := model.NewMessage(model.WithText(ctx.Lang))
	return model.NewUserUpdate(update.Message.Chat.ID, model.WithMessages(msg))
})
_ = entityReg.RegisterAction(userLangAction)

getUsernameBtn := model.NewReplyButton("get username").LinkAction(usernameAction.SelfRef())
getUserLangBtn := model.NewReplyButton("get user lang").LinkAction(userLangAction.SelfRef())

keyboard := model.NewKeyboard[*model.ReplyButton](model.TwoPerRow, getUsernameBtn, getUserLangBtn)

initializer := model.NewMessage(
	model.WithText("Choose an option:"),
	model.WithReplyKeyboard(keyboard),
)

stage := model.NewStage("main",
	model.WithInitializer(model.NewStaticStageInitializer(initializer)),
)
_ = entityReg.RegisterStage(stage)
```

---

### Callback Handler

Handle inline button presses (callback queries):

```go
greetAction := model.NewAction("greet", func(ctx *model.UserContext, update *tgbotapi.Update) model.Updater {
	return model.NewUserUpdate(
		int64(update.CallbackQuery.From.ID),
		model.WithCallbackQueryResponse(&model.CallbackQueryResponse{
			Text:            "Hello " + ctx.Name + "!",
			CallbackQueryID: update.CallbackQuery.ID,
		}),
	)
})
_ = entityReg.RegisterAction(greetAction)

greetCb := model.NewCallbackHandler("greet", greetAction.SelfRef())
_ = entityReg.RegisterCallbackHandler(greetCb)

// Link in a message:
btn := model.NewInlineButton("greet").LinkCallbackHandler(greetCb.SelfRef())
ik := model.NewKeyboard[*model.InlineButton](model.OnePerRow, btn)

msg := model.NewMessage(
	model.WithText("Tap the button:"),
	model.WithInlineKeyboard(ik),
)
```

---

## 🚀 Features

- **Entity-driven design** — register commands, stages, callbacks, and actions.
- **Session management** — built-in in-memory implementation; pluggable backends supported.
- **Declarative workflows** — model complex, multi-step conversations.
- **Inline & reply keyboards** — helpers for navigation.
- **Callback queries** — first-class support.
- **Authentication hooks** — custom authorization per user.

---

## 🔒 Authentication

Provide your own `Auther`:

```go
type Auther interface {
	Authorize(userID int64) error
}
```

If `Authorize` returns an error, the `Processor` will ignore the user update.

Built-ins:
- `auth.BlacklistAuther`
- `auth.WhitelistAuther`

Attach via:
```go
processor.WithAuthHandler(myAuther)
```

---

## Best Practices

- **Always register `/start`** before starting the processor.
- **Keep `Misc` serializable** (JSON/Proto) if you plan to persist or fan out across clusters.
- **Use `clean` wisely** in `WithTransit` to prevent chat clutter.
- **Isolate side effects** in actions; return `Updater` to describe what to send/do.
- **Split flows into stages** for clarity and easier testing.

---

## Troubleshooting

- **Processor won’t start**
    - Ensure `/start` command is registered.
    - Verify `BOT_TOKEN` is correct and the Telegram API is reachable.

- **Old messages not deleted**
    - Make sure you used `WithTransit(target, true)` and your session backend tracks `StageMessages`.

- **Callback button does nothing**
    - Confirm the `CallbackHandler` is registered and the inline button is linked with `.LinkCallbackHandler(...)`.

---

## 🧭 Roadmap

- [ ] Persistent session storage (e.g., Postgres)
- [ ] Middleware hooks (logging/metrics)
- [ ] Dynamic stage injection from external config

---

## License

MIT (unless stated otherwise in the repository).
