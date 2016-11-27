package pack

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestMarshalError(t *testing.T) {
	d := struct {
		Compact bool
		Schema  uint8
	}{Compact: true, Schema: 0}

	buf, err := Marshal(d)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	other, _ := hex.DecodeString("82a7636f6d70616374c3a6736368656d6100")
	if !bytes.Equal(buf, other) {
		t.Errorf("%x != %x", buf, other)
	}
}
