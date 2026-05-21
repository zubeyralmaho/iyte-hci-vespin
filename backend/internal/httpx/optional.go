package httpx

import "encoding/json"

// Optional represents a JSON value with tri-state semantics: omitted,
// present-with-null, or present-with-value. Used for PATCH request bodies
// where "field absent" and "field set to null" mean different things.
//
//	if req.Field.Set {
//	    if req.Field.Null {
//	        // clear the underlying value
//	    } else {
//	        // apply req.Field.Value
//	    }
//	}
//	// else: field was omitted, leave the current value alone
type Optional[T any] struct {
	Set   bool
	Null  bool
	Value T
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	if string(data) == "null" {
		o.Null = true
		var zero T
		o.Value = zero
		return nil
	}
	return json.Unmarshal(data, &o.Value)
}

func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.Set || o.Null {
		return []byte("null"), nil
	}
	return json.Marshal(o.Value)
}
