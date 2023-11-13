package typegen

type Assignments []*Assignment

type Assignment struct {
	LHS string
	Op  string
	RHS string
}
