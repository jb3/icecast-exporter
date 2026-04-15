package main

import (
	"testing"
)

func TestParseIcecastConfig(t *testing.T) {
	configPath := "config.xml"
	config, err := ParseIcecastConfig(configPath)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config == nil {
		t.Fatal("Expected config not to be nil")
	}
	expectedUser := "admin"
	expectedPass := "hackme"

	if config.AdminUser != expectedUser {
		t.Errorf("Expected AdminUser %s, got %s", expectedUser, config.AdminUser)
	}

	if config.AdminPassword != expectedPass {
		t.Errorf("Expected AdminPassword %s, got %s", expectedPass, config.AdminPassword)
	}
}
