module github.com/mikeschinkel/go-typegen

go 1.21

replace github.com/mikeschinkel/go-serr => ../go-serr

replace github.com/mikeschinkel/go-lib => ../go-lib

replace github.com/mikeschinkel/go-diffator => ../go-diffator

require (
	github.com/mikeschinkel/go-diffator v0.0.0-00010101000000-000000000000
	github.com/mikeschinkel/go-lib v0.0.0-20240105150559-6b08a12c3c43
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
