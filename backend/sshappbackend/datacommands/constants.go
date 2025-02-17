package datacommands

type ProxyStatusRequest struct {
	ProxyID uint16
}

type ProxyStatusResponse struct {
	ProxyID  uint16
	IsActive bool
}

type RemoveProxy struct {
	ProxyID uint16
}

type ProxyInstanceResponse struct {
	Proxies []uint16
}

type ProxyConnectionsRequest struct {
	ProxyID uint16
}

type ProxyConnectionsResponse struct {
	Connections []uint16
}

type TCPConnectionOpened struct {
	ProxyID      uint16
	ConnectionID uint16
}

type TCPConnectionClosed struct {
	ProxyID      uint16
	ConnectionID uint16
}

type TCPProxyData struct {
	ProxyID      uint16
	ConnectionID uint16
	DataLength   uint16
}

type UDPProxyData struct {
	ProxyID    uint16
	ClientIP   string
	ClientPort uint16
	DataLength uint16
}

type ProxyInformationRequest struct {
	ProxyID uint16
}

type ProxyInformationResponse struct {
	Exists     bool
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type ProxyConnectionInformationRequest struct {
	ProxyID      uint16
	ConnectionID uint16
}

type ProxyConnectionInformationResponse struct {
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
