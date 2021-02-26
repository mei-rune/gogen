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
	"reflect"
	"strings"
)

type (
	SourceContext struct {
		FileSet *token.FileSet
		Pkg     *ast.Ident
		Imports []*ast.ImportSpec
		Classes []Class
		Types   []*ast.TypeSpec
	}

	parseVisitor struct {
		src *SourceContext
	}

	genDeclVisitor struct {
		src      *SourceContext
		node     *ast.GenDecl
		Comments []string
	}

	typeSpecVisitor struct {
		src         *SourceContext
		node        *ast.TypeSpec
		isInterface bool
		iface       *Class
		name        *ast.Ident
		comments    []string
	}

	Class struct {
		Ctx         *SourceContext `json:"-"`
		Node        *ast.TypeSpec
		Name        *ast.Ident
		Comments    []string
		IsInterface bool

		Annotations []Annotation
		Methods     []Method
		Fields      []Field
	}

	Field struct {
		Ctx   *SourceContext `json:"-"`
		Clazz *Class         `json:"-"`
		Node  *ast.Field
		Name  *ast.Ident
		Typ   ast.Expr      // field/method/parameter type
		Tag   *ast.BasicLit // field tag; or nil
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
		fields  []Field
	}

	Method struct {
		Ctx         *SourceContext `json:"-"`
		Clazz       *Class         `json:"-"`
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
			for idx := range v.src.Types {
				if name == v.src.Types[idx].Name.Name {
					return nil
				}
			}

			log.Fatalln(errors.New(v.src.PostionFor(rn.Pos()).String() + ": 请先定义类型，后定义 方法"))
		}

		mv := &methodVisitor{node: &ast.Field{Doc: rn.Doc, Names: []*ast.Ident{rn.Name}, Type: rn.Type}, list: &class.Methods}
		ast.Walk(mv, mv.node)
		return nil
	case *ast.GenDecl:
		if rn.Tok == token.TYPE {
			return &genDeclVisitor{src: v.src, node: rn}
		}
		return v
	default:
		return v
	}
}

