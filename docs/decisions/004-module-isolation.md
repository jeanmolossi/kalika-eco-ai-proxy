# ADR 004: Isolamento de módulos e bancos

## Contexto
Migrations estavam todas no gateway e com FK cruzando módulos, quebrando independência.

## Decisão
- Cada módulo mantém migrations e conexão próprias (`database/<module>`), sem fallback global.
- Proibido criar FK entre bancos/schemas de módulos distintos.
- Qualquer dado compartilhado deve ser sincronizado via API/evento e armazenado localmente pelo consumidor.

## Consequências
- Facilita deploy/rollback independente por serviço.
- Aumenta necessidade de contratos bem definidos e consistência eventual.
