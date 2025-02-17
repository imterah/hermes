package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/datacommands"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/remote-code/backendutil_custom"
	"github.com/charmbracelet/log"
)

type TCPProxy struct {
	connectionIDIndex uint16
	connectionIDLock  sync.Mutex

	proxyInformation *commonbackend.AddProxy
	connections      map[uint16]net.Conn
	server           net.Listener
}

type UDPProxy struct {
	server           *net.UDPConn
	proxyInformation *commonbackend.AddProxy
}

type SSHRemoteAppBackend struct {
	proxyIDIndex uint16
	proxyIDLock  sync.Mutex

	tcpProxies map[uint16]*TCPProxy
	udpProxies map[uint16]*UDPProxy

	sock net.Conn
}

func (backend *SSHRemoteAppBackend) StartBackend(byte []byte) (bool, error) {
	backend.tcpProxies = map[uint16]*TCPProxy{}
	backend.udpProxies = map[uint16]*UDPProxy{}

	return true, nil
}

func (backend *SSHRemoteAppBackend) StopBackend() (bool, error) {
	for tcpProxyIndex, tcpProxy := range backend.tcpProxies {
		for _, tcpConnection := range tcpProxy.connections {
			tcpConnection.Close()
		}

		tcpProxy.server.Close()
		delete(backend.tcpProxies, tcpProxyIndex)
	}

	for udpProxyIndex, udpProxy := range backend.udpProxies {
		udpProxy.server.Close()
		delete(backend.udpProxies, udpProxyIndex)
	}

	return true, nil
}

func (backend *SSHRemoteAppBackend) GetBackendStatus() (bool, error) {
	return true, nil
}

func (backend *SSHRemoteAppBackend) StartProxy(command *commonbackend.AddProxy) (uint16, bool, error) {
	// Allocate a new proxy ID
	backend.proxyIDLock.Lock()
	proxyID := backend.proxyIDIndex
	backend.proxyIDIndex++
	backend.proxyIDLock.Unlock()

	if command.Protocol == "tcp" {
		backend.tcpProxies[proxyID] = &TCPProxy{
			connections:      map[uint16]net.Conn{},
			proxyInformation: command,
		}

		server, err := net.Listen("tcp", fmt.Sprintf(":%d", command.DestPort))

		if err != nil {
			return 0, false, fmt.Errorf("failed to open server: %s", err.Error())
		}

		backend.tcpProxies[proxyID].server = server

		go func() {
			for {
				conn, err := server.Accept()

				if err != nil {
					log.Warnf("failed to accept connection: %s", err.Error())
					return
				}

				go func() {
					backend.tcpProxies[proxyID].connectionIDLock.Lock()
					connectionID := backend.tcpProxies[proxyID].connectionIDIndex
					backend.tcpProxies[proxyID].connectionIDIndex++
					backend.tcpProxies[proxyID].connectionIDLock.Unlock()

					dataBuf := make([]byte, 65535)

					onConnection := &datacommands.TCPConnectionOpened{
						ProxyID:      proxyID,
						ConnectionID: connectionID,
					}

					connectionCommandMarshalled, err := datacommands.Marshal(onConnection)

					if err != nil {
						log.Errorf("failed to marshal connection message: %s", err.Error())
					}

					backend.sock.Write(connectionCommandMarshalled)

					tcpData := &datacommands.TCPProxyData{
						ProxyID:      proxyID,
						ConnectionID: connectionID,
					}

					for {
						len, err := conn.Read(dataBuf)

						if err != nil {
							if errors.Is(err, net.ErrClosed) {
								return
							} else if err.Error() != "EOF" {
								log.Warnf("failed to read from sock: %s", err.Error())
							}

							conn.Close()
							break
						}

						tcpData.DataLength = uint16(len)
						marshalledMessageCommand, err := datacommands.Marshal(tcpData)

						if err != nil {
							log.Warnf("failed to marshal message data: %s", err.Error())

							conn.Close()
							break
						}

						if _, err := backend.sock.Write(marshalledMessageCommand); err != nil {
							log.Warnf("failed to send marshalled message data: %s", err.Error())

							conn.Close()
							break
						}

						if _, err := backend.sock.Write(dataBuf[:len]); err != nil {
							log.Warnf("failed to send raw message data: %s", err.Error())

							conn.Close()
							break
						}
					}

					onDisconnect := &datacommands.TCPConnectionClosed{
						ProxyID:      proxyID,
						ConnectionID: connectionID,
					}

					disconnectionCommandMarshalled, err := datacommands.Marshal(onDisconnect)

					if err != nil {
						log.Errorf("failed to marshal disconnection message: %s", err.Error())
					}

					backend.sock.Write(disconnectionCommandMarshalled)
				}()
			}
		}()
	} else if command.Protocol == "udp" {
		backend.udpProxies[proxyID] = &UDPProxy{
			proxyInformation: command,
		}

		server, err := net.ListenUDP("udp", &net.UDPAddr{
			IP:   net.IPv4(0, 0, 0, 0),
			Port: int(command.DestPort),
		})

		if err != nil {
			return 0, false, fmt.Errorf("failed to open server: %s", err.Error())
		}

		backend.udpProxies[proxyID].server = server
		dataBuf := make([]byte, 65535)

		udpProxyData := &datacommands.UDPProxyData{
			ProxyID: proxyID,
		}

		go func() {
			for {
				len, addr, err := server.ReadFromUDP(dataBuf)

				if err != nil {
					log.Warnf("failed to read from UDP socket: %s", err.Error())
					continue
				}

				udpProxyData.ClientIP = addr.IP.String()
				udpProxyData.ClientPort = uint16(addr.Port)
				udpProxyData.DataLength = uint16(len)

				marshalledMessageCommand, err := datacommands.Marshal(udpProxyData)

				if err != nil {
					log.Warnf("failed to marshal message data: %s", err.Error())
					continue
				}

				if _, err := backend.sock.Write(marshalledMessageCommand); err != nil {
					log.Warnf("failed to send marshalled message data: %s", err.Error())
					continue
				}

				if _, err := backend.sock.Write(dataBuf[:len]); err != nil {
					log.Warnf("failed to send raw message data: %s", err.Error())
					continue
				}
			}
		}()
	}

	return proxyID, true, nil
}

