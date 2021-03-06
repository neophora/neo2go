package compiler

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"

	"golang.org/x/tools/go/loader"
)

var (
	// Go language builtin functions and custom builtin utility functions.
	builtinFuncs = []string{
		"len", "append", "SHA256",
		"SHA1", "Hash256", "Hash160",
		"VerifySignature", "AppCall",
		"FromAddress", "Equals",
		"panic", "DynAppCall",
		"delete", "Remove",
	}
)

// typeAndValueForField returns a zero initialized typeAndValue for the given type.Var.
func typeAndValueForField(fld *types.Var) (types.TypeAndValue, error) {
	switch t := fld.Type().(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Int:
			return types.TypeAndValue{
				Type:  t,
				Value: constant.MakeInt64(0),
			}, nil
		case types.String:
			return types.TypeAndValue{
				Type:  t,
				Value: constant.MakeString(""),
			}, nil
		case types.Bool, types.UntypedBool:
			return types.TypeAndValue{
				Type:  t,
				Value: constant.MakeBool(false),
			}, nil
		default:
			return types.TypeAndValue{}, fmt.Errorf("could not initialize struct field %s to zero, type: %s", fld.Name(), t)
		}
	default:
		return types.TypeAndValue{Type: t}, nil
	}
}

// countGlobals counts the global variables in the program to add
// them with the stack size of the function.
func countGlobals(f ast.Node) (i int64) {
	ast.Inspect(f, func(node ast.Node) bool {
		switch node.(type) {
		// Skip all function declarations.
		case *ast.FuncDecl:
			return false
		// After skipping all funcDecls we are sure that each value spec
		// is a global declared variable or constant.
		case *ast.ValueSpec:
			i++
		}
		return true
	})
	return
}

// isIdentBool looks if the given ident is a boolean.
func isIdentBool(ident *ast.Ident) bool {
	return ident.Name == "true" || ident.Name == "false"
}

// isExprNil looks if the given expression is a `nil`.
func isExprNil(e ast.Expr) bool {
	v, ok := e.(*ast.Ident)
	return ok && v.Name == "nil"
}

// makeBoolFromIdent creates a bool type from an *ast.Ident.
func makeBoolFromIdent(ident *ast.Ident, tinfo *types.Info) (types.TypeAndValue, error) {
	var b bool
	switch ident.Name {
	case "true":
		b = true
	case "false":
		b = false
	default:
		return types.TypeAndValue{}, fmt.Errorf("givent identifier cannot be converted to a boolean => %s", ident.Name)
	}
	return types.TypeAndValue{
		Type:  tinfo.ObjectOf(ident).Type(),
		Value: constant.MakeBool(b),
	}, nil
}

// resolveEntryPoint returns the function declaration of the entrypoint and the corresponding file.
func resolveEntryPoint(entry string, pkg *loader.PackageInfo) (*ast.FuncDecl, *ast.File) {
	var (
		main *ast.FuncDecl
		file *ast.File
	)
	for _, f := range pkg.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.FuncDecl:
				if t.Name.Name == entry {
					main = t
					file = f
					return false
				}
			}
			return true
		})
	}
	return main, file
}

// indexOfStruct returns the index of the given field inside that struct.
// If the struct does not contain that field it will return -1.
func indexOfStruct(strct *types.Struct, fldName string) int {
	for i := 0; i < strct.NumFields(); i++ {
		if strct.Field(i).Name() == fldName {
			return i
		}
	}
	return -1
}

type funcUsage map[string]bool

func (f funcUsage) funcUsed(name string) bool {
	_, ok := f[name]
	return ok
}

// lastStmtIsReturn checks if last statement of the declaration was return statement..
func lastStmtIsReturn(decl *ast.FuncDecl) (b bool) {
	if l := len(decl.Body.List); l != 0 {
		_, ok := decl.Body.List[l-1].(*ast.ReturnStmt)
		return ok
	}
	return false
}

func analyzeFuncUsage(pkgs map[*types.Package]*loader.PackageInfo) funcUsage {
	usage := funcUsage{}

	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			ast.Inspect(f, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.CallExpr:
					switch t := n.Fun.(type) {
					case *ast.Ident:
						usage[t.Name] = true
					case *ast.SelectorExpr:
						usage[t.Sel.Name] = true
					}
				}
				return true
			})
		}
	}
	return usage
}

func isBuiltin(expr ast.Expr) bool {
	var name string

	switch t := expr.(type) {
	case *ast.Ident:
		name = t.Name
	case *ast.SelectorExpr:
		name = t.Sel.Name
	default:
		return false
	}

	for _, n := range builtinFuncs {
		if name == n {
			return true
		}
	}
	return false
}

func (c *codegen) isCompoundArrayType(t ast.Expr) bool {
	switch s := t.(type) {
	case *ast.ArrayType:
		return true
	case *ast.Ident:
		arr, ok := c.typeInfo.Types[s].Type.Underlying().(*types.Slice)
		return ok && !isByte(arr.Elem())
	}
	return false
}

func isByte(t types.Type) bool {
	e, ok := t.(*types.Basic)
	return ok && e.Kind() == types.Byte
}

func (c *codegen) isStructType(t ast.Expr) (int, bool) {
	switch s := t.(type) {
	case *ast.StructType:
		return s.Fields.NumFields(), true
	case *ast.Ident:
		st, ok := c.typeInfo.Types[s].Type.Underlying().(*types.Struct)
		if ok {
			return st.NumFields(), true
		}
	}
	return 0, false
}

func isByteArray(lit *ast.CompositeLit, tInfo *types.Info) bool {
	if len(lit.Elts) == 0 {
		if typ, ok := lit.Type.(*ast.ArrayType); ok {
			if name, ok := typ.Elt.(*ast.Ident); ok {
				return name.Name == "byte" || name.Name == "uint8"
			}
		}

		return false
	}

	typ := tInfo.Types[lit.Elts[0]].Type.Underlying()
	return isByte(typ)
}

func isSyscall(fun *funcScope) bool {
	if fun.selector == nil {
		return false
	}
	_, ok := syscalls[fun.selector.Name][fun.name]
	return ok
}

func isByteArrayType(t types.Type) bool {
	return t.String() == "[]byte"
}

func isStringType(t types.Type) bool {
	return t.String() == "string"
}
