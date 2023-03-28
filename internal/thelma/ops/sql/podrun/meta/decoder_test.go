package meta

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_FromMapIgnoresNonmatchingPrefix(t *testing.T) {
	type annotations struct {
		A         string
		B         string
		MiXeDcAsE string
		Empty     string `k8smeta:",omitempty"`
		CustomKey string `k8smeta:"woohoo"`
	}

	m := map[string]string{
		"p/A":         "one",
		"p/B":         "two",
		"p/C":         "not in struct, should be ignored",
		"p/MiXeDcAsE": "three",
		"whatever":    "missing prefix, should be ignored",
		"p/woohoo":    "four",
	}

	var d Decoder[annotations]
	d.Prefix = "p/"

	expected := annotations{
		A:         "one",
		B:         "two",
		MiXeDcAsE: "three",
		CustomKey: "four",
	}
	actual, err := d.FromMap(m)
	require.NoError(t, err)
	assert.Equal(t, expected, *actual)

	m2, err := d.ToMap(expected)
	require.NoError(t, err)
	assert.Equal(t, m2, map[string]string{
		"p/A":         "one",
		"p/B":         "two",
		"p/MiXeDcAsE": "three",
		"p/woohoo":    "four",
		// Empty field should be omitted
	})
}

func Test_Annotations(t *testing.T) {
	type MyStruct struct {
		Foo string `json:"fooval"`
		OK  bool   `json:"okval"`
	}

	type annotations struct {
		Bool        bool                 `annotation:"boolval"`
		EmptyStruct struct{}             `annotation:"emptystructval"`
		Float       float64              `annotation:"floatval"`
		Int         int                  `annotation:"intval"`
		MapVal      map[string]time.Time `annotation:"mapval"`
		SliceVal    []int                `annotation:"sliceval"`
		String      string               `annotation:"stringval"`
		Struct      MyStruct             `annotation:"mystructval"`
		Time        time.Time            `annotation:"timestamp"`
	}

	d := Decoder[annotations]{
		Prefix:  "my.annotation.prefix/",
		TagName: "annotation",
	}

	now := time.Now().Round(time.Millisecond)
	nowbs, err := json.Marshal(now)
	require.NoError(t, err)
	nows := string(nowbs)

	mapval := map[string]time.Time{"now": now}
	mapvalbs, err := json.Marshal(mapval)
	require.NoError(t, err)

	a := annotations{
		Bool:        true,
		EmptyStruct: struct{}{},
		Float:       1.2,
		Int:         42,
		MapVal:      mapval,
		SliceVal:    []int{1, 2, 3},
		String:      "a string",
		Struct: MyStruct{
			Foo: "Bar",
			OK:  true,
		},
		Time: now,
	}

	m := map[string]string{
		"my.annotation.prefix/boolval":        "true",
		"my.annotation.prefix/emptystructval": "{}",
		"my.annotation.prefix/floatval":       "1.2",
		"my.annotation.prefix/intval":         "42",
		"my.annotation.prefix/mapval":         string(mapvalbs),
		"my.annotation.prefix/mystructval":    `{"fooval":"Bar","okval":true}`,
		"my.annotation.prefix/sliceval":       `[1,2,3]`,
		"my.annotation.prefix/stringval":      "a string",
		"my.annotation.prefix/timestamp":      nows,
	}

	m2, err := d.ToMap(a)
	require.NoError(t, err)
	assert.Equal(t, m, m2)

	a2, err := d.FromMap(m)
	require.NoError(t, err)
	assert.Equal(t, *a2, a)
}
