# Estratégia de Modularização do Proxy de IA

## Visão de Produto
Historicamente, o proxy de IA concentrou responsabilidades operacionais no módulo `platform`, o que dificultava a evolução de cada capability de forma independente. A estratégia abaixo organiza o domínio em bounded contexts e mantém as implementações alinhadas a cada contexto dentro de `internal/{tenant,guardrails,ratelimit,cache,llm,observability,database}` para facilitar a extração futura como serviços autônomos.

## Inventário atual por bounded context
- **tenant**: gerencia stores e políticas de tenants, chave de autenticação e revogação agendada.
- **guardrails**: validação e saneamento de prompts/respostas, com potencial para políticas específicas por domínio.
- **llm**: abstrai provedores, pool, router/tokenizer e métricas de chamadas de modelos; já opera com aliases e defaults configuráveis.
- **observability**: publishers para eventos de uso e auditoria (arquivo ou Kafka) consolidados com precificação.
- **cache** e **ratelimit**: fornecem semântica de cache/no-op e rate limiting local.
- **database**: conexão Postgres compartilhada pelos módulos.
- **toolkit**: `httpx`, `logger`, `config`, `apperr` expostos como utilitários compartilhados.

## Bounded contexts propostos (DDD)
1. **Tenant & Identity**: cadastro, chaves, políticas de modelo e limites por tenant. Expõe API de gerenciamento e eventos de ciclo de vida (criação, revogação, alteração de planos).
2. **Guardrails**: políticas de moderação e transformação de prompts/respostas, configuráveis por tenant. Deve consumir eventos de configuração ou consultar o serviço de Tenant.
3. **LLM Gateway**: roteamento e orquestração de chamadas LLM, abstraindo provedores e modelos. Depende de contratos de Tenant (allowlists, defaults) e de Guardrails (políticas pré/pós chamada).
4. **Observability & Billing**: coleta de métricas de uso, auditoria e custos por tenant. Consome eventos do Gateway e do Tenant.
5. **Developer Experience** (SDK/CLI): pacotes auxiliares (`httpx`, `logger`, clients) que atuam como anti-corruption layer para consumo dos serviços.

## Backlog de modularização
1. **Isolar contratos de Tenant**
   - Extrair interfaces de store e políticas para `pkg/tenant` com eventos de domínio (TenantCreated, ApiKeyRevoked).
   - Expor API/cliente separado consumido pelo Gateway e Guardrails.
2. **Separar Guardrails como serviço**
   - Definir contrato síncrono (gRPC/HTTP) para validação e pós-processamento de prompts/respostas.
   - Criar fila/eventos para atualizações de políticas e conectar com Tenant (permissões por tenant/modelo).
3. **Formalizar o LLM Gateway**
   - Manter apenas roteamento e retries no módulo principal; mover providers e pool para `pkg/llm` com contrato público.
   - Introduzir Anti-Corruption Layer para provedores externos e mecanismo de feature flags por tenant.
4. **Observability & Billing**
   - Substituir publishers de `usage` e `audit` por um serviço próprio que recebe eventos do Gateway.
   - Padronizar esquema de eventos (request/response IDs, custos, tokens, modelo) e backends pluggable (Kafka, HTTP).
5. **Tokenizer e custos**
   - Transformar `tokenizer` em serviço/biblioteca compartilhada com catálogo de modelos/tokenizers, versionado.
   - Expor API para contagem e cálculo de custo unitário por modelo, consumida por Observability.
6. **Infra compartilhada**
   - Consolidar `config`, `logger`, `httpx` e `apperr` como toolkit comum para serviços satélites, com guidelines de logging e tracing.
7. **Migration roadmap**
   - Iniciar com extração de contratos e clients (`pkg/*`), depois mover implementação para serviços externos mantendo compatibilidade.
   - Introduzir pact tests/contratos entre Gateway ↔ Tenant/Guardrails/Observability para garantir evolução independente.

## Progresso inicial (refatoração corrente)
- Contratos de domínio re-exportados em `pkg/{tenant,guardrails,llm,tokenizer}` para que módulos e futuros serviços dependam de caminhos estáveis alinhados aos bounded contexts.
- Rotas HTTP, casos de uso e roteadores internos já consomem os contratos de `pkg/*`, preparando o código para substituição gradual por clients externos.
- Infraestrutura compartilhada de configuração, logging, servidor HTTP e erros de aplicação realocada para `pkg/toolkit` para que serviços satélites possam reutilizar contratos estáveis.
- O executável do gateway passou a viver em `apps/gateway` e o módulo de domínio em `internal/gateway`, refletindo a topologia orientada a serviços descrita na visão de diretórios.
- Eventos de uso e auditoria, além da precificação por token, foram consolidados em `pkg/observability`, permitindo que o gateway publique métricas e custos via contrato público antes da extração do serviço de billing.
- O runtime agora registra módulos separados para tenant, guardrails, rate limiting/cache, LLM/router/tokenizer, observability e database, substituindo o antigo agregador `platform` e aproximando o layout da futura divisão em serviços.
- Cada bounded context possui um executável próprio em `apps/{gateway,tenant,guardrails,observability}`, com registries mínimos para permitir a operação isolada e servir de ponte para futuras extrações como serviços externos.

## TODO para concluir o isolamento completo
- ~~**Banco e storage por serviço**: mover schemas e migrações para pastas específicas de cada app e garantir que `docker-compose` provisiona instâncias independentes ou databases separados.~~ Cada app agora aceita `*_POSTGRES_DB` dedicado, provisionado automaticamente pelo `postgres-master` em `docker-compose`, isolando os schemas.
- ~~**Autenticação/assinatura entre serviços**: padronizar autenticação mTLS ou tokens de serviço para chamadas HTTP/gRPC entre gateway, guardrails, tenant e observability.~~ O tráfego HTTP entre serviços usa `SERVICES_AUTH_TOKEN`, checado em middleware compartilhado e propagado pelos clients remotos.
- **Contratos versionados e pact tests**: publicar OpenAPI/gRPC por serviço e criar testes de contrato (gateway ↔ tenant/guardrails/observability) para permitir deploys independentes.
- ~~**Circuit breakers e retries cross-service**: adicionar camadas de fallback/resiliência específicas em cada client remoto para evitar acoplamento via erro cascata.~~ Os clients remotos usam retries configuráveis e transporte com circuit breaker compartilhado para proteger chamadas entre bounded contexts.
- **Build/pipeline por serviço**: separar pipelines de build/deploy (imagens, Helm charts) para cada app em `apps/`, removendo as dependências cruzadas ainda existentes no Makefile.

## Itens para abrir como issues no MaiGo
- Expor um construtor nativo para obter um `*http.Client` (ou adaptador) a partir do client MaiGo para uso em bibliotecas que não aceitam interfaces.
- Permitir inicializar clientes sem exigir base URL (modo opcional) apenas para reaproveitar configuração de timeout/transport em cenários onde a base é fornecida em runtime externo.

## Critérios de pronto
- Cada bounded context possui contrato versionado (OpenAPI/gRPC) e SDK em `pkg`.
- O runtime do gateway depende apenas de módulos com contratos finos (`pkg/*`), aptos a serem substituídos por clients externos sem alterar rotas HTTP ou casos de uso.
- Telemetria e custos são calculados fora do processo principal, com IDs de rastreamento propagados de ponta a ponta.
