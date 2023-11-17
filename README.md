# go-typegen
Simple generator of Golang code to instantiate a type given an instantiated object. 

## Intended Use-case 

To create type instantiation code for use in unit testing from within code that generates over complex objects that would be hard to manually convert to use in testing.

**_NOTE:_** While good testing practices create tests with small data targeted specific edge-cases, `typegen` is not intended for typical unit testing. `typegen` is for diagnosing a bug in a huge data structure where it is hard to determine what part of the data is causing error. Perfect examples for this are heavily recursive data structures, which typegen should handle with aplomb. 

## Usage

```go
package main

import "github.com/mikeschinkel/go-typegen"

func main() {
  value := []int{1, 2, 3}
  funcName := "getData"
  // Replace w/package name where you will use getdata() func.
  omitPkg := "typegen_test"
  nb := typegen.NewNodeBuilder(value, funcName, omitPkg)
  nb.Build()
  code := nb.Generate()
  println(code)
}
```
The above code will print the following:
```go
func getData() []int {
  var1 := []int{1,2,3,}
  return var1
}
```

See it run [in the playground](https://goplay.tools/snippet/7SOrqjjpQTj).

## Stability
This is brand new and likely has many rough edges. 

However, I'd like to make it robust so if you find issues either submit a PR or a bug report, please.

## FAQ

- Q: What not just use `fmt.Sprintf("%#v", value)`?  
- A: Because it does not recurse through complex structures.
----
- Q: Your question goes [here](https://github.com/mikeschinkel/go-typegen/issues/new).
- A: ???

## License
Apache 2.0
