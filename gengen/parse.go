package gengen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"

	"github.com/pkg/errors"
)

type (
	SourceContext struct {
		Pkg        *ast.Ident
		Imports    []*ast.ImportSpec
		Interfaces []Interface
		Types      []*ast.TypeSpec
	}

	parseVisitor struct {
		src *SourceContext
	}

	typeSpecVisitor struct {
		src   *SourceContext
		node  *ast.TypeSpec
		iface *Interface
		name  *ast.Ident
	}

	Interface struct {
		ctx  *SourceContext `json:"-"`
		Node *ast.TypeSpec
		Name *ast.Ident

		Methods []Method
	}

	interfaceTypeVisitor struct {
		node    *ast.TypeSpec
		ts      *typeSpecVisitor
		methods []Method
	}

	Method struct {
		Itf         *Interface `json:"-"`
		Node        *ast.Field
		Name        *ast.Ident
		Comments    []string
		Annotations []Annotation
		Params      *Params
		Results     *Results
	}

	methodVisitor struct {
		depth    int
		node     *ast.Field
		list     *[]Method
		name     *ast.Ident
		params   *Params
		results  *Results
		isMethod bool
	}

	Params struct {
		Method *Method `json:"-"`
		List   []Param
	}

	argListVisitor struct {
		list *Params
	}

	Param struct {
		Method     *Method `json:"-"`
		Name       *ast.Ident
		IsVariadic bool
		Typ        ast.Expr
	}

	argVisitor struct {
		node  *ast.TypeSpec
		parts []ast.Expr
		list  *Params
	}

	Results struct {
		Method *Method `json:"-"`
		List   []Result
	}

	resultListVisitor struct {
		list *Results
	}

	Result struct {
		Method *Method `json:"-"`
		Name   *ast.Ident
		Typ    ast.Expr
	}

	resultVisitor struct {
		node  *ast.TypeSpec
		parts []ast.Expr
		list  *Results
	}
)

func (v *parseVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	case *ast.File:
		v.src.Pkg = rn.Name
		return v
	case *ast.ImportSpec:
		v.src.Imports = append(v.src.Imports, rn)
		return nil
	case *ast.TypeSpec:
		switch rn.Type.(type) {
		case *ast.InterfaceType:
		default:
			v.src.Types = append(v.src.Types, rn)
		}
		return &typeSpecVisitor{src: v.src, node: rn}
	default:
		return v
	}
}

/*
package foo

type FooService interface {
	Bar(ctx context.Context, i int, s string) (string, error)
}
*/

func (v *typeSpecVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	case *ast.Ident:
		if v.name == nil {
			v.name = rn
		}
		return v
	case *ast.InterfaceType:
		return &interfaceTypeVisitor{ts: v, methods: []Method{}}
	case nil:
		if v.iface != nil {
			v.iface.ctx = v.src
			v.iface.Node = v.node
			v.iface.Name = v.name
			v.src.Interfaces = append(v.src.Interfaces, *v.iface)
		}
		return nil
	default:
		return v
	}
}

func (v *interfaceTypeVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.Field:
		return &methodVisitor{node: rn, list: &v.methods}
	case nil:
		v.ts.iface = &Interface{Methods: v.methods}
		for idx := range v.ts.iface.Methods {
			v.ts.iface.Methods[idx].Itf = v.ts.iface

			v.ts.iface.Methods[idx].Params.Method = &v.ts.iface.Methods[idx]
			for j := range v.ts.iface.Methods[idx].Params.List {
				v.ts.iface.Methods[idx].Params.List[j].Method = &v.ts.iface.Methods[idx]
			}
			v.ts.iface.Methods[idx].Results.Method = &v.ts.iface.Methods[idx]

			for j := range v.ts.iface.Methods[idx].Results.List {
				v.ts.iface.Methods[idx].Results.List[j].Method = &v.ts.iface.Methods[idx]
			}
		}
		return nil
	}
}

