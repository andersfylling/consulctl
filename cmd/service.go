// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)



type ConsulCheck struct {
	HTTP string `json:"HTTP,omitempty"`
	Interval string `json:"interval,omitempty"`
	Method string `json:"method,omitempty"`
	Timeout string `json:"timeout,omitempty"`
	Args []string `json:"args,omitempty"`
}
type ConsulWeights struct {
	Passing int `json:"passing"`
	Warning int `json:"warning"`
}

// ServiceDef is a Consul Service Definition scheme
type ServiceDef struct {
	// Name - Required - Specifies the logical name of the service. Many service instances may share the same logical service name.
	Name string `json:"name"`

	// ID Specifies a unique ID for this service. This must be unique per agent. This defaults to the Name parameter if not provided.
	ID string `json:"ID,omitempty"`

	// Tags Specifies a list of tags to assign to the service. These tags can be used for later filtering and are exposed via the APIs.
	Tags []string `json:"tags,omitempty"`

	// Address Specifies the address of the service. If not provided, the agent's address is used as the address for the service during DNS queries.
	Address string `json:"address,omitempty"`

	// Meta Specifies arbitrary KV metadata linked to the service instance.
	Meta map[string]string `json:"meta,omitempty"`

	// Port Specifies the port of the service.
	Port uint16 `json:"port,omitempty"`

	// Kind The kind of service. Defaults to "" which is a typical consul service. This value may also be "connect-proxy" for services that are Connect-capable proxies representing another service.
	Kind string `json:"kind,omitempty"`

	// Proxy (Proxy: nil) - From 1.2.3 on, specifies the configuration for a Connect proxy instance. This is only valid if Kind == "connect-proxy". See the Proxy documentation for full details.
	Proxy interface{} `json:"proxy,omitempty"`

	// Connect (Connect: nil) - Specifies the configuration for Connect. See the Connect Structure section below for supported fields.
	Connect interface{} `json:"connect,omitempty"`

	// Check (Check: nil) - Specifies a check. Please see the check documentation for more information about the accepted fields. If you don't provide a name or id for the check then they will be generated. To provide a custom id and/or name set the CheckID and/or Name field.
	Check *ConsulCheck `json:"check,omitempty"`

	// Checks (array<Check>: nil) - Specifies a list of checks. Please see the check documentation for more information about the accepted fields. If you don't provide a name or id for the check then they will be generated. To provide a custom id and/or name set the CheckID and/or Name field. The automatically generated Name and CheckID depend on the position of the check within the array, so even though the behavior is deterministic, it is recommended for all checks to either let consul set the CheckID by leaving the field empty/omitting it or to provide a unique value.
	Checks []*ConsulCheck `json:"checks,omitempty"`

	// EnableTagOverride (bool: false) - Specifies to disable the anti-entropy feature for this service's tags. If EnableTagOverride is set to true then external agents can update this service in the catalog and modify the tags. Subsequent local sync operations by this agent will ignore the updated tags. For instance, if an external agent modified both the tags and the port for this service and EnableTagOverride was set to true then after the next sync cycle the service's port would revert to the original value but the tags would maintain the updated value. As a counter example, if an external agent modified both the tags and port for this service and EnableTagOverride was set to false then after the next sync cycle the service's port and the tags would revert to the original value and all modifications would be lost.
	EnableTagOverride bool `json:"enable_tag_override,omitempty"`

	// Weights (Weights: nil) - Specifies weights for the service. Please see the service documentation for more information about weights. If this field is not provided weights will default to {"Passing": 1, "Warning": 1}.
	Weights *ConsulWeights `json:"weights,omitempty"`

	// file path for reading and writing
	path string
}

func (s *ServiceDef) Load(path string) error {
	s.path = path
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, s)
	if err != nil {
		return err
	}

	if s.ID == "" {
		// assumes the env var is unique from the service name
		s.ID = os.Getenv("CONSUL_SERVICE_ID")
		if s.ID == "" {
			s.ID = s.Name
		}
	}

	return nil
}

func (s *ServiceDef) SaveChanges() error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(s.path, b, 0644)
}

const ENVSrvDefPath = "CONSULCTL_SERVICE_DEFINITION_PATH"

var name string
var id string
var newTag string
var register bool
var deregister bool
var path string
var srvDef = &ServiceDef{}


// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "For managing your own service definition",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// load service definition file
		if path == "" {
			path = os.Getenv(ENVSrvDefPath)
			if path == "" {
				path = "./service.json"
			}
		}

		err := srvDef.Load(path)
		if err != nil {
			fmt.Println(fmt.Errorf(err.Error() + ". Creating service.json file"))
		}
	},
	Run: serviceAction,
}


func init() {
	rootCmd.AddCommand(serviceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serviceCmd.Flags().StringVarP(&name, "name", "", "",
		"Set the service name")

	serviceCmd.Flags().StringVarP(&id, "id", "", "",
		"Set the service id (unique on a per node basis). " +
		"Defaults to HOSTNAME environment variable if needed")

	serviceCmd.Flags().StringVarP(&newTag, "add-tag", "", "",
		"Add a new tag to your consul service")

	serviceCmd.Flags().BoolVarP(&register, "register", "", false,
		"Register the service with Consul")

	serviceCmd.Flags().BoolVarP(&deregister, "deregister", "", false,
		"Deregister the service with Consul")

	serviceCmd.Flags().StringVarP(&path, "definition-path", "", "",
		"Specify the path where you keep the service definition json file. " +
		"Default is the same dir as you call this command")
}

func serviceAction(cmd *cobra.Command, args []string) {
	// validate
	fmt.Print(baseURL)


	if name != "" {
		srvDef.Name = name
	}

	if id != "" {
		srvDef.ID = id
	}

	if newTag != "" {
		exists := false
		for i := range srvDef.Tags {
			exists = srvDef.Tags[i] == newTag
			if exists {
				fmt.Println("duplicate tag not added")
				break
			}
		}

		if !exists {
			srvDef.Tags = append(srvDef.Tags, newTag)
		}
	}


	// register
	if register {
		if err := registerService(baseURL, srvDef); err != nil {
			panic(err)
		}
	}

	// deregister
	if deregister {
		if err := deregisterService(baseURL, srvDef); err != nil {
			panic(err)
		}
	}

	// done
	srvDef.SaveChanges()
}

func registerService(url string, srv *ServiceDef) error {
	if srv.Name == "" {
		return errors.New("service name is empty")
	}

	data, err := json.Marshal(srv)
	if err != nil {
		return err
	}

	const endpoint = "/agent/service/register"
	resp, err := sendConsulRequest(http.MethodPut, url + endpoint, data)

	if resp.StatusCode != 200 {
		return errors.New("did not get successful http response: " + resp.Status)
	}

	return nil
}

func deregisterService(url string, srv *ServiceDef) error {
	if srv.Name == "" {
		return errors.New("service name is empty")
	}

	endpoint := "/agent/service/deregister/" + srv.ID
	_, err := sendConsulRequest(http.MethodPut, url + endpoint, nil)
	return err
}

func sendConsulRequest(method, endpoint string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}