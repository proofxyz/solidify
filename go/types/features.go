// Package types defines frequently used data primitives for the solidify
// toolchain.
package types

// FeatureGroup is an implementation of storage.FeatureGroup treating the zero-
// and non-zero values separately.
type FeatureGroup struct {
	Type          string
	NonZeroValues []string
	zeroValue     *string
}

// Name returns the name (i.e. type) of the FeatureGroup.
// E.g. "Background"
func (g FeatureGroup) Name() string {
	return g.Type
}

// NumValues returns the number of values in this group.
// E.g. 3 for Backgrounds being either "None", "White" or "Black".
func (g FeatureGroup) NumValues() uint8 {
	return uint8(len(g.NonZeroValues) + 1)
}

// SetZeroValue sets the zero value of this feature. Defaults to "None"
func (g *FeatureGroup) SetZeroValue(zero string) {
	g.zeroValue = &zero
}

// ZeroValue returns te current zero value
func (g FeatureGroup) ZeroValue() string {
	if g.zeroValue != nil {
		return *g.zeroValue
	}

	return "None"
}

// Values returns the values in that the feature of given type can take.
// The zero value is always at first position.
func (g FeatureGroup) Values() []string {
	vs := []string{g.ZeroValue()}
	vs = append(vs, g.NonZeroValues...)
	return vs
}
