package gengen

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type (
	SourceContext struct {
		Pkg     *ast.Ident
		Imports []*ast.ImportSpec
		Classes []Class
		Types   []*ast.TypeSpec
	}

	parseVisitor struct {
		src *SourceContext
	}

	typeSpecVisitor struct {
		src         *SourceContext
		node        *ast.TypeSpec
		isInterface bool
		iface       *Class
		name        *ast.Ident
	}

	Class struct {
		ctx         *SourceContext `json:"-"`
		Node        *ast.TypeSpec
		Name        *ast.Ident
		Comments    []string
		IsInterface bool

		Annotations []Annotation
		Methods     []Method
	}

	interfaceTypeVisitor struct {
		node    *ast.TypeSpec
		ts      *typeSpecVisitor
		methods []Method
	}

	structVisitor struct {
		node    *ast.TypeSpec
		ts      *typeSpecVisitor
		methods []Method
	}

	Method struct {
		Itf         *Class `json:"-"`
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
	case *ast.FuncDecl:
		if rn.Recv == nil || len(rn.Recv.List) == 0 {
			return nil
		}

		var name string
		if star, ok := rn.Recv.List[0].Type.(*ast.StarExpr); ok {
			name = star.X.(*ast.Ident).Name
		} else if ident, ok := rn.Recv.List[0].Type.(*ast.Ident); ok {
			name = ident.Name
		} else {
			log.Fatalln(fmt.Errorf("func.recv is unknown type - %T", rn.Recv.List[0].Type))
		}
		var class *Class
		for idx := range v.src.Classes {
			if name == v.src.Classes[idx].Name.Name {
				class = &v.src.Classes[idx]
				break
			}
		}

		if class == nil {
			log.Fatalln(errors.New(strconv.Itoa(int(rn.Pos())) + ": 请先定义类型，后定义 方法"))
		}

		mv := &methodVisitor{node: &ast.Field{Doc: rn.Doc, Names: []*ast.Ident{rn.Name}, Type: rn.Type}, list: &class.Methods}
		ast.Walk(mv, mv.node)
		return nil
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
	case *ast.StructType:
		v.isInterface = false
		return &structVisitor{ts: v, methods: []Method{}}
	case *ast.InterfaceType:
		v.isInterface = true
		return &interfaceTypeVisitor{ts: v, methods: []Method{}}
	case nil:
		if v.iface != nil {
			v.iface.IsInterface = v.isInterface
			v.iface.ctx = v.src
			v.iface.Node = v.node
			v.iface.Name = v.name

			if v.node.Comment != nil {
				for _, a := range v.node.Comment.List {
					v.iface.Comments = append(v.iface.Comments, a.Text)
				}
			}

			if v.node.Doc != nil {
				for _, a := range v.node.Doc.List {
					v.iface.Comments = append(v.iface.Comments, a.Text)
				}
			}

			v.src.Classes = append(v.src.Classes, *v.iface)
		}
		return nil
	default:
		return v
	}
}

func (v *structVisitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	default:
		return v
	case *ast.FieldList:
		return nil
	case *ast.Field:
		return nil
	case nil:
		v.ts.iface = &Class{Methods: v.methods}
		return nil
	}
}

func (v *interfaceTypeVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.Field:
		return &methodVisitor{node: rn, list: &v.methods}
	case nil:
		v.ts.iface = &Class{Methods: v.methods}
		for idx := range v.ts.iface.Methods {
			v.ts.iface.Methods[idx].init(v.ts.iface)
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
	case *ast.FuncLit:
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
	case *ast.BlockStmt:
		return nil
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
	for _, i := range sc.Classes {
		for _, m := range i.Methods {
			if len(m.Results.List) < 1 {
				return fmt.Errorf("method %q of interface %q has no result types", m.Name, i.Name)
			}
		}
	}
	return nil
}

func (method *Method) init(iface *Class) {
	method.Itf = iface
	method.Params.Method = method
	for j := range method.Params.List {
		method.Params.List[j].Method = method
	}

	method.Results.Method = method
	for j := range method.Results.List {
		method.Results.List[j].Method = method
	}
}

func Parse(filename string, source io.Reader) (*SourceContext, error) {
	f, err := parser.ParseFile(token.NewFileSet(), filename, source, parser.DeclarationErrors|parser.ParseComments)
	if err != nil {
		return nil, errors.New("parsing input file '" + filename + "': " + err.Error())
	}

	context := &SourceContext{}
	visitor := &parseVisitor{src: context}
	ast.Walk(visitor, f)

	for classIdx, itf := range context.Classes {
		for _, comment := range context.Classes[classIdx].Comments {
			ann := parseAnnotation(comment)
			if ann != nil {
				context.Classes[classIdx].Annotations = append(context.Classes[classIdx].Annotations, *ann)
			}
		}

		for idx := range itf.Methods {
			method := &itf.Methods[idx]
			method.init(&context.Classes[classIdx])
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

func typePrint(typ ast.Node) string {
	fset := token.NewFileSet()
	var buf strings.Builder
	if err := format.Node(&buf, fset, typ); err != nil {
		log.Fatalln(err)
	}
	return buf.String()
}
