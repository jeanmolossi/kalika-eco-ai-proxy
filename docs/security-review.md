# Revisão de Segurança e Conceitos

## Considerações de Revisão (Segurança)

1. **Vazamento de causa em erros**: `httpx.WriteProblem` sempre popula o campo `Cause` com a mensagem de erro encapsulada, mesmo quando a causa é um erro de infraestrutura ou banco. Esse detalhe volta para o cliente e pode expor informações sensíveis (ex.: mensagens SQL ou erros do provedor). O ideal é suprimir ou redigir a causa nas respostas e manter o detalhe completo apenas em logs.
2. **Mensagens de autenticação inconsistentes**: O handler de embeddings retorna `Unauthorized` usando o erro bruto de `Tenants.FindByAPIKey`, revelando diferenças entre chave inválida, tenant inativo ou falha interna. O endpoint de chat usa mensagem fixa. Normalizar a resposta para todos os caminhos evita enumeração de chaves e exposição de falhas internas.
3. **CORS aberto por padrão**: O servidor HTTP inicia com `AllowOrigins: ["*"]`, permitindo requisições de qualquer origem. Em APIs protegidas por chave isso amplia o risco caso a chave vaze em frontends e dificulta mitigação de CSRF. Prefira allowlists explícitas por ambiente/tenant e desabilite credenciais quando não forem necessárias.
4. **Sem validação de modelos permitidos**: O roteador simples aceita qualquer modelo solicitado (ou fallback stub) sem checar `ModelsAllowed` ou limites de plano. Isso permite uso de modelos não aprovados ou caros, gerando custo indevido ou violando restrições contratuais. É necessário validar o modelo antes de despachar para o cliente de LLM.

## Recomendações (Ação)
- Remover ou ocultar detalhes de `Cause` em respostas e mantê-los apenas em logs/metrics estruturados.
- Padronizar mensagens de falha de autenticação e registrar a causa real apenas no servidor.
- Substituir o coringa CORS por allowlist configurada por ambiente e negar credenciais por padrão.
- Validar modelos contra a política do tenant e retornar `403`/`400` quando o modelo não for permitido; definir defaults por tenant em vez de fallbacks genéricos.

## Documentação (PT-BR)

### Guardrails como firewall de LLM

O mecanismo de guardrails funciona como um firewall de políticas que inspeciona o tráfego de cada tenant antes e depois das chamadas de LLM:

- Constrói um `guardrails.Context` neutro com tenant, chave de API, endpoint, modelo e mensagens normalizadas de entrada/saída para auditoria e avaliação de políticas.【F:internal/platform/guardrails/guardrails.go†L16-L48】
- Carrega as regras do tenant para fase de entrada ou saída, ordena por prioridade e aplica em sequência, garantindo que regras determinísticas rodem antes das mais amplas.【F:internal/platform/guardrails/simple_engine.go†L20-L56】【F:internal/platform/guardrails/simple_engine.go†L64-L77】
- Aplica ações típicas de firewall: **block** (interrompe e retorna erro), **rewrite** (saneia ou reduz conteúdo) ou **allow** (segue o fluxo). A decisão inclui IDs de regra para rastreabilidade e auditoria.【F:internal/platform/guardrails/guardrails.go†L28-L47】【F:internal/platform/guardrails/simple_engine.go†L77-L107】
- Suporta regras de regex para bloquear/regravar e limites máximos de comprimento em payloads de entrada e saída, permitindo filtrar prompts/respostas ou impor tetos de tamanho antes de chegar ao modelo ou ao cliente.【F:internal/platform/guardrails/simple_engine.go†L79-L144】

### Melhorias para o firewall

- **Allowlist/Denylist explícitos**: Adicionar semântica de primeira correspondência para permitir ou negar padrões aprovados (ex.: tenant/endpoint/modelo) antes de regravar via regex; hoje o padrão é permitir quando não há match.
- **Contexto mais rico**: Enriquecer `Context` com papéis das mensagens (system/user/assistant), tipos de mídia e flags de streaming para diferenciar prompts de respostas e detectar modalidades inseguras (imagens, áudio). Hoje apenas strings concatenadas são inspecionadas.【F:internal/platform/guardrails/guardrails.go†L30-L40】
- **Rigor nos erros**: Garantir que decisões de block/rewrite cheguem aos handlers HTTP e evitar “logar e permitir” regras inválidas. Rejeitar regras malformadas no carregamento em vez de seguir com best-effort.【F:internal/platform/guardrails/simple_engine.go†L79-L109】
- **Governança de políticas**: Versionar regras, exigir `IsActive` e validar regex para impedir execução de itens desativados ou malformados. Incluir modo dry-run e amostragem para avaliar impacto antes de aplicar.
- **Sensores específicos de LLM**: Integrar classificadores/detetores (PII, segredos, jailbreak) antes das regex, alimentando ações de block/rewrite. Combinar com severidade por tenant e negar por padrão se o detector falhar.
- **Telemetria completa**: Emitir logs de auditoria estruturados com `RuleIDs`, IDs de requisição e diffs de conteúdo reescrito (com redactions) para forense e compliance. Adicionar métricas de taxas de block/rewrite por tenant para detectar abuso ou má configuração.
