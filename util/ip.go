package util

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
)

var ()

func ResolveIP(masterServer string) net.IP {
	url := masterServer + "/ip"

	// Master server request
	res, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil
	}

	// JSON decoding
	type IPData struct {
		Success bool
		Ip      string
	}
	var data IPData
	err = json.Unmarshal(body, &data)
	if err != nil || !data.Success {
		return nil
	}

	return net.ParseIP(data.Ip)
}
