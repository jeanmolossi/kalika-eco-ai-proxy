# Template - HTTP Handler (Go)

```go
func (h *Handlers) Handle<Operation>() echo.HandlerFunc {
    return func(c echo.Context) error {
        ctx := c.Request().Context()
        req := <RequestDTO>{}

        if err := c.Bind(&req); err != nil {
            return problem.BadRequest("invalid_body", "payload inválido")
        }

        if err := req.Validate(); err != nil {
            return problem.BadRequest("invalid_input", err.Error())
        }

        resp, err := h.usecase.Execute(ctx, req)
        if err != nil {
            return problem.FromError(err) // mapeia erro de domínio para HTTP
        }

        return c.JSON(http.StatusOK, resp)
    }
}
```
