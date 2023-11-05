# go-typegen
Simple generator of Golang code to instantiate a type given an instantiated object. 

Use-case is to create type instantiation code for use in unit testing.

## Usage

```go
value := map[string]int{"Foo": 1, "Bar": 2, "Baz": 3} 
cb := typegen.NewCodeBuilder()
cb.Prettify = true
code := cb.Marshal(value)
println(code)  // Prints: map[string]int{"Bar":2,"Baz":3,"Foo":1,} 
```

## Stability
This is brand new and likely has many rough edge. 

However, I'd like to make it robust so if you find issues either submit a PR or a bug report, please.

## Roadmap
Prettification coming soon

## License
Apache 2.0
