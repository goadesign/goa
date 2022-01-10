package expr

import (
	"fmt"
	"strings"
	"testing"
)

var (
	testrand = NewRandom("test")

	simpleResult        = resultType("a", String, "b", Int, view("default", "a", String, "b", Int), view("link", "a", String))
	simpleResultDefault = resultType("a", String, "b", Int)
	simpleResultLink    = resultType("a", String)

	embeddedResult        = resultType("r", simpleResult, view("default", "r:link", AsObject(simpleResult)))
	embeddedResultDefault = resultType("r", simpleResultLink)

	collectionResult        = collection(simpleResult)
	collectionResultDefault = collection(simpleResultDefault)
	collectionResultLink    = collection(simpleResultLink)

	collectionLinkView = object(String)
	compositeResult    = resultType("a", object(collectionResult), "b", String,
		view("default", "a", object(String), "b", String),
		view("link", "a", collectionLinkView))
	compositeResultDefault = resultType("a", object(collectionResultDefault), "b", String)
	compositeResultLink    = resultType("a", object(collectionResultLink))

	recursiveResult         = resultRecursive("a", String, view("default", "a", object(String)))
	embeddedRecursiveResult = resultType("a", String, "rec", recursiveResult)
)

func init() {
	vobj := (*collectionLinkView)[0]
	vobj.Attribute.Meta = map[string][]string{"view": {"link"}}
}

func TestProject(t *testing.T) {
	cases := []struct {
		Name     string
		Result   *ResultTypeExpr
		View     string
		Expected *ResultTypeExpr
	}{
		{"default", simpleResult, "default", simpleResultDefault},
		{"link", simpleResult, "link", simpleResultLink},
		{"embedded", embeddedResult, "default", embeddedResultDefault},
		{"collection-default", collectionResult, "default", collectionResultDefault},
		{"collection-link", collectionResult, "link", collectionResultLink},
		{"composite-default", compositeResult, "default", compositeResultDefault},
		{"composite-link", compositeResult, "link", compositeResultLink},
		{"recursive", recursiveResult, "default", embeddedRecursiveResult},
	}
	for _, k := range cases {
		t.Run(k.Name, func(t *testing.T) {
			projected, err := Project(k.Result, k.View)
			if err != nil {
				t.Fatal(err)
			}
			if !Equal(projected, k.Expected) {
				projected.AttributeExpr.Debug("got")
				k.Expected.AttributeExpr.Debug("expected")
				t.Errorf("got: %s, expected: %s\n", Hash(projected, false, true, true), Hash(k.Expected, false, true, true))
			}
			if pobj := AsObject(projected.AttributeExpr.Type); pobj != nil {
				for _, att := range *pobj {
					att2 := k.Expected.AttributeExpr.Find(att.Name)
					if att2 == nil {
						continue
					}
					if att.Attribute.Description != att2.Description {
						t.Errorf("got description %q, expected %q", att.Attribute.Description, att2.Description)
					}
				}
			}
		})
	}
}

// view is a helper function for building view expressions used in tests. name
// is the name of the view, attributes list the names of the attributes rendered
// by the view. name may use the format "name:view" in which case view is the
// name of the view used to render the attribute (when its type is a result
// type).
func view(name string, params ...interface{}) *ViewExpr {
	var obj Object = make([]*NamedAttributeExpr, len(params)/2)
	for i := 0; i < len(params); i += 2 {
		var (
			attName string
			attView string
		)
		{
			n := params[i].(string)
			elems := strings.Split(n, ":")
			attName = elems[0]
			if len(elems) > 1 {
				attView = elems[1]
			}
		}
		att := &AttributeExpr{Type: params[i+1].(DataType)}
		if attView != "" {
			att.Meta = MetaExpr{"view": []string{attView}}
		}
		obj[i/2] = &NamedAttributeExpr{Name: attName, Attribute: att}
	}
	att := &AttributeExpr{Type: &obj}
	return &ViewExpr{Name: name, AttributeExpr: att}
}

// resultType is a helper function that builds result type expressions used in
// tests. The arguments is a list of attribute name and type pairs followed by a
// list of view expressions, e.g.:
//
//    resultType("attr1", String, "attr2", Int, view1, view2)
//
func resultType(params ...interface{}) *ResultTypeExpr {
	var (
		views []*ViewExpr
		obj   Object
	)
	for i, p := range params {
		switch pt := p.(type) {
		case string:
			obj = append(obj, &NamedAttributeExpr{
				Name: params[i].(string),
				Attribute: &AttributeExpr{
					Type:        params[i+1].(DataType),
					Description: fmt.Sprintf("desc %s", params[i]),
				}})
		case *ViewExpr:
			views = append(views, pt)
		}
	}

	t := testrand.String()
	return &ResultTypeExpr{
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{Type: &obj},
			TypeName:      t,
		},
		Identifier: "vnd.application." + t,
		Views:      views,
	}
}

func collection(elemType *ResultTypeExpr) *ResultTypeExpr {
	return &ResultTypeExpr{
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: &Array{
					ElemType: &AttributeExpr{Type: elemType},
				},
			},
		},
		Views: elemType.Views,
	}
}

func resultRecursive(params ...interface{}) *ResultTypeExpr {
	rt := resultType(params...)
	recAtt := &NamedAttributeExpr{Name: "rec", Attribute: &AttributeExpr{Type: rt, Description: "desc rec"}}
	obj := AsObject(rt)
	*obj = append(*obj, recAtt)
	for _, v := range rt.Views {
		vObj := v.Type.(*Object)
		*vObj = append(*vObj, recAtt)
	}
	return rt
}
