package web_test

import (
	json2 "encoding/json"
	"github.com/stretchr/testify/assert"
	"io-engine-backend/src/client"
	"testing"
)

func TestParseCircleRendererComponent(t *testing.T) {

	json := `{"Type":"CircleRendererComponent", "Size":[2,2], "Radius":4.3, "Color":"#001121" }`

	var comp client.CircleRendererComponent

	err := json2.Unmarshal([]byte(json), &comp)

	assert.NoError(t, err)

	assert.True(t, comp.Color.R < 0.00001)
	assert.True(t, comp.Color.G > 0.0001)

}
