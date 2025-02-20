package datacommands

import (
	"bytes"
	"log"
	"os"
	"testing"
)

var logLevel = os.Getenv("HERMES_LOG_LEVEL")

func TestProxyStatusRequest(t *testing.T) {
	commandInput := &ProxyStatusRequest{
		ProxyID: 19132,
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyStatusRequest)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.ProxyID != commandUnmarshalled.ProxyID {
		t.Fail()
		log.Printf("ProxyID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ProxyID, commandUnmarshalled.ProxyID)
	}
}

func TestProxyStatusResponse(t *testing.T) {
	commandInput := &ProxyStatusResponse{
		ProxyID:  19132,
		IsActive: true,
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyStatusResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.ProxyID != commandUnmarshalled.ProxyID {
		t.Fail()
		log.Printf("ProxyID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ProxyID, commandUnmarshalled.ProxyID)
	}

	if commandInput.IsActive != commandUnmarshalled.IsActive {
		t.Fail()
		log.Printf("IsActive's are not equal (orig: '%t', unmsh: '%t')", commandInput.IsActive, commandUnmarshalled.IsActive)
	}
}

func TestRemoveProxy(t *testing.T) {
	commandInput := &RemoveProxy{
		ProxyID: 19132,
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*RemoveProxy)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.ProxyID != commandUnmarshalled.ProxyID {
		t.Fail()
		log.Printf("ProxyID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ProxyID, commandUnmarshalled.ProxyID)
	}
}

func TestProxyConnectionsRequest(t *testing.T) {
	commandInput := &ProxyConnectionsRequest{
		ProxyID: 19132,
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyConnectionsRequest)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.ProxyID != commandUnmarshalled.ProxyID {
		t.Fail()
		log.Printf("ProxyID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ProxyID, commandUnmarshalled.ProxyID)
	}
}

func TestProxyConnectionsResponse(t *testing.T) {
	commandInput := &ProxyConnectionsResponse{
		Connections: []uint16{12831, 9455, 64219, 12, 32},
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyConnectionsResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	for connectionIndex, originalConnection := range commandInput.Connections {
		remoteConnection := commandUnmarshalled.Connections[connectionIndex]

		if originalConnection != remoteConnection {
			t.Fail()
			log.Printf("(in #%d) SourceIP's are not equal (orig: %d, unmsh: %d)", connectionIndex, originalConnection, connectionIndex)
		}
	}
}

func TestProxyInstanceResponse(t *testing.T) {
	commandInput := &ProxyInstanceResponse{
		Proxies: []uint16{12831, 9455, 64219, 12, 32},
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyInstanceResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	for proxyIndex, originalProxy := range commandInput.Proxies {
		remoteProxy := commandUnmarshalled.Proxies[proxyIndex]

		if originalProxy != remoteProxy {
			t.Fail()
			log.Printf("(in #%d) Proxy IDs are not equal (orig: %d, unmsh: %d)", proxyIndex, originalProxy, remoteProxy)
		}
	}
}

func TestTCPConnectionOpened(t *testing.T) {
	commandInput := &TCPConnectionOpened{
		ProxyID:      19132,
		ConnectionID: 25565,
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*TCPConnectionOpened)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.ProxyID != commandUnmarshalled.ProxyID {
		t.Fail()
		log.Printf("ProxyID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ProxyID, commandUnmarshalled.ProxyID)
	}

	if commandInput.ConnectionID != commandUnmarshalled.ConnectionID {
		t.Fail()
		log.Printf("ConnectionID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ConnectionID, commandUnmarshalled.ConnectionID)
	}
}

func TestTCPConnectionClosed(t *testing.T) {
	commandInput := &TCPConnectionClosed{
		ProxyID:      19132,
		ConnectionID: 25565,
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*TCPConnectionClosed)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.ProxyID != commandUnmarshalled.ProxyID {
		t.Fail()
		log.Printf("ProxyID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ProxyID, commandUnmarshalled.ProxyID)
	}

	if commandInput.ConnectionID != commandUnmarshalled.ConnectionID {
		t.Fail()
		log.Printf("ConnectionID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ConnectionID, commandUnmarshalled.ConnectionID)
	}
}

func TestTCPProxyData(t *testing.T) {
	commandInput := &TCPProxyData{
		ProxyID:      19132,
		ConnectionID: 25565,
		DataLength:   1234,
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*TCPProxyData)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.ProxyID != commandUnmarshalled.ProxyID {
		t.Fail()
		log.Printf("ProxyID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ProxyID, commandUnmarshalled.ProxyID)
	}

	if commandInput.ConnectionID != commandUnmarshalled.ConnectionID {
		t.Fail()
		log.Printf("ConnectionID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ConnectionID, commandUnmarshalled.ConnectionID)
	}

	if commandInput.DataLength != commandUnmarshalled.DataLength {
		t.Fail()
		log.Printf("DataLength's are not equal (orig: '%d', unmsh: '%d')", commandInput.DataLength, commandUnmarshalled.DataLength)
	}
}

func TestUDPProxyData(t *testing.T) {
	commandInput := &UDPProxyData{
		ProxyID:    19132,
		ClientIP:   "68.51.23.54",
		ClientPort: 28173,
		DataLength: 1234,
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*UDPProxyData)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.ProxyID != commandUnmarshalled.ProxyID {
		t.Fail()
		log.Printf("ProxyID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ProxyID, commandUnmarshalled.ProxyID)
	}

	if commandInput.ClientIP != commandUnmarshalled.ClientIP {
		t.Fail()
		log.Printf("ClientIP's are not equal (orig: '%s', unmsh: '%s')", commandInput.ClientIP, commandUnmarshalled.ClientIP)
	}

	if commandInput.ClientPort != commandUnmarshalled.ClientPort {
		t.Fail()
		log.Printf("ClientPort's are not equal (orig: '%d', unmsh: '%d')", commandInput.ClientPort, commandUnmarshalled.ClientPort)
	}

	if commandInput.DataLength != commandUnmarshalled.DataLength {
		t.Fail()
		log.Printf("DataLength's are not equal (orig: '%d', unmsh: '%d')", commandInput.DataLength, commandUnmarshalled.DataLength)
	}
}

func TestProxyInformationRequest(t *testing.T) {
	commandInput := &ProxyInformationRequest{
		ProxyID: 19132,
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyInformationRequest)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.ProxyID != commandUnmarshalled.ProxyID {
		t.Fail()
		log.Printf("ProxyID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ProxyID, commandUnmarshalled.ProxyID)
	}
}

func TestProxyInformationResponseExists(t *testing.T) {
	commandInput := &ProxyInformationResponse{
		Exists:     true,
		SourceIP:   "192.168.0.139",
		SourcePort: 19132,
		DestPort:   19132,
		Protocol:   "tcp",
	}

	commandMarshalled, err := Marshal(commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyInformationResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Exists != commandUnmarshalled.Exists {
		t.Fail()
		log.Printf("Exists's are not equal (orig: '%t', unmsh: '%t')", commandInput.Exists, commandUnmarshalled.Exists)
	}

	if commandInput.SourceIP != commandUnmarshalled.SourceIP {
		t.Fail()
		log.Printf("SourceIP's are not equal (orig: %s, unmsh: %s)", commandInput.SourceIP, commandUnmarshalled.SourceIP)
	}

	if commandInput.SourcePort != commandUnmarshalled.SourcePort {
		t.Fail()
		log.Printf("SourcePort's are not equal (orig: %d, unmsh: %d)", commandInput.SourcePort, commandUnmarshalled.SourcePort)
	}

	if commandInput.DestPort != commandUnmarshalled.DestPort {
		t.Fail()
		log.Printf("DestPort's are not equal (orig: %d, unmsh: %d)", commandInput.DestPort, commandUnmarshalled.DestPort)
	}

	if commandInput.Protocol != commandUnmarshalled.Protocol {
		t.Fail()
		log.Printf("Protocols are not equal (orig: %s, unmsh: %s)", commandInput.Protocol, commandUnmarshalled.Protocol)
	}
}

func TestProxyInformationResponseNoExist(t *testing.T) {
	commandInput := &ProxyInformationResponse{
		Exists: false,
	}

	commandMarshalled, err := Marshal(commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyInformationResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Exists != commandUnmarshalled.Exists {
		t.Fail()
		log.Printf("Exists's are not equal (orig: '%t', unmsh: '%t')", commandInput.Exists, commandUnmarshalled.Exists)
	}
}

func TestProxyConnectionInformationRequest(t *testing.T) {
	commandInput := &ProxyConnectionInformationRequest{
		ProxyID:      19132,
		ConnectionID: 25565,
	}

	commandMarshalled, err := Marshal(commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyConnectionInformationRequest)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.ProxyID != commandUnmarshalled.ProxyID {
		t.Fail()
		log.Printf("ProxyID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ProxyID, commandUnmarshalled.ProxyID)
	}

	if commandInput.ConnectionID != commandUnmarshalled.ConnectionID {
		t.Fail()
		log.Printf("ConnectionID's are not equal (orig: '%d', unmsh: '%d')", commandInput.ConnectionID, commandUnmarshalled.ConnectionID)
	}
}

func TestProxyConnectionInformationResponseExists(t *testing.T) {
	commandInput := &ProxyConnectionInformationResponse{
		Exists:     true,
		ClientIP:   "192.168.0.139",
		ClientPort: 19132,
	}

	commandMarshalled, err := Marshal(commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyConnectionInformationResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Exists != commandUnmarshalled.Exists {
		t.Fail()
		log.Printf("Exists's are not equal (orig: '%t', unmsh: '%t')", commandInput.Exists, commandUnmarshalled.Exists)
	}

	if commandInput.ClientIP != commandUnmarshalled.ClientIP {
		t.Fail()
		log.Printf("SourceIP's are not equal (orig: %s, unmsh: %s)", commandInput.ClientIP, commandUnmarshalled.ClientIP)
	}

	if commandInput.ClientPort != commandUnmarshalled.ClientPort {
		t.Fail()
		log.Printf("ClientPort's are not equal (orig: %d, unmsh: %d)", commandInput.ClientPort, commandUnmarshalled.ClientPort)
	}
}

func TestProxyConnectionInformationResponseNoExists(t *testing.T) {
	commandInput := &ProxyConnectionInformationResponse{
		Exists: false,
	}

	commandMarshalled, err := Marshal(commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyConnectionInformationResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Exists != commandUnmarshalled.Exists {
		t.Fail()
		log.Printf("Exists's are not equal (orig: '%t', unmsh: '%t')", commandInput.Exists, commandUnmarshalled.Exists)
	}
}
