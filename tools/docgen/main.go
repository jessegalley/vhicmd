package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const docTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>VHI API Documentation</title>
    <style>
        body {
            font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', monospace;
            background-color: #1E1E2E; /* Base */
            color: #CDD6F4;           /* Text */
            line-height: 1.7;
            padding: 24px;
            margin: 0;
        }
        h1, h2, h3 {
            color: #FAB387; /* Peach */
            font-weight: 600;
            margin-bottom: 0.5em;
        }
        pre {
            background-color: #313244; /* Surface0 */
            padding: 12px;
            border-radius: 8px;
            font-size: 0.95rem;
            overflow-x: auto;
        }
        .function-name {
            color: #89B4FA; /* Blue */
            font-weight: 600;
        }
        .function-doc {
            color: #A6E3A1; /* Green */
            margin-bottom: 6px;
            font-style: italic;
        }
        .param {
            color: #F9E2AF; /* Yellow */
        }
        a {
            color: #89B4FA; /* Blue */
            text-decoration: underline;
            transition: color 0.2s;
        }
        a:hover {
            color: #74C7EC; /* Sky */
        }
        ul {
            padding-left: 20px;
        }
        li {
            margin-bottom: 4px;
        }
        code {
            font-family: 'JetBrains Mono', 'Fira Code', monospace;
            font-size: 0.95rem;
            background-color: #45475A; /* Surface1 */
            padding: 2px 4px;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <h1>VHI API Documentation</h1>
    {{ .Content }}

    <script>
    document.addEventListener('DOMContentLoaded', function() {
        // Make each group (h2 and its subsequent content) collapsible.
        // For every h2, wrap all sibling nodes until the next h2 in a div.
        var h2s = document.querySelectorAll('body > h2');
        h2s.forEach(function(header) {
            // Create an arrow span if not already present.
            var arrowSpan = document.createElement('span');
            arrowSpan.className = 'arrow';
            // Groups are expanded by default.
            arrowSpan.textContent = ' ▼';
            header.appendChild(arrowSpan);

            // Create a container for all content following the header until the next h2.
            var groupContent = document.createElement('div');
            groupContent.className = 'group-content';
            var next = header.nextSibling;
            while (next && !(next.nodeType === 1 && next.tagName === 'H2')) {
                var current = next;
                next = next.nextSibling;
                groupContent.appendChild(current);
            }
            header.parentNode.insertBefore(groupContent, next);

            // Toggle collapse/expand on header click.
            header.addEventListener('click', function() {
                if (groupContent.style.display === 'none') {
                    groupContent.style.display = 'block';
                    arrowSpan.textContent = ' ▼';
                } else {
                    groupContent.style.display = 'none';
                    arrowSpan.textContent = ' ►';
                }
            });
        });
    });
    </script>
</body>
</html>
`

type FunctionDoc struct {
	Name    string
	Params  string
	Returns string
	Comment string
}

type StructDoc struct {
	Name    string
	Comment string
	Fields  []FieldDoc
}

type FieldDoc struct {
	Name    string
	Type    string
	Tag     string
	Comment string
}

func main() {
	// Generate the documentation text (in HTML form).
	docContent, err := generateDocumentation()
	if err != nil {
		log.Fatalf("Error generating documentation: %v", err)
	}

	// Ensure docs directory exists.
	if err := os.MkdirAll("docs", 0755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	// Create "docs/api.html" file.
	outFile, err := os.Create(filepath.Join("docs", "api.html"))
	if err != nil {
		log.Fatalf("Failed to create HTML file: %v", err)
	}
	defer outFile.Close()

	// Render docContent into the HTML template.
	tmpl := template.Must(template.New("doc").Parse(docTemplate))
	data := struct {
		Content template.HTML
	}{
		Content: template.HTML(docContent), // The doc we generated.
	}

	if err := tmpl.Execute(outFile, data); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	log.Println("Documentation written to docs/api.html")
}

// generateDocumentation parses the `api` directory using Go’s AST,
// grouping exported functions and structs by file name, and produces HTML content.
func generateDocumentation() (string, error) {
	var buf bytes.Buffer

	// Path to your package containing API-related code.
	apiDir := filepath.Join(".", "api")

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, apiDir, func(fi os.FileInfo) bool {
		// Skip doc.go to avoid re-including generated docs.
		return fi.Name() != "doc.go"
	}, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse package: %v", err)
	}

	// Intro / Package-level doc.
	buf.WriteString(`<p>Package <strong>api</strong> provides a VHI/OpenStack client implementation.</p>`)

	// Define the groups of files you want to document.
	components := map[string][]string{
		"Authentication":   {"auth.go"},
		"Catalog":          {"catalog.go"},
		"Domains":          {"domains.go"},
		"Flavors":          {"flavors.go"},
		"Images":           {"images.go", "image_structs.go"},
		"Networks":         {"networks.go", "ports.go"},
		"Projects":         {"projects.go"},
		"Storage":          {"volumes.go", "volume_structs.go"},
		"Virtual Machines": {"vm.go", "vm_structs.go"},
	}

	// Collect and sort group names alphabetically.
	groupNames := make([]string, 0, len(components))
	for groupName := range components {
		groupNames = append(groupNames, groupName)
	}
	sort.Strings(groupNames)

	// Iterate over groups in alphabetical order.
	for _, groupName := range groupNames {
		buf.WriteString(fmt.Sprintf("<h2>%s</h2>\n", groupName))

		var functions []FunctionDoc
		var structs []StructDoc

		// Collect functions and structs from the files in the group.
		for _, targetFile := range components[groupName] {
			for _, pkg := range pkgs {
				for filename, file := range pkg.Files {
					if !strings.HasSuffix(filename, targetFile) {
						continue
					}

					// Examine each top-level declaration.
					for _, decl := range file.Decls {
						switch d := decl.(type) {
						case *ast.FuncDecl:
							// Only consider exported functions without receivers.
							if !d.Name.IsExported() || d.Recv != nil {
								continue
							}
							params := getParamList(d.Type.Params)
							returns := getParamList(d.Type.Results)
							comment := ""
							if d.Doc != nil {
								comment = strings.ReplaceAll(strings.TrimSpace(d.Doc.Text()), "\n", "<br/>")
							}
							functions = append(functions, FunctionDoc{
								Name:    d.Name.Name,
								Params:  params,
								Returns: returns,
								Comment: comment,
							})
						case *ast.GenDecl:
							// Look for type declarations (structs).
							if d.Tok == token.TYPE {
								for _, spec := range d.Specs {
									if ts, ok := spec.(*ast.TypeSpec); ok {
										if st, ok := ts.Type.(*ast.StructType); ok {
											var fields []FieldDoc
											// Process each field in the struct.
											for _, field := range st.Fields.List {
												fieldType := getTypeString(field.Type)
												tag := ""
												if field.Tag != nil {
													tag = field.Tag.Value
												}
												fieldComment := ""
												if field.Comment != nil {
													fieldComment = strings.TrimSpace(field.Comment.Text())
												}
												// If no names are provided, it is an embedded field.
												if len(field.Names) == 0 {
													fields = append(fields, FieldDoc{
														Name:    fieldType,
														Type:    fieldType,
														Tag:     tag,
														Comment: fieldComment,
													})
												} else {
													for _, name := range field.Names {
														fields = append(fields, FieldDoc{
															Name:    name.Name,
															Type:    fieldType,
															Tag:     tag,
															Comment: fieldComment,
														})
													}
												}
											}
											structComment := ""
											if d.Doc != nil {
												structComment = strings.TrimSpace(d.Doc.Text())
											}
											structs = append(structs, StructDoc{
												Name:    ts.Name.Name,
												Comment: structComment,
												Fields:  fields,
											})
										}
									}
								}
							}
						}
					}
				}
			}
		}

		// Sort functions and structs alphabetically by name.
		sort.Slice(functions, func(i, j int) bool {
			return functions[i].Name < functions[j].Name
		})
		sort.Slice(structs, func(i, j int) bool {
			return structs[i].Name < structs[j].Name
		})

		// Render function documentation.
		for _, fn := range functions {
			buf.WriteString("<div style='margin-bottom: 1em;'>\n")
			if fn.Comment != "" {
				buf.WriteString(fmt.Sprintf("<p class='function-doc'>%s</p>\n", fn.Comment))
			}
			buf.WriteString(fmt.Sprintf(
				"<code><span class='function-name'>%s</span>(%s) (%s)</code>\n",
				fn.Name, fn.Params, fn.Returns,
			))
			buf.WriteString("</div>\n")
		}

		// Render struct documentation as code blocks.
		for _, s := range structs {
			buf.WriteString("<div style='margin-bottom: 1em;'>\n")
			if s.Comment != "" {
				buf.WriteString(fmt.Sprintf("<p class='function-doc'>%s</p>\n", s.Comment))
			}
			buf.WriteString("<pre><code>")
			buf.WriteString(fmt.Sprintf("type %s struct {\n", s.Name))
			// Render each field.
			for _, field := range s.Fields {
				line := "    "
				// For a normal field, output "Name Type".
				// For an embedded field, Name and Type will be identical.
				if field.Name != field.Type {
					line += fmt.Sprintf("%s %s", field.Name, field.Type)
				} else {
					line += field.Name
				}
				if field.Tag != "" {
					line += " " + field.Tag
				}
				if field.Comment != "" {
					line += " // " + field.Comment
				}
				line += "\n"
				buf.WriteString(line)
			}
			buf.WriteString("}\n")
			buf.WriteString("</code></pre>\n")
			buf.WriteString("</div>\n")
		}
	}

	// Common parameters summary.
	buf.WriteString(`
<h2>Common Parameters</h2>
<ul>
  <li><strong>computeURL</strong>: Nova API endpoint</li>
  <li><strong>storageURL</strong>: Cinder API endpoint</li>
  <li><strong>networkURL</strong>: Neutron API endpoint</li>
  <li><strong>imageURL</strong>: Glance API endpoint</li>
  <li><strong>token</strong>: Authentication token</li>
  <li><strong>queryParams</strong>: Optional URL parameters (<code>map[string]string</code>)</li>
</ul>
`)

	return buf.String(), nil
}

// getParamList constructs a parameter string from the AST FieldList.
func getParamList(fields *ast.FieldList) string {
	if fields == nil {
		return ""
	}
	var params []string
	for _, field := range fields.List {
		typ := getTypeString(field.Type)
		if len(field.Names) == 0 {
			// unnamed parameter.
			params = append(params, typ)
		} else {
			for _, name := range field.Names {
				// <span> styling on param names.
				params = append(params, fmt.Sprintf("<span class='param'>%s</span> %s", name.Name, typ))
			}
		}
	}
	return strings.Join(params, ", ")
}

// getTypeString returns a string representation of a Go AST Expr.
func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", getTypeString(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeString(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getTypeString(t.Key), getTypeString(t.Value))
	default:
		return fmt.Sprintf("%T", expr)
	}
}
