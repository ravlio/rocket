package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
	"rockets/internal/http/gen"
	"rockets/internal/rocket"
	"time"
)

// ListenEchoServer starts the Echo server and listens on the specified address.
func ListenEchoServer(_ context.Context, e *echo.Echo, addr string) func() error {
	return func() error {
		log.Info().Msgf("Listening http server on %s", addr)
		err := e.Start(addr)
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("can't listen http server on %s: %w", addr, err)
		}
		log.Info().Msgf("Http server stopped listening")

		return nil
	}
}

// ShutDownEchoServer gracefully shuts down the Echo server when the context is done.
func ShutDownEchoServer(ctx context.Context, echo *echo.Echo) func() error {
	return func() error {
		<-ctx.Done()
		log.Info().Msgf("Shutting down http server on %s...", echo.Server.Addr)

		// stop the http
		httpCtx, httpCancel := context.WithTimeout(ctx, time.Second*10)
		defer httpCancel()
		if err := echo.Shutdown(httpCtx); err != nil {
			return fmt.Errorf("can't shutdown http server: %w", err)
		}

		return nil
	}
}

// NewEcho creates a new Echo instance with the necessary middleware and routes.
func NewEcho() *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/ready", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           int((12 * time.Hour).Seconds()),
	}))

	return e
}

// StrictServer implements the gen.StrictServerInterface for handling API requests.
type StrictServer struct {
	echo   *echo.Echo
	rocket rocket.Service
}

var _ gen.StrictServerInterface = (*StrictServer)(nil)

func (s *StrictServer) IngestMessage(ctx context.Context, request gen.IngestMessageRequestObject) (gen.IngestMessageResponseObject, error) {
	var msgType rocket.MessageType

	switch request.Body.Metadata.MessageType {
	case gen.RocketExploded:
		msgType = rocket.MessageTypeExploded
	case gen.RocketLaunched:
		msgType = rocket.MessageTypeLaunched
	case gen.RocketSpeedIncreased:
		msgType = rocket.MessageTypeSpeedIncreased
	case gen.RocketSpeedDecreased:
		msgType = rocket.MessageTypeSpeedDecreased
	case gen.RocketMissionChanged:
		msgType = rocket.MessageTypeMissionChanged
	default:
		return gen.IngestMessage400JSONResponse{
			Code:    "unknown_message_type",
			Message: fmt.Sprintf("unknown message type: %s", request.Body.Metadata.MessageType),
		}, nil
	}

	msg := rocket.TelemetryMessage{
		Metadata: rocket.MessageMetadata{
			Channel:       request.Body.Metadata.Channel,
			MessageNumber: request.Body.Metadata.MessageNumber,
			MessageTime:   request.Body.Metadata.MessageTime,
			MessageType:   msgType,
		},
		Message: rocket.Message{
			By:          request.Body.Message.By,
			LaunchSpeed: request.Body.Message.LaunchSpeed,
			Mission:     request.Body.Message.Mission,
			NewMission:  request.Body.Message.NewMission,
			Reason:      request.Body.Message.Reason,
			Type:        request.Body.Message.Type,
		},
	}

	err := s.rocket.ProcessMessage(ctx, msg)
	if err != nil {
		return gen.IngestMessage500JSONResponse{
			Code:    "unknown",
			Message: err.Error(),
		}, nil
	}

	return gen.IngestMessage202JSONResponse{}, nil
}

func (s *StrictServer) ListRockets(ctx context.Context, request gen.ListRocketsRequestObject) (gen.ListRocketsResponseObject, error) {
	var rockets []gen.RocketState

	var sortBy string
	if request.Params.SortBy != nil {
		switch *request.Params.SortBy {
		case gen.Id:
			sortBy = "id"
		case gen.Mission:
			sortBy = "mission"
		case gen.Speed:
			sortBy = "speed"
		case gen.Type:
			sortBy = "type"
		default:
			return gen.ListRockets400JSONResponse{
				Code:    "unknown_sort_by",
				Message: fmt.Sprintf("unknown sort by: %s", *request.Params.SortBy),
			}, nil
		}
	} else {
		sortBy = "id"
	}

	var sortOrder string
	if request.Params.SortOrder != nil {
		switch *request.Params.SortOrder {
		case gen.Asc:
			sortOrder = "ASC"
		case gen.Desc:
			sortOrder = "DESC"
		default:
			return gen.ListRockets400JSONResponse{
				Code:    "unknown_sort_order",
				Message: fmt.Sprintf("unknown sort order: %s", *request.Params.SortOrder),
			}, nil
		}
	} else {
		sortOrder = "asc"
	}

	resp := s.rocket.ListAllRockets(ctx, sortBy, sortOrder)
	for _, state := range resp {
		rockets = append(rockets, stateToServer(state))
	}

	return gen.ListRockets200JSONResponse(rockets), nil
}

func (s *StrictServer) GetRocketState(ctx context.Context, request gen.GetRocketStateRequestObject) (gen.GetRocketStateResponseObject, error) {
	state, ok := s.rocket.GetRocketState(ctx, request.Id)
	if !ok {
		return gen.GetRocketState404JSONResponse{
			Code:    "not_found",
			Message: fmt.Sprintf("rocket with id %s not found", request.Id),
		}, nil
	}

	return gen.GetRocketState200JSONResponse(stateToServer(state)), nil
}
