package maps

func Values[K comparable, V any](m map[K]V) []V {
	var vs []V
	for _, v := range m {
		vs = append(vs, v)
	}
	return vs
}

func ValuesFlattened[K comparable, V any](m map[K][]V) []V {
	var vs []V
	for _, v := range m {
		vs = append(vs, v...)
	}
	return vs
}

func Keys[K comparable, V any](m map[K]V) []K {
	var ks []K
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func TransformKeys[K comparable, V any, T comparable](m map[K]V, transformFn func(K) T) map[T]V {
	r := make(map[T]V)
	for k, v := range m {
		r[transformFn(k)] = v
	}
	return r
}

func TransformValues[K comparable, V any, T comparable](m map[K]V, transformFn func(V) T) map[K]T {
	r := make(map[K]T)
	for k, v := range m {
		r[k] = transformFn(v)
	}
	return r
}

func Transform[K1 comparable, V1 any, K2 comparable, V2 any](m map[K1]V1, transformFn func(K1, V1) (K2, V2)) map[K2]V2 {
	r := make(map[K2]V2)
	for k1, v1 := range m {
		k2, v2 := transformFn(k1, v1)
		r[k2] = v2
	}
	return r
}
