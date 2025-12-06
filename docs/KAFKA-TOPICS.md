# Kafka Topics

## Contexto
O `docker-compose` local inclui um cluster Kafka (com Zookeeper) para trafegar os eventos críticos do proxy de IA. A criação automática de tópicos está desabilitada para evitar divergências de configuração; por isso, os tópicos abaixo devem ser criados antes de testar fluxos que publicam eventos.

## Tópicos obrigatórios
- **ai-proxy.audit.events**
  - Finalidade: armazenar trilhas de auditoria com request ID, tenant, usuário (quando disponível), modelo resolvido, veredictos de guardrails e hashes/parâmetros relevantes da requisição.
  - Sugestão de configuração: 3 partitions, `cleanup.policy=delete`, `retention.ms=604800000` (7 dias) para ambientes de dev; ajustar retenção conforme políticas internas em produção.
- **ai-proxy.usage.events**
  - Finalidade: registrar contagens de tokens, custos estimados em USD, modelo utilizado e timestamps para billing e capacidade de prever consumo.
  - Sugestão de configuração: 3 partitions, `cleanup.policy=compact,delete` (para manter último estado por request ID) com `retention.ms` de pelo menos 14 dias.
- **ai-proxy.guardrails.verdicts**
  - Finalidade: coletar decisões de firewall/guardrails (allow, block, rewrite) nos estágios de input/output para monitoramento de segurança e tuning de regras.
  - Sugestão de configuração: 1–3 partitions (baixa cardinalidade), `cleanup.policy=delete`, `retention.ms=1209600000` (14 dias) para análise retroativa.

## Comandos de criação (ambiente local)
Executar dentro do container Kafka ou com `KAFKA_ADVERTISED_LISTENERS` acessível via `localhost:29092`:

```bash
kafka-topics --bootstrap-server localhost:29092 \
  --create --if-not-exists --topic ai-proxy.audit.events \
  --partitions 3 --replication-factor 1 \
  --config cleanup.policy=delete --config retention.ms=604800000

kafka-topics --bootstrap-server localhost:29092 \
  --create --if-not-exists --topic ai-proxy.usage.events \
  --partitions 3 --replication-factor 1 \
  --config cleanup.policy=compact,delete --config retention.ms=1209600000

kafka-topics --bootstrap-server localhost:29092 \
  --create --if-not-exists --topic ai-proxy.guardrails.verdicts \
  --partitions 1 --replication-factor 1 \
  --config cleanup.policy=delete --config retention.ms=1209600000
```

### Inicialização via Makefile
Com o cluster do `docker compose` ativo, execute um único comando que cria todos os tópicos obrigatórios:

```bash
make kafka-topics-init
```

## UI de gerenciamento
- O container `kafka-ui` sobe automaticamente no `docker compose` e fica disponível em `http://localhost:8082`.
- Ele usa o broker interno `kafka:9092` e o Zookeeper `zookeeper:2181`; não há autenticação habilitada por padrão em dev.
- Use a UI para conferir offsets, partições, mensagens e verificar se os tópicos foram criados corretamente.

## Observações
- Ajuste o `replication-factor` em ambientes com múltiplos brokers para manter tolerância a falhas.
- Caso uma aplicação externa precise consumir via host, use `localhost:29092`; containers no mesmo `docker-compose` podem usar `kafka:9092` como bootstrap.
- Padronize o uso de `request_id` como chave de partição para alinhar auditoria, custos e veredictos no mesmo shard.
- Para publicar eventos direto no Kafka (em vez de arquivos locais), configure `USAGE_MODE=kafka`/`AUDIT_MODE=kafka` com `KAFKA_BROKERS` e os tópicos adequados no ambiente.
