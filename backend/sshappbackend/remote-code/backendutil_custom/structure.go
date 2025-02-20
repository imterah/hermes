package backendutil_custom

import (
	"net"

	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/datacommands"
)

type BackendInterface interface {
	StartBackend(arguments []byte) (bool, error)
	StopBackend() (bool, error)
	GetBackendStatus() (bool, error)
	StartProxy(command *commonbackend.AddProxy) (uint16, bool, error)
	StopProxy(command *datacommands.RemoveProxy) (bool, error)
	GetAllProxies() []uint16
	ResolveProxy(proxyID uint16) *datacommands.ProxyInformationResponse
	GetAllClientConnections(proxyID uint16) []uint16
	ResolveConnection(proxyID, connectionID uint16) *datacommands.ProxyConnectionInformationResponse
	CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse
	CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse
	OnTCPConnectionClosed(proxyID, connectionID uint16)
	HandleTCPMessage(message *datacommands.TCPProxyData, data []byte)
	HandleUDPMessage(message *datacommands.UDPProxyData, data []byte)
	OnSocketConnection(sock net.Conn)
}
