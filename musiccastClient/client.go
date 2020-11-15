package musiccastClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"math"
	"net"
	"net/http"
	"strings"
)

type MusiccastClient struct {
	client *http.Client
	log    *logrus.Logger
}

func New(log *logrus.Logger) *MusiccastClient {
	return &MusiccastClient{
		client: http.DefaultClient,
		log:    log,
	}

}

func (c *MusiccastClient) PowerOn(host, zone string) error {
	return c.power(host, zone, "on")
}
func (c *MusiccastClient) PowerOff(host, zone string) error {
	return c.power(host, zone, "standby")
}

func (c *MusiccastClient) power(host, zone, powerStatus string) error {
	response, err := c.client.Get(fmt.Sprintf("http://%s/YamahaExtendedControl/v1/%s/setPower?power=%s", host, zone, powerStatus))
	if err != nil {
		return err
	}
	return MapStatusResponse(c.log, response)
}

func (c *MusiccastClient) areDevicesCompatible(hosts ...string) error {
	var versions []float64
	for _, host := range hosts {
		features, err := c.getFeatures(host)
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			versions = append(versions, features.CompatibleClient...)
		} else {
			if !contains(versions, features.Version) {
				return fmt.Errorf("versions are not compatible '%.2f is not in %.2f'", features.Version, versions)
			}
		}
	}
	return nil
}

func (c *MusiccastClient) getFeatures(host string) (*Distribution, error) {
	response, err := c.client.Get(fmt.Sprintf("http://%s/YamahaExtendedControl/v1/system/getFeatures", host))
	if err != nil {
		return nil, err
	}
	return MapFeaturesResponse(c.log, response)
}

func (c *MusiccastClient) Link(master string, links []string) error {
	err := c.areDevicesCompatible(append([]string{master}, links...)...)
	if err != nil {
		return err
	}
	groudId := strings.ReplaceAll(uuid.New().String(), "-", "")

	err = c.prepareLinks(groudId, links)
	if err != nil {
		return err
	}

	err = c.prepareMaster(groudId, master, links)
	if err != nil {
		return err
	}

	err = c.startDistribution(master)
	return err

}

func (c *MusiccastClient) prepareLinks(groupId string, links []string) error {
	for _, link := range links {
		err := c.prepareLink(groupId, link)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *MusiccastClient) prepareLink(groupId string, host string) error {
	linkRequest, err := json.Marshal(LinkRequest{
		GroupId: groupId,
		Zones:   []string{"main"},
	})
	if err != nil {
		return err
	}
	response, err := c.client.Post(fmt.Sprintf("http://%s/YamahaExtendedControl/v1/dist/setClientInfo", host), "application/json", bytes.NewBuffer(linkRequest))
	return MapStatusResponse(c.log, response)
}

func (c *MusiccastClient) prepareMaster(groupId string, master string, links []string) error {
	var linksIps []string
	for _, link := range links {
		ip, err := net.LookupIP(link)
		if err != nil {
			return err
		}
		if len(ip) > 0 {
			linksIps = append(linksIps, ip[0].String())
		}
	}

	masterRequest, err := json.Marshal(MasterLinkRequest{
		GroupId:    groupId,
		Zone:       "zone2",
		Type:       "add",
		ClientList: linksIps,
	})
	if err != nil {
		return err
	}
	response, err := c.client.Post(fmt.Sprintf("http://%s/YamahaExtendedControl/v1/dist/setServerInfo", master), "application/json", bytes.NewBuffer(masterRequest))
	return MapStatusResponse(c.log, response)
}

func (c *MusiccastClient) startDistribution(master string) error {
	response, err := c.client.Get(fmt.Sprintf("http://%s/YamahaExtendedControl/v1/dist/startDistribution?num=0", master))
	if err != nil {
		return err
	}
	return MapStatusResponse(c.log, response)
}

func contains(s []float64, e float64) bool {
	for _, a := range s {
		if withTolerane(a, e) {
			return true
		}
	}
	return false
}

func withTolerane(a, b float64) bool {
	tolerance := 0.10
	if diff := math.Abs(a - b); diff < tolerance {
		return true
	} else {
		return false
	}
}
