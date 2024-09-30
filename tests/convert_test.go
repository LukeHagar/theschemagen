package theschemagen_test

import (
	"os"
	"testing"

	"github.com/lukehagar/theschemagen"
	"github.com/stretchr/testify/require"
)

func TestConvertJSONToOAS(t *testing.T) {
	// assert := assert.New(t)
	require := require.New(t)

	testJson, err := os.ReadFile("./test-files/test.json")
	if err != nil {
		panic(err)
	}

	exampleJSON := string(testJson)

	schema := theschemagen.ConvertJSONToOAS(exampleJSON)
	// root level check
	require.Equal(schema.Type, "object")

	require.Equal(schema.Properties["stringsMock"].Properties["stringTest"].Type, "string")

	require.Equal(schema.Properties["stringsMock"].Properties["isoDate"].Type, "string")
	require.Equal(schema.Properties["stringsMock"].Properties["isoDate"].Format, "date")

	require.Equal(schema.Properties["stringsMock"].Properties["isoDateTime"].Type, "string")
	require.Equal(schema.Properties["stringsMock"].Properties["isoDateTime"].Format, "date-time")

	require.Equal(schema.Properties["numbersMock"].Properties["smallInt"].Type, "integer")
	require.Equal(schema.Properties["numbersMock"].Properties["smallInt"].Format, "int32")
}

func BenchmarkConvertJSONToOAS(b *testing.B) {
	testJson, err := os.ReadFile("./test-files/test.json")
	if err != nil {
		panic(err)
	}

	exampleJSON := string(testJson)

	for i := 0; i < b.N; i++ {
		theschemagen.ConvertJSONToOAS(exampleJSON)
	}
}

func TestConvertObject(t *testing.T) {
	testJson, err := os.ReadFile("./test-files/test.json")
	if err != nil {
		panic(err)
	}

	exampleJSON := string(testJson)

	schema := theschemagen.ConvertJSONToOAS(exampleJSON)
	theschemagen.PrettyPrint(schema, "yaml")
}
