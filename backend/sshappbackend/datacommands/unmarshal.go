package datacommands

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// Unmarshal reads from the provided connection and returns
// the message type (as a string), the unmarshalled struct, or an error.
func Unmarshal(conn io.Reader) (string, interface{}, error) {
	// Every command starts with a 1-byte command ID.
	header := make([]byte, 1)
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", nil, fmt.Errorf("couldn't read command ID: %w", err)
	}

	cmdID := header[0]
	switch cmdID {
	// ProxyStatusRequest: 1 byte ID + 2 bytes ProxyID.
	case ProxyStatusRequestID:
		buf := make([]byte, 2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyStatusRequest ProxyID: %w", err)
		}

		proxyID := binary.BigEndian.Uint16(buf)

		return "proxyStatusRequest", &ProxyStatusRequest{
			Type:    "proxyStatusRequest",
			ProxyID: proxyID,
		}, nil

	// ProxyStatusResponse: 1 byte ID + 2 bytes ProxyID + 1 byte IsActive.
	case ProxyStatusResponseID:
		buf := make([]byte, 2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyStatusResponse ProxyID: %w", err)
		}

		proxyID := binary.BigEndian.Uint16(buf)
		boolBuf := make([]byte, 1)

		if _, err := io.ReadFull(conn, boolBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyStatusResponse IsActive: %w", err)
		}

		isActive := boolBuf[0] != 0

		return "proxyStatusResponse", &ProxyStatusResponse{
			Type:     "proxyStatusResponse",
			ProxyID:  proxyID,
			IsActive: isActive,
		}, nil

	// RemoveProxy: 1 byte ID + 2 bytes ProxyID.
	case RemoveProxyID:
		buf := make([]byte, 2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read RemoveProxy ProxyID: %w", err)
		}

		proxyID := binary.BigEndian.Uint16(buf)

		return "removeProxy", &RemoveProxy{
			Type:    "removeProxy",
			ProxyID: proxyID,
		}, nil

	// ProxyConnectionsRequest: 1 byte ID + 2 bytes ProxyID.
	case ProxyConnectionsRequestID:
		buf := make([]byte, 2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyConnectionsRequest ProxyID: %w", err)
		}

		proxyID := binary.BigEndian.Uint16(buf)

		return "proxyConnectionsRequest", &ProxyConnectionsRequest{
			Type:    "proxyConnectionsRequest",
			ProxyID: proxyID,
		}, nil

	// ProxyConnectionsResponse: 1 byte ID + 2 bytes Connections length + 2 bytes for each Connection in Connections.
	case ProxyConnectionsResponseID:
		buf := make([]byte, 2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyConnectionsResponse length: %w", err)
		}

		length := binary.BigEndian.Uint16(buf)
		connections := make([]uint16, length)

		var failedDuringReading error

		for connectionIndex := range connections {
			if _, err := io.ReadFull(conn, buf); err != nil {
				failedDuringReading = fmt.Errorf("couldn't read ProxyConnectionsResponse with position of %d: %w", connectionIndex, err)
				break
			}

			connections[connectionIndex] = binary.BigEndian.Uint16(buf)
		}

		return "proxyConnectionsResponse", &ProxyConnectionsResponse{
			Type:        "proxyConnectionsResponse",
			Connections: connections,
		}, failedDuringReading

	// ProxyInstanceResponse: 1 byte ID + 2 bytes Proxies length + 2 bytes for each Proxy in Proxies.
	case ProxyInstanceResponseID:
		buf := make([]byte, 2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyConnectionsResponse length: %w", err)
		}

		length := binary.BigEndian.Uint16(buf)
		proxies := make([]uint16, length)

		var failedDuringReading error

		for connectionIndex := range proxies {
			if _, err := io.ReadFull(conn, buf); err != nil {
				failedDuringReading = fmt.Errorf("couldn't read ProxyConnectionsResponse with position of %d: %w", connectionIndex, err)
				break
			}

			proxies[connectionIndex] = binary.BigEndian.Uint16(buf)
		}

		return "proxyInstanceResponse", &ProxyInstanceResponse{
			Type:    "proxyInstanceResponse",
			Proxies: proxies,
		}, failedDuringReading

	// TCPConnectionOpened: 1 byte ID + 2 bytes ProxyID + 2 bytes ConnectionID.
	case TCPConnectionOpenedID:
		buf := make([]byte, 2+2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read TCPConnectionOpened fields: %w", err)
		}

		proxyID := binary.BigEndian.Uint16(buf[0:2])
		connectionID := binary.BigEndian.Uint16(buf[2:4])

		return "tcpConnectionOpened", &TCPConnectionOpened{
			Type:         "tcpConnectionOpened",
			ProxyID:      proxyID,
			ConnectionID: connectionID,
		}, nil

	// TCPConnectionClosed: 1 byte ID + 2 bytes ProxyID + 2 bytes ConnectionID.
	case TCPConnectionClosedID:
		buf := make([]byte, 2+2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read TCPConnectionClosed fields: %w", err)
		}

		proxyID := binary.BigEndian.Uint16(buf[0:2])
		connectionID := binary.BigEndian.Uint16(buf[2:4])

		return "tcpConnectionClosed", &TCPConnectionClosed{
			Type:         "tcpConnectionClosed",
			ProxyID:      proxyID,
			ConnectionID: connectionID,
		}, nil

	// TCPProxyData: 1 byte ID + 2 bytes ProxyID + 2 bytes ConnectionID + 2 bytes DataLength.
	case TCPProxyDataID:
		buf := make([]byte, 2+2+2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read TCPProxyData fields: %w", err)
		}

		proxyID := binary.BigEndian.Uint16(buf[0:2])
		connectionID := binary.BigEndian.Uint16(buf[2:4])
		dataLength := binary.BigEndian.Uint16(buf[4:6])

		return "tcpProxyData", &TCPProxyData{
			Type:         "tcpProxyData",
			ProxyID:      proxyID,
			ConnectionID: connectionID,
			DataLength:   dataLength,
		}, nil

	// UDPProxyData:
	// Format: 1 byte ID + 2 bytes ProxyID + 2 bytes ConnectionID +
	//         1 byte IP version + IP bytes + 2 bytes ClientPort + 2 bytes DataLength.
	case UDPProxyDataID:
		// Read 2 bytes ProxyID + 2 bytes ConnectionID.
		buf := make([]byte, 2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read UDPProxyData ProxyID/ConnectionID: %w", err)
		}

		proxyID := binary.BigEndian.Uint16(buf)

		// Read IP version.
		ipVerBuf := make([]byte, 1)

		if _, err := io.ReadFull(conn, ipVerBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read UDPProxyData IP version: %w", err)
		}

		var ipSize int

		if ipVerBuf[0] == 4 {
			ipSize = IPv4Size
		} else if ipVerBuf[0] == 6 {
			ipSize = IPv6Size
		} else {
			return "", nil, fmt.Errorf("invalid IP version received: %v", ipVerBuf[0])
		}

		// Read the IP bytes.
		ipBytes := make([]byte, ipSize)
		if _, err := io.ReadFull(conn, ipBytes); err != nil {
			return "", nil, fmt.Errorf("couldn't read UDPProxyData IP bytes: %w", err)
		}
		clientIP := net.IP(ipBytes).String()

		// Read ClientPort.
		portBuf := make([]byte, 2)

		if _, err := io.ReadFull(conn, portBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read UDPProxyData ClientPort: %w", err)
		}

		clientPort := binary.BigEndian.Uint16(portBuf)

		// Read DataLength.
		dataLengthBuf := make([]byte, 2)

		if _, err := io.ReadFull(conn, dataLengthBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read UDPProxyData DataLength: %w", err)
		}

		dataLength := binary.BigEndian.Uint16(dataLengthBuf)

		return "udpProxyData", &UDPProxyData{
			Type:       "udpProxyData",
			ProxyID:    proxyID,
			ClientIP:   clientIP,
			ClientPort: clientPort,
			DataLength: dataLength,
		}, nil

	// ProxyInformationRequest: 1 byte ID + 2 bytes ProxyID.
	case ProxyInformationRequestID:
		buf := make([]byte, 2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyInformationRequest ProxyID: %w", err)
		}

		proxyID := binary.BigEndian.Uint16(buf)

		return "proxyInformationRequest", &ProxyInformationRequest{
			Type:    "proxyInformationRequest",
			ProxyID: proxyID,
		}, nil

	// ProxyInformationResponse:
	// Format: 1 byte ID + 1 byte Exists +
	//         1 byte IP version + IP bytes + 2 bytes SourcePort + 2 bytes DestPort + 1 byte Protocol.
	case ProxyInformationResponseID:
		// Read Exists flag.
		boolBuf := make([]byte, 1)

		if _, err := io.ReadFull(conn, boolBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyInformationResponse Exists flag: %w", err)
		}

		exists := boolBuf[0] != 0

		if !exists {
			return "proxyInformationResponse", &ProxyInformationResponse{
				Type:   "proxyInformationResponse",
				Exists: exists,
			}, nil
		}

		// Read IP version.
		ipVerBuf := make([]byte, 1)

		if _, err := io.ReadFull(conn, ipVerBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyInformationResponse IP version: %w", err)
		}

		var ipSize int

		if ipVerBuf[0] == 4 {
			ipSize = IPv4Size
		} else if ipVerBuf[0] == 6 {
			ipSize = IPv6Size
		} else {
			return "", nil, fmt.Errorf("invalid IP version in ProxyInformationResponse: %v", ipVerBuf[0])
		}

		// Read the source IP bytes.
		ipBytes := make([]byte, ipSize)

		if _, err := io.ReadFull(conn, ipBytes); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyInformationResponse IP bytes: %w", err)
		}

		sourceIP := net.IP(ipBytes).String()

		// Read SourcePort and DestPort.
		portsBuf := make([]byte, 2+2)

		if _, err := io.ReadFull(conn, portsBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyInformationResponse ports: %w", err)
		}

		sourcePort := binary.BigEndian.Uint16(portsBuf[0:2])
		destPort := binary.BigEndian.Uint16(portsBuf[2:4])

		// Read protocol.
		protoBuf := make([]byte, 1)

		if _, err := io.ReadFull(conn, protoBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyInformationResponse protocol: %w", err)
		}
		var protocol string
		if protoBuf[0] == TCP {
			protocol = "tcp"
		} else if protoBuf[0] == UDP {
			protocol = "udp"
		} else {
			return "", nil, fmt.Errorf("invalid protocol value in ProxyInformationResponse: %d", protoBuf[0])
		}

		return "proxyInformationResponse", &ProxyInformationResponse{
			Type:       "proxyInformationResponse",
			Exists:     exists,
			SourceIP:   sourceIP,
			SourcePort: sourcePort,
			DestPort:   destPort,
			Protocol:   protocol,
		}, nil

	// ProxyConnectionInformationRequest: 1 byte ID + 2 bytes ProxyID + 2 bytes ConnectionID.
	case ProxyConnectionInformationRequestID:
		buf := make([]byte, 2+2)

		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyConnectionInformationRequest fields: %w", err)
		}

		proxyID := binary.BigEndian.Uint16(buf[0:2])
		connectionID := binary.BigEndian.Uint16(buf[2:4])

		return "proxyConnectionInformationRequest", &ProxyConnectionInformationRequest{
			Type:         "proxyConnectionInformationRequest",
			ProxyID:      proxyID,
			ConnectionID: connectionID,
		}, nil

	// ProxyConnectionInformationResponse:
	// Format: 1 byte ID + 1 byte Exists + 1 byte IP version + IP bytes + 2 bytes ClientPort.
	case ProxyConnectionInformationResponseID:
		// Read Exists flag.
		boolBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, boolBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyConnectionInformationResponse Exists flag: %w", err)
		}

		exists := boolBuf[0] != 0

		if !exists {
			return "proxyConnectionInformationResponse", &ProxyConnectionInformationResponse{
				Type:   "proxyConnectionInformationResponse",
				Exists: exists,
			}, nil
		}

		// Read IP version.
		ipVerBuf := make([]byte, 1)

		if _, err := io.ReadFull(conn, ipVerBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyConnectionInformationResponse IP version: %w", err)
		}

		if ipVerBuf[0] != 4 && ipVerBuf[0] != 6 {
			return "", nil, fmt.Errorf("invalid IP version in ProxyConnectionInformationResponse: %v", ipVerBuf[0])
		}

		var ipSize int

		if ipVerBuf[0] == 4 {
			ipSize = IPv4Size
		} else {
			ipSize = IPv6Size
		}

		// Read IP bytes.
		ipBytes := make([]byte, ipSize)

		if _, err := io.ReadFull(conn, ipBytes); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyConnectionInformationResponse IP bytes: %w", err)
		}

		clientIP := net.IP(ipBytes).String()

		// Read ClientPort.
		portBuf := make([]byte, 2)

		if _, err := io.ReadFull(conn, portBuf); err != nil {
			return "", nil, fmt.Errorf("couldn't read ProxyConnectionInformationResponse ClientPort: %w", err)
		}

		clientPort := binary.BigEndian.Uint16(portBuf)

		return "proxyConnectionInformationResponse", &ProxyConnectionInformationResponse{
			Type:       "proxyConnectionInformationResponse",
			Exists:     exists,
			ClientIP:   clientIP,
			ClientPort: clientPort,
		}, nil
	default:
		return "", nil, fmt.Errorf("unknown command id: %v", cmdID)
	}
}
