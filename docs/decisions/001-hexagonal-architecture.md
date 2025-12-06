# ADR 001: Hexagonal Architecture como padrão

## Contexto
Precisamos de isolamento claro entre domínio e infraestrutura para manter módulos independentes e testáveis.

## Decisão
- Usaremos DDD + Ports & Adapters em todos os módulos.
- Interfaces (ports) residem no domínio; adapters em `infra` ou `remote`.
- Aplicações `apps/*` apenas fazem wiring.

## Consequências
- Facilita troca de adapters (ex.: Postgres -> outro storage) sem alterar casos de uso.
- Exige disciplina de boundaries e revisões de PR para impedir vazamento de infra para domínio.
