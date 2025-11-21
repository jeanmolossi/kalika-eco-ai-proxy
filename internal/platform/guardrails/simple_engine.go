package guardrails

import (
	"context"
	"log/slog"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type simpleEngine struct {
	repo   RuleRepository
	logger *slog.Logger
}

func NewSimpleEngine(repo RuleRepository, logger *slog.Logger) Engine {
	if logger == nil {
		logger = slog.Default()
	}

	return &simpleEngine{
		repo:   repo,
		logger: logger.With("component", "guardrails.simpleEngine"),
	}
}

func (e *simpleEngine) EvaluateInput(ctx context.Context, gx Context) (Decision, error) {
	rules, err := e.repo.ListRulesForTenant(ctx, gx.TenantID, PhaseInput)
	if err != nil {
		return Decision{}, err
	}

	return e.applyRules(ctx, gx, rules, true)
}

func (e *simpleEngine) EvaluateOutput(ctx context.Context, gx Context) (Decision, error) {
	rules, err := e.repo.ListRulesForTenant(ctx, gx.TenantID, PhaseOutput)
	if err != nil {
		return Decision{}, err
	}

	return e.applyRules(ctx, gx, rules, false)
}

func (e *simpleEngine) applyRules(
	ctx context.Context,
	gx Context,
	rules []Rule,
	isInput bool,
) (Decision, error) {
	// Ordena por prioridade ascendente
	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	decision := Decision{
		Action:   ActionAllow,
		Metadata: map[string]any{},
	}

	messages := gx.InputMessages
	if !isInput {
		messages = gx.OutputMessages
	}

	rewritten := make([]string, len(messages))
	copy(rewritten, messages)

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		switch rule.Kind {
		case RuleKindRegexBlock:
			hit, err := ruleMatchesRegex(rule, messages)
			if err != nil {
				e.logger.Warn("invalid regex rule", "rule_id", rule.ID, "err", err)
				continue
			}

			if hit {
				decision.Action = ActionBlock
				decision.Reason = "blocked_by_regex"
				decision.RuleIDs = append(decision.RuleIDs, rule.ID)

				return decision, nil
			}

		case RuleKindRegexRewrite:
			rw, hit, err := ruleRewriteRegex(rule, rewritten)
			if err != nil {
				e.logger.Warn("invalid regex rewrite rule", "rule_id", rule.ID, "err", err)
				continue
			}

			if hit {
				rewritten = rw
				decision.Action = ActionRewrite
				decision.Reason = "rewritten_by_regex"
				decision.RuleIDs = append(decision.RuleIDs, rule.ID)
			}

		case RuleKindMaxLength:
			maxLen, err := strconv.Atoi(strings.TrimSpace(rule.Pattern))
			if err != nil {
				e.logger.Warn("invalid max_length rule", "rule_id", rule.ID, "pattern", rule.Pattern, "err", err)
				continue
			}

			totalLen := 0
			for _, m := range rewritten {
				totalLen += len(m)
			}

			if totalLen > maxLen {
				switch rule.Action {
				case ActionBlock:
					decision.Action = ActionBlock
					decision.Reason = "blocked_by_max_length"
					decision.RuleIDs = append(decision.RuleIDs, rule.ID)

					return decision, nil
				case ActionRewrite:
					trimmed := trimToMaxLength(rewritten, maxLen)
					rewritten = trimmed
					decision.Action = ActionRewrite
					decision.Reason = "trimmed_by_max_length"
					decision.RuleIDs = append(decision.RuleIDs, rule.ID)
				default:
					// se action == allow, só loga
					e.logger.Info("max_length rule matched but action=allow", "rule_id", rule.ID)
				}
			}
		}
	}

	if decision.Action == ActionRewrite {
		if isInput {
			decision.RewrittenInputMessages = rewritten
		} else {
			decision.RewrittenOutputMessages = rewritten
		}
	}

	return decision, nil
}

func ruleMatchesRegex(rule Rule, msgs []string) (bool, error) {
	re, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return false, err
	}

	return slices.ContainsFunc(msgs, re.MatchString), nil
}

func ruleRewriteRegex(rule Rule, msgs []string) ([]string, bool, error) {
	re, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return nil, false, err
	}

	hit := false

	out := make([]string, len(msgs))
	for i, m := range msgs {
		if re.MatchString(m) {
			hit = true
			out[i] = re.ReplaceAllString(m, rule.Replacement)
		} else {
			out[i] = m
		}
	}

	return out, hit, nil
}

func trimToMaxLength(msgs []string, max int) []string {
	out := make([]string, 0, len(msgs))

	remaining := max
	for _, m := range msgs {
		if remaining <= 0 {
			break
		}

		if len(m) <= remaining {
			out = append(out, m)
			remaining -= len(m)
			continue
		}

		out = append(out, m[:remaining])
		remaining = 0
	}

	return out
}
