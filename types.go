package main

import (
	v3 "code.cloudfoundry.org/cli/actor/v3action"
)

//YAML config
type Manifest struct {
	Policies []NetworkPolicy `yaml:"network-policies"`
}

type NetworkPolicy struct {
	SrcApp    string `yaml:"src"`
	SrcSpace  string `yaml:"src-space,omitempty"`
	DestApp   string `yaml:"dest"`
	DestSpace string `yaml:"dest-space,omitempty"`
	Ports     string `yaml:"ports"`
	Protocol  string `yaml:"protocol,omitempty"`
}

//Network policy POST request
type NetworkPolicyData struct {
	Policies []Policy `json:"policies"`
}

type Policy struct {
	Destination Destination `json:"destination"`
	Source      Source      `json:"source"`
}

type Destination struct {
	Id       string `json:"id"`
	Ports    Ports  `json:"ports"`
	Protocol string `json:"protocol"`
}

type Ports struct {
	From int `json:"start"`
	To   int `json:"end"`
}

type Source struct {
	Id string `json:"id"`
}

//cf curl /v3/apps endpoint
type AppResponse struct {
	Resources []v3.Application `json:"resources"`
}
