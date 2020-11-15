package musiccastClient

type LinkRequest struct {
	GroupId string   `json:"group_id"`
	Zones   []string `json:"zone"`
}

type MasterLinkRequest struct {
	GroupId    string   `json:"group_id"`
	Zone       string   `json:"zone"`
	Type       string   `json:"type"`
	ClientList []string `json:"client_list"`
}
