// This file verifies that cascading a View(...) onto collection attributes
// propagates through service-to-view conversions, view-to-service conversions,
// and view validators across multi-level, map, user-type, union, and recursive
// result designs.
package service

import (
	"bytes"
	"go/format"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

// TestCascadedResultViews verifies generated service conversions and view
// validators when parent views select nested views on ArrayOf, MapOf, and
// CollectionOf attributes.
func TestCascadedResultViews(t *testing.T) {
	for _, tc := range []struct {
		name               string
		dsl                func()
		wantServiceSubstrs []string
		omitServiceSubstrs []string
		wantViewsSubstrs   []string
		omitViewsSubstrs   []string
	}{
		{
			name: "three-level array cascade (#4078)",
			dsl: func() {
				label := dsl.ResultType("application/vnd.test.label", func() {
					dsl.TypeName("Label")
					dsl.Attributes(func() {
						dsl.Attribute("id", dsl.String)
						dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
						dsl.Required("id")
					})
					dsl.View("default", func() {
						dsl.Attribute("id")
					})
					dsl.View("extended", func() {
						dsl.Attribute("id")
						dsl.Attribute("name")
					})
				})
				item := dsl.ResultType("application/vnd.test.item", func() {
					dsl.TypeName("Item")
					dsl.Attributes(func() {
						dsl.Attribute("id", dsl.String)
						dsl.Attribute("labels", dsl.ArrayOf(label))
						dsl.Required("id")
					})
					dsl.View("default", func() {
						dsl.Attribute("id")
						dsl.Attribute("labels")
					})
					dsl.View("extended", func() {
						dsl.Attribute("id")
						dsl.Attribute("labels", func() {
							dsl.View("extended")
						})
					})
				})
				order := dsl.ResultType("application/vnd.test.order", func() {
					dsl.TypeName("Order")
					dsl.Attributes(func() {
						dsl.Attribute("id", dsl.String)
						dsl.Attribute("items", dsl.ArrayOf(item))
						dsl.Required("id")
					})
					dsl.View("default", func() {
						dsl.Attribute("id")
						dsl.Attribute("items")
					})
					dsl.View("extended", func() {
						dsl.Attribute("id")
						dsl.Attribute("items", func() {
							dsl.View("extended")
						})
					})
				})
				dsl.Service("Orders", func() {
					dsl.Method("Get", func() {
						dsl.Result(order)
					})
				})
			},
			wantServiceSubstrs: []string{
				"newItemView(val)",
				"newItemViewExtended(val)",
				"newLabelView(val)",
				"newLabelViewExtended(val)",
				"newItem(val)",
				"newItemExtended(val)",
				"newLabel(val)",
				"newLabelExtended(val)",
			},
			wantViewsSubstrs: []string{
				"ValidateItemView(e)",
				"ValidateItemViewExtended(e)",
				"ValidateLabelView(e)",
				"ValidateLabelViewExtended(e)",
			},
		},
		{
			name: "map and nested array-of-map cascade with attribute view override",
			dsl: func() {
				label := dsl.ResultType("application/vnd.test.label", func() {
					dsl.TypeName("Label")
					dsl.Attributes(func() {
						dsl.Attribute("id", dsl.String)
						dsl.Attribute("code", dsl.String)
						dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
					})
					dsl.View("tiny", func() {
						dsl.Attribute("id")
					})
					dsl.View("default", func() {
						dsl.Attribute("id")
						dsl.Attribute("code")
					})
					dsl.View("extended", func() {
						dsl.Attribute("id")
						dsl.Attribute("code")
						dsl.Attribute("name")
					})
				})
				order := dsl.ResultType("application/vnd.test.order", func() {
					dsl.TypeName("Order")
					dsl.Attributes(func() {
						dsl.Attribute("by_key", dsl.MapOf(dsl.String, label, func() {
							dsl.Elem(func() { dsl.View("tiny") })
						}))
						dsl.Attribute("groups", dsl.ArrayOf(dsl.MapOf(dsl.String, label)))
					})
					dsl.View("default", func() {
						dsl.Attribute("by_key")
						dsl.Attribute("groups")
					})
					dsl.View("extended", func() {
						dsl.Attribute("by_key", func() { dsl.View("extended") })
						dsl.Attribute("groups", func() { dsl.View("extended") })
					})
				})
				dsl.Service("Orders", func() {
					dsl.Method("Get", func() {
						dsl.Result(order)
					})
				})
			},
			wantServiceSubstrs: []string{
				"newLabelViewTiny(val)",
				"newLabelView(val)",
				"newLabelViewExtended(val)",
				"newLabelTiny(val)",
				"newLabel(val)",
				"newLabelExtended(val)",
			},
			wantViewsSubstrs: []string{
				"func ValidateOrderViewExtended(",
				"ValidateLabelViewExtended(v)",
			},
			omitViewsSubstrs: []string{
				"func ValidateOrderView(",
				"func ValidateLabelView(",
				"func ValidateLabelViewTiny(",
			},
		},
		{
			name: "shared CollectionOf result type across sibling fields with distinct views",
			dsl: func() {
				item := dsl.ResultType("application/vnd.test.item", func() {
					dsl.TypeName("Item")
					dsl.Attributes(func() {
						dsl.Attribute("id", dsl.String)
						dsl.Attribute("detail", dsl.String, func() { dsl.MinLength(2) })
						dsl.Required("id")
					})
					dsl.View("default", func() { dsl.Attribute("id") })
					dsl.View("extended", func() {
						dsl.Attribute("id")
						dsl.Attribute("detail")
					})
				})
				itemCollection := dsl.CollectionOf(item)
				order := dsl.ResultType("application/vnd.test.order", func() {
					dsl.TypeName("Order")
					dsl.Attributes(func() {
						dsl.Attribute("summary_items", itemCollection)
						dsl.Attribute("detailed_items", itemCollection)
					})
					dsl.View("default", func() {
						dsl.Attribute("summary_items")
						dsl.Attribute("detailed_items", func() { dsl.View("extended") })
					})
					dsl.View("extended", func() {
						dsl.Attribute("summary_items", func() { dsl.View("extended") })
						dsl.Attribute("detailed_items")
					})
				})
				dsl.Service("Orders", func() {
					dsl.Method("Get", func() {
						dsl.Result(order)
					})
				})
			},
			wantServiceSubstrs: []string{
				"vres.SummaryItems = newItemCollectionView(res.SummaryItems)",
				"vres.DetailedItems = newItemCollectionViewExtended(res.DetailedItems)",
				"vres.SummaryItems = newItemCollectionViewExtended(res.SummaryItems)",
				"vres.DetailedItems = newItemCollectionView(res.DetailedItems)",
				"res.SummaryItems = newItemCollection(vres.SummaryItems)",
				"res.DetailedItems = newItemCollectionExtended(vres.DetailedItems)",
				"res.SummaryItems = newItemCollectionExtended(vres.SummaryItems)",
				"res.DetailedItems = newItemCollection(vres.DetailedItems)",
			},
			wantViewsSubstrs: []string{
				"ValidateItemCollectionView(result.SummaryItems)",
				"ValidateItemCollectionViewExtended(result.DetailedItems)",
				"ValidateItemCollectionViewExtended(result.SummaryItems)",
				"ValidateItemCollectionView(result.DetailedItems)",
			},
		},
		{
			name: "recursive result type via ArrayOf cascades extended view",
			dsl: func() {
				node := dsl.ResultType("application/vnd.test.node", func() {
					dsl.TypeName("Node")
					dsl.Attributes(func() {
						dsl.Attribute("id", dsl.String)
						dsl.Attribute("detail", dsl.String, func() { dsl.MinLength(2) })
						dsl.Attribute("children", dsl.ArrayOf("Node"))
						dsl.Required("id")
					})
					dsl.View("default", func() {
						dsl.Attribute("id")
						dsl.Attribute("children")
					})
					dsl.View("extended", func() {
						dsl.Attribute("id")
						dsl.Attribute("detail")
						dsl.Attribute("children", func() {
							dsl.View("extended")
						})
					})
				})
				dsl.Service("Tree", func() {
					dsl.Method("Get", func() {
						dsl.Result(node)
					})
				})
			},
			wantServiceSubstrs: []string{
				"newNodeView(val)",
				"newNodeViewExtended(val)",
				"newNode(val)",
				"newNodeExtended(val)",
			},
			wantViewsSubstrs: []string{
				"ValidateNodeView(e)",
				"ValidateNodeViewExtended(e)",
			},
		},
		{
			name: "unvalidated default child view inside user type and union does not require missing validator",
			dsl: func() {
				label := dsl.ResultType("application/vnd.test.label", func() {
					dsl.TypeName("Label")
					dsl.Attributes(func() {
						dsl.Attribute("id", dsl.String)
						dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
					})
					dsl.View("default", func() { dsl.Attribute("id") })
					dsl.View("extended", func() {
						dsl.Attribute("id")
						dsl.Attribute("name")
					})
				})
				wrapper := dsl.Type("LabelWrapper", func() {
					dsl.Attribute("default_label", label)
					dsl.Attribute("extended_label", label, func() { dsl.View("extended") })
				})
				order := dsl.ResultType("application/vnd.test.order", func() {
					dsl.TypeName("Order")
					dsl.Attributes(func() {
						dsl.Attribute("wrapper", wrapper)
						dsl.OneOf("choice", func() {
							dsl.Attribute("plain", label)
							dsl.Attribute("rich", label, func() { dsl.View("extended") })
						})
					})
					dsl.View("default", func() {
						dsl.Attribute("wrapper")
						dsl.Attribute("choice")
					})
				})
				dsl.Service("Orders", func() {
					dsl.Method("Get", func() {
						dsl.Result(order)
					})
				})
			},
			wantViewsSubstrs: []string{
				"func ValidateLabelWrapperView(",
				"ValidateLabelViewExtended(",
			},
			omitViewsSubstrs: []string{
				"func ValidateLabelView(",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := codegen.RunDSL(t, tc.dsl)
			plan := mustServicePlan(t, root)
			require.Len(t, root.Services, 1)

			files := mustServiceFiles(t, plan)
			var svcFile *codegen.File
			for _, f := range files {
				if filepath.Base(f.Path) == "service.go" {
					svcFile = f
					break
				}
			}
			require.NotNil(t, svcFile)
			serviceCode := renderFormattedSections(t, svcFile.SectionTemplates)

			vf := viewsFile(plan, plan.facts.services[0])
			require.NotNil(t, vf)
			viewsCode := renderFormattedSections(t, vf.SectionTemplates)

			for _, sub := range tc.wantServiceSubstrs {
				assert.Contains(t, serviceCode, sub)
			}
			for _, sub := range tc.omitServiceSubstrs {
				assert.NotContains(t, serviceCode, sub)
			}
			for _, sub := range tc.wantViewsSubstrs {
				assert.Contains(t, viewsCode, sub)
			}
			for _, sub := range tc.omitViewsSubstrs {
				assert.NotContains(t, viewsCode, sub)
			}
		})
	}
}

// renderFormattedSections renders all template sections of a generated Go file
// and verifies that go/format accepts the output.
func renderFormattedSections(t *testing.T, sections []*codegen.SectionTemplate) string {
	t.Helper()
	var buf bytes.Buffer
	for _, section := range sections {
		require.NoError(t, section.Write(&buf))
	}
	formatted, err := format.Source(buf.Bytes())
	require.NoError(t, err, buf.String())
	return string(formatted)
}
