/*
 * Copyright 2018. Akamai Technologies, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"os"
	"fmt"
	"log"
        "github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
        "github.com/akamai/AkamaiOPEN-edgegrid-golang/client-v1"

) 

type ListOfFiredAlerts struct {
	Data []*ListOfFiredAlertsItems
}

type FieldMap struct {
	Cpcode string
	AlertType string
}

type ListOfFiredAlertsItems struct {
  FieldMap *FieldMap `json:"fieldMap"`
  Name string `json:"name"`
  Service string `json:"service"`
  StartTime string `json:"startTime"`
}

type Activation struct {
	Network string `json:"network"`
}

const LOW_ALERT = "Low Traffic -- Content Delivery"
const HIGH_ALERT = "High Traffic -- Content Delivery"

func main() {

	// Command line params
	cpcode := flag.Int("cpcode", 0, "CP Code for alert")
	policyid := flag.Int("policyid", 0, "Visitor Prioritisation policy id")
	enableversion := flag.Int("enableversion", 0, "Visitor Prioritization policy version for enabling")
	disableversion := flag.Int("disableversion", 0, "Visitor Prioritization policy version for disabling")
	section := flag.String("section", "default", "edgerc section")
	flag.Parse()

	if (*cpcode == 0 || *policyid == 0 || *enableversion == 0 || *disableversion == 0) {
		fmt.Printf("cpcode = %d, policyid = %d, enableversion = %d, disableversion = %d\n", *cpcode, *policyid, *enableversion, *disableversion)
		flag.PrintDefaults()
		os.Exit(1)
	}
	

	// Initialize Edgegrid
        config, err := edgegrid.Init("~/.edgerc", *section)
        if err != nil {
		log.Fatal(err)
        }

	// Find active alerts
	req, _ := client.NewRequest(config, "GET", "/alerts/v2/alert-firings/active", nil)
	resp, err := client.Do(config, req)
        if err != nil {
		log.Fatal(err)
        }

	defer resp.Body.Close()
	byt, _ := ioutil.ReadAll(resp.Body)

	var alerts ListOfFiredAlerts
	err = json.Unmarshal(byt, &alerts)

	// Iterate the fired alerts and find the ones that relate to our CPCode
	var versionToActivate = 0
	for _, element := range alerts.Data {
		if (element.FieldMap.Cpcode != fmt.Sprintf("%d",*cpcode)) {
			continue
		}

		res, _ := json.MarshalIndent(element, "", "  ")
		fmt.Println(string(res))


		if (element.FieldMap.AlertType == LOW_ALERT) {
			versionToActivate = *disableversion
		}

		if (element.FieldMap.AlertType == HIGH_ALERT) {
			versionToActivate = *enableversion
		}
	}

	// Activate VP if necessary
	if (versionToActivate != 0) {
		var activation Activation
		activation.Network = "STAGING"
		a,_ := json.Marshal(activation)
		p := fmt.Sprintf("/cloudlets/api/v2/policies/%d/versions/%d/activations",  *policyid, versionToActivate)
		fmt.Println(p)
		fmt.Println(string(a))
		req, _ := client.NewRequest(config, "POST", p, bytes.NewBuffer(a))
		resp, err := client.Do(config, req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		byt, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(byt))
	}
}
