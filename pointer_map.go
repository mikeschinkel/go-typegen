package typegen

// PointerMap will index any pointer values for *Node value where
// Node.Type==PointerNode. Used for for quicker comparison when looking for nodes
// already generated. Set in `NodeBuilder.register()` and checked in
// `NodeBuilder.isRegister()`.
type PointerMap map[uintptr]*Node
