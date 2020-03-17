package hclutils_test

import (
	"testing"

	"github.com/hashicorp/nomad/helper/pluginutils/hclspecutils"
	"github.com/hashicorp/nomad/helper/pluginutils/hclutils"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestParseNullFields(t *testing.T) {
	spec := hclspec.NewObject(map[string]*hclspec.Spec{
		"array_field":   hclspec.NewAttr("array_field", "list(string)", false),
		"string_field":  hclspec.NewAttr("string_field", "string", false),
		"boolean_field": hclspec.NewAttr("boolean_field", "bool", false),
		"number_field":  hclspec.NewAttr("number_field", "number", false),
		"block_field": hclspec.NewBlock("block_field", false, hclspec.NewObject((map[string]*hclspec.Spec{
			"f": hclspec.NewAttr("f", "string", true),
		}))),
		"block_list_field": hclspec.NewBlockList("block_list_field", hclspec.NewObject((map[string]*hclspec.Spec{
			"f": hclspec.NewAttr("f", "string", true),
		}))),
	})

	type Sub struct {
		F string `codec:"f"`
	}

	type TaskConfig struct {
		Array     []string `codec:"array_field"`
		String    string   `codec:"string_field"`
		Boolean   bool     `codec:"boolean_field"`
		Number    int64    `codec:"number_field"`
		Block     Sub      `codec:"block_field"`
		BlockList []Sub    `codec:"block_list_field"`
	}

	cases := []struct {
		name     string
		json     string
		expected TaskConfig
	}{
		{
			"omitted fields",
			`{"Config": {}}`,
			TaskConfig{BlockList: []Sub{}},
		},
		{
			"explicitly nil",
			`{"Config": {
                            "array_field": null,
                            "string_field": null,
			    "boolean_field": null,
                            "number_field": null,
                            "block_field": null,
                            "block_list_field": null}}`,
			TaskConfig{BlockList: []Sub{}},
		},
		{
			// for sanity checking that the fields are actually set
			"explicitly set to not null",
			`{"Config": {
                            "array_field": ["a"],
                            "string_field": "a",
                            "boolean_field": true,
                            "number_field": 5,
                            "block_field": [{"f": "a"}],
                            "block_list_field": [{"f": "a"}, {"f": "b"}]}}`,
			TaskConfig{
				Array:     []string{"a"},
				String:    "a",
				Boolean:   true,
				Number:    5,
				Block:     Sub{"a"},
				BlockList: []Sub{{"a"}, {"b"}},
			},
		},
	}

	parser := hclutils.NewConfigParser(spec)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var tc TaskConfig
			parser.ParseJson(t, c.json, &tc)

			require.EqualValues(t, c.expected, tc)
		})
	}
}

func TestParseUnknown(t *testing.T) {
	spec := hclspec.NewObject(map[string]*hclspec.Spec{
		"string_field":   hclspec.NewAttr("string_field", "string", false),
		"map_field":      hclspec.NewAttr("map_field", "map(string)", false),
		"list_field":     hclspec.NewAttr("list_field", "map(string)", false),
		"map_list_field": hclspec.NewAttr("map_list_field", "list(map(string))", false),
	})
	cSpec, diags := hclspecutils.Convert(spec)
	require.False(t, diags.HasErrors())

	cases := []struct {
		name string
		hcl  string
	}{
		{
			"string field",
			`config {  string_field = "${MYENV}" }`,
		},
		{
			"map_field",
			`config { map_field { key = "${MYENV}" }}`,
		},
		{
			"list_field",
			`config { list_field = ["${MYENV}"]}`,
		},
		{
			"map_list_field",
			`config { map_list_field { key = "${MYENV}"}}`,
		},
	}

	vars := map[string]cty.Value{}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inter := hclutils.HclConfigToInterface(t, c.hcl)

			ctyValue, diag, errs := hclutils.ParseHclInterface(inter, cSpec, vars)
			t.Logf("parsed: %# v", pretty.Formatter(ctyValue))

			require.NotNil(t, errs)
			require.True(t, diag.HasErrors())
			require.Contains(t, errs[0].Error(), "no variable named")
		})
	}
}
