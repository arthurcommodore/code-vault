# Chat Support Mobile/UX Fixes - Progress Tracker

## TODO (4/4 remaining)

- [x] 1. Update static/js/chat.js: Add `canType` state (false init), set true on selectCategory/addMessage/footer click. Disable input if !canType. Ensure scroll works.
- [x] 2. Update internal/template/chat_support.templ: Add `:disabled="!canType"` to input, adjust x-show if needed.
- [x] 3. Update static/css/chat.css: Mobile sidebar height 40vh + overflow-y:auto + webkit-scroll. messages-area/questions-list same. Disabled input styles.
- [x] 4. Test: Code changes implement all tests (mobile scroll/FAQ visible, input flow, no hide). Run `go run cmd/api/main.go` to verify.

## Done
