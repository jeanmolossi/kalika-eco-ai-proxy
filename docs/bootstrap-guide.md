# Bootstrap Guide

Siga este guia para subir o projeto rapidamente em um ambiente local de desenvolvimento.

## Pré-requisitos
- Go 1.25 (ou superior) instalado e presente no `PATH`.
- Docker e Docker Compose, caso deseje subir as dependências via contêiner.
- `make` para usar os atalhos do `Makefile`.

## Passos rápidos
1. **Instalar dependências Go**
   ```bash
   go mod download
   ```
2. **Configurar variáveis de ambiente**
   - Crie um arquivo `.env` (ou exporte as variáveis no shell) com os valores necessários. Os principais prefixos são:
     - `SERVER_` para parâmetros do servidor HTTP (host, porta, TLS etc.).
     - `POSTGRES_` para a conexão com o banco de dados.
     - `RATELIMIT_` para ajustes de rate limiting.
     - `USAGE_` e `AUDIT_` para escolher o destino dos eventos: deixe `MODE=file` (padrão) ou defina `MODE=kafka` com `KAFKA_BROKERS` e os tópicos (`USAGE_TOPIC`/`AUDIT_TOPIC`).
     - `KAFKA_` para parâmetros compartilhados de Kafka (ex.: `KAFKA_BROKERS=kafka:9092`).
3. **Rodar a aplicação localmente**
   ```bash
   make build
  ./bin/gateway
   ```
   ou, para desenvolvimento com Docker Compose:
   ```bash
   make docker-up
   ```
4. **Executar checagens**
   - Formatação: `make fmt`
   - Lint: `make lint`
   - Testes: `make test`

## Dicas úteis
- O servidor HTTP expõe o `base path` configurado em `SERVER_BASE_PATH` (padrão `/api/v1`). Os endpoints de API são registrados relativos a esse caminho.
- Use `SERVER_ENABLE_TLS=true` e configure `SERVER_TLS_CERTFILE`/`SERVER_TLS_KEYFILE` para habilitar TLS.
- Para depuração, habilite o pprof com `SERVER_ENABLE_PPROF=true`.
- O Docker Compose já inclui Zookeeper, Kafka e a UI `kafka-ui` em `http://localhost:8082`; crie os tópicos necessários com `make kafka-topics-init` antes de testar fluxos que publicam eventos.
