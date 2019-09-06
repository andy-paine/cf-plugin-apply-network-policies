package main

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"

	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
)

var _ = Describe("NetworkPolicy", func() {
	cliConn := &pluginfakes.FakeCliConnection{}

	Context("getting spaces", func() {
		It("should generate a map of all spaces", func() {
			space1 := plugin_models.GetSpaces_Model{
				Name: "space-1",
				Guid: "space-1-guid",
			}
			cliConn.GetSpacesReturns([]plugin_models.GetSpaces_Model{space1}, nil)
			spaces, err := getSpaces(cliConn)
			Expect(err).ToNot(HaveOccurred())
			Expect(spaces).To(HaveKeyWithValue("space-1", space1))
		})

		It("should return CLI errors", func() {
			cliErr := fmt.Errorf("Fake CLI error")
			cliConn.GetSpacesReturns(nil, cliErr)
			_, err := getSpaces(cliConn)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Fake CLI error"))
		})
	})

	Context("getting app GUIDs", func() {
		space := plugin_models.GetSpaces_Model{
			Name: "space",
			Guid: "space-guid",
		}
		It("should return the app GUID", func() {
			v3App := `{"resources": [
				{"guid": "app-1-guid", "name": "app-1"}
			]}`
			cliConn.CliCommandWithoutTerminalOutputReturns(strings.Split(v3App, "\n"), nil)
			appName, err := getAppGuid(cliConn, space, "app-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(appName).To(Equal("app-1-guid"))
		})

		It("should return CLI errors", func() {
			cliErr := fmt.Errorf("Fake CLI error")
			cliConn.CliCommandWithoutTerminalOutputReturns(nil, cliErr)
			_, err := getAppGuid(cliConn, space, "app-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Fake CLI error"))
		})
	})
})
