package attributes

import (
	"testing"

	"github.com/divergencetech/ethier/erc721"
	"github.com/h-fam/errdiff"
)

func TestValidateAttributes(t *testing.T) {
	tests := []struct {
		name          string
		test, ref     []erc721.Attribute
		wantErrSubstr string
	}{
		{
			name: "Ok",
			test: []erc721.Attribute{
				{TraitType: "A", Value: "a"},
				{TraitType: "B", Value: "b"},
			},
			ref: []erc721.Attribute{
				{TraitType: "A", Value: "a"},
				{TraitType: "B", Value: "b"},
			},
		},
		{
			name: "With display type",
			test: []erc721.Attribute{
				{TraitType: "A", Value: "a"},
			},
			ref: []erc721.Attribute{
				{TraitType: "A", Value: "a", DisplayType: erc721.DisplayNumber},
			},
			wantErrSubstr: "attributes diff",
		},
		{
			name: "Commutative",
			test: []erc721.Attribute{
				{TraitType: "A", Value: "a"},
				{TraitType: "B", Value: "b"},
			},
			ref: []erc721.Attribute{
				{TraitType: "B", Value: "b"},
				{TraitType: "A", Value: "a"},
			},
		},
		{
			name: "Case sensitive",
			test: []erc721.Attribute{
				{TraitType: "A", Value: "a"},
			},
			ref: []erc721.Attribute{
				{TraitType: "A", Value: "A"},
			},
			wantErrSubstr: "attributes diff",
		},
		{
			name: "Different types",
			test: []erc721.Attribute{
				{TraitType: "A", Value: "a"},
			},
			ref: []erc721.Attribute{
				{TraitType: "B", Value: "a"},
			},
			wantErrSubstr: "attributes diff",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := make([]*erc721.Attribute, len(tt.test))
			for i, _ := range tt.test {
				g[i] = &tt.test[i]
			}

			w := make([]*erc721.Attribute, len(tt.ref))
			for i, _ := range tt.ref {
				w[i] = &tt.ref[i]
			}

			err := ValidateAttributes(w, g)
			if diff := errdiff.Substring(err, tt.wantErrSubstr); diff != "" {
				t.Errorf("unexpected error: %v", diff)
			}
		})
	}
}
