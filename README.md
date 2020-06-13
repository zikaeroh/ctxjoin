# ctxjoin

A library to join contexts in various manners.

```go
func doSomething() {
    ctx, cancel := ctxjoin.AddCancel(ctx, someExtraCancelableContext)
    defer cancel()

    // ...
}
```
