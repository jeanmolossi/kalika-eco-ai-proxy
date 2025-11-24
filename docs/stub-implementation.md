# Plano de implementação dos stubs

Este guia lista os stubs ainda presentes no projeto e sugere como substituí-los por implementações reais para produção.

## LLM e roteamento

- `internal/platform/llm/stub_client.go`: substituir pelo(s) cliente(s) de provedores reais (ex.: OpenAI, Azure OpenAI, Ollama) com seleção via configuração de tenant. Implementar:
  - suporte a streaming de respostas e timeouts por chamada;
  - tradução de erros do provedor para códigos HTTP estáveis;
  - enriquecimento da resposta com IDs do provedor e metadados para auditoria.
- `internal/platform/router/simple_router.go`: remover fallbacks para `stub-model`/`stub-embed-model` e exigir que o tenant declare `DefaultModel` e `ModelsAllowed`. Acrescentar:
  - validação de modelo contra allowlist do tenant antes de rotear;
  - roteamento por capacidade (chat/embedding) e afinidade regional quando houver múltiplos provedores;
  - telemetria de latência e retries com backoff.

## Tokenização

- `internal/platform/tokenizer/openai_tiktoken.go`: o alias atual cobre apenas `stub-model`. Carregar o mapa de aliases de configuração por tenant e alinhar com `ModelsAllowed`/provedor real. Validar a presença de tokenizer compatível ao subir o serviço e falhar rápido se o modelo não for suportado.

## Custos, auditoria e publicação

- `internal/modules/aiproxy/app/audit_usage.go`: hoje o custo é fixo em zero e o `RequestID` fica vazio. Implementar:
  - tabela de preços por modelo (prompt/completion) e cálculo por tokens retornados;
  - geração e propagação de `RequestID`/`TraceID` desde o handler HTTP até os publishers;
  - inclusão de currency e arredondamento consistente; métricas de custo por tenant.
- `internal/platform/module.go`: atualmente publica auditoria/uso em log e usa cache sem efeito. Trocar por:
  - publisher assíncrono para fila (Kafka/SNS) ou tabela de auditoria persistente;
  - semantic cache real (ex.: Redis) com TTL configurável e chave por tenant/modelo;
  - limiter com backend distribuído para múltiplas réplicas.

## Dependências e container

- `internal/modules/aiproxy/deps.go`: as dependências são descritas como "noop/stub". Endurecer o container com validação de wiring (healthcheck de conexões), injeção de clientes reais e falha explícita quando um módulo obrigatório não estiver configurado.

## Modelos e defaults

- `internal/platform/module.go` e `internal/platform/router/simple_router.go`: substituir o modelo padrão `stub-model` por um default configurável por ambiente/tenant, documentando claramente quais modelos são permitidos. Incluir validação para embeddings (`stub-embed-model`) seguindo o mesmo padrão.
