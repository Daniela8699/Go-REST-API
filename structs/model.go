package structs

// Server struct (Model)
type Server struct {
	Address  string `json:"address"`
	SSLGrade string `json:"ssl_grade"`
	Country  string `json:"country"`
	Owner    string `json:"owner"`
}

// DomainInfo struct
type DomainInfo struct {
	Servers          []Server `json:"servers"`
	ServersChanged   bool         `json:"servers_changed"`
	SSLGrade         string       `json:"ssl_grade"`
	PreviousSSLGrade string       `json:"previous_ssl_grade"`
	Logo             string       `json:"logo"`
	Title            string       `json:"title"`
	IsDown           bool         `json:"is_down"`
	LastUpdated		 string       `json:"last_updated"`
}
// ServersHistory struct
type ServersHistory struct {
	Items []ServersHistoryElement `json:"items"`
}
// ServersHistoryElement struct
type ServersHistoryElement struct {
	Host string        `json:"host"`
	
}
