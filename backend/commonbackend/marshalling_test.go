package commonbackend

import (
	"bytes"
	"log"
	"os"
	"testing"
)

var logLevel = os.Getenv("HERMES_LOG_LEVEL")

func TestStart(t *testing.T) {
	commandInput := &Start{
		Arguments: []byte("Hello from automated testing"),
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*Start)

	if !ok {
		t.Fatal("failed typecast")
	}

	if !bytes.Equal(commandInput.Arguments, commandUnmarshalled.Arguments) {
		log.Fatalf("Arguments are not equal (orig: '%s', unmsh: '%s')", string(commandInput.Arguments), string(commandUnmarshalled.Arguments))
	}
}

func TestStop(t *testing.T) {
	commandInput := &Stop{}

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

	_, ok := commandUnmarshalledRaw.(*Stop)

	if !ok {
		t.Fatal("failed typecast")
	}
}

func TestAddConnection(t *testing.T) {
	commandInput := &AddProxy{
		SourceIP:   "192.168.0.139",
		SourcePort: 19132,
		DestPort:   19132,
		Protocol:   "tcp",
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*AddProxy)

	if !ok {
		t.Fatal("failed typecast")
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

func TestRemoveConnection(t *testing.T) {
	commandInput := &RemoveProxy{
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*RemoveProxy)

	if !ok {
		t.Fatal("failed typecast")
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

func TestGetAllConnections(t *testing.T) {
	commandInput := &ProxyConnectionsResponse{
		Connections: []*ProxyClientConnection{
			{
				SourceIP:   "127.0.0.1",
				SourcePort: 19132,
				DestPort:   19132,
				ClientIP:   "127.0.0.1",
				ClientPort: 12321,
			},
			{
				SourceIP:   "127.0.0.1",
				SourcePort: 19132,
				DestPort:   19132,
				ClientIP:   "192.168.0.168",
				ClientPort: 23457,
			},
			{
				SourceIP:   "127.0.0.1",
				SourcePort: 19132,
				DestPort:   19132,
				ClientIP:   "68.42.203.47",
				ClientPort: 38721,
			},
		},
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyConnectionsResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	for commandIndex, originalConnection := range commandInput.Connections {
		remoteConnection := commandUnmarshalled.Connections[commandIndex]

		if originalConnection.SourceIP != remoteConnection.SourceIP {
			t.Fail()
			log.Printf("(in #%d) SourceIP's are not equal (orig: %s, unmsh: %s)", commandIndex, originalConnection.SourceIP, remoteConnection.SourceIP)
		}

		if originalConnection.SourcePort != remoteConnection.SourcePort {
			t.Fail()
			log.Printf("(in #%d) SourcePort's are not equal (orig: %d, unmsh: %d)", commandIndex, originalConnection.SourcePort, remoteConnection.SourcePort)
		}

		if originalConnection.DestPort != remoteConnection.DestPort {
			t.Fail()
			log.Printf("(in #%d) DestPort's are not equal (orig: %d, unmsh: %d)", commandIndex, originalConnection.DestPort, remoteConnection.DestPort)
		}

		if originalConnection.ClientIP != remoteConnection.ClientIP {
			t.Fail()
			log.Printf("(in #%d) ClientIP's are not equal (orig: %s, unmsh: %s)", commandIndex, originalConnection.ClientIP, remoteConnection.ClientIP)
		}

		if originalConnection.ClientPort != remoteConnection.ClientPort {
			t.Fail()
			log.Printf("(in #%d) ClientPort's are not equal (orig: %d, unmsh: %d)", commandIndex, originalConnection.ClientPort, remoteConnection.ClientPort)
		}
	}
}

func TestCheckClientParameters(t *testing.T) {
	commandInput := &CheckClientParameters{
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*CheckClientParameters)

	if !ok {
		t.Fatal("failed typecast")
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

func TestCheckServerParameters(t *testing.T) {
	commandInput := &CheckServerParameters{
		Arguments: []byte("Hello from automated testing"),
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*CheckServerParameters)

	if !ok {
		t.Fatal("failed typecast")
	}

	if !bytes.Equal(commandInput.Arguments, commandUnmarshalled.Arguments) {
		log.Fatalf("Arguments are not equal (orig: '%s', unmsh: '%s')", string(commandInput.Arguments), string(commandUnmarshalled.Arguments))
	}
}

func TestCheckParametersResponse(t *testing.T) {
	commandInput := &CheckParametersResponse{
		InResponseTo: "checkClientParameters",
		IsValid:      true,
		Message:      "Hello from automated testing",
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*CheckParametersResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.InResponseTo != commandUnmarshalled.InResponseTo {
		t.Fail()
		log.Printf("InResponseTo's are not equal (orig: %s, unmsh: %s)", commandInput.InResponseTo, commandUnmarshalled.InResponseTo)
	}

	if commandInput.IsValid != commandUnmarshalled.IsValid {
		t.Fail()
		log.Printf("IsValid's are not equal (orig: %t, unmsh: %t)", commandInput.IsValid, commandUnmarshalled.IsValid)
	}

	if commandInput.Message != commandUnmarshalled.Message {
		t.Fail()
		log.Printf("Messages are not equal (orig: %s, unmsh: %s)", commandInput.Message, commandUnmarshalled.Message)
	}
}

func TestBackendStatusRequest(t *testing.T) {
	commandInput := &BackendStatusRequest{}
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

	_, ok := commandUnmarshalledRaw.(*BackendStatusRequest)

	if !ok {
		t.Fatal("failed typecast")
	}
}

func TestBackendStatusResponse(t *testing.T) {
	commandInput := &BackendStatusResponse{
		IsRunning:  true,
		StatusCode: StatusFailure,
		Message:    "Hello from automated testing",
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*BackendStatusResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.IsRunning != commandUnmarshalled.IsRunning {
		t.Fail()
		log.Printf("IsRunning's are not equal (orig: %t, unmsh: %t)", commandInput.IsRunning, commandUnmarshalled.IsRunning)
	}

	if commandInput.StatusCode != commandUnmarshalled.StatusCode {
		t.Fail()
		log.Printf("StatusCodes are not equal (orig: %d, unmsh: %d)", commandInput.StatusCode, commandUnmarshalled.StatusCode)
	}

	if commandInput.Message != commandUnmarshalled.Message {
		t.Fail()
		log.Printf("Messages are not equal (orig: %s, unmsh: %s)", commandInput.Message, commandUnmarshalled.Message)
	}
}

func TestProxyStatusRequest(t *testing.T) {
	commandInput := &ProxyStatusRequest{
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyStatusRequest)

	if !ok {
		t.Fatal("failed typecast")
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

func TestProxyStatusResponse(t *testing.T) {
	commandInput := &ProxyStatusResponse{
		SourceIP:   "192.168.0.139",
		SourcePort: 19132,
		DestPort:   19132,
		Protocol:   "tcp",
		IsActive:   true,
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyStatusResponse)

	if !ok {
		t.Fatal("failed typecast")
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

	if commandInput.IsActive != commandUnmarshalled.IsActive {
		t.Fail()
		log.Printf("IsActive's are not equal (orig: %t, unmsh: %t)", commandInput.IsActive, commandUnmarshalled.IsActive)
	}
}

func TestProxyConnectionRequest(t *testing.T) {
	commandInput := &ProxyInstanceRequest{}

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

	_, ok := commandUnmarshalledRaw.(*ProxyInstanceRequest)

	if !ok {
		t.Fatal("failed typecast")
	}
}

func TestProxyConnectionResponse(t *testing.T) {
	commandInput := &ProxyInstanceResponse{
		Proxies: []*ProxyInstance{
			{
				SourceIP:   "192.168.0.168",
				SourcePort: 25565,
				DestPort:   25565,
				Protocol:   "tcp",
			},
			{
				SourceIP:   "127.0.0.1",
				SourcePort: 19132,
				DestPort:   19132,
				Protocol:   "udp",
			},
			{
				SourceIP:   "68.42.203.47",
				SourcePort: 22,
				DestPort:   2222,
				Protocol:   "tcp",
			},
		},
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyInstanceResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	for proxyIndex, originalProxy := range commandInput.Proxies {
		remoteProxy := commandUnmarshalled.Proxies[proxyIndex]

		if originalProxy.SourceIP != remoteProxy.SourceIP {
			t.Fail()
			log.Printf("(in #%d) SourceIP's are not equal (orig: %s, unmsh: %s)", proxyIndex, originalProxy.SourceIP, remoteProxy.SourceIP)
		}

		if originalProxy.SourcePort != remoteProxy.SourcePort {
			t.Fail()
			log.Printf("(in #%d) SourcePort's are not equal (orig: %d, unmsh: %d)", proxyIndex, originalProxy.SourcePort, remoteProxy.SourcePort)
		}

		if originalProxy.DestPort != remoteProxy.DestPort {
			t.Fail()
			log.Printf("(in #%d) DestPort's are not equal (orig: %d, unmsh: %d)", proxyIndex, originalProxy.DestPort, remoteProxy.DestPort)
		}

		if originalProxy.Protocol != remoteProxy.Protocol {
			t.Fail()
			log.Printf("(in #%d) ClientIP's are not equal (orig: %s, unmsh: %s)", proxyIndex, originalProxy.Protocol, remoteProxy.Protocol)
		}
	}
}
