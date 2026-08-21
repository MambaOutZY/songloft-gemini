package jsplugin

import "testing"

func TestEncodeNetworkAddresses_NilIsEmptyArray(t *testing.T) {
	if got := encodeNetworkAddresses(nil); got != "[]" {
		t.Fatalf("encodeNetworkAddresses(nil) = %q, want []", got)
	}
}

func TestEncodeNetworkAddresses_PreservesAddresses(t *testing.T) {
	got := encodeNetworkAddresses([]string{"http://192.168.1.10:58091"})
	if got != `["http://192.168.1.10:58091"]` {
		t.Fatalf("encodeNetworkAddresses() = %q, want address array", got)
	}
}
