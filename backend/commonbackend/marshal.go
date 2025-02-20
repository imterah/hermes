package commonbackend

import (
	"encoding/binary"
	"fmt"
	"net"
)

func marshalIndividualConnectionStruct(conn *ProxyClientConnection) []byte {
	sourceIPOriginal := net.ParseIP(conn.SourceIP)
	clientIPOriginal := net.ParseIP(conn.ClientIP)

	var serverIPVer uint8
	var sourceIP []byte

	if sourceIPOriginal.To4() == nil {
		serverIPVer = IPv6
		sourceIP = sourceIPOriginal.To16()
	} else {
		serverIPVer = IPv4
		sourceIP = sourceIPOriginal.To4()
	}

	var clientIPVer uint8
	var clientIP []byte

	if clientIPOriginal.To4() == nil {
		clientIPVer = IPv6
		clientIP = clientIPOriginal.To16()
	} else {
		clientIPVer = IPv4
		clientIP = clientIPOriginal.To4()
	}

	connectionBlock := make([]byte, 8+len(sourceIP)+len(clientIP))

	connectionBlock[0] = serverIPVer
	copy(connectionBlock[1:len(sourceIP)+1], sourceIP)

	binary.BigEndian.PutUint16(connectionBlock[1+len(sourceIP):3+len(sourceIP)], conn.SourcePort)
	binary.BigEndian.PutUint16(connectionBlock[3+len(sourceIP):5+len(sourceIP)], conn.DestPort)

	connectionBlock[5+len(sourceIP)] = clientIPVer
	copy(connectionBlock[6+len(sourceIP):6+len(sourceIP)+len(clientIP)], clientIP)
	binary.BigEndian.PutUint16(connectionBlock[6+len(sourceIP)+len(clientIP):8+len(sourceIP)+len(clientIP)], conn.ClientPort)

	return connectionBlock
}

func marshalIndividualProxyStruct(conn *ProxyInstance) ([]byte, error) {
	sourceIPOriginal := net.ParseIP(conn.SourceIP)

	var sourceIPVer uint8
	var sourceIP []byte

	if sourceIPOriginal.To4() == nil {
		sourceIPVer = IPv6
		sourceIP = sourceIPOriginal.To16()
	} else {
		sourceIPVer = IPv4
		sourceIP = sourceIPOriginal.To4()
	}

	proxyBlock := make([]byte, 6+len(sourceIP))

	proxyBlock[0] = sourceIPVer
	copy(proxyBlock[1:len(sourceIP)+1], sourceIP)

	binary.BigEndian.PutUint16(proxyBlock[1+len(sourceIP):3+len(sourceIP)], conn.SourcePort)
	binary.BigEndian.PutUint16(proxyBlock[3+len(sourceIP):5+len(sourceIP)], conn.DestPort)

	var protocolVersion uint8

	if conn.Protocol == "tcp" {
		protocolVersion = TCP
	} else if conn.Protocol == "udp" {
		protocolVersion = UDP
	} else {
		return proxyBlock, fmt.Errorf("invalid protocol recieved")
	}

	proxyBlock[5+len(sourceIP)] = protocolVersion

	return proxyBlock, nil
}

