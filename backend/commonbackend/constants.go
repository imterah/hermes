package commonbackend

type Start struct {
	Arguments []byte
}

type Stop struct {
}

type AddProxy struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type RemoveProxy struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type ProxyStatusRequest struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type ProxyStatusResponse struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
	IsActive   bool
}

type ProxyInstance struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type ProxyInstanceResponse struct {
	Proxies []*ProxyInstance // List of connections
}

type ProxyInstanceRequest struct {
}

type BackendStatusResponse struct {
	IsRunning  bool   // True if running, false if not running
	StatusCode int    // Either the 'Success' or 'Failure' constant
	Message    string // String message from the client (ex. failed to dial TCP)
}

type BackendStatusRequest struct {
}

type ProxyConnectionsRequest struct {
}

// Client's connection to a specific proxy
type ProxyClientConnection struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	ClientIP   string
	ClientPort uint16
}

type ProxyConnectionsResponse struct {
	Connections []*ProxyClientConnection // List of connections
}

type CheckClientParameters struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type CheckServerParameters struct {
	Arguments []byte
}

// Sent as a response to either CheckClientParameters or CheckBackendParameters
type CheckParametersResponse struct {
	InResponseTo string // Will be either 'checkClientParameters' or 'checkServerParameters'
	IsValid      bool   // If true, valid, and if false, invalid
	Message      string // String message from the client (ex. failed to unmarshal JSON: x is not defined)
}

const (
	StartID = iota
	StopID
	AddProxyID
	RemoveProxyID
	ProxyConnectionsResponseID
	CheckClientParametersID
	CheckServerParametersID
	CheckParametersResponseID
	ProxyConnectionsRequestID
	BackendStatusResponseID
	BackendStatusRequestID
	ProxyStatusRequestID
	ProxyStatusResponseID
	ProxyInstanceResponseID
	ProxyInstanceRequestID
)

const (
	TCP = iota
	UDP
)

const (
	StatusSuccess = iota
	StatusFailure
)

const (
	// IP versions
	IPv4 = 4
	IPv6 = 6

	// TODO: net has these constants defined already. We should switch to these
	IPv4Size = 4
	IPv6Size = 16
)
