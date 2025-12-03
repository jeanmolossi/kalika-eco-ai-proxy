package app

import pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"

func flattenChatMessages(msgs []pkgllm.ChatMessage) []string {
	out := make([]string, 0, len(msgs))

	for _, msg := range msgs {
		out = append(out, msg.Content)
	}

	return out
}

func rebuildChatMessages(original []pkgllm.ChatMessage, rewritten []string) []pkgllm.ChatMessage {
	if len(original) != len(rewritten) {
		// Preserve roles where possible when the message count changes. If a
		// rewritten entry exceeds the original length, fall back to the last
		// known role (or user) instead of forcing all messages to RoleUser.
		out := make([]pkgllm.ChatMessage, len(rewritten))

		fallbackRole := pkgllm.RoleUser
		if len(original) > 0 {
			fallbackRole = original[len(original)-1].Role
		}

		for i, c := range rewritten {
			role := fallbackRole
			if i < len(original) {
				role = original[i].Role
			}

			out[i] = pkgllm.ChatMessage{
				Role:    role,
				Content: c,
			}
		}

		return out
	}

	out := make([]pkgllm.ChatMessage, len(original))

	for i, m := range original {
		m.Content = rewritten[i]
		out[i] = m
	}

	return out
}
