package models

import (
	"fmt"
	"math/big"

	"github.com/zclconf/go-cty/cty"
)

func CtyToMap(v cty.Value) (map[string]any, error) {
	if v.IsNull() {
		return map[string]any{}, nil
	}
	if !v.Type().IsObjectType() && !v.Type().IsMapType() {
		return nil, fmt.Errorf("expected an object, got %s", v.Type().FriendlyName())
	}

	converted, err := CtyToGo(v)
	if err != nil {
		return nil, err
	}
	return converted.(map[string]any), nil
}

func CtyToGo(v cty.Value) (any, error) {
	if !v.IsWhollyKnown() {
		return nil, fmt.Errorf("value is not fully known")
	}
	if v.IsNull() {
		return nil, nil
	}

	v, _ = v.Unmark()
	ty := v.Type()

	switch {
	case ty == cty.String:
		return v.AsString(), nil
	case ty == cty.Bool:
		return v.True(), nil
	case ty == cty.Number:
		return ctyNumberToGo(v.AsBigFloat()), nil
	case ty.IsObjectType() || ty.IsMapType():
		result := make(map[string]any, v.LengthInt())
		for it := v.ElementIterator(); it.Next(); {
			key, value := it.Element()
			converted, err := CtyToGo(value)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key.AsString(), err)
			}
			result[key.AsString()] = converted
		}
		return result, nil
	case ty.IsListType() || ty.IsTupleType() || ty.IsSetType():
		result := make([]any, 0, v.LengthInt())
		for it := v.ElementIterator(); it.Next(); {
			_, value := it.Element()
			converted, err := CtyToGo(value)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported value of type %s", ty.FriendlyName())
	}
}

func ctyNumberToGo(f *big.Float) any {
	if f.IsInt() {
		if i, accuracy := f.Int64(); accuracy == big.Exact && int64(int(i)) == i {
			return int(i)
		}
	}
	result, _ := f.Float64()
	return result
}
