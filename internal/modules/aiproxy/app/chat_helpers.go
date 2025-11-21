package app

import "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"

func flattenChatMessages(msgs []llm.ChatMessage) []string {
	out := make([]string, 0, len(msgs))

	for _, msg := range msgs {
		out = append(out, msg.Content)
	}

	return out
}

func rebuildChatMessages(original []llm.ChatMessage, rewritten []string) []llm.ChatMessage {
	if len(original) != len(rewritten) {
		// em caso de divergencia, faz um fallback seguro
		out := make([]llm.ChatMessage, len(rewritten))
		for i, c := range rewritten {
			out[i] = llm.ChatMessage{
				Role:    llm.RoleUser,
				Content: c,
			}
		}

		return out
	}

	out := make([]llm.ChatMessage, len(original))
	for i, m := range original {
		m.Content = rewritten[i]
		out[i] = m
	}

	return out
}
