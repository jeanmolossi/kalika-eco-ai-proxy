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
		// em caso de divergencia, faz um fallback seguro
		out := make([]pkgllm.ChatMessage, len(rewritten))
		for i, c := range rewritten {
			out[i] = pkgllm.ChatMessage{
				Role:    pkgllm.RoleUser,
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
