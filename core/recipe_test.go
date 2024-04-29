package albius

import "testing"

func TestJsonFieldToInt(t *testing.T) {
	vstr := "2"
	vf64 := float64(2)

	val, err := jsonFieldToInt(vstr)
	if err != nil {
		t.Error(err)
	}
	if val != 2 {
		t.Error(err)
	}

	val, err = jsonFieldToInt(vf64)
	if err != nil {
		t.Error(err)
	}
	if val != 2 {
		t.Error(err)
	}
}
