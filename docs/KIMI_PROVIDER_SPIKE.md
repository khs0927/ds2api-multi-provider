# Kimi K3 provider compatibility spike

This branch adds an isolated `internal/providers/kimi` OpenAI-compatible client. It accepts only `kimi-k3` or `moonshotai/Kimi-K3`; it does not add either name to DS2API's DeepSeek alias map and it is not wired into the DeepSeek account/session runtime.

The spike preserves the complete assistant history as raw JSON, including `reasoning_content` and `tool_calls`, forwards `tools`/`tool_choice`, applies K3's fixed sampling policy (`temperature=1.0`, `top_p=0.95`), forwards `reasoning_effort` (`low|high|max`), supports completion-token budgets, and exposes a live SSE body through `Response.ReadStream` without buffering the response in `Chat`. Exact model IDs are enforced without trimming or aliasing.

It is intentionally not a production provider. A production integration still requires an independently authenticated Kimi endpoint, account ownership, rate limits, live synthetic chats, and a review of how the gateway exposes provider health without coupling Kimi credentials to DeepSeek accounts.

Run the spike tests with `go test ./internal/providers/kimi`.
