package backendutil_custom

import (
	"fmt"
	"net"
	"os"

	"git.terah.dev/imterah/hermes/backend/backendutil"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/datacommands"
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

	log.Debug("Sucessfully connected")

	for {
		commandType, commandRaw, err := datacommands.Unmarshal(helper.socket)

		if err != nil && err.Error() != "couldn't match command ID" {
			return err
		}

		switch commandType {
		case "proxyConnectionsRequest":
			proxyConnectionRequest, ok := commandRaw.(*datacommands.ProxyConnectionsRequest)

			if !ok {
				return fmt.Errorf("failed to typecast")
			}

			connections := helper.Backend.GetAllClientConnections(proxyConnectionRequest.ProxyID)

			serverParams := &datacommands.ProxyConnectionsResponse{
				Type:        "proxyConnectionsResponse",
				Connections: connections,
			}

			byteData, err := datacommands.Marshal(serverParams.Type, serverParams)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		case "removeProxy":
			command, ok := commandRaw.(*datacommands.RemoveProxy)

			if !ok {
				return fmt.Errorf("failed to typecast")
			}

			ok, err = helper.Backend.StopProxy(command)
			var hasAnyFailed bool

			if !ok {
				log.Warnf("failed to remove proxy (ID %d): RemoveProxy returned into failure state", command.ProxyID)
				hasAnyFailed = true
			} else if err != nil {
				log.Warnf("failed to remove proxy (ID %d): %s", command.ProxyID, err.Error())
				hasAnyFailed = true
			}

			response := &datacommands.ProxyStatusResponse{
				Type:     "proxyStatusResponse",
				ProxyID:  command.ProxyID,
				IsActive: hasAnyFailed,
			}

			responseMarshalled, err := commonbackend.Marshal(response.Type, response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		default:
			commandType, commandRaw, err := commonbackend.Unmarshal(helper.socket)

			if err != nil {
				return err
			}

			switch commandType {
			case "start":
				command, ok := commandRaw.(*commonbackend.Start)

				if !ok {
					return fmt.Errorf("failed to typecast")
				}

				ok, err = helper.Backend.StartBackend(command.Arguments)

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
					Type:       "backendStatusResponse",
					IsRunning:  ok,
					StatusCode: statusCode,
					Message:    message,
				}

				responseMarshalled, err := commonbackend.Marshal(response.Type, response)

				if err != nil {
					log.Error("failed to marshal response: %s", err.Error())
					continue
				}

				helper.socket.Write(responseMarshalled)
			case "stop":
				_, ok := commandRaw.(*commonbackend.Stop)

				if !ok {
					return fmt.Errorf("failed to typecast")
				}

				ok, err = helper.Backend.StopBackend()

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
					Type:       "backendStatusResponse",
					IsRunning:  !ok,
					StatusCode: statusCode,
					Message:    message,
				}

				responseMarshalled, err := commonbackend.Marshal(response.Type, response)

				if err != nil {
					log.Error("failed to marshal response: %s", err.Error())
					continue
				}

				helper.socket.Write(responseMarshalled)
			case "backendStatusRequest":
				_, ok := commandRaw.(*commonbackend.BackendStatusRequest)

				if !ok {
					return fmt.Errorf("failed to typecast")
				}

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
					Type:       "backendStatusResponse",
					IsRunning:  ok,
					StatusCode: statusCode,
					Message:    message,
				}

				responseMarshalled, err := commonbackend.Marshal(response.Type, response)

				if err != nil {
					log.Error("failed to marshal response: %s", err.Error())
					continue
				}

				helper.socket.Write(responseMarshalled)
			case "addProxy":
				command, ok := commandRaw.(*commonbackend.AddProxy)

				if !ok {
					return fmt.Errorf("failed to typecast")
				}

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
					Type:     "proxyStatusResponse",
					ProxyID:  id,
					IsActive: !hasAnyFailed,
				}

				responseMarshalled, err := commonbackend.Marshal(response.Type, response)

				if err != nil {
					log.Error("failed to marshal response: %s", err.Error())
					continue
				}

				helper.socket.Write(responseMarshalled)
			case "checkClientParameters":
				command, ok := commandRaw.(*commonbackend.CheckClientParameters)

				if !ok {
					return fmt.Errorf("failed to typecast")
				}

				resp := helper.Backend.CheckParametersForConnections(command)
				resp.Type = "checkParametersResponse"
				resp.InResponseTo = "checkClientParameters"

				byteData, err := commonbackend.Marshal(resp.Type, resp)

				if err != nil {
					return err
				}

				if _, err = helper.socket.Write(byteData); err != nil {
					return err
				}
			case "checkServerParameters":
				command, ok := commandRaw.(*commonbackend.CheckServerParameters)

				if !ok {
					return fmt.Errorf("failed to typecast")
				}

				resp := helper.Backend.CheckParametersForBackend(command.Arguments)
				resp.Type = "checkParametersResponse"
				resp.InResponseTo = "checkServerParameters"

				byteData, err := commonbackend.Marshal(resp.Type, resp)

				if err != nil {
					return err
				}

				if _, err = helper.socket.Write(byteData); err != nil {
					return err
				}
			default:
				log.Warn("Unsupported command recieved: %s", commandType)
			}
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
