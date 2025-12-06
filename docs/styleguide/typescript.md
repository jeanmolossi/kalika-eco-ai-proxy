# TypeScript Styleguide

- Use TypeScript estrito (`strict: true`). Não use `any` salvo em fronteiras externas.
- Mantenha módulos por domínio, não por tecnologia.
- Prefira funções puras; side-effects concentrados em adapters.
- Tipos: interfaces para contratos, `type` para aliases. Evite enums; use unions string.
- Formatação: 2 spaces, sem semicolons (seguir `.editorconfig`/linter do projeto caso exista).
- Erros: lance `Error` com mensagem contextual e capture no nível de orquestração.
- HTTP clients: sempre com timeout e circuit breaker quando aplicável.
- Tests com `vitest/jest` em `__tests__` próximos ao código.
- Imports absolutos só via path mapping acordado; nada de caminhos relativos complexos (`../../../`).
