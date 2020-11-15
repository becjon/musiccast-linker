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
	c.log.Infof("standby %s:%s", host, zone)
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
			versions = append(versions, features.Distribution.CompatibleClient...)
		} else {
			if !contains(versions, features.Distribution.Version) {
				return fmt.Errorf("versions are not compatible '%.2f is not in %.2f'", features.Distribution.Version, versions)
			}
		}
	}
	c.log.Info("devices are compatible")
	return nil
}

func (c *MusiccastClient) getFeatures(host string) (*featuresResponse, error) {
	response, err := c.client.Get(fmt.Sprintf("http://%s/YamahaExtendedControl/v1/system/getFeatures", host))
	if err != nil {
		return nil, err
	}
	return MapFeaturesResponse(c.log, response)
}

func (c *MusiccastClient) Link(master string, clients []string) error {
	err := c.areDevicesCompatible(append([]string{master}, clients...)...)
	if err != nil {
		return err
	}
	groudId := strings.ReplaceAll(uuid.New().String(), "-", "")

	err = c.prepareClients(groudId, clients)
	if err != nil {
		return err
	}

	err = c.prepareMaster(groudId, master, clients)
	if err != nil {
		return err
	}

	err = c.startDistribution(master)
	if err != nil {
		return err
	}
	c.log.Infof("linking successful. master:'%s' is now sending to client:'%s'", master, clients)
	return nil
}

func (c *MusiccastClient) prepareClients(groupId string, clients []string) error {
	for _, link := range clients {
		err := c.prepareClient(groupId, link)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *MusiccastClient) prepareClient(groupId string, host string) error {
	clientLinkRequest, err := json.Marshal(ClientLinkRequest{
		GroupId: groupId,
		Zones:   []string{"main"},
	})
	if err != nil {
		return err
	}
	response, err := c.client.Post(fmt.Sprintf("http://%s/YamahaExtendedControl/v1/dist/setClientInfo", host), "application/json", bytes.NewBuffer(clientLinkRequest))
	return MapStatusResponse(c.log, response)
}

func (c *MusiccastClient) prepareMaster(groupId string, master string, clients []string) error {
	var clientIps []string
	for _, client := range clients {
		ip, err := net.LookupIP(client)
		if err != nil {
			return err
		}
		if len(ip) > 0 {
			clientIps = append(clientIps, ip[0].String())
		}
	}

	masterRequest, err := json.Marshal(MasterLinkRequest{
		GroupId:    groupId,
		Zone:       "zone2",
		Type:       "add",
		ClientList: clientIps,
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

func (c *MusiccastClient) ChangeInput(hostname string, zone string, input string) error {
	err := c.setIput(hostname, zone, input)
	if err != nil {
		return err
	}
	c.log.Infof("input for '%s:%s' is set to '%s'", hostname, zone, input)
	return nil
}

func (c *MusiccastClient) setIput(hostname string, zone string, input string) error {
	response, err := c.client.Get(fmt.Sprintf("http://%s/YamahaExtendedControl/v1/%s/setInput?input=%s", hostname, zone, input))
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
