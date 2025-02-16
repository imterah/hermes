package datacommands

type ProxyStatusRequest struct {
	Type    string
	ProxyID uint16
}

type ProxyStatusResponse struct {
	Type     string
	ProxyID  uint16
	IsActive bool
}

type RemoveProxy struct {
	Type    string
	ProxyID uint16
}

type ProxyInstanceResponse struct {
	Type    string
	Proxies []uint16
}

type ProxyConnectionsRequest struct {
	Type    string
	ProxyID uint16
}

type ProxyConnectionsResponse struct {
	Type        string
	Connections []uint16
}

type TCPConnectionOpened struct {
	Type         string
	ProxyID      uint16
	ConnectionID uint16
}

type TCPConnectionClosed struct {
	Type         string
	ProxyID      uint16
	ConnectionID uint16
}

type TCPProxyData struct {
	Type         string
	ProxyID      uint16
	ConnectionID uint16
	DataLength   uint16
}

type UDPProxyData struct {
	Type       string
	ProxyID    uint16
	ClientIP   string
	ClientPort uint16
	DataLength uint16
}

type ProxyInformationRequest struct {
	Type    string
	ProxyID uint16
}

type ProxyInformationResponse struct {
	Type       string
	Exists     bool
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type ProxyConnectionInformationRequest struct {
	Type         string
	ProxyID      uint16
	ConnectionID uint16
}

type ProxyConnectionInformationResponse struct {
	Type       string
	Exists     bool
	ClientIP   string
	ClientPort uint16
}

const (
	ProxyStatusRequestID = iota + 100
	ProxyStatusResponseID
	RemoveProxyID
	ProxyInstanceResponseID
	ProxyConnectionsRequestID
	ProxyConnectionsResponseID
	TCPConnectionOpenedID
	TCPConnectionClosedID
	TCPProxyDataID
	UDPProxyDataID
	ProxyInformationRequestID
	ProxyInformationResponseID
	ProxyConnectionInformationRequestID
	ProxyConnectionInformationResponseID
)
