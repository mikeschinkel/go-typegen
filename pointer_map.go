package typegen

// PointerMap will index any pointer values for *Node value where
// Node.Type==PointerNode. Used for for quicker comparison when looking for nodes
// already generated. Set in `CodeBuilder.register()` and checked in
// `CodeBuilder.isRegister()`.
type PointerMap map[uintptr]*Node
