package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

type IcecastConfig struct {
	XMLName       xml.Name `xml:"icecast"`
	AdminUser     string   `xml:"authentication>admin-user"`
	AdminPassword string   `xml:"authentication>admin-password"`
}

func ParseIcecastConfig(path string) (*IcecastConfig, error) {
	configFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error read file: %w", err)
	}

	var config = &IcecastConfig{}
	if err := xml.Unmarshal(configFile, config); err != nil {
		return nil, fmt.Errorf("error unmarshal file: %w", err)
	}
	return config, nil
}
