package porttranslation

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type connectionData struct {
	udpConn         *net.UDPConn
	buf             []byte
	hasBeenAliveFor time.Time
}

type PortTranslation struct {
	UDPAddr   *net.UDPAddr
	WriteFrom func(ip string, port uint16, data []byte)

	newConnectionLock sync.Mutex
	connections       map[string]map[uint16]*connectionData
}

func (translation *PortTranslation) CleanupPorts() {
	if translation.connections == nil {
		translation.connections = map[string]map[uint16]*connectionData{}
		return
	}

	for connectionIPIndex, connectionPorts := range translation.connections {
		anyAreAlive := false

		for connectionPortIndex, connectionData := range connectionPorts {
			if time.Now().Before(connectionData.hasBeenAliveFor.Add(3 * time.Minute)) {
				anyAreAlive = true
				continue
			}

			connectionData.udpConn.Close()
			delete(connectionPorts, connectionPortIndex)
		}

		if !anyAreAlive {
			delete(translation.connections, connectionIPIndex)
		}
	}
}

func (translation *PortTranslation) StopAllPorts() {
	if translation.connections == nil {
		return
	}

	for connectionIPIndex, connectionPorts := range translation.connections {
		for connectionPortIndex, connectionData := range connectionPorts {
			connectionData.udpConn.Close()
			delete(connectionPorts, connectionPortIndex)
		}

		delete(translation.connections, connectionIPIndex)
	}

	translation.connections = nil
}

func (translation *PortTranslation) WriteTo(ip string, port uint16, data []byte) (int, error) {
	if translation.connections == nil {
		translation.connections = map[string]map[uint16]*connectionData{}
	}

	connectionPortData, ok := translation.connections[ip]

	if !ok {
		translation.connections[ip] = map[uint16]*connectionData{}
		connectionPortData = translation.connections[ip]
	}

	connectionStruct, ok := connectionPortData[port]

	if !ok {
		connectionPortData[port] = &connectionData{}
		connectionStruct = connectionPortData[port]

		udpConn, err := net.DialUDP("udp", nil, translation.UDPAddr)

		if err != nil {
			return 0, fmt.Errorf("failed to initialize UDP socket: %s", err.Error())
		}

		connectionStruct.udpConn = udpConn
		connectionStruct.buf = make([]byte, 65535)

		go func() {
			for {
				n, err := udpConn.Read(connectionStruct.buf)

				if err != nil {
					udpConn.Close()
					delete(connectionPortData, port)

					return
				}

				connectionStruct.hasBeenAliveFor = time.Now()
				translation.WriteFrom(ip, port, connectionStruct.buf[:n])
			}
		}()
	}

	connectionStruct.hasBeenAliveFor = time.Now()
	return connectionStruct.udpConn.Write(data)
}
