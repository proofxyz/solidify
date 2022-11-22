package extractor

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/h-fam/errdiff"
	"github.com/proofxyz/solidify/go/types"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		fMaps []FeaturesMap
		want  []types.FeatureGroup
	}{

		{
			fMaps: []FeaturesMap{
				{"t0": "f01", "t1": "f11"},
			},
			want: []types.FeatureGroup{
				{
					Type:          "t0",
					NonZeroValues: []string{"f01"},
				},
				{
					Type:          "t1",
					NonZeroValues: []string{"f11"},
				},
			},
		},
		{
			fMaps: []FeaturesMap{
				{"t0": "f01", "t1": "f11"},
				{},
			},
			want: []types.FeatureGroup{
				{
					Type:          "t0",
					NonZeroValues: []string{"f01"},
				},
				{
					Type:          "t1",
					NonZeroValues: []string{"f11"},
				},
			},
		},
		{
			fMaps: []FeaturesMap{
				{"t2": "f21", "t0": "f02"},
				{"t1": "f11", "t0": "f01", "t2": "None"},
			},
			want: []types.FeatureGroup{
				{
					Type:          "t0",
					NonZeroValues: []string{"f01", "f02"},
				},
				{
					Type:          "t1",
					NonZeroValues: []string{"f11"},
				},
				{
					Type:          "t2",
					NonZeroValues: []string{"f21"},
				},
			},
		},
	}

	ignore := cmpopts.IgnoreUnexported(types.FeatureGroup{})
	for _, tt := range tests {
		got := ExtractFeatureGroups(tt.fMaps)
		if diff := cmp.Diff(tt.want, got, ignore); diff != "" {
			t.Errorf("ExtractFeatureGroups(%+v) diff (-want +got):\n%s", tt.fMaps, diff)
		}
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		fMaps            []FeaturesMap
		fGroups          []types.FeatureGroup
		want             [][]uint8
		wantErrSubstring string
	}{
		{
			fMaps: []FeaturesMap{
				{"t0": "f01", "t1": "f11"},
			},
			fGroups: []types.FeatureGroup{
				{
					Type:          "t0",
					NonZeroValues: []string{"f01"},
				},
				{
					Type:          "t1",
					NonZeroValues: []string{"f10", "f11"},
				},
			},
			want: [][]uint8{
				{1, 2},
			},
		},
		{
			fMaps: []FeaturesMap{
				{"t0": "f01", "t1": "f11"},
				{},
			},
			fGroups: []types.FeatureGroup{
				{
					Type:          "t0",
					NonZeroValues: []string{"f01"},
				},
				{
					Type:          "t1",
					NonZeroValues: []string{"f11"},
				},
			},
			want: [][]uint8{
				{1, 1},
				{0, 0},
			},
		},
		{
			fMaps: []FeaturesMap{
				{"t0": "XXX"},
			},
			fGroups: []types.FeatureGroup{
				{
					Type:          "t0",
					NonZeroValues: []string{"f01"},
				},
			},
			wantErrSubstring: "unknown feature value",
		},
	}

	for _, tt := range tests {
		got, err := EncodeFeatures(tt.fMaps, tt.fGroups)

		if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
			t.Fatalf("EncodeFeatures([fMaps], [fGroups]) unexpected error %s", diff)
		}

		if tt.wantErrSubstring != "" {
			continue
		}

		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("EncodeFeatures([fMaps], [fGroups]) = %v, want %v", got, tt.want)
		}
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		json string
		want []FeaturesMap
	}{
		{
			`[
				{
					"Specie": "Owl",
					"Eyes": "Angry - Yellow",
					"Outerwear": "Hoodie Down",
					"Headwear": "None"
				},
				{
					"Specie": "Owl",
					"Eyes": "Angry - Purple",
					"Outerwear": "None",
					"Eyewear": "Eyepatch",
					"Headwear": "Skully"
				}
			]`,
			[]FeaturesMap{
				{
					"Specie":    "Owl",
					"Eyes":      "Angry - Yellow",
					"Outerwear": "Hoodie Down",
					"Headwear":  "None",
				},
				{
					"Specie":    "Owl",
					"Eyes":      "Angry - Purple",
					"Headwear":  "Skully",
					"Eyewear":   "Eyepatch",
					"Outerwear": "None",
				},
			},
		},
	}

	for _, tt := range tests {
		got, err := ParseFeaturesJSON(strings.NewReader(tt.json))
		if err != nil {
			t.Fatalf("Failed to parse features json: %s", err)
		}

		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Fatalf("ParseFeaturesJSON([json]) diff (-want +got)\n%s", diff)
		}
	}
}
