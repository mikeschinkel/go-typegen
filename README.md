# go-typegen
Simple generator of Golang code to instantiate a type given an instantiated object. 

Use-case is to create type instantiation code for use in unit testing.

## Usage

```go
cb := typegen.NewCodeBuilder()
cb.Prettify = true
code := cb.Marshal(value)
```

## Stability
This is brand new and likely has many rough edge. 

However, I'd like to make it robust so if you find issues either submit a PR or a bug report, please.

## Roadmap
Prettification coming soon

## License
Apache 2.0
