package gss

import (
	"sort"
	"strconv"
	"strings"

	"github.com/functionalfoundry/graphqlws"
	"github.com/graphql-go/graphql/language/ast"
)

type SubscribeFilter interface {
	ReplaceFieldsFromDocument(subscription *graphqlws.Subscription)
}

type subscribeFilter struct {
	SubscribeFilter
}

func NewSubscribeFilter() *subscribeFilter {
	return &subscribeFilter{}
}

type astArgs struct {
	Key string
	Val string
}

func operationDefinitionsWithOperation(
	doc *ast.Document,
	op string,
) []*ast.OperationDefinition {
	defs := []*ast.OperationDefinition{}
	for _, node := range doc.Definitions {
		if node.GetKind() == "OperationDefinition" {
			if def, ok := node.(*ast.OperationDefinition); ok {
				if def.Operation == op {
					defs = append(defs, def)
				}
			}
		}
	}
	return defs
}

func selectionSetsForOperationDefinitions(
	defs []*ast.OperationDefinition,
) []*ast.SelectionSet {
	sets := []*ast.SelectionSet{}
	for _, def := range defs {
		if set := def.GetSelectionSet(); set != nil {
			sets = append(sets, set)
		}
	}
	return sets
}

func ifToStr(d interface{}) string {
	if v, ok := d.(string); ok {
		return v
	}
	if v, ok := d.(int); ok {
		return strconv.Itoa(v)
	}
	return ""
}

func nameForSelectionSet(variables map[string]interface{}, set *ast.SelectionSet) ([]string, bool) {
	fieldStrs := []string{}
	for _, fieldIF := range set.Selections {
		field, ok := fieldIF.(*ast.Field)
		if !ok {
			continue
		}
		args := []astArgs{}
		for _, arg := range field.Arguments {
			val := arg.Value
			var kk string
			var vv interface{}
			if val.GetKind() == "Variable" {
				valName := val.GetValue().(*ast.Name)
				kk = valName.Value
				vv = variables[valName.Value]
			} else {
				kk = arg.Name.Value
				vv = arg.Value.GetValue()
			}
			if v := ifToStr(vv); v != "" {
				args = append(args, astArgs{
					Key: kk,
					Val: v,
				})
			}
		}
		sort.Slice(args, func(i, j int) bool {
			return args[i].Key <= args[j].Key
		})
		joinedArgs := []string{field.Name.Value}
		for _, a := range args {
			joinedArgs = append(joinedArgs, a.Val)
		}
		fieldStrs = append(fieldStrs, strings.Join(joinedArgs, ":"))
	}
	return fieldStrs, len(fieldStrs) > 0
}

func namesForSelectionSets(variables map[string]interface{}, sets []*ast.SelectionSet) []string {
	nameList := []string{}
	for _, set := range sets {
		if names, ok := nameForSelectionSet(variables, set); ok {
			nameList = append(nameList, names...)
		}
	}
	return nameList
}

func (f *subscribeFilter) ReplaceFieldsFromDocument(subscription *graphqlws.Subscription) {
	defs := operationDefinitionsWithOperation(subscription.Document, "subscription")
	sets := selectionSetsForOperationDefinitions(defs)
	fields := namesForSelectionSets(subscription.Variables, sets)
	subscription.Fields = fields
}
