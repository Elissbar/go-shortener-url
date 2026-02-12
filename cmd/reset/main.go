package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type TemplateData struct {
	Package     string
	PackagePath string
	StructName  string
	Fields      []FieldData
}

type FieldData struct {
	Name string
	Type string
}

func checkComments(comments []*ast.Comment) bool {
	for _, cmnt := range comments {
		if strings.Contains(cmnt.Text, "generate:reset") {
			return true
		}
	}
	return false
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()

	var Template = `
	package {{.Package}}

	{{range .Structs}}
	func (st *{{.StructName}}) Reset() {
		{{range .Fields}}
		{{if eq .Type "string"}}st.{{.Name}} = ""{{end}}
		{{if eq .Type "int"}}st.{{.Name}} = 0{{end}}
		{{if eq .Type "bool"}}st.{{.Name}} = false{{end}}
		{{if eq .Type "IsMap"}}clear(st.{{.Name}}){{end}}
		{{if eq .Type "IsSlice"}}st.{{.Name}} = st.{{.Name}}[:0]{{end}}
		{{end}}
	}
	{{end}}`
	t := template.Must(template.New("reset").Parse(Template))

	packages := make(map[string][]TemplateData)

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".gen.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Failed to parse file %v\n", path)
			return err
		}

		ast.Inspect(f, func(n ast.Node) bool {
			if decl, ok := n.(*ast.GenDecl); ok && decl.Doc != nil && checkComments(decl.Doc.List) {
				for _, spec := range decl.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					if st, ok := ts.Type.(*ast.StructType); ok {
						var fields []FieldData

						for _, field := range st.Fields.List {
							switch t := field.Type.(type) {
							case *ast.Ident:
								fields = append(fields, FieldData{
									Name: field.Names[0].Name,
									Type: t.Name,
								})
							case *ast.MapType:
								fields = append(fields, FieldData{
									Name: field.Names[0].Name,
									Type: "IsMap",
								})
							case *ast.ArrayType:
								fields = append(fields, FieldData{
									Name: field.Names[0].Name,
									Type: "IsSlice",
								})
							}
						}

						packages[f.Name.Name] = append(packages[f.Name.Name], TemplateData{
							Package:     f.Name.Name,
							PackagePath: filepath.Dir(path),
							StructName:  ts.Name.Name,
							Fields:      fields,
						})
					}
				}
			}
			return true
		})
		return nil
	})

	if err != nil {
		fmt.Printf("Walk error: %v\n", err)
	}

	for pkgName, structs := range packages {
		var buf bytes.Buffer

		data := map[string]any{
			"Package": pkgName,
			"Structs": structs,
		}

		if err := t.Execute(&buf, data); err != nil {
			fmt.Printf("Template error: %v\n", err)
			continue
		}

		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			fmt.Printf("Format error: %v\n", err)
			continue
		}

		outputPath := filepath.Join(structs[0].PackagePath, "reset.gen.go")
		if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
			fmt.Printf("Write error: %v\n", err)
			continue
		}

		fmt.Printf("Generated %s with %d struct(s)\n", outputPath, len(structs))
	}
}
