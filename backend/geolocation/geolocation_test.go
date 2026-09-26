package geolocation

import (
	"testing"
)

func TestLookupLocalIP(t *testing.T) {
	// Localhost IPv4
	geo, err := Lookup("127.0.0.1")
	if err != nil {
		t.Fatalf("Expected nil error for local IP, got %v", err)
	}
	if geo == nil || geo.CountryName != "Localhost" {
		t.Errorf("Expected Localhost, got %+v", geo)
	}

	// Localhost IPv6
	geoV6, err := Lookup("::1")
	if err != nil {
		t.Fatalf("Expected nil error for local IPv6, got %v", err)
	}
	if geoV6 == nil || geoV6.CountryCode != "00" {
		t.Errorf("Expected CountryCode 00, got %+v", geoV6)
	}

	// Cached retrieval
	cached, err := Lookup("127.0.0.1")
	if err != nil || cached != geo {
		t.Errorf("Expected cached pointer match, got %p vs %p", cached, geo)
	}
}

func TestLookupUninitialized(t *testing.T) {
	// A public IP when service is uninitialized should return ErrServiceNotInitialized
	_, err := Lookup("8.8.8.8")
	if err != ErrServiceNotInitialized {
		t.Errorf("Expected ErrServiceNotInitialized, got %v", err)
	}
}
