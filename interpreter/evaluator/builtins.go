package evaluator

import (
	"fmt"
	"monkey/interpreter/object"
)

var builtins = map[string]*object.BuiltIn{
	"len":   {Fn: lenFunc},
	"first": {Fn: firstFunc},
	"last":  {Fn: lastFunc},
	"rest":  {Fn: restFunc},
	"push":  {Fn: pushFunc},
	"puts":  {Fn: putsFunc},
}

func lenFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}

	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}

	default:
		return newError("argument to `len` not supported, got %s", arg.Type())
	}
}

func firstFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	if len(arr.Elements) > 0 {
		return arr.Elements[0]
	}

	return NULL
}

func lastFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	if len(arr.Elements) > 0 {
		return arr.Elements[len(arr.Elements)-1]
	}

	return NULL
}

func restFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `rest` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	length := len(arr.Elements)

	if len(arr.Elements) > 0 {
		newEls := make([]object.Object, length-1)
		copy(newEls, arr.Elements[1:length])
		return &object.Array{Elements: newEls}
	}

	return NULL
}

func pushFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}

	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	length := len(arr.Elements)
	el := args[1]

	newEls := make([]object.Object, length+1)
	copy(newEls, arr.Elements)
	newEls[length] = el

	return &object.Array{Elements: newEls}

}

func putsFunc(args ...object.Object) object.Object {
	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}
	return NULL
}