func Marshal(command interface{}) ([]byte, error) {
	switch command := command.(type) {
	case *Start:
		startCommandBytes := make([]byte, 1+2+len(command.Arguments))
		startCommandBytes[0] = StartID
		binary.BigEndian.PutUint16(startCommandBytes[1:3], uint16(len(command.Arguments)))
		copy(startCommandBytes[3:], command.Arguments)

		return startCommandBytes, nil
	case *Stop:
		return []byte{StopID}, nil
	case *AddProxy:
		sourceIP := net.ParseIP(command.SourceIP)

		var ipVer uint8
		var ipBytes []byte

		if sourceIP.To4() == nil {
			ipBytes = sourceIP.To16()
			ipVer = IPv6
		} else {
			ipBytes = sourceIP.To4()
			ipVer = IPv4
		}

		addConnectionBytes := make([]byte, 1+1+len(ipBytes)+2+2+1)

		addConnectionBytes[0] = AddProxyID
		addConnectionBytes[1] = ipVer

		copy(addConnectionBytes[2:2+len(ipBytes)], ipBytes)

		binary.BigEndian.PutUint16(addConnectionBytes[2+len(ipBytes):4+len(ipBytes)], command.SourcePort)
		binary.BigEndian.PutUint16(addConnectionBytes[4+len(ipBytes):6+len(ipBytes)], command.DestPort)

		var protocol uint8

		if command.Protocol == "tcp" {
			protocol = TCP
		} else if command.Protocol == "udp" {
			protocol = UDP
		} else {
			return nil, fmt.Errorf("invalid protocol")
		}

		addConnectionBytes[6+len(ipBytes)] = protocol

		return addConnectionBytes, nil
	case *RemoveProxy:
		sourceIP := net.ParseIP(command.SourceIP)

		var ipVer uint8
		var ipBytes []byte

		if sourceIP.To4() == nil {
			ipBytes = sourceIP.To16()
			ipVer = IPv6
		} else {
			ipBytes = sourceIP.To4()
			ipVer = IPv4
		}

		removeConnectionBytes := make([]byte, 1+1+len(ipBytes)+2+2+1)

		removeConnectionBytes[0] = RemoveProxyID
		removeConnectionBytes[1] = ipVer
		copy(removeConnectionBytes[2:2+len(ipBytes)], ipBytes)
		binary.BigEndian.PutUint16(removeConnectionBytes[2+len(ipBytes):4+len(ipBytes)], command.SourcePort)
		binary.BigEndian.PutUint16(removeConnectionBytes[4+len(ipBytes):6+len(ipBytes)], command.DestPort)

		var protocol uint8

		if command.Protocol == "tcp" {
			protocol = TCP
		} else if command.Protocol == "udp" {
			protocol = UDP
		} else {
			return nil, fmt.Errorf("invalid protocol")
		}

		removeConnectionBytes[6+len(ipBytes)] = protocol

		return removeConnectionBytes, nil
	case *ProxyConnectionsResponse:
		connectionsArray := make([][]byte, len(command.Connections))
		totalSize := 0

		for connIndex, conn := range command.Connections {
			connectionsArray[connIndex] = marshalIndividualConnectionStruct(conn)
			totalSize += len(connectionsArray[connIndex]) + 1
		}

		if totalSize == 0 {
			totalSize = 1
		}

		connectionCommandArray := make([]byte, totalSize+1)
		connectionCommandArray[0] = ProxyConnectionsResponseID

		currentPosition := 1

		for _, connection := range connectionsArray {
			copy(connectionCommandArray[currentPosition:currentPosition+len(connection)], connection)
			connectionCommandArray[currentPosition+len(connection)] = '\r'
			currentPosition += len(connection) + 1
		}

		connectionCommandArray[totalSize] = '\n'
		return connectionCommandArray, nil
	case *CheckClientParameters:
		sourceIP := net.ParseIP(command.SourceIP)

		var ipVer uint8
		var ipBytes []byte

		if sourceIP.To4() == nil {
			ipBytes = sourceIP.To16()
			ipVer = IPv6
		} else {
			ipBytes = sourceIP.To4()
			ipVer = IPv4
		}

		checkClientBytes := make([]byte, 1+1+len(ipBytes)+2+2+1)

		checkClientBytes[0] = CheckClientParametersID
		checkClientBytes[1] = ipVer
		copy(checkClientBytes[2:2+len(ipBytes)], ipBytes)
		binary.BigEndian.PutUint16(checkClientBytes[2+len(ipBytes):4+len(ipBytes)], command.SourcePort)
		binary.BigEndian.PutUint16(checkClientBytes[4+len(ipBytes):6+len(ipBytes)], command.DestPort)

		var protocol uint8

		if command.Protocol == "tcp" {
			protocol = TCP
		} else if command.Protocol == "udp" {
			protocol = UDP
		} else {
			return nil, fmt.Errorf("invalid protocol")
		}

		checkClientBytes[6+len(ipBytes)] = protocol

		return checkClientBytes, nil
	case *CheckServerParameters:
		serverCommandBytes := make([]byte, 1+2+len(command.Arguments))
		serverCommandBytes[0] = CheckServerParametersID
		binary.BigEndian.PutUint16(serverCommandBytes[1:3], uint16(len(command.Arguments)))
		copy(serverCommandBytes[3:], command.Arguments)

		return serverCommandBytes, nil
	case *CheckParametersResponse:
		var checkMethod uint8

		if command.InResponseTo == "checkClientParameters" {
			checkMethod = CheckClientParametersID
		} else if command.InResponseTo == "checkServerParameters" {
			checkMethod = CheckServerParametersID
		} else {
			return nil, fmt.Errorf("invalid mode recieved (must be either checkClientParameters or checkServerParameters)")
		}

		var isValid uint8

		if command.IsValid {
			isValid = 1
		}

		checkResponseBytes := make([]byte, 3+2+len(command.Message))
		checkResponseBytes[0] = CheckParametersResponseID
		checkResponseBytes[1] = checkMethod
		checkResponseBytes[2] = isValid

		binary.BigEndian.PutUint16(checkResponseBytes[3:5], uint16(len(command.Message)))

		if len(command.Message) != 0 {
			copy(checkResponseBytes[5:], []byte(command.Message))
		}

		return checkResponseBytes, nil
	case *BackendStatusResponse:
		var isRunning uint8

		if command.IsRunning {
			isRunning = 1
		} else {
			isRunning = 0
		}

		statusResponseBytes := make([]byte, 3+2+len(command.Message))
		statusResponseBytes[0] = BackendStatusResponseID
		statusResponseBytes[1] = isRunning
		statusResponseBytes[2] = byte(command.StatusCode)

		binary.BigEndian.PutUint16(statusResponseBytes[3:5], uint16(len(command.Message)))

		if len(command.Message) != 0 {
			copy(statusResponseBytes[5:], []byte(command.Message))
		}

		return statusResponseBytes, nil
	case *BackendStatusRequest:
		statusRequestBytes := make([]byte, 1)
		statusRequestBytes[0] = BackendStatusRequestID

		return statusRequestBytes, nil
	case *ProxyStatusRequest:
		sourceIP := net.ParseIP(command.SourceIP)

		var ipVer uint8
		var ipBytes []byte

		if sourceIP.To4() == nil {
			ipBytes = sourceIP.To16()
			ipVer = IPv6
		} else {
			ipBytes = sourceIP.To4()
			ipVer = IPv4
		}

		commandBytes := make([]byte, 1+1+len(ipBytes)+2+2+1)

		commandBytes[0] = ProxyStatusRequestID
		commandBytes[1] = ipVer

		copy(commandBytes[2:2+len(ipBytes)], ipBytes)

		binary.BigEndian.PutUint16(commandBytes[2+len(ipBytes):4+len(ipBytes)], command.SourcePort)
		binary.BigEndian.PutUint16(commandBytes[4+len(ipBytes):6+len(ipBytes)], command.DestPort)

		var protocol uint8

		if command.Protocol == "tcp" {
			protocol = TCP
		} else if command.Protocol == "udp" {
			protocol = UDP
		} else {
			return nil, fmt.Errorf("invalid protocol")
		}

		commandBytes[6+len(ipBytes)] = protocol

		return commandBytes, nil
	case *ProxyStatusResponse:
		sourceIP := net.ParseIP(command.SourceIP)

		var ipVer uint8
		var ipBytes []byte

		if sourceIP.To4() == nil {
			ipBytes = sourceIP.To16()
			ipVer = IPv6
		} else {
			ipBytes = sourceIP.To4()
			ipVer = IPv4
		}

		commandBytes := make([]byte, 1+1+len(ipBytes)+2+2+1+1)

		commandBytes[0] = ProxyStatusResponseID
		commandBytes[1] = ipVer

		copy(commandBytes[2:2+len(ipBytes)], ipBytes)

		binary.BigEndian.PutUint16(commandBytes[2+len(ipBytes):4+len(ipBytes)], command.SourcePort)
		binary.BigEndian.PutUint16(commandBytes[4+len(ipBytes):6+len(ipBytes)], command.DestPort)

		var protocol uint8

		if command.Protocol == "tcp" {
			protocol = TCP
		} else if command.Protocol == "udp" {
			protocol = UDP
		} else {
			return nil, fmt.Errorf("invalid protocol")
		}

		commandBytes[6+len(ipBytes)] = protocol

		var isActive uint8

		if command.IsActive {
			isActive = 1
		} else {
			isActive = 0
		}

		commandBytes[7+len(ipBytes)] = isActive

		return commandBytes, nil
	case *ProxyInstanceResponse:
		proxyArray := make([][]byte, len(command.Proxies))
		totalSize := 0

		for proxyIndex, proxy := range command.Proxies {
			var err error
			proxyArray[proxyIndex], err = marshalIndividualProxyStruct(proxy)

			if err != nil {
				return nil, err
			}

			totalSize += len(proxyArray[proxyIndex]) + 1
		}

		if totalSize == 0 {
			totalSize = 1
		}

		connectionCommandArray := make([]byte, totalSize+1)
		connectionCommandArray[0] = ProxyInstanceResponseID

		currentPosition := 1

		for _, connection := range proxyArray {
			copy(connectionCommandArray[currentPosition:currentPosition+len(connection)], connection)
			connectionCommandArray[currentPosition+len(connection)] = '\r'
			currentPosition += len(connection) + 1
		}

		connectionCommandArray[totalSize] = '\n'

		return connectionCommandArray, nil
	case *ProxyInstanceRequest:
		return []byte{ProxyInstanceRequestID}, nil
	case *ProxyConnectionsRequest:
		return []byte{ProxyConnectionsRequestID}, nil
	}

	return nil, fmt.Errorf("couldn't match command type")
}
