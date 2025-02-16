package main

import (
	"os"
	"sync"

	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/datacommands"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/remote-code/backendutil_custom"
	"github.com/charmbracelet/log"
)

type TCPProxy struct {
	proxyIDIndex uint16
	proxyIDLock  sync.Mutex
}

type UDPProxy struct {
}

type SSHRemoteAppBackend struct {
	connectionIDIndex uint16
	connectionIDLock  sync.Mutex

	tcpProxies map[uint16]*TCPProxy
	udpProxies map[uint16]*UDPProxy
}

func (backend *SSHRemoteAppBackend) StartBackend(byte []byte) (bool, error) {
	backend.tcpProxies = map[uint16]*TCPProxy{}
	backend.udpProxies = map[uint16]*UDPProxy{}

	return true, nil
}

func (backend *SSHRemoteAppBackend) StopBackend() (bool, error) {
	return true, nil
}

func (backend *SSHRemoteAppBackend) GetBackendStatus() (bool, error) {
	return true, nil
}

func (backend *SSHRemoteAppBackend) StartProxy(command *commonbackend.AddProxy) (uint16, bool, error) {
	return 0, true, nil
}

func (backend *SSHRemoteAppBackend) StopProxy(command *datacommands.RemoveProxy) (bool, error) {
	return true, nil
}

func (backend *SSHRemoteAppBackend) GetAllProxies() []uint16 {
	return []uint16{}
}

func (backend *SSHRemoteAppBackend) ResolveProxy(proxyID uint16) *datacommands.ProxyInformationResponse {
	return &datacommands.ProxyInformationResponse{}
}

func (backend *SSHRemoteAppBackend) GetAllClientConnections(proxyID uint16) []uint16 {
	return []uint16{}
}

func (backend *SSHRemoteAppBackend) ResolveConnection(proxyID uint16) *datacommands.ProxyConnectionsResponse {
	return &datacommands.ProxyConnectionsResponse{}
}

func (backend *SSHRemoteAppBackend) CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse {
	return &commonbackend.CheckParametersResponse{
		IsValid: true,
		Message: "Valid!",
	}
}

func (backend *SSHRemoteAppBackend) CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse {
	return &commonbackend.CheckParametersResponse{
		IsValid: true,
		Message: "Valid!",
	}
}

func (backend *SSHRemoteAppBackend) HandleTCPMessage(message *datacommands.TCPProxyData, data []byte) {

}

func (backend *SSHRemoteAppBackend) HandleUDPMessage(message *datacommands.UDPProxyData, data []byte) {

}

func main() {
	logLevel := os.Getenv("HERMES_LOG_LEVEL")

	if logLevel != "" {
		switch logLevel {
		case "debug":
			log.SetLevel(log.DebugLevel)

		case "info":
			log.SetLevel(log.InfoLevel)

		case "warn":
			log.SetLevel(log.WarnLevel)

		case "error":
			log.SetLevel(log.ErrorLevel)

		case "fatal":
			log.SetLevel(log.FatalLevel)
		}
	}

	backend := &SSHRemoteAppBackend{}

	application := backendutil_custom.NewHelper(backend)
	err := application.Start()

	if err != nil {
		log.Fatalf("failed execution in application: %s", err.Error())
	}
}