func (v *methodVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		v.depth++
		return v
	case *ast.Ident:
		v.name = rn
		v.depth++
		return v
	case *ast.FuncType:
		v.depth++
		v.isMethod = true
		return v
	case *ast.FieldList:
		if v.params == nil {
			v.params = &Params{}
			return &argListVisitor{list: v.params}
		}
		if v.results == nil {
			v.results = &Results{}
		}
		return &resultListVisitor{list: v.results}
	case nil:
		v.depth--
		if v.depth == 0 && v.isMethod && v.name != nil {
			var comments []string
			if v.node.Comment != nil {
				for _, a := range v.node.Comment.List {
					comments = append(comments, a.Text)
				}
			}

			if v.node.Doc != nil {
				for _, a := range v.node.Doc.List {
					comments = append(comments, a.Text)
				}
			}

			*v.list = append(*v.list, Method{Node: v.node, Name: v.name, Params: v.params, Results: v.results, Comments: comments})
		}
		return nil
	}
}

func (v *argListVisitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	default:
		return nil
	case *ast.Field:
		return &argVisitor{list: v.list}
	}
}

func (v *argVisitor) Visit(n ast.Node) ast.Visitor {
	switch t := n.(type) {
	case *ast.CommentGroup, *ast.BasicLit:
		return nil
	case *ast.Ident: //Expr -> everything, but clarity
		if t.Name != "_" {
			v.parts = append(v.parts, t)
		}
	case ast.Expr:
		v.parts = append(v.parts, t)
	case nil:
		names := v.parts[:len(v.parts)-1]
		tp := v.parts[len(v.parts)-1]
		if len(names) == 0 {
			v.list.List = append(v.list.List, Param{Typ: tp})
			return nil
		}
		for _, n := range names {
			v.list.List = append(v.list.List, Param{
				Name: n.(*ast.Ident),
				Typ:  tp,
			})
		}
	}
	return nil
}

func (v *resultListVisitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	default:
		return nil
	case *ast.Field:
		return &resultVisitor{list: v.list}
	}
}

func (v *resultVisitor) Visit(n ast.Node) ast.Visitor {
	switch t := n.(type) {
	case *ast.CommentGroup, *ast.BasicLit:
		return nil
	case *ast.Ident: //Expr -> everything, but clarity
		if t.Name != "_" {
			v.parts = append(v.parts, t)
		}
	case ast.Expr:
		v.parts = append(v.parts, t)
	case nil:
		names := v.parts[:len(v.parts)-1]
		tp := v.parts[len(v.parts)-1]
		if len(names) == 0 {
			v.list.List = append(v.list.List, Result{Typ: tp})
			return nil
		}
		for _, n := range names {
			v.list.List = append(v.list.List, Result{
				Name: n.(*ast.Ident),
				Typ:  tp,
			})
		}
	}
	return nil
}

func (sc *SourceContext) validate() error {
	if len(sc.Interfaces) != 1 {
		return fmt.Errorf("found %d interfaces, expecting exactly 1", len(sc.Interfaces))
	}
	for _, i := range sc.Interfaces {
		for _, m := range i.Methods {
			if len(m.Results.List) < 1 {
				return fmt.Errorf("method %q of interface %q has no result types", m.Name, i.Name)
			}
		}
	}
	return nil
}

func Parse(filename string, source io.Reader) (*SourceContext, error) {
	f, err := parser.ParseFile(token.NewFileSet(), filename, source, parser.DeclarationErrors|parser.ParseComments)
	if err != nil {
		return nil, errors.New("parsing input file '" + filename + "': " + err.Error())
	}

	context := &SourceContext{}
	visitor := &parseVisitor{src: context}
	ast.Walk(visitor, f)

	for _, itf := range context.Interfaces {
		for idx := range itf.Methods {
			method := &itf.Methods[idx]
			for _, comment := range method.Comments {
				ann := parseAnnotation(comment)
				if ann != nil {
					method.Annotations = append(method.Annotations, *ann)
				}
			}
		}
	}
	if err := context.validate(); err != nil {
		return nil, errors.New("examining input file '" + filename + "': " + err.Error())
	}
	return context, nil
}

func ParseFile(filename string) (*SourceContext, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("error while opening '" + filename + "': " + err.Error())
	}
	defer file.Close()
	return Parse(filename, file)
}
