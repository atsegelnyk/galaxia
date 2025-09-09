# 🌌 Galaxia — Telegram Bot Constructor

**Galaxia** is a **flexible Telegram bot constructor** written in go for building complex conversational flows based on **entities** and **stages**.  
It provides a high-level abstraction over [Telegram Bot API](https://core.telegram.org/bots/api) and manages **state**, **sessions**, **actions**, and **callback workflows** seamlessly.

---

## ✨ Concept

The **core object** is the [`EntityRegistry`](./entityregistry/registry.go), where users **register workflow entities** like:

- **Commands** — entrypoints into workflows.
- **Stages** — structured conversational steps.
- **Actions** — user-triggered events that return responses.
- **Callbacks** — inline button interactions.

Everything is orchestrated via the **`Processor`**, which acts as the main engine:

```go
gp := processor.NewGalaxiaProcessor(
    processor.WithEntityRegistry(entityReg),
    processor.WithBotToken("<TOKEN>"),
    processor.WithAuthHandler(Auther{}),
    processor.WithSessionRepository(session.NewInMemorySessionRepository()),
)

gp.Start(context.Background())
```

---

## 🚀 Features

- 🔹 **Entity-driven design** — register commands, stages, callbacks, and actions.
- 🔹 **Session management** — built-in in-memory sessions; pluggable backends supported.
- 🔹 **Declarative workflows** — define complex conversation flows in a structured way.
- 🔹 **Inline & reply keyboards** — built-in helpers for user navigation.
- 🔹 **Callback query support** — seamless handling of inline button interactions.
- 🔹 **Authentication hooks** — custom authorization per user.

---

## 📦 Installation

```bash
go get github.com/atsegelnyk/galaxia
```

---

## 🧩 Abstractions

### **1. Processor**
Main engine managing **updates**, **sessions**, and **entity workflows**.

```go
gp := processor.NewGalaxiaProcessor(
    processor.WithEntityRegistry(entityReg),
    processor.WithBotToken("<TOKEN>"),
    processor.WithAuthHandler(Auther{}),
    processor.WithSessionRepository(session.NewInMemorySessionRepository()),
)
gp.Start(context.Background())
```

---

### **2. Entity Registry**
A central object where you register:

- **Commands** → `/start`
- **Stages** → multi-step conversations
- **Actions** → message-driven handlers
- **Callbacks** → inline button responses

Example:

```go
entityReg := entityregistry.New()
_ = entityReg.RegisterCommand(model.NewCommand("start", startHandler))
_ = entityReg.RegisterCallbackHandler(greetCallbackHandler)
_ = entityReg.RegisterStage(defaultStage)
```

---

### **3. Stage**
Represents a **single step** in a workflow.

```go
stage := model.NewStage("default",
    model.WithInitializer(model.NewStaticStageInitializer(
        model.NewMessage(model.WithText("Choose an option:"), model.WithReplyKeyboard(keyboard)),
    )),
)
_ = entityReg.RegisterStage(stage)
```

#### Stage Components:
- **Initializer** → message or keyboard shown when entering the stage.
- **Input handler** → optional function to process free-text responses.
- **Transitions** → move between stages dynamically.

---

### **4. Command**
An **entrypoint** for workflow execution.

```go
startCmd := model.NewCommand("start", startHandler)
_ = entityReg.RegisterCommand(startCmd)
```

⚠️ **Note**: `/start` **must** be registered, or the processor won’t start.

---

### **5. Action**
Triggered when a user **presses a button** or **sends a message** within a stage.

```go
nameButton := model.NewReplyButton("name").LinkAction(nameHandler)
ageButton := model.NewReplyButton("age").LinkAction(ageHandler)
```

An action handler **returns a `Responser`**, which defines what to send back to the user.

---

### **6. Callback**
Inline button **clicks** that respond via callback queries.

```go
greetHandler := model.NewCallbackHandler("greet", func(ctx *model.UserContext, update *tgbotapi.Update) model.Responser {
    return model.NewUserResponse(
        int64(update.CallbackQuery.From.ID),
        model.WithCallbackQueryResponse(&model.CallbackQueryResponse{
            Text: "Hello " + ctx.Name + "!",
            CallbackQueryID: update.CallbackQuery.ID,
        }),
    )
})
_ = entityReg.RegisterCallbackHandler(greetHandler)
```

---


## 🔒 Authentication

Implement `AuthHandler` to restrict access:

```go
type Auther struct{}

func (a Auther) Authorize(userID int64) error {
    if userID != 1234 && userID != 4321 {
        return errors.New("Unauthorized")
    }
    return nil
}
```

---

## 🧠 Future Plans

- [ ] Persistent session storage (Redis, Postgres)
- [ ] Middleware hooks for logging & metrics
- [ ] Dynamic stage injection from external config
