package backendutil_custom

import (
	"io"
	"net"
	"os"

	"git.terah.dev/imterah/hermes/backend/backendutil"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/datacommands"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/gaslighter"
	"github.com/charmbracelet/log"
)

type BackendApplicationHelper struct {
	Backend    BackendInterface
	SocketPath string

	socket net.Conn
}

func (helper *BackendApplicationHelper) Start() error {
	log.Debug("BackendApplicationHelper is starting")
	err := backendutil.ConfigureProfiling()

	if err != nil {
		return err
	}

	log.Debug("Currently waiting for Unix socket connection...")

	helper.socket, err = net.Dial("unix", helper.SocketPath)

	if err != nil {
		return err
	}

	helper.Backend.OnSocketConnection(helper.socket)

	log.Debug("Sucessfully connected")

	gaslighter := &gaslighter.Gaslighter{}
	gaslighter.ProxiedReader = helper.socket

	commandID := make([]byte, 1)

	for {
		if _, err := helper.socket.Read(commandID); err != nil {
			return err
		}

		gaslighter.Byte = commandID[0]
		gaslighter.HasGaslit = false

		var commandRaw interface{}

		if gaslighter.Byte > 100 {
			commandRaw, err = datacommands.Unmarshal(gaslighter)
		} else {
			commandRaw, err = commonbackend.Unmarshal(gaslighter)
		}

		if err != nil {
			return err
		}

		switch command := commandRaw.(type) {
		case *datacommands.ProxyConnectionsRequest:
			connections := helper.Backend.GetAllClientConnections(command.ProxyID)

			serverParams := &datacommands.ProxyConnectionsResponse{
				Connections: connections,
			}

			byteData, err := datacommands.Marshal(serverParams)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		case *datacommands.RemoveProxy:
			ok, err := helper.Backend.StopProxy(command)
			var hasAnyFailed bool

			if !ok {
				log.Warnf("failed to remove proxy (ID %d): RemoveProxy returned into failure state", command.ProxyID)
				hasAnyFailed = true
			} else if err != nil {
				log.Warnf("failed to remove proxy (ID %d): %s", command.ProxyID, err.Error())
				hasAnyFailed = true
			}

			response := &datacommands.ProxyStatusResponse{
				ProxyID:  command.ProxyID,
				IsActive: hasAnyFailed,
			}

			responseMarshalled, err := datacommands.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *datacommands.ProxyInformationRequest:
			response := helper.Backend.ResolveProxy(command.ProxyID)
			responseMarshalled, err := datacommands.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *datacommands.ProxyConnectionInformationRequest:
			response := helper.Backend.ResolveConnection(command.ProxyID, command.ConnectionID)
			responseMarshalled, err := datacommands.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *datacommands.TCPConnectionClosed:
			helper.Backend.OnTCPConnectionClosed(command.ProxyID, command.ConnectionID)
		case *datacommands.TCPProxyData:
			bytes := make([]byte, command.DataLength)
			_, err := io.ReadFull(helper.socket, bytes)

			if err != nil {
				log.Warn("failed to read TCP data")
			}

			helper.Backend.HandleTCPMessage(command, bytes)
		case *datacommands.UDPProxyData:
			bytes := make([]byte, command.DataLength)
			_, err := io.ReadFull(helper.socket, bytes)

			if err != nil {
				log.Warn("failed to read TCP data")
			}

			helper.Backend.HandleUDPMessage(command, bytes)
		case *commonbackend.Start:
			ok, err := helper.Backend.StartBackend(command.Arguments)

			var (
				message    string
				statusCode int
			)

			if err != nil {
				message = err.Error()
				statusCode = commonbackend.StatusFailure
			} else {
				statusCode = commonbackend.StatusSuccess
			}

			response := &commonbackend.BackendStatusResponse{
				IsRunning:  ok,
				StatusCode: statusCode,
				Message:    message,
			}

			responseMarshalled, err := commonbackend.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *commonbackend.Stop:
			ok, err := helper.Backend.StopBackend()

			var (
				message    string
				statusCode int
			)

			if err != nil {
				message = err.Error()
				statusCode = commonbackend.StatusFailure
			} else {
				statusCode = commonbackend.StatusSuccess
			}

			response := &commonbackend.BackendStatusResponse{
				IsRunning:  !ok,
				StatusCode: statusCode,
				Message:    message,
			}

			responseMarshalled, err := commonbackend.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *commonbackend.BackendStatusRequest:
			ok, err := helper.Backend.GetBackendStatus()

			var (
				message    string
				statusCode int
			)

			if err != nil {
				message = err.Error()
				statusCode = commonbackend.StatusFailure
			} else {
				statusCode = commonbackend.StatusSuccess
			}

			response := &commonbackend.BackendStatusResponse{
				IsRunning:  ok,
				StatusCode: statusCode,
				Message:    message,
			}

			responseMarshalled, err := commonbackend.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *commonbackend.AddProxy:
			id, ok, err := helper.Backend.StartProxy(command)
			var hasAnyFailed bool

			if !ok {
				log.Warnf("failed to add proxy (%s:%d -> remote:%d): StartProxy returned into failure state", command.SourceIP, command.SourcePort, command.DestPort)
				hasAnyFailed = true
			} else if err != nil {
				log.Warnf("failed to add proxy (%s:%d -> remote:%d): %s", command.SourceIP, command.SourcePort, command.DestPort, err.Error())
				hasAnyFailed = true
			}

			response := &datacommands.ProxyStatusResponse{
				ProxyID:  id,
				IsActive: !hasAnyFailed,
			}

			responseMarshalled, err := datacommands.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *commonbackend.CheckClientParameters:
			resp := helper.Backend.CheckParametersForConnections(command)
			resp.InResponseTo = "checkClientParameters"

			byteData, err := commonbackend.Marshal(resp)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		case *commonbackend.CheckServerParameters:
			resp := helper.Backend.CheckParametersForBackend(command.Arguments)
			resp.InResponseTo = "checkServerParameters"

			byteData, err := commonbackend.Marshal(resp)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		default:
			log.Warnf("Unsupported command recieved: %T", command)
		}
	}
}

func NewHelper(backend BackendInterface) *BackendApplicationHelper {
	socketPath, ok := os.LookupEnv("HERMES_API_SOCK")

	if !ok {
		log.Warn("HERMES_API_SOCK is not defined! This will cause an issue unless the backend manually overwrites it")
	}

	helper := &BackendApplicationHelper{
		Backend:    backend,
		SocketPath: socketPath,
	}

	return helper
}
