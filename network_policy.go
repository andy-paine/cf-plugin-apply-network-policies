package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
	"gopkg.in/yaml.v2"
)

var Version = ""

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			if Version == "" {
				fmt.Printf("cf-plugin-apply-network-policies (development)\n")
			} else {
				fmt.Printf("cf-plugin-apply-network-policies v%s\n", Version)
			}
			os.Exit(0)
		}
	}
	plugin.Start(&ApplyNetworkPoliciesPlugin{})
}

// ApplyNetworkPoliciesPlugin empty struct for plugin
type ApplyNetworkPoliciesPlugin struct{}

// Run of seeder plugin
func (plugin ApplyNetworkPoliciesPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if len(args) < 2 {
		cliConnection.CliCommand(args[0], "-h")
		log.Fatal("YAML file path is needed")
	}

	if err := plugin.applyNetworkPolicies(cliConnection, args[1]); err != nil {
		log.Fatal("Error applying network policies: ", err)
		os.Exit(1)
	}
}

// GetMetadata of plugin
func (ApplyNetworkPoliciesPlugin) GetMetadata() plugin.PluginMetadata {
	version := plugin.VersionType{Major: 0, Minor: 0, Build: 0}
	versionRE := regexp.MustCompile("([0-9]+).([0-9]+).([0-9]+)")
	versionParts := versionRE.FindStringSubmatch(Version)
	if len(versionParts) == 4 {
		var part int64
		part, _ = strconv.ParseInt(versionParts[1], 10, 32)
		version.Major = int(part)
		part, _ = strconv.ParseInt(versionParts[2], 10, 32)
		version.Minor = int(part)
		part, _ = strconv.ParseInt(versionParts[3], 10, 32)
		version.Build = int(part)
	}
	return plugin.PluginMetadata{
		Name:    "apply-network-policies",
		Version: version,
		Commands: []plugin.Command{
			{
				Name:     "apply-network-policies",
				HelpText: "Manage network policies declaritively",
				UsageDetails: plugin.Usage{
					Usage: "apply-network-policies network-policies.yml",
				},
			},
		},
	}
}

func (plugin ApplyNetworkPoliciesPlugin) applyNetworkPolicies(cliConnection plugin.CliConnection, ymlFile string) error {
	data, err := ioutil.ReadFile(ymlFile)
	if err != nil {
		return err
	}

	var manifest Manifest
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return err
	}

	currentSpace, err := cliConnection.GetCurrentSpace()
	if err != nil {
		return err
	}

	spaces, err := getSpaces(cliConnection)
	if err != nil {
		return err
	}

	for _, policy := range manifest.Policies {
		if policy.SrcApp == "" || policy.DestApp == "" || policy.Ports == "" {
			return fmt.Errorf("src, dest and ports are all required")
		}

		srcSpace := getOrDefault(policy.SrcSpace, currentSpace.Name)
		destSpace := getOrDefault(policy.DestSpace, currentSpace.Name)
		fmt.Fprintf(os.Stdout, "Adding network policy from %s in space %s to %s in space %s... ", policy.SrcApp, srcSpace, policy.DestApp, destSpace)

		srcAppGuid, err := getAppGuid(cliConnection, spaces[srcSpace], policy.SrcApp)
		if err != nil {
			return err
		}

		destAppGuid, err := getAppGuid(cliConnection, spaces[destSpace], policy.DestApp)
		if err != nil {
			return err
		}

		jsonData, err := getNetworkPolicyData(srcAppGuid, destAppGuid, policy)

		output, err := cliConnection.CliCommandWithoutTerminalOutput("curl",
			"-X",
			"POST",
			"/networking/v1/external/policies",
			"-d",
			string(jsonData),
		)
		if err != nil {
			return err
		}

		fmt.Println("✔ DONE")
		fmt.Println(output)
	}
	return nil
}

func getSpaces(cliConnection plugin.CliConnection) (map[string]plugin_models.GetSpaces_Model, error) {
	spaceList, err := cliConnection.GetSpaces()
	if err != nil {
		return nil, err
	}

	spaces := map[string]plugin_models.GetSpaces_Model{}
	for _, space := range spaceList {
		spaces[space.Name] = space
	}
	return spaces, nil
}

func getAppGuid(cliConnection plugin.CliConnection, space plugin_models.GetSpaces_Model, appName string) (string, error) {
	output, err := cliConnection.CliCommandWithoutTerminalOutput(
		"curl",
		fmt.Sprintf("/v3/apps?names=%s&space_guids=%s", appName, space.Guid),
	)
	if err != nil {
		return "", err
	}

	var result AppResponse
	err = json.Unmarshal([]byte(strings.Join(output, " ")), &result)
	if err != nil {
		return "", err
	}

	if len(result.Resources) != 1 {
		return "", fmt.Errorf("Could not find unique application called '%s' in space '%s'", appName, space.Name)
	}

	return result.Resources[0].GUID, nil
}

func getNetworkPolicyData(srcAppGuid, destAppGuid string, policy NetworkPolicy) ([]byte, error) {
	fromPort, err := strconv.Atoi(strings.Split(policy.Ports, "-")[0])
	toPort, err := strconv.Atoi(strings.Split(policy.Ports, "-")[1])
	if err != nil {
		return nil, err
	}

	data := NetworkPolicyData{
		[]Policy{
			Policy{
				Destination: Destination{
					Id: destAppGuid,
					Ports: Ports{
						From: fromPort,
						To:   toPort,
					},
					Protocol: getOrDefault(policy.Protocol, "tcp"),
				},
				Source: Source{
					Id: srcAppGuid,
				},
			},
		},
	}

	return json.Marshal(data)
}

func getOrDefault(value string, def string) string {
	if value == "" {
		return def
	}
	return value
}
