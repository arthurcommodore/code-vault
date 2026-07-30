# Shared Cotarpreco integration points

These fragments belonged to the chat but originally lived in shared files.
They are documented separately so unrelated Cotarpreco code is not copied into
the reusable archive.

## Application registration

`cmd/api/main.go` registered:

```go
routes.ChatActions(private)
routes.ChatPages(pages)
```

## Subscription permission

`internal/logic/subscription.go` defined `ErrChatSubscriptionRequired` and
`IsChatAccessAllowed`, which refreshed subscription usage and required current
Pro access.

## Repository and account lifecycle

`internal/repository/global.go` declared and initialized:

```go
TicketSupportRepo *MongoRepository[model.TicketSupport]
TicketSupportRepo = NewMongoRepository[model.TicketSupport](context.Background(), "ticketSupports")
```

`internal/logic/account_data.go` included `ticketsSupport` and
`currentChatBuffer` in account exports, included ticket IDs while collecting
related audit logs, and deleted the Redis chat history during account deletion.

## Shell route, navigation, titles, and SEO

- `internal/routes/layout_route.go` rendered the `chat` shell page through
  `chatComponentForUser`.
- `internal/routes/handlers/indexing_handler.go` listed `chat` as a private
  robots path.
- `internal/template/layout.templ` included `chat.css` and rendered the chat
  sidebar link/icon.
- `internal/template/page_title.go` mapped `chat` to `page_title_chat`.
- `internal/template/page_title_test.go` tested Portuguese, English, and
  Spanish chat titles.

## Browser registration and cleanup

`static/js/index.js` imported `chat.js` and registered `Alpine.data("chat",
chat)`. `static/js/layout.js` removed `cotarpreco:chat:<lang>` from both
`localStorage` and `sessionStorage` when local account data was cleared.

## Translation fragments outside support.json

Remove or restore together with the feature:

- `nav_chat` in every `layout.json`.
- `page_title_chat` in every `page_titles.json`.
- `message_chat_ai_answered`, `message_chat_ai_empty_answer`,
  `message_chat_history_cleared`, `message_chat_subscription_required`, and
  `message_human_support_ticket_created` in every `messages.json`.
- Chat/OpenAI/history/ticket disclosures in every `privacy.json`.

The complete feature-owned `support.json` files are under `translations/`.

## Configuration and documentation

- `.env.example`: `OPENAI_API_KEY` and `OPENAI_MODEL`.
- `README.md`: feature list, optional OpenAI integration, environment table,
  route lists, and secrets inventory.
- `docs/lgpd-audit.md`: chat/OpenAI/ticket data flow, retention, safeguards,
  findings, and pending decisions.

## Shared dependencies deliberately preserved in Cotarpreco

- Redis client/repository, used by sessions, rate limits, SSE, and other flows.
- Generic rate limiting (`allowHandlerAction`).
- Generic SSE/realtime dispatcher and `/sse` route.
- Notifications and payment/subscription realtime events.
- Generic `Message` model/repository/routes, used for editable quote-sharing
  messages and unrelated to chat conversations.
- Authentication, localization, logging, product/category/quote repositories,
  and account export/deletion infrastructure.
