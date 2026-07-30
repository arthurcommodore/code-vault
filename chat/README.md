# Cotarpreco chat archive

This directory preserves the AI support chat removed from Cotarpreco on
2026-07-30. The source snapshot is commit
`e9d429b12d679511f01623d1300b0842397bdb9c` from the Cotarpreco repository.

No credentials, environment files, user data, database dumps, messages, or
production configuration are included.

## What the feature did

The authenticated chat combined a localized FAQ with an optional OpenAI-backed
assistant. Pro subscribers could search FAQ categories, continue a conversation
with the assistant, keep a short browser and Redis history, rate an answer, and
escalate an unresolved conversation into a human-support ticket.

The browser sent the last messages to the Go API. The backend constructed a
system prompt with a bounded summary of the user's recent products, categories,
and quotes, called OpenAI Chat Completions with `store=false`, evaluated answer
satisfaction in a second bounded call, and returned the answer plus an
escalation recommendation. Input, history, and ticket content were truncated.

## Architecture and directory map

- `backend/logic`: FAQ loading, user context, satisfaction, and ticket creation.
- `backend/gpt`: minimal OpenAI Chat Completions HTTP client and DTOs.
- `backend/models`: MongoDB support-ticket model.
- `backend/tests`: satisfaction and ticket sanitization tests.
- `routes`: localized page routes and authenticated API handlers.
- `templates`: `templ` source and its focused test.
- `javascript`: Alpine chat state, FAQ flow, API calls, and browser persistence.
- `styles`: desktop/mobile chat styling.
- `frontend/images`: navigation/page icon.
- `redis`: remote history implementation.
- `translations`: complete FAQ/UI data in Portuguese, English, and Spanish.
- `integration`: excerpts and a checklist for code that originally lived in
  shared Cotarpreco files.
- `realtime`: records the result of the SSE/realtime audit.

Generated `*_templ.go` files are intentionally not archived; regenerate them
from the `.templ` source.

The nested `go.mod` marks this archive as independent from the Cotarpreco
module. The archive is not expected to compile standalone until the shared
integration dependencies described below are supplied.

## Dependencies

- Go, Gin, `a-h/templ`, MongoDB, and `go-redis/v9`.
- Alpine.js and the shared Cotarpreco `apiJson` browser helper.
- An OpenAI-compatible Chat Completions endpoint as implemented in
  `backend/gpt/service.go`.
- Shared Cotarpreco authentication, translation, logging, rate-limit,
  subscription, repository, product, category, and quote services.

Environment variable names used by the feature:

- `OPENAI_API_KEY`
- `OPENAI_MODEL` (the original fallback was `gpt-4o-mini`)

Do not commit real values for these variables.

## Routes

The route groups added the language prefix automatically.

- `GET /:lang/chat`: standalone chat component.
- `GET /:lang/app/chat`: chat inside the authenticated application shell.
- `GET /api/chat/history`: read remote history.
- `PUT /api/chat/history`: replace remote history.
- `DELETE /api/chat/history`: clear remote history.
- `POST /api/chat/ia`: request an AI answer.
- `POST /api/chat/human-support`: create a MongoDB support ticket.

All API routes required authentication. Chat access also required current Pro
access and used per-user rate limits (`chat-ai-user` and `chat-ticket-user`).

## Redis and browser storage

Redis used a single string key containing JSON:

```text
chat:history:<userID>
```

- TTL: 7 days.
- Maximum messages: 80.
- Maximum message text: 4,000 Unicode code points.
- Operations: `GET`, `SET` with TTL, and `DEL`.
- No streams, sets, hashes, Lua scripts, locks, workers, Pub/Sub channels, or
  chat-specific subscribers were present.

The browser key was:

```text
cotarpreco:chat:<lang>
```

The last implementation used `sessionStorage`; cleanup also removed the legacy
`localStorage` form. The browser TTL marker was 7 days.

## SSE and realtime

The chat did not define or publish any chat-specific SSE/realtime event. The
shared `/sse`, Redis dispatcher, notification events, page reload events, and
subscription/payment events were not part of the chat and must remain shared
when this archive is reused selectively.

## MongoDB

The exclusive collection was:

```text
ticketSupports
```

It stored user identity snapshots, status, origin (`ai_chat`), satisfaction,
human-support need, reason, bounded conversation messages, and timestamps. No
chat-specific index creation was found in the source snapshot.

The collection is not dropped automatically. After confirming that archived
support tickets are no longer operationally or legally required, an operator
may remove `ticketSupports` manually from the appropriate database. Review
retention, audit, and privacy obligations before doing so.

Redis keys matching `chat:history:*` may likewise be removed manually. Do not
use a broad Redis deletion pattern that could affect sessions, rate limits,
notifications, SSE, payments, or other application data.

## Reusing the feature

1. Copy the backend, model, route, template, JavaScript, style, image, Redis,
   and translation files into equivalent packages/directories.
2. Reintroduce the shared integration points documented in
   `integration/REMOVAL_POINTS.md`.
3. Initialize `TicketSupportRepo` for a `ticketSupports` collection.
4. Provide an authenticated Redis client and user/session lookup.
5. Provide localized `support.json` loading from the expected working path.
6. Configure `OPENAI_API_KEY` and optionally `OPENAI_MODEL` outside source
   control.
7. Restore Pro-access checks or replace them with the target application's
   authorization rule.
8. Register the Alpine component and include its CSS/image assets.
9. Regenerate templates with `templ generate`.
10. Add privacy disclosures and retention procedures appropriate to the new
    controller and jurisdiction.
11. Run formatting, tests, static analysis, and a production build.

## Original exclusive files

```text
internal/logic/chat.go
internal/logic/chat_ai_context.go
internal/logic/chat_history.go
internal/logic/chat_satisfaction.go
internal/logic/chat_satisfaction_test.go
internal/logic/chat_support_ticket.go
internal/logic/chat_support_ticket_test.go
internal/logic/gpt/global.go
internal/logic/gpt/service.go
internal/model/ticket_support.go
internal/routes/chat_route.go
internal/routes/handlers/chat_handler.go
internal/template/chat_support.templ
internal/template/chat_support_test.go
static/js/chat.js
static/css/chat.css
static/img/chat.svg
locales/pt/support.json
locales/en/support.json
locales/es/support.json
TODO.md
```

Shared source files that contained smaller integration fragments are listed in
`integration/REMOVAL_POINTS.md`; they were not copied wholesale.
