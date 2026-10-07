// This fixture generates a result view with a required, located named union.
// It compiles the complete service and HTTP transport, then calls the original
// endpoint to prove that valid branches pass and missing or invalid values fail.
package generator

import (
	"testing"

	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestGenerateNamedUnionResultView(t *testing.T) {
	dir := generateViewedTransportModule(t, namedUnionViewDSL)
	writeGeneratedContractTest(t, dir, "checks", namedUnionViewTest)
	runGeneratedPackageTests(t, dir, "./...")
}

func TestGenerateFlattenedNamedUnionResultView(t *testing.T) {
	dir := generateViewedTransportModule(t, func() { namedUnionViewMappingDSL(true) })
	writeGeneratedContractTest(t, dir, "checks", namedUnionViewTest)
	runGeneratedPackageTests(t, dir, "./...")
}

func namedUnionViewDSL() {
	namedUnionViewMappingDSL(false)
}

// namedUnionViewMappingDSL keeps the same located types and result views;
// only the authored JSON mapping differs between the two generated callers.
func namedUnionViewMappingDSL(flatten bool) {
	d.API("named-union-view", func() { d.Description("Verify selected results with a located named union") })
	text := d.Type("Text", func() {
		d.Meta("struct:pkg:path", "content")
		d.Field(1, "value", d.String, "The non-empty text selected by the service", func() { d.MinLength(1) })
		d.Required("value")
	})
	choice := d.Type("Content", &expr.Union{TypeName: "Choice", Flatten: flatten}, func() {
		d.Meta("struct:pkg:path", "content")
		d.Attribute("text", text)
	})
	alias := d.Type("SelectedContent", choice, func() { d.Meta("struct:pkg:path", "content") })
	result := d.ResultType("application/vnd.named-union-view.record", func() {
		d.TypeName("Record")
		d.Field(1, "choice", alias, "The required content selected by the service")
		d.Required("choice")
		d.View("default", func() { d.Attribute("choice") })
	})
	collection := d.CollectionOf(result)
	d.Service("records", func() {
		d.Description("Return selected content through the declared result view")
		d.Method("read", func() {
			d.Description("Read the service-selected content and validate its view")
			d.Result(result)
			d.HTTP(func() { d.GET("/record"); d.Response(d.StatusOK) })
		})
		d.Method("list", func() {
			d.Description("Return an empty collection through its declared view")
			d.Result(collection)
			d.HTTP(func() { d.GET("/records"); d.Response(d.StatusOK) })
		})
	})
}

const namedUnionViewTest = `package checks_test
import (
 "context"
 "errors"
 goa "goa.design/goa/v3/pkg"
 "testing"
 genrecords "generated.local/gen/records"
 gencontent "generated.local/gen/content"
)
type service struct {choice gencontent.SelectedContent; missing bool}
func(s *service)Read(context.Context)(*genrecords.Record,error){if s.missing{return nil,nil};return &genrecords.Record{Choice:s.choice},nil}
func(s *service)List(context.Context)(genrecords.RecordCollection,error){return nil,nil}
func TestMissingObjectAndEmptyCollection(t *testing.T){
 endpoints:=genrecords.NewEndpoints(&service{missing:true})
 result,err:=endpoints.Read(t.Context(),nil)
 var failure *goa.ServiceError
 if !errors.As(err,&failure)||!failure.Fault||result!=nil{t.Errorf("missing result=%+v err=%v",result,err)}
 if _,err:=endpoints.List(t.Context(),nil);err!=nil{t.Errorf("empty collection: %v",err)}
}
func TestNamedUnionViewValidation(t *testing.T){
 for _,test:=range []struct{name string;choice gencontent.SelectedContent;invalid bool}{
  {"valid",gencontent.SelectedContent(gencontent.Content(gencontent.NewChoiceText(&gencontent.Text{Value:"hello"}))),false},
  {"unset",gencontent.SelectedContent{},true},
  {"invalid text",gencontent.SelectedContent(gencontent.Content(gencontent.NewChoiceText(&gencontent.Text{Value:""}))),true},
 }{
  t.Run(test.name,func(t *testing.T){
   endpoints:=genrecords.NewEndpoints(&service{choice:test.choice})
   result,err:=endpoints.Read(t.Context(),nil)
   if (err!=nil)!=test.invalid{t.Errorf("result=%+v err=%v want invalid=%v",result,err,test.invalid)}
  })
 }
}
`