func (v *genDeclVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	case *ast.TypeSpec:
		switch rn.Type.(type) {
		case *ast.InterfaceType:
		default:
			v.src.Types = append(v.src.Types, rn)
		}

		var comments []string
		if v.node.Doc != nil {
			for _, a := range v.node.Doc.List {
				comments = append(comments, a.Text)
			}
		}

		return &typeSpecVisitor{src: v.src, node: rn, comments: comments}
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
			v.iface.Ctx = v.src
			v.iface.Node = v.node
			v.iface.Name = v.name
			v.iface.Comments = v.comments

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
	switch rn := n.(type) {
	default:
		return v
	case *ast.FieldList:
		return v
	case *ast.Field:
		if len(rn.Names) == 0 {
			v.fields = append(v.fields, Field{Node: rn, Name: nil, Typ: rn.Type, Tag: rn.Tag})
		} else {
			for _, name := range rn.Names {
				v.fields = append(v.fields, Field{Node: rn, Name: name, Typ: rn.Type, Tag: rn.Tag})
			}
		}
		return nil
	case nil:
		v.ts.iface = &Class{Methods: v.methods, Fields: v.fields}
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

func (sc *SourceContext) PostionFor(pos token.Pos) token.Position {
	return sc.FileSet.PositionFor(pos, true)
}

func (sc *SourceContext) GetType(name string) *ast.TypeSpec {
	for _, typ := range sc.Types {
		if typ.Name.Name == name {
			return typ
		}
	}
	return nil
}

func (sc *SourceContext) GetClass(name string) *Class {
	for idx := range sc.Classes {
		if sc.Classes[idx].Name.Name == name {
			return &sc.Classes[idx]
		}
	}
	return nil
}

func (sc *SourceContext) validate() error {
	//	for _, i := range sc.Classes {
	//		for _, m := range i.Methods {
	//			if m.Results == nil || len(m.Results.List) < 1 {
	//				return fmt.Errorf("method %q of interface %q has no result types", m.Name, i.Name)
	//			}
	//		}
	//	}
	return nil
}

func (method *Method) init(iface *Class) {
	method.Clazz = iface
	if method.Params != nil {
		method.Params.Method = method
		for j := range method.Params.List {
			method.Params.List[j].Method = method
		}
	}

	if method.Results != nil {
		method.Results.Method = method
		for j := range method.Results.List {
			method.Results.List[j].Method = method
		}
	}
}

func Parse(filename string, source io.Reader) (*SourceContext, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, source, parser.DeclarationErrors|parser.ParseComments)
	if err != nil {
		return nil, errors.New("parsing input file '" + filename + "': " + err.Error())
	}

	context := &SourceContext{
		FileSet: fset,
	}
	visitor := &parseVisitor{src: context}
	ast.Walk(visitor, f)

	for classIdx := range context.Classes {
		context.Classes[classIdx].Ctx = context

		for _, comment := range context.Classes[classIdx].Comments {
			ann := parseAnnotation(comment)
			if ann != nil {
				context.Classes[classIdx].Annotations = append(context.Classes[classIdx].Annotations, *ann)
			}
		}

		for idx := range context.Classes[classIdx].Methods {
			method := &context.Classes[classIdx].Methods[idx]
			method.Ctx = context
			method.init(&context.Classes[classIdx])
			for _, comment := range method.Comments {
				ann := parseAnnotation(comment)
				if ann != nil {
					method.Annotations = append(method.Annotations, *ann)
				}
			}
		}

		for idx := range context.Classes[classIdx].Fields {
			context.Classes[classIdx].Fields[idx].Ctx = context
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

var RangeDefineds = map[string]struct {
	Start ast.Expr
	End   ast.Expr
}{}

func AddRangeDefined(typ, start, end string) {
	var s ast.Expr = &ast.Ident{Name: strings.TrimPrefix(start, "*")}
	var e ast.Expr = &ast.Ident{Name: strings.TrimPrefix(end, "*")}

	if strings.HasPrefix(start, "*") {
		s = &ast.StarExpr{X: s}
	}

	if strings.HasPrefix(end, "*") {
		e = &ast.StarExpr{X: e}
	}

	RangeDefineds[typ] = struct {
		Start ast.Expr
		End   ast.Expr
	}{s, e}
}

func IsRange(classes []Class, typ ast.Expr) (bool, ast.Expr, ast.Expr) {
	name := strings.TrimPrefix(typePrint(typ), "*")

	if value, ok := RangeDefineds[name]; ok {
		return true, value.Start, value.End
	}

	var cls *Class
	for idx := range classes {
		if classes[idx].Name.Name == name {
			cls = &classes[idx]
			break
		}
	}
	if cls == nil {
		return false, nil, nil
	}

	if len(cls.Fields) != 2 {
		return false, nil, nil
	}

	var startType, endType ast.Expr
	for _, field := range cls.Fields {
		if field.Name.Name == "Start" {
			startType = field.Typ
		} else if field.Name.Name == "End" {
			endType = field.Typ
		}
	}
	if startType == nil || endType == nil {
		return false, nil, nil
	}

	aType := strings.TrimPrefix(typePrint(startType), "*")
	bType := strings.TrimPrefix(typePrint(endType), "*")
	if aType != bType {
		return false, nil, nil
	}
	return true, startType, endType
}

func IsPtrType(typ ast.Expr) bool {
	_, ok := typ.(*ast.StarExpr)
	return ok
}

func IsStructType(typ ast.Expr) bool {
	_, ok := typ.(*ast.StructType)
	return ok
}

func IsSliceType(typ ast.Expr) bool {
	_, ok := typ.(*ast.ArrayType)
	return ok
}

func IsArrayType(typ ast.Expr) bool {
	_, ok := typ.(*ast.ArrayType)
	return ok
}

func IsEllipsisType(typ ast.Expr) bool {
	_, ok := typ.(*ast.Ellipsis)
	return ok
}

func IsMapType(typ ast.Expr) bool {
	_, ok := typ.(*ast.MapType)
	return ok
}

func KeyType(typ ast.Expr) ast.Expr {
	m, ok := typ.(*ast.MapType)
	if ok {
		return m.Key
	}
	return nil
}

func ValueType(typ ast.Expr) ast.Expr {
	m, ok := typ.(*ast.MapType)
	if ok {
		return m.Value
	}
	return nil
}

func ElemType(typ ast.Expr) ast.Expr {
	switch t := typ.(type) {
	case *ast.StarExpr:
		return t.X
	case *ast.ArrayType:
		return t.Elt
	case *ast.Ellipsis:
		return t.Elt
	}
	return nil
}

func (f *Field) GetTag(key string) string {
	if f.Tag == nil {
		return ""
	}
	return reflect.StructTag(strings.Trim(f.Tag.Value, "`")).Get(key)
}
