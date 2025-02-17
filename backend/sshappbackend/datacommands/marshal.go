package datacommands

import (
	"encoding/binary"
	"fmt"
	"net"
)

// Example size and protocol constants — adjust as needed.
const (
	IPv4Size = 4
	IPv6Size = 16

	TCP = 1
	UDP = 2
)

// Marshal takes a command (pointer to one of our structs) and converts it to a byte slice.
func Marshal(command interface{}) ([]byte, error) {
	switch cmd := command.(type) {
	// ProxyStatusRequest: 1 byte for the command ID + 2 bytes for the ProxyID.
	case *ProxyStatusRequest:
		buf := make([]byte, 1+2)

		buf[0] = ProxyStatusRequestID
		binary.BigEndian.PutUint16(buf[1:], cmd.ProxyID)

		return buf, nil

	// ProxyStatusResponse: 1 byte for the command ID, 2 bytes for ProxyID, and 1 byte for IsActive.
	case *ProxyStatusResponse:
		buf := make([]byte, 1+2+1)

		buf[0] = ProxyStatusResponseID
		binary.BigEndian.PutUint16(buf[1:], cmd.ProxyID)

		if cmd.IsActive {
			buf[3] = 1
		} else {
			buf[3] = 0
		}

		return buf, nil

	// RemoveProxy: 1 byte for the command ID + 2 bytes for the ProxyID.
	case *RemoveProxy:
		buf := make([]byte, 1+2)

		buf[0] = RemoveProxyID
		binary.BigEndian.PutUint16(buf[1:], cmd.ProxyID)

		return buf, nil

	// ProxyConnectionsRequest: 1 byte for the command ID + 2 bytes for the ProxyID.
	case *ProxyConnectionsRequest:
		buf := make([]byte, 1+2)

		buf[0] = ProxyConnectionsRequestID
		binary.BigEndian.PutUint16(buf[1:], cmd.ProxyID)

		return buf, nil

	// ProxyConnectionsResponse: 1 byte for the command ID + 2 bytes length of the Connections + 2 bytes for each
	// number in the Connection array.
	case *ProxyConnectionsResponse:
		buf := make([]byte, 1+((len(cmd.Connections)+1)*2))

		buf[0] = ProxyConnectionsResponseID
		binary.BigEndian.PutUint16(buf[1:], uint16(len(cmd.Connections)))

		for connectionIndex, connection := range cmd.Connections {
			binary.BigEndian.PutUint16(buf[3+(connectionIndex*2):], connection)
		}

		return buf, nil

	// ProxyConnectionsResponse: 1 byte for the command ID + 2 bytes length of the Proxies + 2 bytes for each
	// number in the Proxies array.
	case *ProxyInstanceResponse:
		buf := make([]byte, 1+((len(cmd.Proxies)+1)*2))

		buf[0] = ProxyInstanceResponseID
		binary.BigEndian.PutUint16(buf[1:], uint16(len(cmd.Proxies)))

		for connectionIndex, connection := range cmd.Proxies {
			binary.BigEndian.PutUint16(buf[3+(connectionIndex*2):], connection)
		}

		return buf, nil

	// TCPConnectionOpened: 1 byte for the command ID + 2 bytes ProxyID + 2 bytes ConnectionID.
	case *TCPConnectionOpened:
		buf := make([]byte, 1+2+2)

		buf[0] = TCPConnectionOpenedID
		binary.BigEndian.PutUint16(buf[1:], cmd.ProxyID)
		binary.BigEndian.PutUint16(buf[3:], cmd.ConnectionID)

		return buf, nil

	// TCPConnectionClosed: 1 byte for the command ID + 2 bytes ProxyID + 2 bytes ConnectionID.
	case *TCPConnectionClosed:
		buf := make([]byte, 1+2+2)

		buf[0] = TCPConnectionClosedID
		binary.BigEndian.PutUint16(buf[1:], cmd.ProxyID)
		binary.BigEndian.PutUint16(buf[3:], cmd.ConnectionID)

		return buf, nil

	// TCPProxyData: 1 byte ID + 2 bytes ProxyID + 2 bytes ConnectionID + 2 bytes DataLength.
	case *TCPProxyData:
		buf := make([]byte, 1+2+2+2)

		buf[0] = TCPProxyDataID
		binary.BigEndian.PutUint16(buf[1:], cmd.ProxyID)
		binary.BigEndian.PutUint16(buf[3:], cmd.ConnectionID)
		binary.BigEndian.PutUint16(buf[5:], cmd.DataLength)

		return buf, nil

	// UDPProxyData:
	// Format: 1 byte ID + 2 bytes ProxyID + 2 bytes ConnectionID +
	//         1 byte IP version + IP bytes + 2 bytes ClientPort + 2 bytes DataLength.
	case *UDPProxyData:
		ip := net.ParseIP(cmd.ClientIP)
		if ip == nil {
			return nil, fmt.Errorf("invalid client IP: %v", cmd.ClientIP)
		}

		var ipVer uint8
		var ipBytes []byte

		if ip4 := ip.To4(); ip4 != nil {
			ipBytes = ip4
			ipVer = 4
		} else if ip16 := ip.To16(); ip16 != nil {
			ipBytes = ip16
			ipVer = 6
		} else {
			return nil, fmt.Errorf("unable to detect IP version for: %v", cmd.ClientIP)
		}

		totalSize := 1 + // id
			2 + // ProxyID
			1 + // IP version
			len(ipBytes) + // client IP bytes
			2 + // ClientPort
			2 // DataLength

		buf := make([]byte, totalSize)
		offset := 0
		buf[offset] = UDPProxyDataID
		offset++

		binary.BigEndian.PutUint16(buf[offset:], cmd.ProxyID)
		offset += 2

		buf[offset] = ipVer
		offset++

		copy(buf[offset:], ipBytes)
		offset += len(ipBytes)

		binary.BigEndian.PutUint16(buf[offset:], cmd.ClientPort)
		offset += 2

		binary.BigEndian.PutUint16(buf[offset:], cmd.DataLength)

		return buf, nil

	// ProxyInformationRequest: 1 byte ID + 2 bytes ProxyID.
	case *ProxyInformationRequest:
		buf := make([]byte, 1+2)
		buf[0] = ProxyInformationRequestID
		binary.BigEndian.PutUint16(buf[1:], cmd.ProxyID)
		return buf, nil

	// ProxyInformationResponse:
	// Format: 1 byte ID + 1 byte Exists + (if exists:)
	//         1 byte IP version + IP bytes + 2 bytes SourcePort + 2 bytes DestPort + 1 byte Protocol.
	// (For simplicity, this marshaller always writes the IP and port info even if !Exists.)
	case *ProxyInformationResponse:
		if !cmd.Exists {
			buf := make([]byte, 1+1)
			buf[0] = ProxyInformationResponseID
			buf[1] = 0 /* false */

			return buf, nil
		}

		ip := net.ParseIP(cmd.SourceIP)

		if ip == nil {
			return nil, fmt.Errorf("invalid source IP: %v", cmd.SourceIP)
		}

		var ipVer uint8
		var ipBytes []byte

		if ip4 := ip.To4(); ip4 != nil {
			ipBytes = ip4
			ipVer = 4
		} else if ip16 := ip.To16(); ip16 != nil {
			ipBytes = ip16
			ipVer = 6
		} else {
			return nil, fmt.Errorf("unable to detect IP version for: %v", cmd.SourceIP)
		}

		totalSize := 1 + // id
			1 + // Exists flag
			1 + // IP version
			len(ipBytes) +
			2 + // SourcePort
			2 + // DestPort
			1 // Protocol

		buf := make([]byte, totalSize)

		offset := 0
		buf[offset] = ProxyInformationResponseID
		offset++

		// We already handle this above
		buf[offset] = 1 /* true */
		offset++

		buf[offset] = ipVer
		offset++

		copy(buf[offset:], ipBytes)
		offset += len(ipBytes)

		binary.BigEndian.PutUint16(buf[offset:], cmd.SourcePort)
		offset += 2

		binary.BigEndian.PutUint16(buf[offset:], cmd.DestPort)
		offset += 2

		// Encode protocol as 1 byte.
		switch cmd.Protocol {
		case "tcp":
			buf[offset] = TCP
		case "udp":
			buf[offset] = UDP
		default:
			return nil, fmt.Errorf("invalid protocol: %v", cmd.Protocol)
		}

		// offset++ (not needed since we are at the end)
		return buf, nil

	// ProxyConnectionInformationRequest: 1 byte ID + 2 bytes ProxyID + 2 bytes ConnectionID.
	case *ProxyConnectionInformationRequest:
		buf := make([]byte, 1+2+2)

		buf[0] = ProxyConnectionInformationRequestID
		binary.BigEndian.PutUint16(buf[1:], cmd.ProxyID)
		binary.BigEndian.PutUint16(buf[3:], cmd.ConnectionID)

		return buf, nil

	// ProxyConnectionInformationResponse:
	// Format: 1 byte ID + 1 byte Exists + (if exists:)
	//         1 byte IP version + IP bytes + 2 bytes ClientPort.
	// This marshaller only writes the rest of the data if Exists.
	case *ProxyConnectionInformationResponse:
		if !cmd.Exists {
			buf := make([]byte, 1+1)
			buf[0] = ProxyConnectionInformationResponseID
			buf[1] = 0 /* false */

			return buf, nil
		}

		ip := net.ParseIP(cmd.ClientIP)

		if ip == nil {
			return nil, fmt.Errorf("invalid client IP: %v", cmd.ClientIP)
		}

		var ipVer uint8
		var ipBytes []byte
		if ip4 := ip.To4(); ip4 != nil {
			ipBytes = ip4
			ipVer = 4
		} else if ip16 := ip.To16(); ip16 != nil {
			ipBytes = ip16
			ipVer = 6
		} else {
			return nil, fmt.Errorf("unable to detect IP version for: %v", cmd.ClientIP)
		}

		totalSize := 1 + // id
			1 + // Exists flag
			1 + // IP version
			len(ipBytes) +
			2 // ClientPort

		buf := make([]byte, totalSize)
		offset := 0
		buf[offset] = ProxyConnectionInformationResponseID
		offset++

		// We already handle this above
		buf[offset] = 1 /* true */
		offset++

		buf[offset] = ipVer
		offset++

		copy(buf[offset:], ipBytes)
		offset += len(ipBytes)

		binary.BigEndian.PutUint16(buf[offset:], cmd.ClientPort)

		return buf, nil

	default:
		return nil, fmt.Errorf("unsupported command type")
	}
}
