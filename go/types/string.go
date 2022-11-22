package types

// StringField is a string implementing the storage.Field interface.
type StringField string

// Encode encodes a string as UTF-8 bytes
func (v StringField) Encode() ([]byte, error) {
	return []byte(v), nil
}
