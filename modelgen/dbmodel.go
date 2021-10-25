package modelgen

import (
	"encoding/json"
	"sort"
	"text/template"

	"github.com/ovn-org/libovsdb/ovsdb"
)

// NewDBTemplate return a new ClientDBModel template
// It includes the following other templates that can be overridden to customize the generated file
// "header"
// "preDBDefinitions"
// "postDBDefinitions"
// It is design to be used with a map[string] interface and some defined keys (see GetDBTemplateData)
func NewDBTemplate() *template.Template {
	return template.Must(template.New("").Funcs(
		template.FuncMap{
			"escape": escape,
		},
	).Parse(`
{{- define "header" }}
// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.
{{- end }}
{{- define "preDBDefinitions" }}
 import (
	"encoding/json"

	"github.com/ovn-org/libovsdb/model"
	"github.com/ovn-org/libovsdb/ovsdb"
)
{{- end }}
{{ define "postDBDefinitions" }}{{ end }}
{{ template "header" . }}

package {{ index . "PackageName" }}

{{ template "preDBDefinitions" }}

// FullDatabaseModel returns the DatabaseModel object to be used in libovsdb
func FullDatabaseModel() (*model.ClientDBModel, error) {
	return model.NewClientDBModel("{{ index . "DatabaseName" }}", map[string]model.Model{
    {{ range index . "Tables" }} "{{ .TableName }}" : &{{ .StructName }}{}, 
    {{ end }}
	})
}

var schema = {{ index . "Schema" | escape }}

func Schema() ovsdb.DatabaseSchema {
	var s ovsdb.DatabaseSchema
	err := json.Unmarshal([]byte(schema), &s)
	if err != nil {
		panic(err)
	}
	return s
}

{{ template "postDBDefinitions" . }}
`))
}

//TableInfo represents the information of a table needed by the Model template
type TableInfo struct {
	TableName  string
	StructName string
}

// GetDBTemplateData returns the map needed to execute the DBTemplate. It has the following keys:
// DatabaseName: (string) the database name
// PackageName : (string) the package name
// Tables: []Table list of Tables that form the Model
func GetDBTemplateData(pkg string, schema *ovsdb.DatabaseSchema) map[string]interface{} {
	data := map[string]interface{}{}
	data["DatabaseName"] = schema.Name
	data["PackageName"] = pkg
	schemaBytes, _ := json.MarshalIndent(schema, "", "  ")
	data["Schema"] = string(schemaBytes)
	tables := []TableInfo{}

	var order sort.StringSlice
	for tableName := range schema.Tables {
		order = append(order, tableName)
	}
	order.Sort()

	for _, tableName := range order {
		tables = append(tables, TableInfo{
			TableName:  tableName,
			StructName: StructName(tableName),
		})
	}
	data["Tables"] = tables
	return data
}

func escape(s string) string {
	return "`" + s + "`"
}