func (backend *SSHRemoteAppBackend) StopProxy(command *datacommands.RemoveProxy) (bool, error) {
	tcpProxy, ok := backend.tcpProxies[command.ProxyID]

	if !ok {
		udpProxy, ok := backend.udpProxies[command.ProxyID]

		if !ok {
			return ok, fmt.Errorf("could not find proxy")
		}

		udpProxy.server.Close()
		delete(backend.udpProxies, command.ProxyID)
	} else {
		for _, tcpConnection := range tcpProxy.connections {
			tcpConnection.Close()
		}

		tcpProxy.server.Close()
		delete(backend.tcpProxies, command.ProxyID)
	}

	return true, nil
}

func (backend *SSHRemoteAppBackend) GetAllProxies() []uint16 {
	proxyList := make([]uint16, len(backend.tcpProxies)+len(backend.udpProxies))

	currentPos := 0

	for tcpProxy := range backend.tcpProxies {
		proxyList[currentPos] = tcpProxy
		currentPos += 1
	}

	for udpProxy := range backend.udpProxies {
		proxyList[currentPos] = udpProxy
		currentPos += 1
	}

	return proxyList
}

func (backend *SSHRemoteAppBackend) ResolveProxy(proxyID uint16) *datacommands.ProxyInformationResponse {
	var proxyInformation *commonbackend.AddProxy
	response := &datacommands.ProxyInformationResponse{}

	tcpProxy, ok := backend.tcpProxies[proxyID]

	if !ok {
		udpProxy, ok := backend.udpProxies[proxyID]

		if !ok {
			response.Exists = false
			return response
		}

		proxyInformation = udpProxy.proxyInformation
	} else {
		proxyInformation = tcpProxy.proxyInformation
	}

	response.Exists = true
	response.SourceIP = proxyInformation.SourceIP
	response.SourcePort = proxyInformation.SourcePort
	response.DestPort = proxyInformation.DestPort
	response.Protocol = proxyInformation.Protocol

	return response
}

func (backend *SSHRemoteAppBackend) GetAllClientConnections(proxyID uint16) []uint16 {
	tcpProxy, ok := backend.tcpProxies[proxyID]

	if !ok {
		return []uint16{}
	}

	connectionsArray := make([]uint16, len(tcpProxy.connections))
	currentPos := 0

	for connectionIndex := range tcpProxy.connections {
		connectionsArray[currentPos] = connectionIndex
		currentPos++
	}

	return connectionsArray
}

func (backend *SSHRemoteAppBackend) ResolveConnection(proxyID, connectionID uint16) *datacommands.ProxyConnectionInformationResponse {
	response := &datacommands.ProxyConnectionInformationResponse{}
	tcpProxy, ok := backend.tcpProxies[proxyID]

	if !ok {
		response.Exists = false
		return response
	}

	connection, ok := tcpProxy.connections[connectionID]

	if !ok {
		response.Exists = false
		return response
	}

	addr := connection.RemoteAddr().String()
	ip := addr[:strings.LastIndex(addr, ":")]
	port, err := strconv.Atoi(addr[strings.LastIndex(addr, ":")+1:])

	if err != nil {
		log.Warnf("failed to parse client port: %s", err.Error())
		response.Exists = false

		return response
	}

	response.ClientIP = ip
	response.ClientPort = uint16(port)

	return response
}

func (backend *SSHRemoteAppBackend) CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse {
	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func (backend *SSHRemoteAppBackend) CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse {
	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func (backend *SSHRemoteAppBackend) HandleTCPMessage(message *datacommands.TCPProxyData, data []byte) {
	tcpProxy, ok := backend.tcpProxies[message.ProxyID]

	if !ok {
		return
	}

	connection, ok := tcpProxy.connections[message.ConnectionID]

	if !ok {
		return
	}

	connection.Write(data)
}

func (backend *SSHRemoteAppBackend) HandleUDPMessage(message *datacommands.UDPProxyData, data []byte) {
	udpProxy, ok := backend.udpProxies[message.ProxyID]

	if !ok {
		return
	}

	udpProxy.server.WriteToUDP(data, &net.UDPAddr{
		IP:   net.ParseIP(message.ClientIP),
		Port: int(message.ClientPort),
	})
}

func (backend *SSHRemoteAppBackend) OnTCPConnectionClosed(proxyID, connectionID uint16) {
	tcpProxy, ok := backend.tcpProxies[proxyID]

	if !ok {
		return
	}

	connection, ok := tcpProxy.connections[connectionID]

	if !ok {
		return
	}

	connection.Close()
	delete(tcpProxy.connections, connectionID)
}

func (backend *SSHRemoteAppBackend) OnSocketConnection(sock net.Conn) {
	backend.sock = sock
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
