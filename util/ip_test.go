package util

import (
	"testing"
)

func TestResolveIP(t *testing.T) {
	ip := ResolveIP("https://akadok.deimos-ga.me")
	if ip == nil {
		t.Log("IP resolving error - COULD BE RELATED TO NETWORK OR MASTER SERVER")
		t.Fail()
	}
}
